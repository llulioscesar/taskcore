# Taskcore Permissions System

**Version:** 2.0 (MVP with Groups & Permission Schemes)
**Inspired by:** Jira Groups, Project Roles & Permission Schemes
**Philosophy:** Professional-grade permissions, extensible for future

---

## Overview

Taskcore uses a **comprehensive permission system** inspired by Jira:

1. **Global Permissions:** System-wide admin flag
2. **Groups:** Global user groups (like `taskcore-developers`)
3. **Permission Schemes:** Reusable sets of permissions (assigned to projects)
4. **Project Roles:** Project-specific roles (Admin, Developer, Reporter)

**Key concepts:**
- **Groups** are **global** (user belongs to group across all projects)
- **Project Roles** are **per-project** (user has a role in a specific project)
- **Permission Schemes** define **what** users/groups/roles can do
- Projects reference a Permission Scheme (reusable)

---

## Global Permissions

### System Administrator (`is_admin` flag in `users` table)

**Who:** Typically the person who installed Taskcore

**Can do:**
- Create, edit, delete any project
- Manage all users (create, deactivate, reset passwords)
- Configure global settings (SMTP, OIDC providers)
- Access admin dashboard
- View system logs
- Assign global admin to other users

**Cannot do:**
- Nothing is restricted for global admins

**Implementation:**
```sql
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;
```

**Check in code:**
```go
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())
        if !user.IsAdmin {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Groups

### What are Groups?

Groups are **global collections of users** that can be:
- Assigned to project roles in bulk
- Granted permissions via permission schemes
- Used across multiple projects

**Default groups:**

| Group                      | Description                              | Typical use                           |
| -------------------------- | ---------------------------------------- | ------------------------------------- |
| `taskcore-administrators`  | Global admins                            | System administrators                 |
| `taskcore-developers`      | Developers across projects               | Engineering team                      |
| `taskcore-users`           | All users with login access              | Everyone with an account              |

### Groups vs Project Roles

| Feature                  | Groups                           | Project Roles                       |
| ------------------------ | -------------------------------- | ----------------------------------- |
| Scope                    | Global (across all projects)     | Per-project                         |
| Membership               | User belongs to group globally   | User has role in specific project   |
| Use case                 | Bulk assign to multiple projects | Fine-grained project access         |
| Example                  | "All developers"                 | "John is developer on Project A"    |

### Creating Groups

```sql
INSERT INTO groups (name, description) VALUES
('backend-team', 'Backend engineering team'),
('frontend-team', 'Frontend engineering team'),
('qa-team', 'Quality assurance team');
```

### Adding Users to Groups

```sql
INSERT INTO group_members (group_id, user_id) VALUES
(backend_group_id, alice_user_id),
(backend_group_id, bob_user_id);
```

### Assigning Groups to Projects

Instead of adding users one-by-one to a project, assign a group:

```sql
INSERT INTO project_members (project_id, group_id, role) VALUES
(project_id, backend_group_id, 'developer');
```

Now **all members of `backend-team`** are developers on this project.

---

## Permission Schemes

### What are Permission Schemes?

Permission Schemes are **reusable sets of permissions** that define **who can do what** in a project.

**Benefits:**
- **Reusable:** One scheme can be used by multiple projects
- **Flexible:** Grant permissions to users, groups, or roles
- **Scalable:** Change permissions in one place, affect many projects

### Default Permission Scheme

Every Taskcore installation has a "Default Permission Scheme" with sensible defaults:

| Permission              | Granted to         | Notes                                     |
| ----------------------- | ------------------ | ----------------------------------------- |
| `browse_project`        | Project role: All  | Everyone in project can view it           |
| `create_issue`          | Project role: All  | Everyone can create issues                |
| `edit_issue`            | Project role: Developer, Admin | Only developers can edit |
| `delete_issue`          | Project role: Admin | Only admins can delete                   |
| `administer_project`    | Project role: Admin | Only admins can configure                |
| `configure_workflow`    | Project role: Admin | Only admins can edit workflows           |
| `manage_sprints`        | Project role: Admin | Only admins manage sprints               |
| `add_comment`           | Project role: All  | Everyone can comment                      |

### Creating Custom Permission Schemes

```sql
-- Create a new scheme
INSERT INTO permission_schemes (name, description) VALUES
('Open Source Project Scheme', 'Public project with restricted admin');

