package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const SessionDuration = 24 * time.Hour * 7 // 7 days

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type Session struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func CreateSession(ctx context.Context, db *sqlx.DB, userID uuid.UUID, duration time.Duration) (*Session, error) {
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}

	session := &Session{
		UserID:       userID,
		RefreshToken: token,
		ExpiresAt:    time.Now().Add(duration),
	}

	query := `
		INSERT INTO sessions (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err = db.QueryRowxContext(ctx, query,
		session.UserID, session.RefreshToken, session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)

	return session, err
}

func GetSessionByToken(ctx context.Context, db *sqlx.DB, token string) (*Session, error) {
	session := &Session{}
	err := db.GetContext(ctx, session, `SELECT * FROM sessions WHERE refresh_token = $1`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	if session.IsExpired() {
		return nil, ErrSessionExpired
	}
	return session, nil
}

func DeleteSession(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func DeleteUserSessions(ctx context.Context, db *sqlx.DB, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func DeleteExpiredSessions(ctx context.Context, db *sqlx.DB) (int64, error) {
	result, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
