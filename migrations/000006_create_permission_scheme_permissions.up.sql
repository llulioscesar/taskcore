-- Create ENUM for permission types
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

-- Create ENUM for grantee types
CREATE TYPE grantee_type AS ENUM ('user', 'group', 'project_role', 'anyone');

-- Create ENUM for project roles (will be used by project_members too)
CREATE TYPE project_role AS ENUM ('admin', 'developer', 'reporter');

-- Create permission_scheme_permissions table
CREATE TABLE permission_scheme_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    permission_scheme_id UUID NOT NULL REFERENCES permission_schemes(id) ON DELETE CASCADE,
    permission permission_type NOT NULL,
    grantee_type grantee_type NOT NULL,
    grantee_id UUID, -- user_id or group_id, NULL for 'anyone' or 'project_role'
    project_role project_role, -- Only used when grantee_type = 'project_role'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(permission_scheme_id, permission, grantee_type, grantee_id, project_role)
);

-- Indexes
CREATE INDEX idx_perm_scheme_perms_scheme_id ON permission_scheme_permissions(permission_scheme_id);
CREATE INDEX idx_perm_scheme_perms_permission ON permission_scheme_permissions(permission);

-- Insert default permissions for the default scheme
-- Get the default scheme ID and insert permissions
DO $$
DECLARE
    default_scheme_id UUID;
BEGIN
    SELECT id INTO default_scheme_id FROM permission_schemes WHERE name = 'Default Permission Scheme';

    -- Admin role gets all permissions
    INSERT INTO permission_scheme_permissions (permission_scheme_id, permission, grantee_type, project_role) VALUES
    (default_scheme_id, 'administer_project', 'project_role', 'admin'),
    (default_scheme_id, 'configure_workflow', 'project_role', 'admin'),
    (default_scheme_id, 'manage_sprints', 'project_role', 'admin'),
    (default_scheme_id, 'delete_issue', 'project_role', 'admin'),
    (default_scheme_id, 'edit_all_comments', 'project_role', 'admin'),
    (default_scheme_id, 'delete_all_comments', 'project_role', 'admin'),
    (default_scheme_id, 'edit_label', 'project_role', 'admin');

    -- Developer role permissions
    INSERT INTO permission_scheme_permissions (permission_scheme_id, permission, grantee_type, project_role) VALUES
    (default_scheme_id, 'browse_project', 'project_role', 'developer'),
    (default_scheme_id, 'create_issue', 'project_role', 'developer'),
    (default_scheme_id, 'edit_issue', 'project_role', 'developer'),
    (default_scheme_id, 'assign_issue', 'project_role', 'developer'),
    (default_scheme_id, 'transition_issue', 'project_role', 'developer'),
    (default_scheme_id, 'close_issue', 'project_role', 'developer'),
    (default_scheme_id, 'add_comment', 'project_role', 'developer'),
    (default_scheme_id, 'edit_own_comment', 'project_role', 'developer'),
    (default_scheme_id, 'delete_own_comment', 'project_role', 'developer'),
    (default_scheme_id, 'add_issue_to_sprint', 'project_role', 'developer'),
    (default_scheme_id, 'create_label', 'project_role', 'developer');

    -- Reporter role permissions
    INSERT INTO permission_scheme_permissions (permission_scheme_id, permission, grantee_type, project_role) VALUES
    (default_scheme_id, 'browse_project', 'project_role', 'reporter'),
    (default_scheme_id, 'create_issue', 'project_role', 'reporter'),
    (default_scheme_id, 'add_comment', 'project_role', 'reporter'),
    (default_scheme_id, 'edit_own_comment', 'project_role', 'reporter'),
    (default_scheme_id, 'delete_own_comment', 'project_role', 'reporter');

    -- Admin also gets developer permissions (inheritance)
    INSERT INTO permission_scheme_permissions (permission_scheme_id, permission, grantee_type, project_role) VALUES
    (default_scheme_id, 'browse_project', 'project_role', 'admin'),
    (default_scheme_id, 'create_issue', 'project_role', 'admin'),
    (default_scheme_id, 'edit_issue', 'project_role', 'admin'),
    (default_scheme_id, 'assign_issue', 'project_role', 'admin'),
    (default_scheme_id, 'transition_issue', 'project_role', 'admin'),
    (default_scheme_id, 'close_issue', 'project_role', 'admin'),
    (default_scheme_id, 'add_comment', 'project_role', 'admin'),
    (default_scheme_id, 'edit_own_comment', 'project_role', 'admin'),
    (default_scheme_id, 'delete_own_comment', 'project_role', 'admin'),
    (default_scheme_id, 'add_issue_to_sprint', 'project_role', 'admin'),
    (default_scheme_id, 'create_label', 'project_role', 'admin');
END $$;