-- Grant permissions
INSERT INTO permission_scheme_permissions
(permission_scheme_id, permission, grantee_type, project_role) VALUES
-- Everyone can view
(scheme_id, 'browse_project', 'anyone', NULL),
-- Only developers can create issues
(scheme_id, 'create_issue', 'project_role', 'developer'),
-- Only admins can delete
(scheme_id, 'delete_issue', 'project_role', 'admin');
```

### Grantee Types

Permissions can be granted to 4 types of entities:

| Grantee Type    | Description                          | Example                                |
| --------------- | ------------------------------------ | -------------------------------------- |
| `user`          | Specific user                        | Grant "delete_issue" to Alice          |
| `group`         | All users in a group                 | Grant "create_issue" to `qa-team`      |
| `project_role`  | All users with a project role        | Grant "edit_issue" to `developer` role |
| `anyone`        | Public (even non-logged-in users)    | Grant "browse_project" to `anyone`     |

**Examples:**

```sql
-- Grant to specific user
INSERT INTO permission_scheme_permissions
(permission_scheme_id, permission, grantee_type, grantee_id) VALUES
(scheme_id, 'administer_project', 'user', alice_user_id);

-- Grant to group
INSERT INTO permission_scheme_permissions
(permission_scheme_id, permission, grantee_type, grantee_id) VALUES
(scheme_id, 'create_issue', 'group', qa_team_group_id);

-- Grant to project role
INSERT INTO permission_scheme_permissions
(permission_scheme_id, permission, grantee_type, project_role) VALUES
(scheme_id, 'edit_issue', 'project_role', 'developer');

