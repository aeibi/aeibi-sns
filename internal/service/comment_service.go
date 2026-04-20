package service

import (
	"aeibi/api"
	"aeibi/internal/async"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CommentService struct {
	db       *db.Queries
	pool     *pgxpool.Pool
	producer *async.Producer
}

func NewCommentService(pool *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) *CommentService {
	return &CommentService{
		db:       db.New(pool),
		pool:     pool,
		producer: async.New(riverClient),
	}
}

func (s *CommentService) CreateTopComment(ctx context.Context, uid string, req *api.CreateTopCommentRequest) (*api.CreateTopCommentResponse, error) {
	commentUid := uuid.New()
	postUid := util.UUID(req.PostUid)
	authorUid := util.UUID(uid)
	var resp *api.CreateTopCommentResponse
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		postRow, err := qtx.GetPostByUid(ctx, postUid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		if err != nil {
			return fmt.Errorf("get post: %w", err)
		}
		if postRow.Visibility == db.PostVisibilityPRIVATE && postRow.AuthorUid != authorUid {
			return fmt.Errorf("post not found")
		}
		images := []string{}
		if req.Images != nil {
			images = req.Images
		}
		_, err = qtx.CreateComment(ctx, db.CreateCommentParams{
			Uid:       commentUid,
			PostUid:   postUid,
			AuthorUid: authorUid,
			RootUid:   commentUid,
			Content:   req.Content,
			Images:    images,
			Ip:        "",
		})
		if err != nil {
			return fmt.Errorf("create comment: %w", err)
		}
		commentCount, err := qtx.IncrementPostCommentCount(ctx, postUid)
		if err != nil {
			return fmt.Errorf("increment post comment count: %w", err)
		}
		if postRow.AuthorUid != authorUid {
			if err := s.producer.EnqueueCommentInboxTx(ctx, tx, async.CommentInboxArgs{
				MessageUID:  uuid.New(),
				ReceiverUID: postRow.AuthorUid,
				ActorUID:    authorUid,
				CommentUID:  commentUid,
				PostUID:     postUid,
				ParentUID:   uuid.Nil,
			}); err != nil {
				return fmt.Errorf("enqueue comment inbox job: %w", err)
			}
		}
		if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
			PostUID: postUid,
			Action:  async.PostSearchActionUpsert,
		}); err != nil {
			return fmt.Errorf("enqueue update post search job: %w", err)
		}
		resp = &api.CreateTopCommentResponse{
			Uid:          commentUid.String(),
			CommentCount: commentCount,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *CommentService) CreateReply(ctx context.Context, uid string, req *api.CreateReplyRequest) (*api.CreateReplyResponse, error) {
	replyUid := uuid.New()
	parentUid := util.UUID(req.ParentUid)
	authorUid := util.UUID(uid)
	var resp *api.CreateReplyResponse
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		parentRow, err := qtx.GetCommentByUid(ctx, parentUid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("parent comment not found")
		}
		if err != nil {
			return fmt.Errorf("get parent comment: %w", err)
		}
		postRow, err := qtx.GetPostByUid(ctx, parentRow.PostUid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		if err != nil {
			return fmt.Errorf("get post: %w", err)
		}
		if postRow.Visibility == db.PostVisibilityPRIVATE && postRow.AuthorUid != authorUid {
			return fmt.Errorf("post not found")
		}

		_, err = qtx.CreateComment(ctx, db.CreateCommentParams{
			Uid:              replyUid,
			PostUid:          parentRow.PostUid,
			RootUid:          parentRow.RootUid,
			ParentUid:        uuid.NullUUID{UUID: parentUid, Valid: true},
			ReplyToAuthorUid: uuid.NullUUID{UUID: parentRow.AuthorUid, Valid: parentRow.RootUid != parentUid},
			AuthorUid:        authorUid,
			Content:          req.Content,
			Ip:               "",
		})
		if err != nil {
			return fmt.Errorf("create reply: %w", err)
		}
		replyCount, err := qtx.IncrementCommentReplyCount(ctx, parentRow.RootUid)
		if err != nil {
			return fmt.Errorf("increment comment reply count: %w", err)
		}
		if parentRow.AuthorUid != authorUid {
			if err := s.producer.EnqueueCommentInboxTx(ctx, tx, async.CommentInboxArgs{
				MessageUID:  uuid.New(),
				ReceiverUID: parentRow.AuthorUid,
				ActorUID:    authorUid,
				CommentUID:  replyUid,
				PostUID:     parentRow.PostUid,
				ParentUID:   parentUid,
			}); err != nil {
				return fmt.Errorf("enqueue comment inbox job: %w", err)
			}
		}
		resp = &api.CreateReplyResponse{
			Uid:        replyUid.String(),
			ReplyCount: replyCount,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *CommentService) ListTopComments(ctx context.Context, viewerUid string, req *api.ListTopCommentsRequest) (*api.ListTopCommentsResponse, error) {
	vuid := util.UUID(viewerUid)
	postUID := util.UUID(req.PostUid)

	token, err := decodeTopCommentsPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}
	postRow, err := s.db.GetPostByUid(ctx, postUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	if postRow.Visibility == db.PostVisibilityPRIVATE && (viewerUid == "" || postRow.AuthorUid != vuid) {
		return nil, fmt.Errorf("post not found")
	}

	rows, err := s.db.ListTopComments(ctx, db.ListTopCommentsParams{
		PostUid:         postUID,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		CursorID:        util.UUID(token.CursorID),
	})
	if err != nil {
		return nil, fmt.Errorf("list top comments: %w", err)
	}

	commentUIDs := make([]uuid.UUID, 0, len(rows))
	authorUIDs := make([]uuid.UUID, 0, len(rows))
	seenAuthorUIDs := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		commentUIDs = append(commentUIDs, row.Uid)
		if _, ok := seenAuthorUIDs[row.AuthorUid]; ok {
			continue
		}
		seenAuthorUIDs[row.AuthorUid] = struct{}{}
		authorUIDs = append(authorUIDs, row.AuthorUid)
	}

	userRows, err := s.db.GetUsersByUIDs(ctx, authorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get comment users: %w", err)
	}
	userMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(userRows))
	for _, row := range userRows {
		userMap[row.Uid] = row
	}

	likedSet := make(map[uuid.UUID]struct{})
	if viewerUid != "" && len(commentUIDs) > 0 {
		likedUIDs, err := s.db.ListLikedCommentUIDsByUserAndCommentUIDs(ctx, db.ListLikedCommentUIDsByUserAndCommentUIDsParams{
			UserUid:     vuid,
			CommentUids: commentUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list liked comment uids: %w", err)
		}
		for _, commentUID := range likedUIDs {
			likedSet[commentUID] = struct{}{}
		}
	}

	comments := make([]*api.Comment, 0, len(rows))
	for _, row := range rows {
		authorRow, ok := userMap[row.AuthorUid]
		if !ok {
			continue
		}
		parentUid := util.NullUUIDString(row.ParentUid)
		_, liked := likedSet[row.Uid]
		comments = append(comments, &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  authorRow.Nickname,
				AvatarUrl: authorRow.AvatarUrl,
			},
			PostUid:    row.PostUid.String(),
			RootUid:    row.RootUid.String(),
			ParentUid:  parentUid,
			Content:    row.Content,
			Images:     row.Images,
			ReplyCount: row.ReplyCount,
			LikeCount:  row.LikeCount,
			Liked:      liked,
			CreatedAt:  row.CreatedAt.Time.Unix(),
			UpdatedAt:  row.UpdatedAt.Time.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeTopCommentsPageToken(topCommentsPageToken{
			CursorCreatedAt: last.CreatedAt.Time.Unix(),
			CursorID:        last.Uid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListTopCommentsResponse{
		Comments:      comments,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *CommentService) ListReplies(ctx context.Context, viewerUid string, req *api.ListRepliesRequest) (*api.ListRepliesResponse, error) {
	vuid := util.UUID(viewerUid)
	rootUID := util.UUID(req.Uid)
	page := req.Page
	if page < 1 {
		page = 1
	}
	rootRow, err := s.db.GetCommentByUid(ctx, rootUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	postRow, err := s.db.GetPostByUid(ctx, rootRow.PostUid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	if postRow.Visibility == db.PostVisibilityPRIVATE && (viewerUid == "" || postRow.AuthorUid != vuid) {
		return nil, fmt.Errorf("comment not found")
	}

	rows, err := s.db.ListReplies(ctx, db.ListRepliesParams{
		RootUid: rootUID,
		Page:    page,
	})
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}

	commentUIDs := make([]uuid.UUID, 0, len(rows))
	userUIDs := make([]uuid.UUID, 0, len(rows)*2)
	seenUserUIDs := make(map[uuid.UUID]struct{}, len(rows)*2)
	for _, row := range rows {
		commentUIDs = append(commentUIDs, row.Uid)
		if _, ok := seenUserUIDs[row.AuthorUid]; !ok {
			seenUserUIDs[row.AuthorUid] = struct{}{}
			userUIDs = append(userUIDs, row.AuthorUid)
		}
		if row.ReplyToAuthorUid.Valid {
			replyUID := row.ReplyToAuthorUid.UUID
			if _, ok := seenUserUIDs[replyUID]; !ok {
				seenUserUIDs[replyUID] = struct{}{}
				userUIDs = append(userUIDs, replyUID)
			}
		}
	}

	userRows, err := s.db.GetUsersByUIDs(ctx, userUIDs)
	if err != nil {
		return nil, fmt.Errorf("get comment users: %w", err)
	}
	userMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(userRows))
	for _, row := range userRows {
		userMap[row.Uid] = row
	}

	likedSet := make(map[uuid.UUID]struct{})
	if viewerUid != "" && len(commentUIDs) > 0 {
		likedUIDs, err := s.db.ListLikedCommentUIDsByUserAndCommentUIDs(ctx, db.ListLikedCommentUIDsByUserAndCommentUIDsParams{
			UserUid:     vuid,
			CommentUids: commentUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list liked comment uids: %w", err)
		}
		for _, commentUID := range likedUIDs {
			likedSet[commentUID] = struct{}{}
		}
	}

	comments := make([]*api.Comment, 0, len(rows))
	for _, row := range rows {
		authorRow, ok := userMap[row.AuthorUid]
		if !ok {
			continue
		}
		parentUid := util.NullUUIDString(row.ParentUid)
		var replyToAuthor *api.CommentAuthor
		if row.ReplyToAuthorUid.Valid {
			replyUser, ok := userMap[row.ReplyToAuthorUid.UUID]
			if ok {
				replyToAuthor = &api.CommentAuthor{
					Uid:       row.ReplyToAuthorUid.UUID.String(),
					Nickname:  replyUser.Nickname,
					AvatarUrl: replyUser.AvatarUrl,
				}
			}
		}
		_, liked := likedSet[row.Uid]
		comments = append(comments, &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  authorRow.Nickname,
				AvatarUrl: authorRow.AvatarUrl,
			},
			PostUid:       row.PostUid.String(),
			RootUid:       row.RootUid.String(),
			ParentUid:     parentUid,
			ReplyToAuthor: replyToAuthor,
			Content:       row.Content,
			Images:        row.Images,
			ReplyCount:    row.ReplyCount,
			LikeCount:     row.LikeCount,
			Liked:         liked,
			CreatedAt:     row.CreatedAt.Time.Unix(),
			UpdatedAt:     row.UpdatedAt.Time.Unix(),
		})
	}

	var total int32
	if len(rows) > 0 {
		total = rows[0].Total
	}

	return &api.ListRepliesResponse{
		Comments: comments,
		Page:     page,
		Total:    total,
	}, nil
}

func (s *CommentService) GetComment(ctx context.Context, viewerUid string, req *api.GetCommentRequest) (*api.GetCommentResponse, error) {
	vuid := util.UUID(viewerUid)

	row, err := s.db.GetCommentByUid(ctx, util.UUID(req.Uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("get comment: %w", err)
	}
	postRow, err := s.db.GetPostByUid(ctx, row.PostUid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	if postRow.Visibility == db.PostVisibilityPRIVATE && (viewerUid == "" || postRow.AuthorUid != vuid) {
		return nil, fmt.Errorf("comment not found")
	}

	userUIDs := make([]uuid.UUID, 0, 2)
	userUIDs = append(userUIDs, row.AuthorUid)
	if row.ReplyToAuthorUid.Valid {
		userUIDs = append(userUIDs, row.ReplyToAuthorUid.UUID)
	}
	userRows, err := s.db.GetUsersByUIDs(ctx, userUIDs)
	if err != nil {
		return nil, fmt.Errorf("get comment users: %w", err)
	}
	userMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(userRows))
	for _, userRow := range userRows {
		userMap[userRow.Uid] = userRow
	}
	authorRow, ok := userMap[row.AuthorUid]
	if !ok {
		return nil, fmt.Errorf("comment not found")
	}

	liked := false
	if viewerUid != "" {
		likedUIDs, err := s.db.ListLikedCommentUIDsByUserAndCommentUIDs(ctx, db.ListLikedCommentUIDsByUserAndCommentUIDsParams{
			UserUid:     vuid,
			CommentUids: []uuid.UUID{row.Uid},
		})
		if err != nil {
			return nil, fmt.Errorf("list liked comment uids: %w", err)
		}
		liked = len(likedUIDs) > 0
	}

	parentUid := util.NullUUIDString(row.ParentUid)
	var replyToAuthor *api.CommentAuthor
	if row.ReplyToAuthorUid.Valid {
		replyUser, ok := userMap[row.ReplyToAuthorUid.UUID]
		if ok {
			replyToAuthor = &api.CommentAuthor{
				Uid:       row.ReplyToAuthorUid.UUID.String(),
				Nickname:  replyUser.Nickname,
				AvatarUrl: replyUser.AvatarUrl,
			}
		}
	}

	return &api.GetCommentResponse{
		Comment: &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  authorRow.Nickname,
				AvatarUrl: authorRow.AvatarUrl,
			},
			PostUid:       row.PostUid.String(),
			RootUid:       row.RootUid.String(),
			ParentUid:     parentUid,
			ReplyToAuthor: replyToAuthor,
			Content:       row.Content,
			Images:        row.Images,
			ReplyCount:    row.ReplyCount,
			LikeCount:     row.LikeCount,
			Liked:         liked,
			CreatedAt:     row.CreatedAt.Time.Unix(),
			UpdatedAt:     row.UpdatedAt.Time.Unix(),
		},
	}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, uid string, req *api.DeleteCommentRequest) error {
	commentUid := util.UUID(req.Uid)
	authorUid := util.UUID(uid)

	commentRow, err := s.db.GetCommentByUid(ctx, commentUid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("get comment: %w", err)
	}

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		affected, err := qtx.ArchiveCommentByUidAndAuthor(ctx, db.ArchiveCommentByUidAndAuthorParams{
			Uid:       commentUid,
			AuthorUid: authorUid,
		})
		if err != nil {
			return fmt.Errorf("archive comment: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("comment not found or no permission")
		}

		if commentRow.RootUid == commentUid {
			if _, err := qtx.DecrementPostCommentCount(ctx, commentRow.PostUid); err != nil {
				return fmt.Errorf("decrement post comment count: %w", err)
			}
			if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
				PostUID: commentRow.PostUid,
				Action:  async.PostSearchActionUpsert,
			}); err != nil {
				return fmt.Errorf("enqueue update post search job: %w", err)
			}
			return nil
		}
		if _, err := qtx.DecrementCommentReplyCount(ctx, commentRow.RootUid); err != nil {
			return fmt.Errorf("decrement comment reply count: %w", err)
		}
		return nil
	})
}

func (s *CommentService) LikeComment(ctx context.Context, uid string, req *api.LikeCommentRequest) (*api.LikeCommentResponse, error) {
	commentUid := util.UUID(req.Uid)
	userUid := util.UUID(uid)

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertCommentLikeEdge(ctx, db.InsertCommentLikeEdgeParams{
				CommentUid: commentUid,
				UserUid:    userUid,
			})
			if err != nil {
				return fmt.Errorf("comment like: insert comment like edge: %w", err)
			}

			if affected > 0 {
				if _, err := qtx.IncrementCommentLikeCount(ctx, commentUid); err != nil {
					return fmt.Errorf("comment like: increment comment like count: %w", err)
				}
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			affected, err := qtx.DeleteCommentLikeEdge(ctx, db.DeleteCommentLikeEdgeParams{
				CommentUid: commentUid,
				UserUid:    userUid,
			})
			if err != nil {
				return fmt.Errorf("comment like: delete comment like edge: %w", err)
			}

			if affected > 0 {
				if _, err := qtx.DecrementCommentLikeCount(ctx, commentUid); err != nil {
					return fmt.Errorf("comment like: decrement comment like count: %w", err)
				}
			}

		default:
			return fmt.Errorf("comment like: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &api.LikeCommentResponse{}, nil
}

type topCommentsPageToken struct {
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

func decodeTopCommentsPageToken(pageToken string) (topCommentsPageToken, error) {
	var token topCommentsPageToken
	if pageToken != "" {
		raw, err := base64.RawURLEncoding.DecodeString(pageToken)
		if err != nil {
			return topCommentsPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		if err := json.Unmarshal(raw, &token); err != nil {
			return topCommentsPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}
	if token.CursorCreatedAt == 0 || token.CursorID == "" {
		token.CursorCreatedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
		token.CursorID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	}
	return token, nil
}

func encodeTopCommentsPageToken(token topCommentsPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
