package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CommentService struct {
	db  *db.Queries
	dbx *sql.DB
}

func NewCommentService(dbx *sql.DB) *CommentService {
	return &CommentService{
		db:  db.New(dbx),
		dbx: dbx,
	}
}

func (s *CommentService) CreateTopComment(ctx context.Context, uid string, req *api.CreateTopCommentRequest) (*api.CreateTopCommentResponse, error) {
	commentUid := uuid.New()
	postUid := util.UUID(req.PostUid)
	authorUid := util.UUID(uid)
	var resp *api.CreateTopCommentResponse
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
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
			_, _ = qtx.CreateCommentInboxMessage(ctx, db.CreateCommentInboxMessageParams{
				ReceiverUid: postRow.Author,
				ActorUid:    authorUid,
				CommentUid:  uuid.NullUUID{UUID: commentUid, Valid: true},
				PostUid:     uuid.NullUUID{UUID: postUid, Valid: true},
				ParentUid:   uuid.NullUUID{UUID: postUid, Valid: true},
			})
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
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
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
			_, _ = qtx.CreateCommentInboxMessage(ctx, db.CreateCommentInboxMessageParams{
				ReceiverUid: commentRow.AuthorUid,
				ActorUid:    authorUid,
				CommentUid:  uuid.NullUUID{UUID: replyUid, Valid: true},
				PostUid:     uuid.NullUUID{UUID: commentRow.PostUid, Valid: true},
				ParentUid:   uuid.NullUUID{UUID: parentUid, Valid: true},
			})
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
	rows, err := s.db.ListTopComments(ctx, db.ListTopCommentsParams{
		Viewer:          uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		PostUid:         util.UUID(req.PostUid),
		CursorCreatedAt: sql.NullTime{Time: time.Unix(req.CursorCreatedAt, 0).UTC(), Valid: req.CursorCreatedAt != 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(req.CursorId), Valid: req.CursorId != ""},
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
			CreatedAt:     row.CreatedAt.Unix(),
			UpdatedAt:     row.UpdatedAt.Unix(),
		})
	}

	var nextCursorCreatedAt int64
	var nextCursorID string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursorCreatedAt = last.CreatedAt.Unix()
		nextCursorID = last.Uid.String()
	}

	return &api.ListTopCommentsResponse{
		Comments:            comments,
		NextCursorCreatedAt: nextCursorCreatedAt,
		NextCursorId:        nextCursorID,
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
			CreatedAt:     row.CreatedAt.Unix(),
			UpdatedAt:     row.UpdatedAt.Unix(),
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
		if errors.Is(err, sql.ErrNoRows) {
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
			CreatedAt:     row.CreatedAt.Unix(),
			UpdatedAt:     row.UpdatedAt.Unix(),
		},
	}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, uid string, req *api.DeleteCommentRequest) error {
	commentUid := util.UUID(req.Uid)
	authorUid := util.UUID(uid)

	commentRow, err := s.db.GetCommentMetaByUid(ctx, commentUid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("get comment: %w", err)
	}

	return db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
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

	var (
		count int32
		err   error
	)

	switch req.Action {
	case api.ToggleAction_TOGGLE_ACTION_ADD:
		count, err = s.db.AddCommentLike(ctx, db.AddCommentLikeParams{
			CommentUid: commentUid,
			UserUid:    userUid,
		})
	default:
		count, err = s.db.RemoveCommentLike(ctx, db.RemoveCommentLikeParams{
			CommentUid: commentUid,
			UserUid:    userUid,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("comment like: %w", err)
	}
	return &api.LikeCommentResponse{
		Count: count,
	}, nil
}
