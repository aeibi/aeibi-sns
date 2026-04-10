package service

import (
	"aeibi/api"
	"aeibi/internal/config"
	"aeibi/internal/repository/db"
	"aeibi/internal/repository/oss"
	"aeibi/util"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db  *db.Queries
	dbx *sql.DB
	oss *oss.OSS
	cfg *config.Config
}

func NewUserService(dbx *sql.DB, ossClient *oss.OSS, cfg *config.Config) *UserService {
	return &UserService{
		db:  db.New(dbx),
		dbx: dbx,
		oss: ossClient,
		cfg: cfg,
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
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		err = qtx.CreateUser(ctx, db.CreateUserParams{
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
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetUser(ctx context.Context, viewerUid string, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	row, err := s.db.GetUserByUid(ctx, util.UUID(req.Uid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	isFollowing := false
	if viewerUid != "" && viewerUid != req.Uid {
		isFollowing, err = s.db.IsFollowing(ctx, db.IsFollowingParams{
			FollowerUid: util.UUID(viewerUid),
			FolloweeUid: util.UUID(req.Uid),
		})
		if err != nil {
			return nil, fmt.Errorf("get follow: %w", err)
		}
	}
	return &api.GetUserResponse{
		User: &api.User{
			Uid: row.Uid.String(),
			// Username:       row.Username,
			Role: string(row.Role),
			// Email:          row.Email,
			Nickname:       row.Nickname,
			AvatarUrl:      row.AvatarUrl,
			FollowersCount: row.FollowersCount,
			FollowingCount: row.FollowingCount,
			IsFollowing:    isFollowing,
			Description:    row.Description,
		},
	}, nil
}

func (s *UserService) SearchUsers(ctx context.Context, req *api.SearchUsersRequest) (*api.SearchUsersResponse, error) {
	rows, err := s.db.SearchUsers(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
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
			Description:    row.Description,
		})
	}

	return &api.SearchUsersResponse{
		Users: users,
	}, nil
}

func (s *UserService) SuggestUsersByPrefix(ctx context.Context, req *api.SuggestUsersByPrefixRequest) (*api.SuggestUsersByPrefixResponse, error) {
	rows, err := s.db.SuggestUsersByNicknamePrefix(ctx, req.Prefix)
	if err != nil {
		return nil, fmt.Errorf("suggest users by prefix: %w", err)
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
			Description:    row.Description,
		})
	}

	return &api.SuggestUsersByPrefixResponse{
		Users: users,
	}, nil
}

func (s *UserService) GetMe(ctx context.Context, uid string) (*api.GetMeResponse, error) {
	row, err := s.db.GetUserByUid(ctx, util.UUID(uid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
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
			IsFollowing:    false,
			Description:    row.Description,
		},
	}, nil
}

func (s *UserService) UpdateMe(ctx context.Context, uid string, req *api.UpdateMeRequest) error {
	params := db.UpdateUserParams{Uid: util.UUID(uid)}
	paths := make(map[string]struct{}, len(req.UpdateMask.GetPaths()))
	for _, path := range req.UpdateMask.GetPaths() {
		paths[path] = struct{}{}
	}
	if _, ok := paths["username"]; ok {
		params.Username = sql.NullString{String: req.User.Username, Valid: true}
	}
	if _, ok := paths["email"]; ok {
		params.Email = sql.NullString{String: req.User.Email, Valid: true}
	}
	if _, ok := paths["nickname"]; ok {
		params.Nickname = sql.NullString{String: req.User.Nickname, Valid: true}
	}
	if _, ok := paths["avatar_url"]; ok {
		params.AvatarUrl = sql.NullString{String: req.User.AvatarUrl, Valid: true}
	}
	err := s.db.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *UserService) ChangePassword(ctx context.Context, uid string, req *api.ChangePasswordRequest) error {
	userUID := util.UUID(uid)
	return db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		passwordHash, err := qtx.GetUserPasswordHashByUid(ctx, userUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("user not found")
			}
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
			Uid:          userUID,
			PasswordHash: string(newPasswordHash),
		})
		if err != nil {
			return fmt.Errorf("update user password: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("user not found")
		}
		if _, err := qtx.DeleteRefreshTokenByUid(ctx, userUID); err != nil {
			return fmt.Errorf("clear refresh token: %w", err)
		}
		return nil
	})
}

func (s *UserService) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	var resp *api.LoginResponse
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		row, err := qtx.GetUserByUsername(ctx, req.Account)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("invalid credentials")
			}
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
			ExpiresAt: time.Now().Add(s.cfg.Auth.RefreshTTL),
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
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		row, err := qtx.GetRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("invalid refresh token")
			}
			return fmt.Errorf("get refresh token: %w", err)
		}

		now := time.Now()
		uid := row.Uid
		accessToken, refreshToken, err := s.genToken(uid.String())
		if err != nil {
			return err
		}

		if err := qtx.UpsertRefreshToken(ctx, db.UpsertRefreshTokenParams{
			Uid:       uid,
			Token:     refreshToken,
			ExpiresAt: now.Add(s.cfg.Auth.RefreshTTL),
		}); err != nil {
			return fmt.Errorf("save refresh token: %w", err)
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
