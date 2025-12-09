package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/store"
)

func create(ctx context.Context, db *sqlx.DB, u *User) error {
	query := `
		INSERT INTO users (email, password_hash, name, avatar, is_active, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query,
		u.Email, u.PasswordHash, u.Name, u.Avatar, u.IsActive, u.IsAdmin,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil && store.IsUniqueViolation(err) {
		return ErrEmailExists
	}
	return err
}

func getByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*User, error) {
	u := &User{}
	err := db.GetContext(ctx, u, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func getByEmail(ctx context.Context, db *sqlx.DB, email string) (*User, error) {
	u := &User{}
	err := db.GetContext(ctx, u, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func list(ctx context.Context, db *sqlx.DB) ([]*User, error) {
	var users []*User
	err := db.SelectContext(ctx, &users, `SELECT * FROM users WHERE is_active = true ORDER BY name`)
	return users, err
}

func update(ctx context.Context, db *sqlx.DB, u *User) error {
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

func delete_(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
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
