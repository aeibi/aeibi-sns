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

type MessageHandler struct {
	api.UnimplementedMessageServiceServer
	svc *service.MessageService
}

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

func (h *MessageHandler) ListCommentInboxMessages(ctx context.Context, req *api.ListCommentInboxMessagesRequest) (*api.ListCommentInboxMessagesResponse, error) {
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
	return h.svc.ListCommentInboxMessages(ctx, uid, req)
}

func (h *MessageHandler) ListFollowInboxMessages(ctx context.Context, req *api.ListFollowInboxMessagesRequest) (*api.ListFollowInboxMessagesResponse, error) {
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
	return h.svc.ListFollowInboxMessages(ctx, uid, req)
}

func (h *MessageHandler) DeleteInboxMessage(ctx context.Context, req *api.DeleteInboxMessageRequest) (*emptypb.Empty, error) {
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
	if err := h.svc.DeleteInboxMessage(ctx, uid, req); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *MessageHandler) MarkAllInboxMessagesRead(ctx context.Context, _ *emptypb.Empty) (*api.MarkAllInboxMessagesReadResponse, error) {
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.MarkAllInboxMessagesRead(ctx, uid)
}

func (h *MessageHandler) CountUnreadInboxMessages(ctx context.Context, _ *emptypb.Empty) (*api.CountUnreadInboxMessagesResponse, error) {
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.CountUnreadInboxMessages(ctx, uid)
}
