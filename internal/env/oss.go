package env

import (
	"context"
	"fmt"
	"time"

	"aeibi/internal/config"
	"aeibi/internal/repository/oss"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// InitOSS ensures the configured bucket exists and returns an OSS client wrapper.
func InitOSS(ctx context.Context, cfg config.OSSConfig) (*oss.OSS, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	bucketCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(bucketCtx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket %q: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(bucketCtx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %q: %w", cfg.Bucket, err)
		}
	}

	return oss.New(client, cfg.Bucket), nil
}
