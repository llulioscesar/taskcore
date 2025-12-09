package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Name         string    `db:"name"`
	Avatar       *string   `db:"avatar"`
	IsActive     bool      `db:"is_active"`
	IsAdmin      bool      `db:"is_admin"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
