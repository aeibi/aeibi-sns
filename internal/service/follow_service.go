package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowService struct {
	db  *db.Queries
	dbx *sql.DB
}

func NewFollowService(dbx *sql.DB) *FollowService {
	return &FollowService{db: db.New(dbx), dbx: dbx}
}

func (s *FollowService) Follow(ctx context.Context, uid string, req *api.FollowRequest) (*api.FollowResponse, error) {
	followerUID := util.UUID(uid)
	followeeUID := util.UUID(req.Uid)

	if followerUID == followeeUID {
		return nil, fmt.Errorf("follow: cannot follow yourself")
	}

	var followingCount int32
	var followersCount int32
	var createdFollowMessage bool

	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			applied, err := qtx.InsertFollowEdge(ctx, db.InsertFollowEdgeParams{
				FollowerUid: followerUID,
				FolloweeUid: followeeUID,
			})
			if err != nil {
				return fmt.Errorf("follow: insert follow edge: %w", err)
			}

			if applied {
				followingCount, err = qtx.IncrementFollowingCount(ctx, followerUID)
				if err != nil {
					return fmt.Errorf("follow: increment following_count: %w", err)
				}

				followersCount, err = qtx.IncrementFollowersCount(ctx, followeeUID)
				if err != nil {
					return fmt.Errorf("follow: increment followers_count: %w", err)
				}

				createdFollowMessage = true
			} else {
				row, err := qtx.GetFollowCounts(ctx, db.GetFollowCountsParams{
					FollowerUid: followerUID,
					FolloweeUid: followeeUID,
				})
				if err != nil {
					return fmt.Errorf("follow: get follow counts: %w", err)
				}
				followingCount = row.FollowingCount
				followersCount = row.FollowersCount
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			applied, err := qtx.DeleteFollowEdge(ctx, db.DeleteFollowEdgeParams{
				FollowerUid: followerUID,
				FolloweeUid: followeeUID,
			})
			if err != nil {
				return fmt.Errorf("follow: delete follow edge: %w", err)
			}

			if applied {
				followingCount, err = qtx.DecrementFollowingCount(ctx, followerUID)
				if err != nil {
					return fmt.Errorf("follow: decrement following_count: %w", err)
				}

				followersCount, err = qtx.DecrementFollowersCount(ctx, followeeUID)
				if err != nil {
					return fmt.Errorf("follow: decrement followers_count: %w", err)
				}
			} else {
				row, err := qtx.GetFollowCounts(ctx, db.GetFollowCountsParams{
					FollowerUid: followerUID,
					FolloweeUid: followeeUID,
				})
				if err != nil {
					return fmt.Errorf("follow: get follow counts: %w", err)
				}
				followingCount = row.FollowingCount
				followersCount = row.FollowersCount
			}

		default:
			return fmt.Errorf("follow: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if createdFollowMessage {
		// TODO: move this to async processing
		_, _ = s.db.CreateFollowInboxMessage(ctx, db.CreateFollowInboxMessageParams{
			Uid:         uuid.New(),
			ReceiverUid: followeeUID,
			ActorUid:    followerUID,
		})
	}

	return &api.FollowResponse{
		FollowingCount: followingCount,
		FollowersCount: followersCount,
	}, nil
}

func (s *FollowService) ListMyFollowers(ctx context.Context, uid string, req *api.ListMyFollowersRequest) (*api.ListMyFollowersResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	cursorCreatedAt, cursorID, err := decodeAndValidateFollowPageToken(req.GetPageToken(), query)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListFollowers(ctx, db.ListFollowersParams{
		Uid:             util.UUID(uid),
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
		Query:           sql.NullString{String: query, Valid: query != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}

	users := make([]*api.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &api.User{
			Uid:            row.Uid.String(),
			Role:           string(row.Role),
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			IsFollowing:    row.Following,
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeFollowPageToken(followPageToken{
			Query:           query,
			CursorCreatedAt: last.FollowedAt.Unix(),
			CursorID:        last.Uid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListMyFollowersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *FollowService) ListMyFollowing(ctx context.Context, uid string, req *api.ListMyFollowingRequest) (*api.ListMyFollowingResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	cursorCreatedAt, cursorID, err := decodeAndValidateFollowPageToken(req.GetPageToken(), query)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListFollowing(ctx, db.ListFollowingParams{
		Uid:             util.UUID(uid),
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
		Query:           sql.NullString{String: query, Valid: query != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("list following: %w", err)
	}

	users := make([]*api.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &api.User{
			Uid:            row.Uid.String(),
			Role:           string(row.Role),
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			IsFollowing:    true,
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeFollowPageToken(followPageToken{
			Query:           query,
			CursorCreatedAt: last.FollowedAt.Unix(),
			CursorID:        last.Uid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListMyFollowingResponse{
		Users:         users,
		NextPageToken: nextPageToken,
	}, nil
}

type followPageToken struct {
	Query           string `json:"query,omitempty"`
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

func decodeAndValidateFollowPageToken(pageToken string, query string) (sql.NullTime, uuid.NullUUID, error) {
	if pageToken == "" {
		return sql.NullTime{}, uuid.NullUUID{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return sql.NullTime{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token followPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return sql.NullTime{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	if token.Query != query {
		return sql.NullTime{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "page_token does not match current filters")
	}
	if token.CursorCreatedAt <= 0 || token.CursorID == "" {
		return sql.NullTime{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	cursorID, err := uuid.Parse(token.CursorID)
	if err != nil {
		return sql.NullTime{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	return sql.NullTime{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true}, uuid.NullUUID{UUID: cursorID, Valid: true}, nil
}

func encodeFollowPageToken(token followPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
