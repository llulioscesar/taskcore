# Taskcore Database Schema

**Version:** 1.0
**Database:** PostgreSQL 15+
**Migration Tool:** golang-migrate

---

## Schema Overview

```
Core Authentication:
├── users (with is_admin flag for global admins)
├── sessions (JWT refresh tokens)
└── oidc_providers (optional P2)

Groups & Global Permissions:
├── groups (global groups like jira-developers, jira-users)
└── group_members (user membership in groups)

Permission & Notification Schemes:
├── permission_schemes (reusable permission sets)
├── permission_scheme_permissions (permissions per scheme)
├── notification_schemes (P1 - optional for MVP)
└── notification_scheme_events (P1 - optional for MVP)

Projects & Organization:
├── projects (with FK to permission_scheme)
└── project_members (users OR groups with roles per project)

Issue Management:
├── issue_types (Epic, Story, Task, Bug, Subtask)
├── workflows
├── workflow_statuses
├── workflow_transitions
├── issues
├── issue_comments
├── issue_relations (blocks, relates to)
├── labels
└── issue_labels (many-to-many)

Scrum/Agile:
├── sprints
└── sprint_issues (many-to-many)

Optional (P2):
├── attachments
└── worklogs
```

---

## Detailed Tables

### 1. Core Authentication

#### `users` (P0) ✅ Created
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_admin BOOLEAN NOT NULL DEFAULT false, -- Global administrator flag
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_admin ON users(is_admin) WHERE is_admin = true;
```

**Notes:**
- `password_hash`: bcrypt hash
- `avatar`: URL or base64 (TBD)
- `is_active`: For soft-deleting users
- `is_admin`: Global administrator (can do everything, like JIRA System Admin)
- Partial index on `is_admin` for fast admin lookups

---

#### `sessions` (P0)
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

**Notes:**
- JWT refresh tokens stored here
- Auto-cleanup of expired tokens via cron job

---

#### `oidc_providers` (P2 - Optional)
```sql
CREATE TABLE oidc_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL, -- "Google", "Keycloak", etc.
    issuer_url VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Notes:**
- Global OIDC configuration
- Linking users via email matching

---

### 2. Groups & Global Permissions

#### `groups` (P0)
```sql
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_groups_name ON groups(name);
```

**Default groups:**
```sql
INSERT INTO groups (name, description) VALUES
('taskcore-administrators', 'Global administrators'),
('taskcore-developers', 'Developers across projects'),
('taskcore-users', 'All users with login access');
```

**Notes:**
- Global groups (like Jira's jira-developers, jira-users)
- Users can belong to multiple groups
- Groups can be assigned to project roles

---

#### `group_members` (P0)
```sql
CREATE TABLE group_members (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
```

**Notes:**
- Many-to-many relationship between users and groups
- Group membership is **global** (not project-specific)

---

### 3. Permission & Notification Schemes

#### `permission_schemes` (P0)
```sql
CREATE TABLE permission_schemes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Default scheme:**
```sql
INSERT INTO permission_schemes (name, description) VALUES
('Default Permission Scheme', 'Default permissions for new projects');
```

**Notes:**
- Permission schemes are reusable sets of permissions
- Projects reference a permission scheme
- Like Jira's permission schemes (but simplified for MVP)

---

#### `permission_scheme_permissions` (P0)
```sql
CREATE TYPE permission_type AS ENUM (
    -- Project permissions
    'administer_project',
    'browse_project',
    -- Issue permissions
    'create_issue',
    'edit_issue',
    'delete_issue',
    'assign_issue',
    'transition_issue',
    'close_issue',
    -- Comment permissions
    'add_comment',
    'edit_own_comment',
    'edit_all_comments',
    'delete_own_comment',
    'delete_all_comments',
    -- Sprint permissions (Scrum)
    'manage_sprints',
    'add_issue_to_sprint',
    -- Workflow permissions
    'configure_workflow',
    -- Label permissions
    'create_label',
    'edit_label'
);

CREATE TYPE grantee_type AS ENUM ('user', 'group', 'project_role', 'anyone');

CREATE TABLE permission_scheme_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    permission_scheme_id UUID NOT NULL REFERENCES permission_schemes(id) ON DELETE CASCADE,
    permission permission_type NOT NULL,
    grantee_type grantee_type NOT NULL,
    grantee_id UUID, -- user_id, group_id, or NULL for 'anyone'/'project_role'
    project_role project_role, -- Only if grantee_type = 'project_role'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(permission_scheme_id, permission, grantee_type, grantee_id, project_role)
);

