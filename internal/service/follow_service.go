package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FollowService struct {
	db *db.Queries
}

func NewFollowService(dbx *sql.DB) *FollowService {
	return &FollowService{db: db.New(dbx)}
}

func (s *FollowService) Follow(ctx context.Context, uid string, req *api.FollowRequest) (*api.FollowResponse, error) {
	var followingCount int32
	var followersCount int32
	switch req.Action {
	case api.ToggleAction_TOGGLE_ACTION_ADD:
		row, err := s.db.AddFollow(ctx, db.AddFollowParams{
			FollowerUid: util.UUID(uid),
			FolloweeUid: util.UUID(req.Uid),
		})
		if err != nil {
			return nil, fmt.Errorf("follow: %w", err)
		}
		_, _ = s.db.CreateFollowInboxMessage(ctx, db.CreateFollowInboxMessageParams{
			ReceiverUid: util.UUID(req.Uid),
			ActorUid:    util.UUID(uid),
		})
		followingCount = row.FollowingCount
		followersCount = row.FollowersCount
	default:
		row, err := s.db.RemoveFollow(ctx, db.RemoveFollowParams{
			FollowerUid: util.UUID(uid),
			FolloweeUid: util.UUID(req.Uid),
		})
		if err != nil {
			return nil, fmt.Errorf("follow: %w", err)
		}
		followingCount = row.FollowingCount
		followersCount = row.FollowersCount
	}

	return &api.FollowResponse{
		FollowingCount: followingCount,
		FollowersCount: followersCount,
	}, nil
}

func (s *FollowService) ListMyFollowers(ctx context.Context, uid string, req *api.ListMyFollowersRequest) (*api.ListMyFollowersResponse, error) {
	rows, err := s.db.ListFollowers(ctx, db.ListFollowersParams{
		Uid:             util.UUID(uid),
		CursorCreatedAt: sql.NullTime{Time: time.Unix(req.CursorCreatedAt, 0).UTC(), Valid: req.CursorCreatedAt != 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(req.CursorId), Valid: req.CursorId != ""},
		Query:           sql.NullString{String: req.Query, Valid: req.Query != ""},
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

	var nextCursorCreatedAt int64
	var nextCursorID string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursorCreatedAt = last.FollowedAt.Unix()
		nextCursorID = last.Uid.String()
	}

	return &api.ListMyFollowersResponse{
		Users:               users,
		NextCursorCreatedAt: nextCursorCreatedAt,
		NextCursorId:        nextCursorID,
	}, nil
}

func (s *FollowService) ListMyFollowing(ctx context.Context, uid string, req *api.ListMyFollowingRequest) (*api.ListMyFollowingResponse, error) {
	rows, err := s.db.ListFollowing(ctx, db.ListFollowingParams{
		Uid:             util.UUID(uid),
		CursorCreatedAt: sql.NullTime{Time: time.Unix(req.CursorCreatedAt, 0).UTC(), Valid: req.CursorCreatedAt != 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(req.CursorId), Valid: req.CursorId != ""},
		Query:           sql.NullString{String: req.Query, Valid: req.Query != ""},
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

	var nextCursorCreatedAt int64
	var nextCursorID string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursorCreatedAt = last.FollowedAt.Unix()
		nextCursorID = last.Uid.String()
	}

	return &api.ListMyFollowingResponse{
		Users:               users,
		NextCursorCreatedAt: nextCursorCreatedAt,
		NextCursorId:        nextCursorID,
	}, nil
}
