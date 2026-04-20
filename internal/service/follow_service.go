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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowService struct {
	db       *db.Queries
	pool     *pgxpool.Pool
	producer *async.Producer
}

func NewFollowService(pool *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) *FollowService {
	return &FollowService{
		db:       db.New(pool),
		pool:     pool,
		producer: async.New(riverClient),
	}
}

func (s *FollowService) Follow(ctx context.Context, uid string, req *api.FollowRequest) (*api.FollowResponse, error) {
	followerUID := util.UUID(uid)
	followeeUID := util.UUID(req.Uid)

	if followerUID == followeeUID {
		return nil, fmt.Errorf("follow: cannot follow yourself")
	}

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		if _, err := qtx.GetUserByUid(ctx, followeeUID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return status.Error(codes.NotFound, "user not found")
			}
			return fmt.Errorf("follow: get target user: %w", err)
		}

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertFollowEdge(ctx, db.InsertFollowEdgeParams{
				FollowerUid: followerUID,
				FolloweeUid: followeeUID,
			})
			if err != nil {
				return fmt.Errorf("follow: insert follow edge: %w", err)
			}

			if affected > 0 {
				if _, err := qtx.IncrementFollowingCount(ctx, followerUID); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return status.Error(codes.NotFound, "user not found")
					}
					return fmt.Errorf("follow: increment following_count: %w", err)
				}

				if _, err := qtx.IncrementFollowersCount(ctx, followeeUID); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return status.Error(codes.NotFound, "user not found")
					}
					return fmt.Errorf("follow: increment followers_count: %w", err)
				}

				if err := s.producer.EnqueueFollowInboxTx(ctx, tx, async.FollowInboxArgs{
					MessageUID:  uuid.New(),
					ReceiverUID: followeeUID,
					ActorUID:    followerUID,
				}); err != nil {
					return fmt.Errorf("follow: enqueue follow inbox job: %w", err)
				}
			}
		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			affected, err := qtx.DeleteFollowEdge(ctx, db.DeleteFollowEdgeParams{
				FollowerUid: followerUID,
				FolloweeUid: followeeUID,
			})
			if err != nil {
				return fmt.Errorf("follow: delete follow edge: %w", err)
			}

			if affected > 0 {
				if _, err := qtx.DecrementFollowingCount(ctx, followerUID); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return status.Error(codes.NotFound, "user not found")
					}
					return fmt.Errorf("follow: decrement following_count: %w", err)
				}

				if _, err := qtx.DecrementFollowersCount(ctx, followeeUID); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return status.Error(codes.NotFound, "user not found")
					}
					return fmt.Errorf("follow: decrement followers_count: %w", err)
				}
			}

		default:
			return fmt.Errorf("follow: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &api.FollowResponse{}, nil
}

func (s *FollowService) ListMyFollowers(ctx context.Context, uid string, req *api.ListMyFollowersRequest) (*api.ListMyFollowersResponse, error) {
	vuid := util.UUID(uid)
	query := strings.TrimSpace(req.GetQuery())
	queryText := pgtype.Text{String: query, Valid: query != ""}

	token, err := decodeFollowPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListFollowers(ctx, db.ListFollowersParams{
		Uid:             vuid,
		Query:           queryText,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		CursorID:        util.UUID(token.CursorID),
	})
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}

	userUIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		userUIDs = append(userUIDs, row.Uid)
	}

	followingUIDs, err := s.db.ListFollowingUIDsByFollowerAndFolloweeUIDs(ctx, db.ListFollowingUIDsByFollowerAndFolloweeUIDsParams{
		FollowerUid:  vuid,
		FolloweeUids: userUIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("list following uids by follower and followee uids: %w", err)
	}

	followingSet := make(map[uuid.UUID]struct{}, len(followingUIDs))
	for _, followeeUID := range followingUIDs {
		followingSet[followeeUID] = struct{}{}
	}

	users := make([]*api.User, 0, len(rows))
	for _, row := range rows {
		_, isFollowing := followingSet[row.Uid]

		users = append(users, &api.User{
			Uid:            row.Uid.String(),
			Role:           string(row.Role),
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			IsFollowing:    isFollowing,
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextPageToken, err = encodeFollowPageToken(followPageToken{
			CursorCreatedAt: last.FollowedAt.Time.Unix(),
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
	vuid := util.UUID(uid)
	query := strings.TrimSpace(req.GetQuery())
	queryText := pgtype.Text{String: query, Valid: query != ""}

	token, err := decodeFollowPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListFollowing(ctx, db.ListFollowingParams{
		Uid:             vuid,
		Query:           queryText,
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		CursorID:        util.UUID(token.CursorID),
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
			CursorCreatedAt: last.FollowedAt.Time.Unix(),
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
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

func decodeFollowPageToken(pageToken string) (followPageToken, error) {
	var token followPageToken
	if pageToken != "" {
		raw, err := base64.RawURLEncoding.DecodeString(pageToken)
		if err != nil {
			return followPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		if err := json.Unmarshal(raw, &token); err != nil {
			return followPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}

	if token.CursorCreatedAt == 0 || token.CursorID == "" {
		token.CursorCreatedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
		token.CursorID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	}
	return token, nil
}

func encodeFollowPageToken(token followPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