CREATE INDEX idx_perm_scheme_perms_scheme_id ON permission_scheme_permissions(permission_scheme_id);
CREATE INDEX idx_perm_scheme_perms_permission ON permission_scheme_permissions(permission);
```

**Notes:**
- Each row grants a specific permission to a user, group, or project role
- `grantee_type`:
  - `'user'`: Direct user grant (grantee_id = user_id)
  - `'group'`: Group grant (grantee_id = group_id)
  - `'project_role'`: Role-based (e.g., all 'developers')
  - `'anyone'`: Public access
- Example: "Grant 'create_issue' to project_role 'developer'"

---

#### `notification_schemes` (P1 - Optional for MVP)
```sql
CREATE TABLE notification_schemes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Notes:**
- Optional for MVP
- Defines who gets notified for what events

---

#### `notification_scheme_events` (P1 - Optional for MVP)
```sql
CREATE TYPE notification_event AS ENUM (
    'issue_created',
    'issue_updated',
    'issue_assigned',
    'issue_commented',
    'issue_transitioned',
    'issue_closed'
);

CREATE TABLE notification_scheme_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_scheme_id UUID NOT NULL REFERENCES notification_schemes(id) ON DELETE CASCADE,
    event notification_event NOT NULL,
    notify_users UUID[], -- Array of user IDs
    notify_groups UUID[], -- Array of group IDs
    notify_roles project_role[], -- Array of project roles
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notif_scheme_events_scheme_id ON notification_scheme_events(notification_scheme_id);
```

**Notes:**
- Optional for MVP
- Can notify specific users, groups, or roles
- Uses PostgreSQL arrays for simplicity

---

### 4. Projects & Organization

#### `projects` (P0)
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(10) NOT NULL UNIQUE, -- e.g., "TASK", "BUG"
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'kanban', -- 'kanban' | 'scrum'
    avatar TEXT,
    permission_scheme_id UUID NOT NULL REFERENCES permission_schemes(id),
    notification_scheme_id UUID REFERENCES notification_schemes(id), -- Optional (P1)
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_key ON projects(key);
CREATE INDEX idx_projects_is_archived ON projects(is_archived);
CREATE INDEX idx_projects_permission_scheme_id ON projects(permission_scheme_id);
```

**Notes:**
- `key`: Prefix for issues (e.g., TASK-123)
- `type`: Determines available features (sprints only for scrum)
- `permission_scheme_id`: Required FK to permission scheme
- `notification_scheme_id`: Optional FK to notification scheme (P1)
- No `organization` table in MVP (can add later)

---

#### `project_members` (P0)
```sql
CREATE TYPE project_role AS ENUM ('admin', 'developer', 'reporter');

CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    role project_role NOT NULL DEFAULT 'reporter',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CHECK (
        (user_id IS NOT NULL AND group_id IS NULL) OR
        (user_id IS NULL AND group_id IS NOT NULL)
    ),
    UNIQUE(project_id, user_id),
    UNIQUE(project_id, group_id)
);

CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_group_id ON project_members(group_id);
```

**Notes:**
- Supports **both** users and groups as members
- **Either** `user_id` OR `group_id` must be set (CHECK constraint)
- If a group is assigned a role, all users in that group inherit it
- **Roles:**
  - `admin`: Configure workflows, manage members
  - `developer`: CRUD issues, comment, assign
  - `reporter`: Create issues, comment only
- **Permission resolution:**
  1. Check direct user assignment
  2. Check group assignments (user inherits highest role from all groups)

---

### 5. Issue Management

#### `issue_types` (P0)
```sql
CREATE TABLE issue_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE, -- 'Epic', 'Story', 'Task', 'Bug', 'Subtask'
    description TEXT,
    icon VARCHAR(50), -- e.g., 'epic', 'story', 'task', 'bug', 'subtask'
    color VARCHAR(7), -- Hex color, e.g., '#6554C0'
    is_subtask BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Default data:**
