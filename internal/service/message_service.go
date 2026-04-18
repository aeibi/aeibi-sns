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
	receiverUID := util.UUID(uid)

	token, err := decodeInboxPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	isReadFilter := readFilterToIsReadFilter(req.ReadFilter)
	rows, err := s.db.ListCommentInboxMessages(ctx, db.ListCommentInboxMessagesParams{
		ReceiverUid:     receiverUID,
		IsRead:          isReadFilter,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		CursorID:        util.UUID(token.CursorID),
	})
	if err != nil {
		return nil, fmt.Errorf("list comment inbox messages: %w", err)
	}

	if len(rows) > 0 && req.ReadFilter != api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_READ {
		messageUids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			messageUids = append(messageUids, row.Uid)
		}
		if _, err := s.db.MarkCommentInboxMessagesReadByUIDsAndReceiver(ctx, db.MarkCommentInboxMessagesReadByUIDsAndReceiverParams{
			ReceiverUid: receiverUID,
			Uids:        messageUids,
		}); err != nil {
			return nil, fmt.Errorf("mark comment inbox messages read: %w", err)
		}
	}

	actorUIDs := make([]uuid.UUID, 0, len(rows))
	commentUIDs := make([]uuid.UUID, 0, len(rows)*2)
	postUIDs := make([]uuid.UUID, 0, len(rows)*2)
	seenActorUIDs := make(map[uuid.UUID]struct{}, len(rows))
	seenCommentUIDs := make(map[uuid.UUID]struct{}, len(rows)*2)
	seenPostUIDs := make(map[uuid.UUID]struct{}, len(rows)*2)
	for _, row := range rows {
		if _, ok := seenActorUIDs[row.ActorUid]; !ok {
			seenActorUIDs[row.ActorUid] = struct{}{}
			actorUIDs = append(actorUIDs, row.ActorUid)
		}
		if _, ok := seenCommentUIDs[row.CommentUid]; !ok {
			seenCommentUIDs[row.CommentUid] = struct{}{}
			commentUIDs = append(commentUIDs, row.CommentUid)
		}
		if _, ok := seenPostUIDs[row.PostUid]; !ok {
			seenPostUIDs[row.PostUid] = struct{}{}
			postUIDs = append(postUIDs, row.PostUid)
		}
		if row.ParentCommentUid.Valid {
			parentUID := row.ParentCommentUid.UUID
			if _, ok := seenCommentUIDs[parentUID]; !ok {
				seenCommentUIDs[parentUID] = struct{}{}
				commentUIDs = append(commentUIDs, parentUID)
			}
			// Backward compatibility: legacy records may store post uid in parent_comment_uid.
			if _, ok := seenPostUIDs[parentUID]; !ok {
				seenPostUIDs[parentUID] = struct{}{}
				postUIDs = append(postUIDs, parentUID)
			}
		}
	}

	userRows, err := s.db.GetUsersByUIDs(ctx, actorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get inbox actors: %w", err)
	}
	userMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(userRows))
	for _, row := range userRows {
		userMap[row.Uid] = row
	}

	commentRows, err := s.db.GetCommentsByUIDs(ctx, commentUIDs)
	if err != nil {
		return nil, fmt.Errorf("get inbox comments: %w", err)
	}
	commentMap := make(map[uuid.UUID]db.GetCommentsByUIDsRow, len(commentRows))
	for _, row := range commentRows {
		commentMap[row.Uid] = row
	}

	postRows, err := s.db.GetPostsByUIDs(ctx, postUIDs)
	if err != nil {
		return nil, fmt.Errorf("get inbox posts: %w", err)
	}
	postMap := make(map[uuid.UUID]db.Post, len(postRows))
	for _, row := range postRows {
		postMap[row.Uid] = row
	}

	messages := make([]*api.CommentInboxMessage, 0, len(rows))
	for _, row := range rows {
		actor, ok := userMap[row.ActorUid]
		if !ok {
			continue
		}

		commentContent := ""
		if commentRow, ok := commentMap[row.CommentUid]; ok {
			commentContent = commentRow.Content
		}

		parentUID := util.NullUUIDString(row.ParentCommentUid)
		parentContent := ""
		if row.ParentCommentUid.Valid {
			puid := row.ParentCommentUid.UUID
			if parentComment, ok := commentMap[puid]; ok {
				parentContent = parentComment.Content
			} else if parentPost, ok := postMap[puid]; ok {
				parentContent = parentPost.Text
			}
		}

		messages = append(messages, &api.CommentInboxMessage{
			Uid:            row.Uid.String(),
			IsRead:         row.ReadAt.Valid,
			CreatedAt:      row.CreatedAt.Time.Unix(),
			CommentUid:     row.CommentUid.String(),
			CommentContent: commentContent,
			PostUid:        row.PostUid.String(),
			ParentUid:      parentUID,
			ParentContent:  parentContent,
			Actor: &api.InboxMessageActor{
				Uid:       row.ActorUid.String(),
				Nickname:  actor.Nickname,
				AvatarUrl: actor.AvatarUrl,
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
	receiverUID := util.UUID(uid)

	token, err := decodeInboxPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	isReadFilter := readFilterToIsReadFilter(req.ReadFilter)
	rows, err := s.db.ListFollowInboxMessages(ctx, db.ListFollowInboxMessagesParams{
		ReceiverUid:     receiverUID,
		IsRead:          isReadFilter,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		CursorID:        util.UUID(token.CursorID),
	})
	if err != nil {
		return nil, fmt.Errorf("list follow inbox messages: %w", err)
	}

	if len(rows) > 0 && req.ReadFilter != api.InboxMessageReadFilter_INBOX_MESSAGE_READ_FILTER_READ {
		messageUids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			messageUids = append(messageUids, row.Uid)
		}
		if _, err := s.db.MarkFollowInboxMessagesReadByUIDsAndReceiver(ctx, db.MarkFollowInboxMessagesReadByUIDsAndReceiverParams{
			ReceiverUid: receiverUID,
			Uids:        messageUids,
		}); err != nil {
			return nil, fmt.Errorf("mark follow inbox messages read: %w", err)
		}
	}

	actorUIDs := make([]uuid.UUID, 0, len(rows))
	seenActorUIDs := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seenActorUIDs[row.ActorUid]; ok {
			continue
		}
		seenActorUIDs[row.ActorUid] = struct{}{}
		actorUIDs = append(actorUIDs, row.ActorUid)
	}

	userRows, err := s.db.GetUsersByUIDs(ctx, actorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get inbox actors: %w", err)
	}
	userMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(userRows))
	for _, row := range userRows {
		userMap[row.Uid] = row
	}

	messages := make([]*api.FollowInboxMessage, 0, len(rows))
	for _, row := range rows {
		actor, ok := userMap[row.ActorUid]
		if !ok {
			continue
		}
		messages = append(messages, &api.FollowInboxMessage{
			Uid:       row.Uid.String(),
			IsRead:    row.ReadAt.Valid,
			CreatedAt: row.CreatedAt.Time.Unix(),
			Actor: &api.InboxMessageActor{
				Uid:       row.ActorUid.String(),
				Nickname:  actor.Nickname,
				AvatarUrl: actor.AvatarUrl,
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
	messageUID := util.UUID(req.Uid)
	receiverUID := util.UUID(uid)

	commentAffected, err := s.db.ArchiveCommentInboxMessageByUIDAndReceiver(ctx, db.ArchiveCommentInboxMessageByUIDAndReceiverParams{
		Uid:         messageUID,
		ReceiverUid: receiverUID,
	})
	if err != nil {
		return fmt.Errorf("delete comment inbox message: %w", err)
	}

	followAffected, err := s.db.ArchiveFollowInboxMessageByUIDAndReceiver(ctx, db.ArchiveFollowInboxMessageByUIDAndReceiverParams{
		Uid:         messageUID,
		ReceiverUid: receiverUID,
	})
	if err != nil {
		return fmt.Errorf("delete follow inbox message: %w", err)
	}

	if commentAffected+followAffected == 0 {
		return fmt.Errorf("message not found or no permission")
	}
	return nil
}

func (s *MessageService) MarkAllInboxMessagesRead(ctx context.Context, uid string) (*api.MarkAllInboxMessagesReadResponse, error) {
	receiverUID := util.UUID(uid)

	commentAffected, err := s.db.MarkAllCommentInboxMessagesReadByReceiver(ctx, receiverUID)
	if err != nil {
		return nil, fmt.Errorf("mark all comment inbox messages read: %w", err)
	}

	followAffected, err := s.db.MarkAllFollowInboxMessagesReadByReceiver(ctx, receiverUID)
	if err != nil {
		return nil, fmt.Errorf("mark all follow inbox messages read: %w", err)
	}

	return &api.MarkAllInboxMessagesReadResponse{
		UpdatedCount: int32(commentAffected + followAffected),
	}, nil
}

func (s *MessageService) CountUnreadInboxMessages(ctx context.Context, uid string) (*api.CountUnreadInboxMessagesResponse, error) {
	receiverUID := util.UUID(uid)

	commentUnreadCount, err := s.db.CountUnreadCommentInboxMessagesByReceiver(ctx, receiverUID)
	if err != nil {
		return nil, fmt.Errorf("count unread comment inbox messages: %w", err)
	}

	followUnreadCount, err := s.db.CountUnreadFollowInboxMessagesByReceiver(ctx, receiverUID)
	if err != nil {
		return nil, fmt.Errorf("count unread follow inbox messages: %w", err)
	}

	return &api.CountUnreadInboxMessagesResponse{
		UnreadCount:        commentUnreadCount + followUnreadCount,
		FollowUnreadCount:  followUnreadCount,
		CommentUnreadCount: commentUnreadCount,
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
	var token inboxPageToken
	if pageToken != "" {
		raw, err := base64.RawURLEncoding.DecodeString(pageToken)
		if err != nil {
			return inboxPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		if err := json.Unmarshal(raw, &token); err != nil {
			return inboxPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}
	if token.CursorCreatedAt == 0 || token.CursorID == "" {
		token.CursorCreatedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
		token.CursorID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
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
