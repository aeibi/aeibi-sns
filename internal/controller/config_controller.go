package controller

import (
	"aeibi/api"
	"aeibi/internal/service"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConfigHandler struct {
	api.UnimplementedConfigServiceServer
	svc *service.ConfigService
}

func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

func (h *ConfigHandler) GetFrontendConfig(ctx context.Context, req *api.GetFrontendConfigRequest) (*api.GetFrontendConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	return h.svc.GetFrontendConfig(ctx, req)
}

func (h *ConfigHandler) GetUploadConfig(ctx context.Context, req *api.GetUploadConfigRequest) (*api.GetUploadConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	return h.svc.GetUploadConfig(ctx, req)
}