```sql
INSERT INTO issue_types (name, icon, color, is_subtask) VALUES
('Epic', 'epic', '#6554C0', false),
('Story', 'story', '#0052CC', false),
('Task', 'task', '#2684FF', false),
('Bug', 'bug', '#FF5630', false),
('Subtask', 'subtask', '#5E6C84', true);
```

---

#### `workflows` (P0)
```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_type_id UUID NOT NULL REFERENCES issue_types(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, issue_type_id)
);

CREATE INDEX idx_workflows_project_id ON workflows(project_id);
```

**Notes:**
- Each project can have different workflows per issue type
- Example: Bug workflow different from Task workflow

---

#### `workflow_statuses` (P0)
```sql
CREATE TABLE workflow_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    category VARCHAR(20) NOT NULL, -- 'todo', 'in_progress', 'done'
    position INTEGER NOT NULL, -- For ordering columns in board
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_statuses_workflow_id ON workflow_statuses(workflow_id);
```

**Default statuses:**
- To Do (category: 'todo')
- In Progress (category: 'in_progress')
- In Review (category: 'in_progress')
- Done (category: 'done')

---

#### `workflow_transitions` (P0)
```sql
CREATE TABLE workflow_transitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    from_status_id UUID NOT NULL REFERENCES workflow_statuses(id) ON DELETE CASCADE,
    to_status_id UUID NOT NULL REFERENCES workflow_statuses(id) ON DELETE CASCADE,
    allowed_role project_role, -- NULL means any role
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(workflow_id, from_status_id, to_status_id)
);

CREATE INDEX idx_workflow_transitions_workflow_id ON workflow_transitions(workflow_id);
```

**Notes:**
- Defines valid state transitions
- `allowed_role`: Optional permission check (e.g., only admins can move to "Done")

---

#### `issues` (P0)
```sql
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_number INTEGER NOT NULL, -- Auto-increment per project
    key VARCHAR(50) NOT NULL UNIQUE, -- e.g., "TASK-123" (generated)
    issue_type_id UUID NOT NULL REFERENCES issue_types(id),
    status_id UUID NOT NULL REFERENCES workflow_statuses(id),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    reporter_id UUID NOT NULL REFERENCES users(id),
    assignee_id UUID REFERENCES users(id),
    parent_issue_id UUID REFERENCES issues(id), -- For Epics > Stories > Subtasks
    priority VARCHAR(20) DEFAULT 'medium', -- 'lowest', 'low', 'medium', 'high', 'highest'
    story_points INTEGER, -- Optional, for Scrum
    sprint_id UUID REFERENCES sprints(id), -- NULL if in backlog
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_issues_project_id ON issues(project_id);
CREATE INDEX idx_issues_key ON issues(key);
CREATE INDEX idx_issues_assignee_id ON issues(assignee_id);
CREATE INDEX idx_issues_status_id ON issues(status_id);
CREATE INDEX idx_issues_parent_issue_id ON issues(parent_issue_id);
CREATE INDEX idx_issues_sprint_id ON issues(sprint_id);

-- Auto-generate issue key
CREATE UNIQUE INDEX idx_issues_project_number ON issues(project_id, issue_number);
```

**Notes:**
- `key`: Auto-generated as `{project.key}-{issue_number}`
- `issue_number`: Sequence per project (e.g., 1, 2, 3...)
- `parent_issue_id`: Hierarchy (Epic → Story → Subtask)
- `sprint_id`: NULL means in backlog

---

#### `issue_comments` (P0)
```sql
CREATE TABLE issue_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL, -- Markdown supported
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issue_comments_issue_id ON issue_comments(issue_id);
CREATE INDEX idx_issue_comments_author_id ON issue_comments(author_id);
```

---

