package storage

import "fmt"

// New creates a storage instance based on configuration
func New(cfg Config) (Storage, error) {
	switch cfg.Type {
	case "local", "":
		if cfg.LocalPath == "" {
			cfg.LocalPath = "./data/attachments"
		}
		return NewLocal(cfg.LocalPath)

	case "s3", "minio":
		if cfg.S3Bucket == "" {
			return nil, fmt.Errorf("S3_BUCKET is required for S3 storage")
		}
		if cfg.S3Endpoint == "" {
			return nil, fmt.Errorf("S3_ENDPOINT is required for S3 storage")
		}
		return NewS3(cfg)

	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}
}
