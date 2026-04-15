package controller

import (
	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/service"
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PostHandler struct {
	api.UnimplementedPostServiceServer
	svc *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) CreatePost(ctx context.Context, req *api.CreatePostRequest) (*api.CreatePostResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.CreatePost(ctx, uid, req)
}

func (h *PostHandler) GetPost(ctx context.Context, req *api.GetPostRequest) (*api.GetPostResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.GetPost(ctx, viewerUid, req)
}

func (h *PostHandler) ListPosts(ctx context.Context, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if strings.TrimSpace(req.Query) != "" {
		return nil, status.Error(codes.InvalidArgument, "query is not supported")
	}
	req.Query = ""
	req.TagName = strings.TrimSpace(req.TagName)
	req.AuthorUid = strings.TrimSpace(req.AuthorUid)
	if req.AuthorUid != "" {
		if _, err := uuid.Parse(req.AuthorUid); err != nil {
			return nil, status.Error(codes.InvalidArgument, "author_uid is invalid")
		}
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.ListPosts(ctx, viewerUid, req)
}

func (h *PostHandler) SearchPosts(ctx context.Context, req *api.SearchPostsRequest) (*api.ListPostsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	req.Query = strings.TrimSpace(req.Query)
	req.TagName = strings.TrimSpace(req.TagName)
	req.AuthorUid = strings.TrimSpace(req.AuthorUid)
	if req.AuthorUid != "" {
		if _, err := uuid.Parse(req.AuthorUid); err != nil {
			return nil, status.Error(codes.InvalidArgument, "author_uid is invalid")
		}
	}
	viewerUid, _ := auth.SubjectFromContext(ctx)
	return h.svc.SearchPosts(ctx, viewerUid, req)
}

func (h *PostHandler) ListMyCollections(ctx context.Context, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if strings.TrimSpace(req.Query) != "" || strings.TrimSpace(req.AuthorUid) != "" || strings.TrimSpace(req.TagName) != "" {
		return nil, status.Error(codes.InvalidArgument, "query, author_uid, tag_name are not supported")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return h.svc.ListMyCollections(ctx, uid, req)
}

func (h *PostHandler) SearchTags(ctx context.Context, req *api.SearchTagsRequest) (*api.SearchTagsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}
	return h.svc.SearchTags(ctx, req)
}

func (h *PostHandler) SuggestTagsByPrefix(ctx context.Context, req *api.SuggestTagsByPrefixRequest) (*api.SuggestTagsByPrefixResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Prefix == "" {
		return nil, status.Error(codes.InvalidArgument, "prefix is required")
	}
	return h.svc.SuggestTagsByPrefix(ctx, req)
}

func (h *PostHandler) UpdatePost(ctx context.Context, req *api.UpdatePostRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Uid == "" {
		return nil, status.Error(codes.InvalidArgument, "uid is required")
	}
	if req.Post == nil {
		return nil, status.Error(codes.InvalidArgument, "post is required")
	}
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, status.Error(codes.InvalidArgument, "update_mask is required")
	}
	uid, ok := auth.SubjectFromContext(ctx)
	if !ok || uid == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if err := h.svc.UpdatePost(ctx, uid, req); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *PostHandler) DeletePost(ctx context.Context, req *api.DeletePostRequest) (*emptypb.Empty, error) {
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
	if err := h.svc.DeletePost(ctx, uid, req); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *PostHandler) LikePost(ctx context.Context, req *api.LikePostRequest) (*api.LikePostResponse, error) {
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
	return h.svc.LikePost(ctx, uid, req)
}

func (h *PostHandler) CollectPost(ctx context.Context, req *api.CollectPostRequest) (*api.CollectPostResponse, error) {
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
	return h.svc.CollectPost(ctx, uid, req)
}
