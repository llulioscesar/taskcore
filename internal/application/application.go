package application

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("application not found")
	ErrNoAccess = errors.New("user does not have application access")
)

// Application represents a module/feature set that can be enabled/disabled
type Application struct {
	ID          uuid.UUID `db:"id"`
	Key         string    `db:"key"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
}

// Default applications (modules)
var (
	AppCore        = "core"        // Base issue tracking
	AppAgile       = "agile"       // Scrum, Kanban boards, sprints
	AppServiceDesk = "servicedesk" // IT service management (future)
)

// UserAccess grants a user access to an application/module
type UserAccess struct {
	ID            uuid.UUID  `db:"id"`
	UserID        uuid.UUID  `db:"user_id"`
	ApplicationID uuid.UUID  `db:"application_id"`
	GrantedBy     *uuid.UUID `db:"granted_by"`
	CreatedAt     time.Time  `db:"created_at"`
}

// GroupAccess grants all members of a group access to an application
type GroupAccess struct {
	ID            uuid.UUID  `db:"id"`
	GroupID       uuid.UUID  `db:"group_id"`
	ApplicationID uuid.UUID  `db:"application_id"`
	GrantedBy     *uuid.UUID `db:"granted_by"`
	CreatedAt     time.Time  `db:"created_at"`
}