-- Grant to anyone (public)
INSERT INTO permission_scheme_permissions
(permission_scheme_id, permission, grantee_type) VALUES
(scheme_id, 'browse_project', 'anyone');
```

### Assigning Permission Scheme to Project

```sql
UPDATE projects
SET permission_scheme_id = custom_scheme_id
WHERE id = project_id;
```

---

## Project Permissions

### Project Roles

Every user in a project has **exactly one role** (stored in `project_members.role`):

| Role             | Description                                  |
| ---------------- | -------------------------------------------- |
| `admin`          | Project administrator                        |
| `developer`      | Team member who works on issues              |
| `reporter`       | Can create and comment on issues, read-only  |

---

### Role: Project Admin

**Typical use case:** Project manager, tech lead

**Project Permissions:**

| Permission                  | Allowed |
| --------------------------- | :-----: |
| View project                |    ✔    |
| Edit project settings       |    ✔    |
| Archive project             |    ✔    |
| Manage project members      |    ✔    |
| Configure workflows         |    ✔    |
| Create/edit/delete sprints  |    ✔    |
| Create issues               |    ✔    |
| Edit any issue              |    ✔    |
| Delete any issue            |    ✔    |
| Assign issues               |    ✔    |
| Transition issues           |    ✔    |
| Comment on issues           |    ✔    |
| Edit/delete any comment     |    ✔    |
| Create/edit/delete labels   |    ✔    |
| Manage issue relations      |    ✔    |

**Cannot do (requires Global Admin):**
- Delete project
- Transfer project ownership

---

### Role: Developer

**Typical use case:** Software engineer, designer, QA tester

**Project Permissions:**

| Permission                     | Allowed |
| ------------------------------ | :-----: |
| View project                   |    ✔    |
| Edit project settings          |    ✖    |
| Manage project members         |    ✖    |
| Configure workflows            |    ✖    |
| View sprints                   |    ✔    |
| Create/edit/delete sprints     |    ✖    |
| Create issues                  |    ✔    |
| Edit own issues                |    ✔    |
| Edit issues assigned to them   |    ✔    |
| Delete own issues              |    ✔    |
| Assign issues                  |    ✔    |
| Transition issues              |    ✔    |
| Comment on issues              |    ✔    |
| Edit/delete own comments       |    ✔    |
| Create labels                  |    ✔    |
| Manage issue relations         |    ✔    |

---

### Role: Reporter

**Typical use case:** External stakeholder, customer, support agent

**Project Permissions:**

| Permission                  | Allowed |
| --------------------------- | :-----: |
| View project                |    ✔    |
| Edit project settings       |    ✖    |
| Manage project members      |    ✖    |
| Configure workflows         |    ✖    |
| View sprints                |    ✔    |
| Create issues               |    ✔    |
| Edit own issues (unassigned)|    ✔    |
| Delete own issues           |    ✖    |
| Assign issues               |    ✖    |
| Transition issues           |    ✖    |
| Comment on issues           |    ✔    |
| Edit/delete own comments    |    ✔    |
| Create labels               |    ✖    |
| Manage issue relations      |    ✖    |

**Notes:**
- Reporters can only edit issues they created **before** they are assigned
- Once assigned, only developers/admins can edit

---

## Permission Matrix (Complete)

| Action                          | Reporter | Developer | Project Admin | Global Admin |
| ------------------------------- | :------: | :-------: | :-----------: | :----------: |
| **Projects**                    |          |           |               |              |
| View project                    |    ✔     |     ✔     |       ✔       |      ✔       |
| Create project                  |    ✖     |     ✖     |       ✖       |      ✔       |
| Edit project settings           |    ✖     |     ✖     |       ✔       |      ✔       |
| Archive project                 |    ✖     |     ✖     |       ✔       |      ✔       |
| Delete project                  |    ✖     |     ✖     |       ✖       |      ✔       |
| **Members**                     |          |           |               |              |
| View members                    |    ✔     |     ✔     |       ✔       |      ✔       |
| Add members                     |    ✖     |     ✖     |       ✔       |      ✔       |
| Remove members                  |    ✖     |     ✖     |       ✔       |      ✔       |
| Change member role              |    ✖     |     ✖     |       ✔       |      ✔       |
| **Issues**                      |          |           |               |              |
| View issues                     |    ✔     |     ✔     |       ✔       |      ✔       |
| Create issues                   |    ✔     |     ✔     |       ✔       |      ✔       |
| Edit own issues                 |    ✔*    |     ✔     |       ✔       |      ✔       |
| Edit any issue                  |    ✖     |     ✔**   |       ✔       |      ✔       |
| Delete own issues               |    ✖     |     ✔     |       ✔       |      ✔       |
| Delete any issue                |    ✖     |     ✖     |       ✔       |      ✔       |
| Assign issues                   |    ✖     |     ✔     |       ✔       |      ✔       |
| Transition issues               |    ✖     |     ✔***  |       ✔       |      ✔       |
| **Comments**                    |          |           |               |              |
| Add comments                    |    ✔     |     ✔     |       ✔       |      ✔       |
| Edit own comments               |    ✔     |     ✔     |       ✔       |      ✔       |
| Delete own comments             |    ✔     |     ✔     |       ✔       |      ✔       |
| Edit any comment                |    ✖     |     ✖     |       ✔       |      ✔       |
| Delete any comment              |    ✖     |     ✖     |       ✔       |      ✔       |
| **Sprints (Scrum projects)**    |          |           |               |              |
| View sprints                    |    ✔     |     ✔     |       ✔       |      ✔       |
| Create sprints                  |    ✖     |     ✖     |       ✔       |      ✔       |
| Start/close sprints             |    ✖     |     ✖     |       ✔       |      ✔       |
| Add issues to sprint            |    ✖     |     ✔     |       ✔       |      ✔       |
| **Workflows**                   |          |           |               |              |
| View workflow                   |    ✔     |     ✔     |       ✔       |      ✔       |
| Configure workflow              |    ✖     |     ✖     |       ✔       |      ✔       |
| **Labels**                      |          |           |               |              |
| View labels                     |    ✔     |     ✔     |       ✔       |      ✔       |
| Create labels                   |    ✖     |     ✔     |       ✔       |      ✔       |
| Edit/delete labels              |    ✖     |     ✖     |       ✔       |      ✔       |

**Notes:**
- `*` Reporters can only edit own issues if unassigned
- `**` Developers can edit issues assigned to them or created by them
- `***` Transitions may be restricted by workflow rules (see below)

---

## Workflow Transition Permissions

Workflows can optionally restrict transitions to specific roles via `workflow_transitions.allowed_role`.

**Examples:**

```sql
-- Anyone with "Transition Issues" permission can move from "To Do" to "In Progress"
INSERT INTO workflow_transitions (workflow_id, from_status_id, to_status_id, allowed_role)
VALUES (workflow_id, todo_status_id, in_progress_status_id, NULL);

