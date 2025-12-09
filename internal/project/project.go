package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("project not found")
	ErrKeyExists = errors.New("project key already exists")
)

type Project struct {
	ID                 uuid.UUID  `db:"id"`
	TemplateID         *uuid.UUID `db:"template_id"`
	Key                string     `db:"key"`
	Name               string     `db:"name"`
	Description        *string    `db:"description"`
	LeadID             uuid.UUID  `db:"lead_id"`
	DefaultAssigneeID  *uuid.UUID `db:"default_assignee_id"`
	PermissionSchemeID uuid.UUID  `db:"permission_scheme_id"`
	WorkflowID         uuid.UUID  `db:"workflow_id"`
	IssueCounter       int        `db:"issue_counter"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleReporter  Role = "reporter"
)

type Member struct {
	ID        uuid.UUID  `db:"id"`
	ProjectID uuid.UUID  `db:"project_id"`
	UserID    *uuid.UUID `db:"user_id"`
	GroupID   *uuid.UUID `db:"group_id"`
	Role      Role       `db:"role"`
	CreatedAt time.Time  `db:"created_at"`
}
