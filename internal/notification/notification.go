package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("notification not found")

type Type string

const (
	TypeAssigned        Type = "assigned"
	TypeMentioned       Type = "mentioned"
	TypeWatching        Type = "watching"
	TypeCommentAdded    Type = "comment_added"
	TypeStatusChanged   Type = "status_changed"
	TypeIssueUpdated    Type = "issue_updated"
	TypeSprintStarted   Type = "sprint_started"
	TypeSprintCompleted Type = "sprint_completed"
)

type Notification struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Type      Type       `db:"type"`
	Title     string     `db:"title"`
	Message   string     `db:"message"`
	EntityID  *uuid.UUID `db:"entity_id"`
	IsRead    bool       `db:"is_read"`
	CreatedAt time.Time  `db:"created_at"`
}

type Preference struct {
	UserID   uuid.UUID `db:"user_id"`
	Type     Type      `db:"type"`
	Email    bool      `db:"email"`
	InApp    bool      `db:"in_app"`
}
