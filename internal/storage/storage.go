package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound      = errors.New("file not found")
	ErrInvalidKey    = errors.New("invalid storage key")
	ErrUploadFailed  = errors.New("upload failed")
	ErrDeleteFailed  = errors.New("delete failed")
	ErrNotConfigured = errors.New("storage not configured")
)

// Storage defines the interface for file storage backends
type Storage interface {
	// Upload stores a file and returns the storage key
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error

	// Download retrieves a file by key
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a file by key
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a URL to access the file (may be signed/temporary for S3)
	URL(ctx context.Context, key string) (string, error)

	// Type returns the storage type identifier
	Type() string
}

// Config holds storage configuration
type Config struct {
	Type string // "local" or "s3"

	// Local storage
	LocalPath string // Base path for local storage

	// S3/MinIO storage
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKey       string
	S3SecretKey       string
	S3UseSSL          bool
	S3PresignedExpiry int // URL expiry in seconds
}
