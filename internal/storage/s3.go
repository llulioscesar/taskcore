package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 implements Storage using S3-compatible storage (AWS S3, MinIO, etc.)
type S3 struct {
	client        *minio.Client
	bucket        string
	presignExpiry time.Duration
}

// NewS3 creates a new S3-compatible storage
func NewS3(cfg Config) (*S3, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, err
	}

	expiry := time.Duration(cfg.S3PresignedExpiry) * time.Second
	if expiry == 0 {
		expiry = 1 * time.Hour // Default 1 hour
	}

	return &S3{
		client:        client,
		bucket:        cfg.S3Bucket,
		presignExpiry: expiry,
	}, nil
}

func (s *S3) Type() string {
	return "s3"
}

func (s *S3) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, opts)
	return err
}

func (s *S3) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}

	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	// Check if object exists
	_, err = obj.Stat()
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return obj, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}

	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3) URL(ctx context.Context, key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}

	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, s.presignExpiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// EnsureBucket creates the bucket if it doesn't exist
func (s *S3) EnsureBucket(ctx context.Context, region string) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: region})
	}
	return nil
}
