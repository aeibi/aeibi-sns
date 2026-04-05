package controller

import (
	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/service"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CommentHandler struct {
	api.UnimplementedCommentServiceServer
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) CreateTopComment(ctx context.Context, req *api.CreateTopCommentRequest) (*api.CreateTopCommentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.PostUid == "" {
		return nil, status.Error(codes.InvalidArgument, "post_uid is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.CreateTopComment(ctx, uid, req)
}

func (h *CommentHandler) CreateReply(ctx context.Context, req *api.CreateReplyRequest) (*api.CreateReplyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.ParentUid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.CreateReply(ctx, uid, req)
}

func (h *CommentHandler) ListTopComments(ctx context.Context, req *api.ListTopCommentsRequest) (*api.ListTopCommentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.PostUid == "" {
		return nil, status.Error(codes.InvalidArgument, "post_uid is required")
	}
	if (req.CursorCreatedAt == 0 && req.CursorId != "") || (req.CursorCreatedAt != 0 && req.CursorId == "") {
		return nil, status.Error(codes.InvalidArgument, "cursor is required")
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.ListTopComments(ctx, viewerUid, req)
}

func (h *CommentHandler) ListReplies(ctx context.Context, req *api.ListRepliesRequest) (*api.ListRepliesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.ListReplies(ctx, viewerUid, req)
}

func (h *CommentHandler) GetComment(ctx context.Context, req *api.GetCommentRequest) (*api.GetCommentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.GetComment(ctx, viewerUid, req)
}

func (h *CommentHandler) DeleteComment(ctx context.Context, req *api.DeleteCommentRequest) (*emptypb.Empty, error) {
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
	if err := h.svc.DeleteComment(ctx, uid, req); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *CommentHandler) LikeComment(ctx context.Context, req *api.LikeCommentRequest) (*api.LikeCommentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	switch req.Action {
	case api.ToggleAction_TOGGLE_ACTION_ADD, api.ToggleAction_TOGGLE_ACTION_REMOVE:
	default:
		return nil, status.Error(codes.InvalidArgument, "action is invalid")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.LikeComment(ctx, uid, req)
}
