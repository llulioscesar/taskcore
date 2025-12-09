package attachment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("attachment not found")
	ErrInvalidFile = errors.New("invalid file")
	ErrFileTooLarge = errors.New("file too large")
	ErrInvalidType  = errors.New("file type not allowed")
)

// Attachment represents file metadata (actual file stored via storage.Storage)
type Attachment struct {
	ID          uuid.UUID  `db:"id"`
	IssueID     *uuid.UUID `db:"issue_id"`     // Attached to issue
	CommentID   *uuid.UUID `db:"comment_id"`   // Attached to comment
	Filename    string     `db:"filename"`     // Original filename
	StorageKey  string     `db:"storage_key"`  // Key in storage backend
	ContentType string     `db:"content_type"` // MIME type
	Size        int64      `db:"size"`         // Size in bytes
	UploadedBy  uuid.UUID  `db:"uploaded_by"`  // User who uploaded
	CreatedAt   time.Time  `db:"created_at"`
}

// Thumbnail for image attachments (optional, stored separately)
type Thumbnail struct {
	ID           uuid.UUID `db:"id"`
	AttachmentID uuid.UUID `db:"attachment_id"`
	StorageKey   string    `db:"storage_key"`
	Width        int       `db:"width"`
	Height       int       `db:"height"`
	CreatedAt    time.Time `db:"created_at"`
}

// Common MIME types
var (
	ImageTypes = []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	DocumentTypes = []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"text/csv",
	}

	ArchiveTypes = []string{
		"application/zip",
		"application/x-tar",
		"application/gzip",
		"application/x-7z-compressed",
	}
)

// IsImage checks if attachment is an image
func (a *Attachment) IsImage() bool {
	for _, t := range ImageTypes {
		if a.ContentType == t {
			return true
		}
	}
	return false
}

// GenerateStorageKey creates a unique storage key
func GenerateStorageKey(issueID uuid.UUID, filename string) string {
	return issueID.String() + "/" + uuid.New().String() + "/" + filename
}
