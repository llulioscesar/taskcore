-- Applications/Modules (like Jira Software, Service Desk, etc)
-- Open source: modules can be enabled/disabled, no commercial licensing
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User Application Access (direct assignment)
CREATE TABLE application_user_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, application_id)
);

CREATE INDEX idx_application_user_access_user ON application_user_access(user_id);
CREATE INDEX idx_application_user_access_app ON application_user_access(application_id);

-- Group Application Access (all group members get access)
CREATE TABLE application_group_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, application_id)
);

CREATE INDEX idx_application_group_access_group ON application_group_access(group_id);
CREATE INDEX idx_application_group_access_app ON application_group_access(application_id);

-- Default Applications (auto-grant to new users)
CREATE TABLE application_defaults (
    application_id UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

-- Insert default applications/modules
INSERT INTO applications (key, name, description, is_active) VALUES
('core', 'Taskcore', 'Core issue tracking functionality', true),
('agile', 'Taskcore Agile', 'Agile boards, sprints, and software development tools', true),
('servicedesk', 'Taskcore Service Desk', 'IT service management and customer support', false);

-- Set core as default (all new users get access)
INSERT INTO application_defaults (application_id, is_default)
SELECT id, true FROM applications WHERE key = 'core';

-- Grant all existing users access to core
INSERT INTO application_user_access (user_id, application_id)
SELECT u.id, a.id FROM users u, applications a WHERE a.key = 'core'
ON CONFLICT DO NOTHING;
