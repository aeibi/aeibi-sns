package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/internal/repository/oss"
	"aeibi/util"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileService struct {
	db                 *db.Queries
	pool               *pgxpool.Pool
	oss                *oss.OSS
	maxUploadSizeBytes int
}

func NewFileService(pool *pgxpool.Pool, ossClient *oss.OSS, maxUploadSizeKB int) *FileService {
	return &FileService{
		db:                 db.New(pool),
		pool:               pool,
		oss:                ossClient,
		maxUploadSizeBytes: maxUploadSizeKB * 1024,
	}
}

func (s *FileService) UploadFile(ctx context.Context, uploader string, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	if s.maxUploadSizeBytes > 0 && len(req.Data) > s.maxUploadSizeBytes {
		return nil, status.Errorf(codes.InvalidArgument, "file size exceeds %dKB", s.maxUploadSizeBytes/1024)
	}

	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	url := "/file/" + uuid.NewString() + path.Ext(req.Name)
	objectKey := strings.TrimPrefix(url, "/")
	if _, err := s.oss.PutObject(ctx, objectKey, req.Data, contentType); err != nil {
		return nil, fmt.Errorf("upload object: %w", err)
	}
	row, err := s.db.CreateFile(ctx, db.CreateFileParams{
		Url:         url,
		Name:        req.Name,
		ContentType: contentType,
		Size:        int64(len(req.Data)),
		Checksum:    req.Checksum,
		Uploader:    util.UUID(uploader),
	})
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	return &api.UploadFileResponse{
		File: &api.File{
			Name:        row.Name,
			ContentType: row.ContentType,
			Size:        row.Size,
			Checksum:    row.Checksum,
			Uploader:    row.Uploader.String(),
		},
		Url: row.Url,
	}, nil
}

func (s *FileService) GetFileMeta(ctx context.Context, viewerUID string, req *api.GetFileMetaRequest) (*api.GetFileMetaResponse, error) {
	row, err := s.db.GetFileByURL(ctx, req.Url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("get file: %w", err)
	}

	return &api.GetFileMetaResponse{
		File: &api.File{
			Name:        row.Name,
			ContentType: row.ContentType,
			Size:        row.Size,
			Checksum:    row.Checksum,
			Uploader:    row.Uploader.String(),
			CreatedAt:   row.CreatedAt.Time.Unix(),
		},
		Url: row.Url,
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, viewerUID string, req *api.GetFileRequest) (*httpbody.HttpBody, error) {
	reader, _, err := s.oss.GetObject(ctx, strings.TrimPrefix(req.Url, "/"))
	if err != nil {
		if errors.Is(err, oss.ErrObjectNotFound) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	contentType := "application/octet-stream"
	return &httpbody.HttpBody{
		ContentType: contentType,
		Data:        data,
	}, nil
}
