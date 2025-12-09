-- Project Roles (dynamic, like Jira)
CREATE TABLE project_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Project Role Actors (users/groups assigned to roles per project)
CREATE TABLE project_role_actors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES project_roles(id) ON DELETE CASCADE,
    actor_type VARCHAR(10) NOT NULL, -- 'user' or 'group'
    actor_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, role_id, actor_type, actor_id)
);

CREATE INDEX idx_project_role_actors_project ON project_role_actors(project_id);
CREATE INDEX idx_project_role_actors_role ON project_role_actors(role_id);
CREATE INDEX idx_project_role_actors_actor ON project_role_actors(actor_type, actor_id);

-- Default Role Actors (applied to new projects)
CREATE TABLE default_role_actors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES project_roles(id) ON DELETE CASCADE,
    actor_type VARCHAR(10) NOT NULL,
    actor_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (role_id, actor_type, actor_id)
);

-- Global Roles (system-wide)
CREATE TABLE global_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Global Role Members
CREATE TABLE global_role_members (
    global_role_id UUID NOT NULL REFERENCES global_roles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (global_role_id, user_id)
);

CREATE INDEX idx_global_role_members_user ON global_role_members(user_id);

-- Global Permission Grants
CREATE TABLE global_permission_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission VARCHAR(50) NOT NULL,
    global_role_id UUID REFERENCES global_roles(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX idx_global_permission_grants_permission ON global_permission_grants(permission);

-- Insert default project roles
INSERT INTO project_roles (name, description) VALUES
('Administrators', 'Full access to project configuration and all project permissions'),
('Developers', 'Can work on issues, log work, and transition issues'),
('Viewers', 'Can browse and view issues but cannot edit');

-- Insert default global roles
INSERT INTO global_roles (name, description) VALUES
('system_admin', 'Full system administration access'),
('admin', 'Jira administration access'),
('user', 'Regular user access');

-- Grant default global permissions
INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'administer_system', id FROM global_roles WHERE name = 'system_admin';

INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'administer_jira', id FROM global_roles WHERE name = 'admin';

INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'manage_users', id FROM global_roles WHERE name = 'admin';

INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'manage_groups', id FROM global_roles WHERE name = 'admin';

INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'create_projects', id FROM global_roles WHERE name = 'admin';

INSERT INTO global_permission_grants (permission, global_role_id)
SELECT 'browse_users', id FROM global_roles WHERE name = 'user';

-- Migrate existing project_members to new role system
-- Add project lead as Administrator for each project
INSERT INTO project_role_actors (project_id, role_id, actor_type, actor_id)
SELECT p.id, pr.id, 'user', p.lead_id
FROM projects p, project_roles pr
WHERE pr.name = 'Administrators'
ON CONFLICT DO NOTHING;
