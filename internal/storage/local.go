package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local implements Storage using the local filesystem
type Local struct {
	basePath string
}

// NewLocal creates a new local filesystem storage
func NewLocal(basePath string) (*Local, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &Local{basePath: basePath}, nil
}

func (l *Local) Type() string {
	return "local"
}

func (l *Local) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	fullPath := filepath.Join(l.basePath, key)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (l *Local) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(l.basePath, key)
	file, err := os.Open(fullPath)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return file, err
}

func (l *Local) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	fullPath := filepath.Join(l.basePath, key)
	err := os.Remove(fullPath)
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	return err
}

func (l *Local) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}

	fullPath := filepath.Join(l.basePath, key)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (l *Local) URL(ctx context.Context, key string) (string, error) {
	// For local storage, return relative path (API will serve it)
	if err := validateKey(key); err != nil {
		return "", err
	}
	return "/api/v1/attachments/download/" + key, nil
}

// validateKey ensures the key doesn't contain path traversal
func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	// Prevent path traversal
	if strings.Contains(key, "..") {
		return ErrInvalidKey
	}
	return nil
}
