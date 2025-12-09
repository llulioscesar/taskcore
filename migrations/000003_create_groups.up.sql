-- Create groups table
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index
CREATE INDEX idx_groups_name ON groups(name);

-- Insert default groups
INSERT INTO groups (name, description) VALUES
('taskcore-administrators', 'Global administrators with full system access'),
('taskcore-developers', 'Developers across all projects'),
('taskcore-users', 'All users with login access');