-- Only admins can move to "Done"
INSERT INTO workflow_transitions (workflow_id, from_status_id, to_status_id, allowed_role)
VALUES (workflow_id, in_review_status_id, done_status_id, 'admin');
```

**Permission check logic:**
```go
func CanTransition(user User, issue Issue, toStatus Status) bool {
    role := GetUserProjectRole(user.ID, issue.ProjectID)

    // Project admins can always transition
    if role == "admin" {
        return true
    }

    // Check workflow transition rules
    transition := GetTransition(issue.Workflow, issue.Status, toStatus)
    if transition == nil {
        return false // No valid transition exists
    }

    if transition.AllowedRole != nil && role != *transition.AllowedRole {
        return false // User's role doesn't match required role
    }

    return true
}
```

---

## Implementation Guide

### 1. Permission Service (Core)

```go
// internal/service/permission/service.go
type Service struct {
    projectMemberRepo     repository.ProjectMemberRepository
    permissionSchemeRepo  repository.PermissionSchemeRepository
    groupMemberRepo       repository.GroupMemberRepository
}

// HasPermission checks if a user has a specific permission in a project
func (s *Service) HasPermission(ctx context.Context, userID, projectID uuid.UUID, perm string) (bool, error) {
    // 1. Global admins bypass all checks
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return false, err
    }
    if user.IsAdmin {
        return true, nil
    }

    // 2. Get project's permission scheme
    project, err := s.projectRepo.GetByID(ctx, projectID)
    if err != nil {
        return false, err
    }

    // 3. Get all permission grants for this permission in the scheme
    grants, err := s.permissionSchemeRepo.GetPermissionGrants(ctx, project.PermissionSchemeID, perm)
    if err != nil {
        return false, err
    }

    // 4. Check each grant
    for _, grant := range grants {
        switch grant.GranteeType {
        case "anyone":
            return true, nil

        case "user":
            if grant.GranteeID == userID {
                return true, nil
            }

        case "group":
            // Check if user is in this group
            isMember, err := s.groupMemberRepo.IsUserInGroup(ctx, userID, grant.GranteeID)
            if err != nil {
                return false, err
            }
            if isMember {
                return true, nil
            }

        case "project_role":
            // Check if user has this project role (directly or via group)
            hasRole, err := s.UserHasProjectRole(ctx, userID, projectID, grant.ProjectRole)
            if err != nil {
                return false, err
            }
            if hasRole {
                return true, nil
            }
        }
    }

    return false, nil
}

