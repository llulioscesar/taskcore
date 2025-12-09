package activity

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionCreated         Action = "created"
	ActionUpdated         Action = "updated"
	ActionDeleted         Action = "deleted"
	ActionStatusChanged   Action = "status_changed"
	ActionAssigned        Action = "assigned"
	ActionCommentAdded    Action = "comment_added"
	ActionCommentUpdated  Action = "comment_updated"
	ActionCommentDeleted  Action = "comment_deleted"
	ActionAttachmentAdded Action = "attachment_added"
	ActionWorkLogged      Action = "work_logged"
	ActionLabelAdded      Action = "label_added"
	ActionLabelRemoved    Action = "label_removed"
	ActionSprintChanged   Action = "sprint_changed"
)

type EntityType string

const (
	EntityIssue   EntityType = "issue"
	EntityProject EntityType = "project"
	EntitySprint  EntityType = "sprint"
)

type Log struct {
	ID         uuid.UUID  `db:"id"`
	EntityType EntityType `db:"entity_type"`
	EntityID   uuid.UUID  `db:"entity_id"`
	UserID     uuid.UUID  `db:"user_id"`
	Action     Action     `db:"action"`
	OldValue   *string    `db:"old_value"`
	NewValue   *string    `db:"new_value"`
	CreatedAt  time.Time  `db:"created_at"`
}
