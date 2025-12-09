package audit

import (
	"time"

	"github.com/google/uuid"
)

// ActionType represents the type of audited action
type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
	ActionLogin  ActionType = "login"
	ActionLogout ActionType = "logout"
	ActionExport ActionType = "export"
	ActionImport ActionType = "import"
)

// ResourceType represents the type of resource being audited
type ResourceType string

const (
	ResourceUser        ResourceType = "user"
	ResourceGroup       ResourceType = "group"
	ResourceProject     ResourceType = "project"
	ResourceIssue       ResourceType = "issue"
	ResourceComment     ResourceType = "comment"
	ResourceAttachment  ResourceType = "attachment"
	ResourceWorkflow    ResourceType = "workflow"
	ResourcePermission  ResourceType = "permission"
	ResourceRole        ResourceType = "role"
	ResourceSprint      ResourceType = "sprint"
	ResourceBoard       ResourceType = "board"
	ResourceFilter      ResourceType = "filter"
	ResourceApplication ResourceType = "application"
	ResourceSystem      ResourceType = "system"
)

// Entry represents an audit log entry
type Entry struct {
	ID           uuid.UUID    `db:"id"`
	UserID       *uuid.UUID   `db:"user_id"`       // Who performed the action (null for system)
	Action       ActionType   `db:"action"`        // What action was performed
	ResourceType ResourceType `db:"resource_type"` // Type of resource affected
	ResourceID   *uuid.UUID   `db:"resource_id"`   // ID of affected resource (if applicable)
	ResourceName *string      `db:"resource_name"` // Human-readable name (for deleted resources)
	ProjectID    *uuid.UUID   `db:"project_id"`    // Project context (if applicable)
	OldValue     *string      `db:"old_value"`     // JSON of previous state
	NewValue     *string      `db:"new_value"`     // JSON of new state
	IPAddress    *string      `db:"ip_address"`    // Client IP
	UserAgent    *string      `db:"user_agent"`    // Client user agent
	Details      *string      `db:"details"`       // Additional JSON details
	CreatedAt    time.Time    `db:"created_at"`
}

// SearchParams for querying audit logs
type SearchParams struct {
	UserID       *uuid.UUID
	Action       *ActionType
	ResourceType *ResourceType
	ResourceID   *uuid.UUID
	ProjectID    *uuid.UUID
	FromDate     *time.Time
	ToDate       *time.Time
	IPAddress    *string
	Limit        int
	Offset       int
}
