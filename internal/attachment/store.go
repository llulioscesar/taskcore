package attachment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, a *Attachment) error {
	query := `
		INSERT INTO attachments (issue_id, comment_id, filename, storage_key, content_type, size, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query,
		a.IssueID, a.CommentID, a.Filename, a.StorageKey, a.ContentType, a.Size, a.UploadedBy).
		Scan(&a.ID, &a.CreatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Attachment, error) {
	a := &Attachment{}
	err := db.GetContext(ctx, a, `SELECT * FROM attachments WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func GetByStorageKey(ctx context.Context, db *sqlx.DB, key string) (*Attachment, error) {
	a := &Attachment{}
	err := db.GetContext(ctx, a, `SELECT * FROM attachments WHERE storage_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func ListByIssue(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Attachment, error) {
	var attachments []*Attachment
	err := db.SelectContext(ctx, &attachments,
		`SELECT * FROM attachments WHERE issue_id = $1 ORDER BY created_at DESC`, issueID)
	return attachments, err
}

func ListByComment(ctx context.Context, db *sqlx.DB, commentID uuid.UUID) ([]*Attachment, error) {
	var attachments []*Attachment
	err := db.SelectContext(ctx, &attachments,
		`SELECT * FROM attachments WHERE comment_id = $1 ORDER BY created_at DESC`, commentID)
	return attachments, err
}

func ListByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID, limit, offset int) ([]*Attachment, error) {
	var attachments []*Attachment
	err := db.SelectContext(ctx, &attachments,
		`SELECT * FROM attachments WHERE uploaded_by = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	return attachments, err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteByIssue(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]string, error) {
	// Return storage keys for cleanup
	var keys []string
	err := db.SelectContext(ctx, &keys,
		`DELETE FROM attachments WHERE issue_id = $1 RETURNING storage_key`, issueID)
	return keys, err
}

// CountByIssue returns total attachments and size for an issue
func CountByIssue(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) (count int64, totalSize int64, err error) {
	row := db.QueryRowxContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(size), 0) FROM attachments WHERE issue_id = $1`, issueID)
	err = row.Scan(&count, &totalSize)
	return
}

// CountByUser returns total attachments and size for a user
func CountByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (count int64, totalSize int64, err error) {
	row := db.QueryRowxContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(size), 0) FROM attachments WHERE uploaded_by = $1`, userID)
	err = row.Scan(&count, &totalSize)
	return
}

// Thumbnail operations

func CreateThumbnail(ctx context.Context, db *sqlx.DB, t *Thumbnail) error {
	query := `
		INSERT INTO attachment_thumbnails (attachment_id, storage_key, width, height)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query, t.AttachmentID, t.StorageKey, t.Width, t.Height).
		Scan(&t.ID, &t.CreatedAt)
}

func GetThumbnail(ctx context.Context, db *sqlx.DB, attachmentID uuid.UUID) (*Thumbnail, error) {
	t := &Thumbnail{}
	err := db.GetContext(ctx, t,
		`SELECT * FROM attachment_thumbnails WHERE attachment_id = $1`, attachmentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func DeleteThumbnail(ctx context.Context, db *sqlx.DB, attachmentID uuid.UUID) (*string, error) {
	var key string
	err := db.GetContext(ctx, &key,
		`DELETE FROM attachment_thumbnails WHERE attachment_id = $1 RETURNING storage_key`, attachmentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &key, err
}
