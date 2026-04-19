package service

import (
	"aeibi/api"
	"aeibi/internal/async"
	"aeibi/internal/config"
	"aeibi/internal/repository/db"
	"aeibi/internal/repository/oss"
	searchrepo "aeibi/internal/repository/search"
	"aeibi/util"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db       *db.Queries
	pool     *pgxpool.Pool
	oss      *oss.OSS
	search   *searchrepo.Search
	cfg      *config.Config
	producer *async.Producer
}

func NewUserService(pool *pgxpool.Pool, ossClient *oss.OSS, search *searchrepo.Search, cfg *config.Config, riverClient *river.Client[pgx.Tx]) *UserService {
	return &UserService{
		db:       db.New(pool),
		pool:     pool,
		oss:      ossClient,
		search:   search,
		producer: async.New(riverClient),
		cfg:      cfg,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *api.CreateUserRequest) error {
	uid := uuid.New()

	avatar, err := util.GenerateDefaultAvatar(uid.String())
	if err != nil {
		return fmt.Errorf("generate default avatar: %w", err)
	}
	avatarURL := fmt.Sprintf("/file/avatars/%s.png", uid)
	avatarObjectKey := strings.TrimPrefix(avatarURL, "/")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		_, err = qtx.CreateUser(ctx, db.CreateUserParams{
			Uid:          uid,
			Username:     req.Username,
			Email:        req.Email,
			Nickname:     req.Nickname,
			PasswordHash: string(passwordHash),
			AvatarUrl:    avatarURL,
		})
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		if _, err = s.oss.PutObject(ctx, avatarObjectKey, avatar, "image/png"); err != nil {
			return fmt.Errorf("upload avatar: %w", err)
		}
		if err := s.producer.EnqueueUpdateUserSearchTx(ctx, tx, async.UpdateUserSearchArgs{
			UserUID: uid,
			Action:  async.UserSearchActionUpsert,
		}); err != nil {
			return fmt.Errorf("enqueue update user search job: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetUser(ctx context.Context, viewerUid string, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	vuid := util.UUID(viewerUid)

	uid := util.UUID(req.Uid)

	row, err := s.db.GetUserByUid(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	isFollowing := false
	if viewerUid != "" && viewerUid != req.Uid {
		isFollowing, err = s.db.IsFollowing(ctx, db.IsFollowingParams{
			FollowerUid: vuid,
			FolloweeUid: uid,
		})
		if err != nil {
			return nil, fmt.Errorf("get follow: %w", err)
		}
	}

	return &api.GetUserResponse{
		User: &api.User{
			Uid:            row.Uid.String(),
			Role:           string(row.Role),
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			IsFollowing:    isFollowing,
			Description:    row.Description,
		},
	}, nil
}

func (s *UserService) SearchUsers(_ context.Context, req *api.SearchUsersRequest) (*api.SearchUsersResponse, error) {
	result, err := s.search.SearchUsers(searchrepo.SearchUsersParams{
		Query: req.Query,
		Limit: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}

	users := make([]*api.User, 0, len(result.Hits))
	for _, hit := range result.Hits {
		users = append(users, &api.User{
			Uid:         hit.UID,
			Nickname:    hit.Nickname,
			AvatarUrl:   hit.AvatarUrl,
			Description: hit.Description,
		})
	}

	return &api.SearchUsersResponse{
		Users: users,
	}, nil
}

func (s *UserService) SuggestUsersByPrefix(_ context.Context, req *api.SuggestUsersByPrefixRequest) (*api.SuggestUsersByPrefixResponse, error) {
	result, err := s.search.SuggestUsersByNickname(req.Prefix, 10)
	if err != nil {
		return nil, fmt.Errorf("suggest users by prefix: %w", err)
	}

	users := make([]*api.User, 0, len(result.Hits))
	for _, hit := range result.Hits {
		users = append(users, &api.User{
			Uid:         hit.UID,
			Nickname:    hit.Nickname,
			AvatarUrl:   hit.AvatarUrl,
			Description: hit.Description,
		})
	}

	return &api.SuggestUsersByPrefixResponse{
		Users: users,
	}, nil
}

func (s *UserService) GetMe(ctx context.Context, uid string) (*api.GetMeResponse, error) {
	vuid := util.UUID(uid)

	row, err := s.db.GetUserByUid(ctx, vuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &api.GetMeResponse{
		User: &api.User{
			Uid:            row.Uid.String(),
			Username:       row.Username,
			Role:           string(row.Role),
			Email:          row.Email,
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			Description:    row.Description,
		},
	}, nil
}

func (s *UserService) UpdateMe(ctx context.Context, uid string, req *api.UpdateMeRequest) error {
	vuid := util.UUID(uid)

	params := db.UpdateUserParams{Uid: vuid}
	paths := make(map[string]struct{}, len(req.UpdateMask.GetPaths()))
	for _, path := range req.UpdateMask.GetPaths() {
		paths[path] = struct{}{}
	}
	if _, ok := paths["username"]; ok {
		params.Username = pgtype.Text{String: req.User.Username, Valid: true}
	}
	if _, ok := paths["email"]; ok {
		params.Email = pgtype.Text{String: req.User.Email, Valid: true}
	}
	if _, ok := paths["nickname"]; ok {
		params.Nickname = pgtype.Text{String: req.User.Nickname, Valid: true}
	}
	if _, ok := paths["avatar_url"]; ok {
		params.AvatarUrl = pgtype.Text{String: req.User.AvatarUrl, Valid: true}
	}

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		_, err := qtx.UpdateUser(ctx, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("user not found")
			}
			return fmt.Errorf("update user: %w", err)
		}

		if err := s.producer.EnqueueUpdateUserSearchTx(ctx, tx, async.UpdateUserSearchArgs{
			UserUID: vuid,
			Action:  async.UserSearchActionUpsert,
		}); err != nil {
			return fmt.Errorf("enqueue update user search job: %w", err)
		}

		return nil
	})
}

func (s *UserService) ChangePassword(ctx context.Context, uid string, req *api.ChangePasswordRequest) error {
	vuid := util.UUID(uid)

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		passwordHash, err := qtx.GetUserPasswordHashByUid(ctx, vuid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		if err != nil {
			return fmt.Errorf("get user password: %w", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.OldPassword)); err != nil {
			return fmt.Errorf("invalid old password")
		}

		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}

		affected, err := qtx.UpdateUserPasswordByUid(ctx, db.UpdateUserPasswordByUidParams{
			Uid:          vuid,
			PasswordHash: string(newPasswordHash),
		})
		if err != nil {
			return fmt.Errorf("update user password: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("user not found")
		}

		if _, err := qtx.DeleteRefreshTokenByUid(ctx, vuid); err != nil {
			return fmt.Errorf("clear refresh token: %w", err)
		}

		return nil
	})
}

func (s *UserService) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	var resp *api.LoginResponse

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		row, err := qtx.GetUserByUsername(ctx, req.Account)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("invalid credentials")
		}
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(req.Password)); err != nil {
			return fmt.Errorf("invalid credentials")
		}
		accessToken, refreshToken, err := s.genToken(row.Uid.String())
		if err != nil {
			return err
		}

		if err := qtx.UpsertRefreshToken(ctx, db.UpsertRefreshTokenParams{
			Uid:       row.Uid,
			Token:     refreshToken,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.cfg.Auth.RefreshTTL), Valid: true},
		}); err != nil {
			return fmt.Errorf("save refresh token: %w", err)
		}

		resp = &api.LoginResponse{
			Tokens: &api.TokenPair{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserService) RefreshToken(ctx context.Context, req *api.RefreshTokenRequest) (*api.RefreshTokenResponse, error) {
	var resp *api.RefreshTokenResponse

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		refreshToken, err := util.RandomString64()
		if err != nil {
			return fmt.Errorf("generate refresh token: %w", err)
		}

		tokenUID, err := qtx.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
			NewToken:  refreshToken,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.cfg.Auth.RefreshTTL), Valid: true},
			OldToken:  req.RefreshToken,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("invalid refresh token")
		}
		if err != nil {
			return fmt.Errorf("rotate refresh token: %w", err)
		}

		accessToken, err := util.GenerateJWT(tokenUID.String(), s.cfg.Auth.JWTSecret, s.cfg.Auth.JWTIssuer, s.cfg.Auth.JWTTTL)
		if err != nil {
			return fmt.Errorf("generate access token: %w", err)
		}

		resp = &api.RefreshTokenResponse{
			Tokens: &api.TokenPair{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserService) Logout(ctx context.Context, uid string) error {
	if _, err := s.db.DeleteRefreshTokenByUid(ctx, util.UUID(uid)); err != nil {
		return fmt.Errorf("clear refresh token: %w", err)
	}
	return nil
}

func (s *UserService) genToken(uid string) (string, string, error) {
	accessToken, err := util.GenerateJWT(uid, s.cfg.Auth.JWTSecret, s.cfg.Auth.JWTIssuer, s.cfg.Auth.JWTTTL)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err := util.RandomString64()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}