// UserHasProjectRole checks if a user has a specific role in a project
// (either directly or via group membership)
func (s *Service) UserHasProjectRole(ctx context.Context, userID, projectID uuid.UUID, role string) (bool, error) {
    // Check direct user assignment
    directMember, err := s.projectMemberRepo.GetByUserAndProject(ctx, userID, projectID)
    if err == nil && directMember.Role == role {
        return true, nil
    }

    // Check group assignments
    userGroups, err := s.groupMemberRepo.GetUserGroups(ctx, userID)
    if err != nil {
        return false, err
    }

    for _, group := range userGroups {
        groupMember, err := s.projectMemberRepo.GetByGroupAndProject(ctx, group.ID, projectID)
        if err == nil && groupMember.Role == role {
            return true, nil
        }
    }

    return false, nil
}
```

### 2. Using Permissions in Handlers

```go
// internal/handler/issue.go
func (h *IssueHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
    user := GetUserFromContext(r.Context())
    issueID := chi.URLParam(r, "issueID")

    issue, err := h.issueService.GetByID(r.Context(), issueID)
    if err != nil {
        http.Error(w, "Issue not found", http.StatusNotFound)
        return
    }

    // Check permission via permission scheme
    canEdit, err := h.permService.HasPermission(r.Context(), user.ID, issue.ProjectID, "edit_issue")
    if err != nil || !canEdit {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Additional business logic checks
    // (e.g., reporters can only edit their own unassigned issues)
    if !h.canEditIssueBusinessLogic(user, issue) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // ... proceed with update
}

func (h *IssueHandler) canEditIssueBusinessLogic(user *models.User, issue *models.Issue) bool {
    // This handles role-specific constraints not in permission scheme
    role, _ := h.permService.GetUserProjectRole(user.ID, issue.ProjectID)

    switch role {
    case "admin", "developer":
        return true
    case "reporter":
        // Reporters can only edit their own issues if unassigned
        return issue.ReporterID == user.ID && issue.AssigneeID == nil
    default:
        return false
    }
}
```

### 3. Middleware for Permission Checking

```go
// internal/middleware/permission.go
func RequireProjectPermission(permService *permission.Service, perm string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := GetUserFromContext(r.Context())
            projectID := chi.URLParam(r, "projectID")

            hasPermission, err := permService.HasPermission(r.Context(), user.ID, uuid.Must Parse(projectID), perm)
            if err != nil || !hasPermission {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
r.Route("/projects/{projectID}", func(r chi.Router) {
    r.Use(middleware.RequireProjectPermission(permService, "browse_project"))

    r.Get("/", handler.GetProject)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireProjectPermission(permService, "administer_project"))
        r.Put("/settings", handler.UpdateProjectSettings)
    })
})
```

---

## Future Enhancements (Post-MVP)

These are **not** in the MVP but can be added later:

### Issue Security Levels
- Restrict issue visibility to specific roles/users
- Useful for sensitive bugs

### Custom Roles
- Beyond the 3 default roles
- Project-specific role definitions

### Field-Level Permissions
- Control who can edit specific fields (e.g., only admins edit "Story Points")

### Notification Permissions
- Control who receives notifications for what events

---

## Comparison with Jira

| Feature                       | Jira                           | Taskcore MVP                    |
| ----------------------------- | ------------------------------ | ------------------------------- |
| Global Permissions            | ✔ (JIRA Admins, JIRA Users)    | ✔ (is_admin flag)               |
| Groups                        | ✔ (Global groups)              | ✔ (Global groups)               |
| Project Roles                 | ✔ (Customizable)               | ✔ (3 fixed roles)               |
| Permission Schemes            | ✔ (Complex, reusable)          | ✔ (Simplified, reusable)        |
| Grantee Types                 | ✔ (User, Group, Role, Anyone)  | ✔ (User, Group, Role, Anyone)   |
| Workflow Transition Rules     | ✔ (Per transition)             | ✔ (Role-based)                  |
| Issue Security Levels         | ✔                              | ✖ (Post-MVP)                    |
| Custom Roles                  | ✔                              | ✖ (Post-MVP)                    |
| Field-Level Permissions       | ✔ (Via schemes)                | ✖ (Post-MVP)                    |

---

## Summary

Taskcore MVP uses a **professional, flexible permission model** inspired by Jira:

- **Global Admin:** `is_admin` flag for system administrators
- **Groups:** Global user collections (like `taskcore-developers`)
- **Permission Schemes:** Reusable permission sets (assigned to projects)
- **4 Grantee Types:** User, Group, Project Role, Anyone
- **3 Project Roles:** Admin, Developer, Reporter
- **Workflow-based** transition controls

**Key advantages:**
- **Flexible:** Grant permissions to users, groups, or roles
- **Reusable:** One permission scheme for multiple projects
- **Scalable:** Bulk-assign users via groups
- **Professional:** Matches Jira's proven security model

This provides **90% of Jira's permission features** with a clean, maintainable implementation.
