package notification

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, n *Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, message, entity_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_read, created_at`
	return db.QueryRowxContext(ctx, query,
		n.UserID, n.Type, n.Title, n.Message, n.EntityID,
	).Scan(&n.ID, &n.IsRead, &n.CreatedAt)
}

func CreateBulk(ctx context.Context, db *sqlx.DB, userIDs []uuid.UUID, notifType Type, title, message string, entityID *uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := `
		INSERT INTO notifications (user_id, type, title, message, entity_id)
		SELECT unnest($1::uuid[]), $2, $3, $4, $5`
	_, err := db.ExecContext(ctx, query, userIDs, notifType, title, message, entityID)
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Notification, error) {
	n := &Notification{}
	err := db.GetContext(ctx, n, `SELECT * FROM notifications WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return n, err
}

func ListByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID, limit, offset int) ([]*Notification, error) {
	var notifications []*Notification
	query := `
		SELECT * FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	err := db.SelectContext(ctx, &notifications, query, userID, limit, offset)
	return notifications, err
}

func ListUnreadByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Notification, error) {
	var notifications []*Notification
	query := `SELECT * FROM notifications WHERE user_id = $1 AND is_read = false ORDER BY created_at DESC`
	err := db.SelectContext(ctx, &notifications, query, userID)
	return notifications, err
}

func CountUnread(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (int, error) {
	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID)
	return count, err
}

func MarkAsRead(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`UPDATE notifications SET is_read = true WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func MarkAllAsRead(ctx context.Context, db *sqlx.DB, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`, userID)
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteOlderThan(ctx context.Context, db *sqlx.DB, days int) (int64, error) {
	result, err := db.ExecContext(ctx,
		`DELETE FROM notifications WHERE created_at < NOW() - INTERVAL '1 day' * $1`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetPreference(ctx context.Context, db *sqlx.DB, userID uuid.UUID, notifType Type) (*Preference, error) {
	p := &Preference{}
	err := db.GetContext(ctx, p,
		`SELECT * FROM notification_preferences WHERE user_id = $1 AND type = $2`, userID, notifType)
	if errors.Is(err, sql.ErrNoRows) {
		return &Preference{UserID: userID, Type: notifType, Email: true, InApp: true}, nil
	}
	return p, err
}

func ListPreferences(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Preference, error) {
	var prefs []*Preference
	err := db.SelectContext(ctx, &prefs,
		`SELECT * FROM notification_preferences WHERE user_id = $1`, userID)
	return prefs, err
}

func SetPreference(ctx context.Context, db *sqlx.DB, p *Preference) error {
	query := `
		INSERT INTO notification_preferences (user_id, type, email, in_app)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, type) DO UPDATE SET email = EXCLUDED.email, in_app = EXCLUDED.in_app`
	_, err := db.ExecContext(ctx, query, p.UserID, p.Type, p.Email, p.InApp)
	return err
}
