package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, u *User) error {
	query := `
		INSERT INTO users (email, password_hash, name, avatar, is_active, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query,
		u.Email, u.PasswordHash, u.Name, u.Avatar, u.IsActive, u.IsAdmin,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil && isUniqueViolation(err) {
		return ErrEmailExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*User, error) {
	u := &User{}
	err := db.GetContext(ctx, u, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func GetByEmail(ctx context.Context, db *sqlx.DB, email string) (*User, error) {
	u := &User{}
	err := db.GetContext(ctx, u, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*User, error) {
	var users []*User
	err := db.SelectContext(ctx, &users, `SELECT * FROM users WHERE is_active = true ORDER BY name`)
	return users, err
}

func Update(ctx context.Context, db *sqlx.DB, u *User) error {
	query := `
		UPDATE users SET email = $2, name = $3, avatar = $4, is_active = $5, is_admin = $6
		WHERE id = $1 RETURNING updated_at`

	err := db.QueryRowxContext(ctx, query,
		u.ID, u.Email, u.Name, u.Avatar, u.IsActive, u.IsAdmin,
	).Scan(&u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func UpdatePassword(ctx context.Context, db *sqlx.DB, id uuid.UUID, passwordHash string) error {
	result, err := db.ExecContext(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`UPDATE users SET is_active = false WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
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

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
