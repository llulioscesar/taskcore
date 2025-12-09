package role

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("role not found")
	ErrNameExists     = errors.New("role name already exists")
	ErrActorNotFound  = errors.New("role actor not found")
	ErrGlobalNotFound = errors.New("global role not found")
)

// Project Role - configurable roles per project (like Jira)
type Role struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Default project roles
var (
	RoleAdministrators = "Administrators"
	RoleDevelopers     = "Developers"
	RoleViewers        = "Viewers"
)

// ActorType defines what type of actor is assigned to a role
type ActorType string

const (
	ActorUser  ActorType = "user"
	ActorGroup ActorType = "group"
)

// RoleActor - users/groups assigned to a project role
type RoleActor struct {
	ID        uuid.UUID `db:"id"`
	ProjectID uuid.UUID `db:"project_id"`
	RoleID    uuid.UUID `db:"role_id"`
	ActorType ActorType `db:"actor_type"`
	ActorID   uuid.UUID `db:"actor_id"` // user_id or group_id
	CreatedAt time.Time `db:"created_at"`
}

// DefaultRoleActor - default actors for a role (applied to new projects)
type DefaultRoleActor struct {
	ID        uuid.UUID `db:"id"`
	RoleID    uuid.UUID `db:"role_id"`
	ActorType ActorType `db:"actor_type"`
	ActorID   uuid.UUID `db:"actor_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Global Role - system-wide roles (like Jira system admin, etc)
type GlobalRoleType string

const (
	GlobalRoleSystemAdmin GlobalRoleType = "system_admin"
	GlobalRoleAdmin       GlobalRoleType = "admin"
	GlobalRoleUser        GlobalRoleType = "user"
)

type GlobalRole struct {
	ID          uuid.UUID      `db:"id"`
	Name        GlobalRoleType `db:"name"`
	Description *string        `db:"description"`
}

// GlobalRoleMember - users assigned to global roles
type GlobalRoleMember struct {
	GlobalRoleID uuid.UUID `db:"global_role_id"`
	UserID       uuid.UUID `db:"user_id"`
	CreatedAt    time.Time `db:"created_at"`
}

// Global Permissions - what global roles can do
type GlobalPermission string

const (
	GlobalPermAdministerSystem GlobalPermission = "administer_system"
	GlobalPermAdministerJira   GlobalPermission = "administer_jira"
	GlobalPermManageUsers      GlobalPermission = "manage_users"
	GlobalPermManageGroups     GlobalPermission = "manage_groups"
	GlobalPermCreateProjects   GlobalPermission = "create_projects"
	GlobalPermBrowseUsers      GlobalPermission = "browse_users"
	GlobalPermShareDashboards  GlobalPermission = "share_dashboards"
	GlobalPermManageFilters    GlobalPermission = "manage_filters"
	GlobalPermBulkChange       GlobalPermission = "bulk_change"
)

type GlobalPermissionGrant struct {
	ID           uuid.UUID        `db:"id"`
	Permission   GlobalPermission `db:"permission"`
	GlobalRoleID *uuid.UUID       `db:"global_role_id"` // NULL = anyone logged in
	GroupID      *uuid.UUID       `db:"group_id"`
}

func AllGlobalPermissions() []GlobalPermission {
	return []GlobalPermission{
		GlobalPermAdministerSystem,
		GlobalPermAdministerJira,
		GlobalPermManageUsers,
		GlobalPermManageGroups,
		GlobalPermCreateProjects,
		GlobalPermBrowseUsers,
		GlobalPermShareDashboards,
		GlobalPermManageFilters,
		GlobalPermBulkChange,
	}
}
