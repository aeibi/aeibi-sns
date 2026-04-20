package controller

import (
	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/service"
	"context"

	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileHandler struct {
	api.UnimplementedFileServiceServer
	svc *service.FileService
}

func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

func (h *FileHandler) UploadFile(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "file data is empty")
	}

	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return h.svc.UploadFile(ctx, uid, req)
}

func (h *FileHandler) GetFileMeta(ctx context.Context, req *api.GetFileMetaRequest) (*api.GetFileMetaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.GetFileMeta(ctx, viewerUid, req)
}

func (h *FileHandler) GetFile(ctx context.Context, req *api.GetFileRequest) (*httpbody.HttpBody, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.GetFile(ctx, viewerUid, req)
}