#### `issue_relations` (P1)
```sql
CREATE TYPE issue_relation_type AS ENUM ('blocks', 'is_blocked_by', 'relates_to', 'duplicates', 'is_duplicated_by');

CREATE TABLE issue_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    relation_type issue_relation_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(source_issue_id, target_issue_id, relation_type)
);

CREATE INDEX idx_issue_relations_source ON issue_relations(source_issue_id);
CREATE INDEX idx_issue_relations_target ON issue_relations(target_issue_id);
```

---

#### `labels` (P0)
```sql
CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL, -- Hex color
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_labels_project_id ON labels(project_id);
```

---

#### `issue_labels` (P0)
```sql
CREATE TABLE issue_labels (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issue_id, label_id)
);

CREATE INDEX idx_issue_labels_issue_id ON issue_labels(issue_id);
CREATE INDEX idx_issue_labels_label_id ON issue_labels(label_id);
```

---

### 6. Scrum/Agile

#### `sprints` (P0)
```sql
CREATE TYPE sprint_state AS ENUM ('future', 'active', 'closed');

CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    goal TEXT,
    state sprint_state NOT NULL DEFAULT 'future',
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sprints_project_id ON sprints(project_id);
CREATE INDEX idx_sprints_state ON sprints(state);
```

**Notes:**
- Only one sprint can be 'active' per project at a time (enforce in app logic)
- `future`: Planning, `active`: In progress, `closed`: Completed

---

### 7. Optional (P2)

#### `attachments` (P2)
```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id),
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL, -- bytes
    mime_type VARCHAR(100),
    storage_path TEXT NOT NULL, -- S3 key or local path
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_issue_id ON attachments(issue_id);
```

---

#### `worklogs` (P2)
```sql
CREATE TABLE worklogs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    time_spent INTEGER NOT NULL, -- seconds
    description TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_worklogs_issue_id ON worklogs(issue_id);
CREATE INDEX idx_worklogs_user_id ON worklogs(user_id);
```

---

## Migration Order

Create migrations in this dependency order:

**Core Auth & Users:**
1. ✅ `000001_create_users.sql` (with is_admin flag)
2. `000002_create_sessions.sql`
3. (Optional P2) `000003_create_oidc_providers.sql`

**Groups:**
4. `000004_create_groups.sql` (with default groups insert)
5. `000005_create_group_members.sql`

**Permission & Notification Schemes:**
6. `000006_create_permission_schemes.sql` (with default scheme insert)
7. `000007_create_permission_scheme_permissions.sql`
8. (Optional P1) `000008_create_notification_schemes.sql`
9. (Optional P1) `000009_create_notification_scheme_events.sql`

**Projects:**
10. `000010_create_projects.sql` (references permission_scheme_id)
11. `000011_create_project_members.sql` (supports users AND groups)

**Issue Types & Workflows:**
12. `000012_create_issue_types.sql` (with default types insert)
13. `000013_create_workflows.sql`
14. `000014_create_workflow_statuses.sql`
15. `000015_create_workflow_transitions.sql`

**Sprints:**
16. `000016_create_sprints.sql`

**Issues:**
17. `000017_create_issues.sql` (references sprints, workflows, issue_types)
18. `000018_create_issue_comments.sql`
19. `000019_create_issue_relations.sql`

**Labels:**
20. `000020_create_labels.sql`
21. `000021_create_issue_labels.sql`

**Optional (P2):**
22. (Optional) `000022_create_attachments.sql`
23. (Optional) `000023_create_worklogs.sql`

---

## Notes

### Auto-Incrementing Issue Numbers

PostgreSQL sequences per project:
```sql
CREATE SEQUENCE issues_seq_${project_key};
```

Or use application-level logic:
```go
func getNextIssueNumber(projectID uuid.UUID) int {
    // SELECT MAX(issue_number) + 1 FROM issues WHERE project_id = ?
}
```

### Audit Logging

All critical tables have `created_at` and `updated_at` timestamps.
For full audit trail, consider adding `issue_history` table (post-MVP).

### Soft Deletes

Consider adding `deleted_at` columns for soft deletes (post-MVP).

---

## Diagram

See `docs/schema.png` for visual ER diagram (TODO: generate with `dbdiagram.io` or similar).
