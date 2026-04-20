package controller

import (
	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/service"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReportHandler struct {
	api.UnimplementedReportServiceServer
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) CreateReport(ctx context.Context, req *api.CreateReportRequest) (*api.CreateReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	switch req.ReportTargetType {
	case api.ReportTargetType_REPORT_TARGET_TYPE_POST, api.ReportTargetType_REPORT_TARGET_TYPE_COMMENT, api.ReportTargetType_REPORT_TARGET_TYPE_USER:
	default:
		return nil, status.Error(codes.InvalidArgument, "report_target_type is invalid")
	}
	if req.TargetUid == "" {
		return nil, status.Error(codes.InvalidArgument, "target_uid is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	resp, err := h.svc.CreateReport(ctx, uid, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
