package permission

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSchemeNotFound   = errors.New("permission scheme not found")
	ErrSchemeNameExists = errors.New("permission scheme name already exists")
)

type Permission string

const (
	PermBrowseProjects       Permission = "browse_projects"
	PermAdministerProjects   Permission = "administer_projects"
	PermManageSprintsBacklog Permission = "manage_sprints_backlog"
	PermCreateIssues         Permission = "create_issues"
	PermEditIssues           Permission = "edit_issues"
	PermDeleteIssues         Permission = "delete_issues"
	PermAssignIssues         Permission = "assign_issues"
	PermAssignableUser       Permission = "assignable_user"
	PermCloseIssues          Permission = "close_issues"
	PermResolveIssues        Permission = "resolve_issues"
	PermTransitionIssues     Permission = "transition_issues"
	PermScheduleIssues       Permission = "schedule_issues"
	PermModifyReporter       Permission = "modify_reporter"
	PermLinkIssues           Permission = "link_issues"
	PermMoveIssues           Permission = "move_issues"
	PermSetIssueSecurity     Permission = "set_issue_security"
	PermAddComments          Permission = "add_comments"
	PermEditAllComments      Permission = "edit_all_comments"
	PermEditOwnComments      Permission = "edit_own_comments"
	PermDeleteAllComments    Permission = "delete_all_comments"
	PermDeleteOwnComments    Permission = "delete_own_comments"
	PermManageWatchers       Permission = "manage_watchers"
	PermViewWatchers         Permission = "view_watchers"
	PermViewVoters           Permission = "view_voters"
	PermWorkOnIssues         Permission = "work_on_issues"
	PermEditOwnWorkLogs      Permission = "edit_own_work_logs"
	PermEditAllWorkLogs      Permission = "edit_all_work_logs"
	PermDeleteOwnWorkLogs    Permission = "delete_own_work_logs"
	PermDeleteAllWorkLogs    Permission = "delete_all_work_logs"
)

type GranteeType string

const (
	GranteeUser        GranteeType = "user"
	GranteeGroup       GranteeType = "group"
	GranteeProjectRole GranteeType = "project_role"
	GranteeAnyone      GranteeType = "anyone"
)

type Scheme struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type SchemePermission struct {
	ID          uuid.UUID   `db:"id"`
	SchemeID    uuid.UUID   `db:"scheme_id"`
	Permission  Permission  `db:"permission"`
	GranteeType GranteeType `db:"grantee_type"`
	GranteeID   *uuid.UUID  `db:"grantee_id"`
}

func AllPermissions() []Permission {
	return []Permission{
		PermBrowseProjects, PermAdministerProjects, PermManageSprintsBacklog,
		PermCreateIssues, PermEditIssues, PermDeleteIssues, PermAssignIssues,
		PermAssignableUser, PermCloseIssues, PermResolveIssues, PermTransitionIssues,
		PermScheduleIssues, PermModifyReporter, PermLinkIssues, PermMoveIssues,
		PermSetIssueSecurity, PermAddComments, PermEditAllComments, PermEditOwnComments,
		PermDeleteAllComments, PermDeleteOwnComments, PermManageWatchers, PermViewWatchers,
		PermViewVoters, PermWorkOnIssues, PermEditOwnWorkLogs, PermEditAllWorkLogs,
		PermDeleteOwnWorkLogs, PermDeleteAllWorkLogs,
	}
}
