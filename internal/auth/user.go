package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserInfo contains the user data needed for authentication
type UserInfo struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	Name         string    `db:"name"`
	Avatar       *string   `db:"avatar"`
	IsActive     bool      `db:"is_active"`
	IsAdmin      bool      `db:"is_admin"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// ErrUserNotFound is returned when user is not found
var ErrUserNotFound = errors.New("user not found")

func getUserByEmail(ctx context.Context, db *sqlx.DB, email string) (*UserInfo, error) {
	u := &UserInfo{}
	err := db.GetContext(ctx, u, `SELECT id, email, name, avatar, is_active, is_admin, password_hash, created_at FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func getUserByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*UserInfo, error) {
	u := &UserInfo{}
	err := db.GetContext(ctx, u, `SELECT id, email, name, avatar, is_active, is_admin, password_hash, created_at FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}
