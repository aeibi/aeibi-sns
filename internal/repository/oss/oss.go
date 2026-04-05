package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type OSS struct {
	client *minio.Client
	bucket string
}

func New(client *minio.Client, bucket string) *OSS {
	return &OSS{
		client: client,
		bucket: bucket,
	}
}

var ErrObjectNotFound = errors.New("object not found")

// PutObject uploads data to the configured bucket and returns the object key.
func (o *OSS) PutObject(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	if o == nil || o.client == nil {
		return "", errors.New("oss client is nil")
	}
	if o.bucket == "" {
		return "", errors.New("bucket is empty")
	}
	if objectName == "" {
		return "", errors.New("object name is empty")
	}
	if len(data) == 0 {
		return "", errors.New("object data is empty")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	reader := bytes.NewReader(data)
	if _, err := o.client.PutObject(ctx, o.bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return objectName, nil
}

// GetObject fetches an object reader and its metadata.
func (o *OSS) GetObject(ctx context.Context, objectName string) (io.ReadCloser, minio.ObjectInfo, error) {
	var emptyInfo minio.ObjectInfo

	if o == nil || o.client == nil {
		return nil, emptyInfo, errors.New("oss client is nil")
	}
	if o.bucket == "" {
		return nil, emptyInfo, errors.New("bucket is empty")
	}
	if objectName == "" {
		return nil, emptyInfo, errors.New("object name is empty")
	}

	obj, err := o.client.GetObject(ctx, o.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, emptyInfo, fmt.Errorf("get object: %w", err)
	}

	info, err := obj.Stat()
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.Code == "NotFound" {
			return nil, emptyInfo, ErrObjectNotFound
		}
		return nil, emptyInfo, fmt.Errorf("stat object: %w", err)
	}

	return obj, info, nil
}
