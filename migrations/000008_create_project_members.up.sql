-- Create project_members table (supports both users and groups)
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    role project_role NOT NULL DEFAULT 'reporter',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- Either user_id OR group_id must be set, not both
    CONSTRAINT project_members_user_or_group CHECK (
        (user_id IS NOT NULL AND group_id IS NULL) OR
        (user_id IS NULL AND group_id IS NOT NULL)
    ),
    -- Unique constraints
    CONSTRAINT project_members_unique_user UNIQUE (project_id, user_id),
    CONSTRAINT project_members_unique_group UNIQUE (project_id, group_id)
);

-- Indexes
CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_group_id ON project_members(group_id);
