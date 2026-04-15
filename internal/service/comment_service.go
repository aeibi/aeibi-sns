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

		postRow, err := qtx.GetPostByUid(ctx, db.GetPostByUidParams{
			Uid:    postUid,
			Viewer: uuid.NullUUID{UUID: authorUid, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("get post: %w", err)
		}
		if postRow.Visibility == db.PostVisibilityPRIVATE && postRow.Author != authorUid {
			return fmt.Errorf("post not found")
		}
		_, err = qtx.CreateComment(ctx, db.CreateCommentParams{
			Uid:       commentUid,
			PostUid:   postUid,
			AuthorUid: authorUid,
			RootUid:   commentUid,
			Content:   req.Content,
			Images:    req.Images,
		})
		if err != nil {
			return fmt.Errorf("create comment: %w", err)
		}
		commentCount, err := qtx.IncrementPostCommentCount(ctx, postUid)
		if err != nil {
			return fmt.Errorf("increment post comment count: %w", err)
		}
		if postRow.Author != authorUid {
			if err := s.producer.EnqueueCommentInboxTx(ctx, tx, async.CommentInboxArgs{
				MessageUID:  uuid.New(),
				ReceiverUID: postRow.Author,
				ActorUID:    authorUid,
				CommentUID:  commentUid,
				PostUID:     postUid,
				ParentUID:   postUid,
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
	commentRow, err := s.db.GetCommentMetaByUid(ctx, parentUid)
	if err != nil {
		return nil, fmt.Errorf("get parent comment meta: %w", err)
	}
	var resp *api.CreateReplyResponse
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		_, err = qtx.CreateComment(ctx, db.CreateCommentParams{
			Uid:              replyUid,
			PostUid:          commentRow.PostUid,
			RootUid:          commentRow.RootUid,
			ParentUid:        uuid.NullUUID{UUID: parentUid, Valid: true},
			ReplyToAuthorUid: uuid.NullUUID{UUID: commentRow.AuthorUid, Valid: commentRow.RootUid != parentUid},
			AuthorUid:        authorUid,
			Content:          req.Content,
		})
		if err != nil {
			return fmt.Errorf("create reply: %w", err)
		}
		replyCount, err := qtx.IncrementCommentReplyCount(ctx, commentRow.RootUid)
		if err != nil {
			return fmt.Errorf("increment comment reply count: %w", err)
		}
		if commentRow.AuthorUid != authorUid {
			if err := s.producer.EnqueueCommentInboxTx(ctx, tx, async.CommentInboxArgs{
				MessageUID:  uuid.New(),
				ReceiverUID: commentRow.AuthorUid,
				ActorUID:    authorUid,
				CommentUID:  replyUid,
				PostUID:     commentRow.PostUid,
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
	cursorCreatedAt, cursorID, err := decodeAndValidateTopCommentsPageToken(req.GetPageToken(), req.GetPostUid())
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListTopComments(ctx, db.ListTopCommentsParams{
		Viewer:          uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		PostUid:         util.UUID(req.PostUid),
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list top comments: %w", err)
	}

	comments := make([]*api.Comment, 0, len(rows))
	for _, row := range rows {
		parentUid := util.NullUUIDString(row.ParentUid)
		var replyToAuthor *api.CommentAuthor
		if row.ReplyToAuthorUid.Valid && row.ReplyToAuthorNickname.Valid && row.ReplyToAuthorAvatarUrl.Valid {
			replyToAuthor = &api.CommentAuthor{
				Uid:       util.NullUUIDString(row.ReplyToAuthorUid),
				Nickname:  row.ReplyToAuthorNickname.String,
				AvatarUrl: row.ReplyToAuthorAvatarUrl.String,
			}
		}
		comments = append(comments, &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  row.AuthorNickname,
				AvatarUrl: row.AuthorAvatarUrl,
			},
			PostUid:       row.PostUid.String(),
			RootUid:       row.RootUid.String(),
			ParentUid:     parentUid,
			ReplyToAuthor: replyToAuthor,
			Content:       row.Content,
			Images:        row.Images,
			ReplyCount:    row.ReplyCount,
			LikeCount:     row.LikeCount,
			Liked:         row.Liked,
			CreatedAt:     row.CreatedAt.Time.Unix(),
			UpdatedAt:     row.UpdatedAt.Time.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeTopCommentsPageToken(topCommentsPageToken{
			PostUID:         req.GetPostUid(),
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
	rows, err := s.db.ListReplies(ctx, db.ListRepliesParams{
		Viewer:  uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		RootUid: util.UUID(req.Uid),
		Page:    req.Page,
	})
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}

	comments := make([]*api.Comment, 0, len(rows))
	for _, row := range rows {
		parentUid := util.NullUUIDString(row.ParentUid)
		var replyToAuthor *api.CommentAuthor
		if row.ReplyToAuthorUid.Valid && row.ReplyToAuthorNickname.Valid && row.ReplyToAuthorAvatarUrl.Valid {
			replyToAuthor = &api.CommentAuthor{
				Uid:       util.NullUUIDString(row.ReplyToAuthorUid),
				Nickname:  row.ReplyToAuthorNickname.String,
				AvatarUrl: row.ReplyToAuthorAvatarUrl.String,
			}
		}
		comments = append(comments, &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  row.AuthorNickname,
				AvatarUrl: row.AuthorAvatarUrl,
			},
			PostUid:       row.PostUid.String(),
			RootUid:       row.RootUid.String(),
			ParentUid:     parentUid,
			ReplyToAuthor: replyToAuthor,
			Content:       row.Content,
			Images:        row.Images,
			ReplyCount:    row.ReplyCount,
			LikeCount:     row.LikeCount,
			Liked:         row.Liked,
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
		Page:     req.Page,
		Total:    total,
	}, nil
}

func (s *CommentService) GetComment(ctx context.Context, viewerUid string, req *api.GetCommentRequest) (*api.GetCommentResponse, error) {
	row, err := s.db.GetCommentByUid(ctx, db.GetCommentByUidParams{
		Viewer: uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		Uid:    util.UUID(req.Uid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("get comment: %w", err)
	}

	parentUid := util.NullUUIDString(row.ParentUid)
	var replyToAuthor *api.CommentAuthor
	if row.ReplyToAuthorUid.Valid && row.ReplyToAuthorNickname.Valid && row.ReplyToAuthorAvatarUrl.Valid {
		replyToAuthor = &api.CommentAuthor{
			Uid:       util.NullUUIDString(row.ReplyToAuthorUid),
			Nickname:  row.ReplyToAuthorNickname.String,
			AvatarUrl: row.ReplyToAuthorAvatarUrl.String,
		}
	}

	return &api.GetCommentResponse{
		Comment: &api.Comment{
			Uid: row.Uid.String(),
			Author: &api.CommentAuthor{
				Uid:       row.AuthorUid.String(),
				Nickname:  row.AuthorNickname,
				AvatarUrl: row.AuthorAvatarUrl,
			},
			PostUid:       row.PostUid.String(),
			RootUid:       row.RootUid.String(),
			ParentUid:     parentUid,
			ReplyToAuthor: replyToAuthor,
			Content:       row.Content,
			Images:        row.Images,
			ReplyCount:    row.ReplyCount,
			LikeCount:     row.LikeCount,
			Liked:         row.Liked,
			CreatedAt:     row.CreatedAt.Time.Unix(),
			UpdatedAt:     row.UpdatedAt.Time.Unix(),
		},
	}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, uid string, req *api.DeleteCommentRequest) error {
	commentUid := util.UUID(req.Uid)
	authorUid := util.UUID(uid)

	commentRow, err := s.db.GetCommentMetaByUid(ctx, commentUid)
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

	var count int32

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
				count, err = qtx.IncrementCommentLikeCount(ctx, commentUid)
				if err != nil {
					return fmt.Errorf("comment like: increment comment like count: %w", err)
				}
			} else {
				count, err = qtx.GetCommentLikeCount(ctx, commentUid)
				if err != nil {
					return fmt.Errorf("comment like: get comment like count: %w", err)
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
				count, err = qtx.DecrementCommentLikeCount(ctx, commentUid)
				if err != nil {
					return fmt.Errorf("comment like: decrement comment like count: %w", err)
				}
			} else {
				count, err = qtx.GetCommentLikeCount(ctx, commentUid)
				if err != nil {
					return fmt.Errorf("comment like: get comment like count: %w", err)
				}
			}

		default:
			return fmt.Errorf("comment like: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &api.LikeCommentResponse{
		Count: count,
	}, nil
}

type topCommentsPageToken struct {
	PostUID         string `json:"post_uid,omitempty"`
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

func decodeAndValidateTopCommentsPageToken(pageToken string, postUID string) (pgtype.Timestamptz, uuid.NullUUID, error) {
	if pageToken == "" {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token topCommentsPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	if token.PostUID != postUID {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "page_token does not match current filters")
	}
	if token.CursorCreatedAt <= 0 || token.CursorID == "" {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	cursorID, err := uuid.Parse(token.CursorID)
	if err != nil {
		return pgtype.Timestamptz{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	return pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true}, uuid.NullUUID{UUID: cursorID, Valid: true}, nil
}

func encodeTopCommentsPageToken(token topCommentsPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
