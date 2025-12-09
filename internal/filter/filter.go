package filter

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("filter not found")
	ErrNameExists = errors.New("filter name already exists for user")
)

type Filter struct {
	ID          uuid.UUID  `db:"id"`
	OwnerID     uuid.UUID  `db:"owner_id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	JQL         string     `db:"jql"`
	IsPublic    bool       `db:"is_public"`
	IsFavorite  bool       `db:"is_favorite"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Share struct {
	FilterID  uuid.UUID  `db:"filter_id"`
	UserID    *uuid.UUID `db:"user_id"`
	GroupID   *uuid.UUID `db:"group_id"`
	ProjectID *uuid.UUID `db:"project_id"`
	CreatedAt time.Time  `db:"created_at"`
}

type Subscription struct {
	ID        uuid.UUID `db:"id"`
	FilterID  uuid.UUID `db:"filter_id"`
	UserID    uuid.UUID `db:"user_id"`
	Schedule  string    `db:"schedule"`
	CreatedAt time.Time `db:"created_at"`
}
