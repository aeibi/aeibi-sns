package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MessageService struct {
	db *db.Queries
}

func NewMessageService(pool *pgxpool.Pool) *MessageService {
	return &MessageService{db: db.New(pool)}
}

func (s *MessageService) ListCommentInboxMessages(ctx context.Context, uid string, req *api.ListCommentInboxMessagesRequest) (*api.ListCommentInboxMessagesResponse, error) {
	token, err := decodeInboxPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	isReadFilter := readFilterToIsReadFilter(req.ReadFilter)
	rows, err := s.db.ListCommentInboxMessages(ctx, db.ListCommentInboxMessagesParams{
		ReceiverUid:     util.UUID(uid),
		IsRead:          isReadFilter,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: token.CursorCreatedAt > 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(token.CursorID), Valid: token.CursorID != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("list comment inbox messages: %w", err)
	}

	if len(rows) > 0 && req.ReadFilter != api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_READ {
		messageUids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			messageUids = append(messageUids, row.Uid)
		}
		if _, err := s.db.MarkInboxMessagesReadByUidsAndReceiver(ctx, db.MarkInboxMessagesReadByUidsAndReceiverParams{
			ReceiverUid: util.UUID(uid),
			Uids:        messageUids,
		}); err != nil {
			return nil, fmt.Errorf("mark comment inbox messages read: %w", err)
		}
	}

	messages := make([]*api.CommentInboxMessage, 0, len(rows))
	for _, row := range rows {
		if !row.CommentUid.Valid {
			continue
		}
		messages = append(messages, &api.CommentInboxMessage{
			Uid:            row.Uid.String(),
			IsRead:         row.IsRead,
			CreatedAt:      row.CreatedAt.Time.Unix(),
			CommentUid:     util.NullUUIDString(row.CommentUid),
			CommentContent: row.CommentContent.String,
			PostUid:        util.NullUUIDString(row.PostUid),
			ParentUid:      util.NullUUIDString(row.ParentUid),
			ParentContent:  row.ParentContent,
			Actor: &api.InboxMessageActor{
				Uid:       row.ActorUid.String(),
				Nickname:  row.ActorNickname,
				AvatarUrl: row.ActorAvatarUrl,
			},
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeInboxPageToken(inboxPageToken{
			CursorCreatedAt: last.CreatedAt.Time.Unix(),
			CursorID:        last.Uid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListCommentInboxMessagesResponse{
		Messages:      messages,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *MessageService) ListFollowInboxMessages(ctx context.Context, uid string, req *api.ListFollowInboxMessagesRequest) (*api.ListFollowInboxMessagesResponse, error) {
	token, err := decodeInboxPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	isReadFilter := readFilterToIsReadFilter(req.ReadFilter)
	rows, err := s.db.ListFollowInboxMessages(ctx, db.ListFollowInboxMessagesParams{
		ReceiverUid:     util.UUID(uid),
		IsRead:          isReadFilter,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: token.CursorCreatedAt > 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(token.CursorID), Valid: token.CursorID != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("list follow inbox messages: %w", err)
	}

	if len(rows) > 0 && req.ReadFilter != api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_READ {
		messageUids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			messageUids = append(messageUids, row.Uid)
		}
		if _, err := s.db.MarkInboxMessagesReadByUidsAndReceiver(ctx, db.MarkInboxMessagesReadByUidsAndReceiverParams{
			ReceiverUid: util.UUID(uid),
			Uids:        messageUids,
		}); err != nil {
			return nil, fmt.Errorf("mark follow inbox messages read: %w", err)
		}
	}

	messages := make([]*api.FollowInboxMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, &api.FollowInboxMessage{
			Uid:       row.Uid.String(),
			IsRead:    row.IsRead,
			CreatedAt: row.CreatedAt.Time.Unix(),
			Actor: &api.InboxMessageActor{
				Uid:       row.ActorUid.String(),
				Nickname:  row.ActorNickname,
				AvatarUrl: row.ActorAvatarUrl,
			},
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeInboxPageToken(inboxPageToken{
			CursorCreatedAt: last.CreatedAt.Time.Unix(),
			CursorID:        last.Uid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListFollowInboxMessagesResponse{
		Messages:      messages,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *MessageService) DeleteInboxMessage(ctx context.Context, uid string, req *api.DeleteInboxMessageRequest) error {
	affected, err := s.db.ArchiveInboxMessageByUidAndReceiver(ctx, db.ArchiveInboxMessageByUidAndReceiverParams{
		Uid:         util.UUID(req.Uid),
		ReceiverUid: util.UUID(uid),
	})
	if err != nil {
		return fmt.Errorf("delete inbox message: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("message not found or no permission")
	}
	return nil
}

func (s *MessageService) MarkAllInboxMessagesRead(ctx context.Context, uid string) (*api.MarkAllInboxMessagesReadResponse, error) {
	affected, err := s.db.MarkAllInboxMessagesReadByReceiver(ctx, util.UUID(uid))
	if err != nil {
		return nil, fmt.Errorf("mark all inbox messages read: %w", err)
	}
	return &api.MarkAllInboxMessagesReadResponse{
		UpdatedCount: int32(affected),
	}, nil
}

func (s *MessageService) CountUnreadInboxMessages(ctx context.Context, uid string) (*api.CountUnreadInboxMessagesResponse, error) {
	counts, err := s.db.CountUnreadInboxMessagesByReceiver(ctx, util.UUID(uid))
	if err != nil {
		return nil, fmt.Errorf("count unread inbox messages: %w", err)
	}
	return &api.CountUnreadInboxMessagesResponse{
		UnreadCount:        counts.UnreadCount,
		FollowUnreadCount:  counts.FollowUnreadCount,
		CommentUnreadCount: counts.CommentUnreadCount,
	}, nil
}

func readFilterToIsReadFilter(readFilter api.InboxMessageReadFilter) pgtype.Bool {
	switch readFilter {
	case api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_UNREAD:
		return pgtype.Bool{Bool: false, Valid: true}
	case api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_READ:
		return pgtype.Bool{Bool: true, Valid: true}
	default:
		return pgtype.Bool{Valid: false}
	}
}

type inboxPageToken struct {
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

func decodeInboxPageToken(pageToken string) (inboxPageToken, error) {
	if pageToken == "" {
		return inboxPageToken{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return inboxPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token inboxPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return inboxPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}
	return token, nil
}

func encodeInboxPageToken(token inboxPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
