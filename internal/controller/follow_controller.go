package controller

import (
	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/service"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowHandler struct {
	api.UnimplementedFollowServiceServer
	svc *service.FollowService
}

func NewFollowHandler(svc *service.FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
}

func (h *FollowHandler) Follow(ctx context.Context, req *api.FollowRequest) (*api.FollowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if uid == req.Uid {
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}
	return h.svc.Follow(ctx, uid, req)
}

func (h *FollowHandler) ListMyFollowers(ctx context.Context, req *api.ListMyFollowersRequest) (*api.ListMyFollowersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if (req.CursorCreatedAt == 0 && req.CursorId != "") || (req.CursorCreatedAt != 0 && req.CursorId == "") {
		return nil, status.Error(codes.InvalidArgument, "cursor is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.ListMyFollowers(ctx, uid, req)
}

func (h *FollowHandler) ListMyFollowing(ctx context.Context, req *api.ListMyFollowingRequest) (*api.ListMyFollowingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if (req.CursorCreatedAt == 0 && req.CursorId != "") || (req.CursorCreatedAt != 0 && req.CursorId == "") {
		return nil, status.Error(codes.InvalidArgument, "cursor is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.ListMyFollowing(ctx, uid, req)
}
