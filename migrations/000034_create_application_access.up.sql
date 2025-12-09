-- Applications (like Jira Software, Service Desk, etc)
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- License
CREATE TABLE licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_key VARCHAR(255) NOT NULL,
    license_type VARCHAR(20) NOT NULL DEFAULT 'community',
    max_users INTEGER NOT NULL DEFAULT 10,
    expires_at TIMESTAMPTZ,
    licensed_to VARCHAR(255) NOT NULL,
    support_expires TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User Application Access
CREATE TABLE application_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, application_id)
);

CREATE INDEX idx_application_access_user ON application_access(user_id);
CREATE INDEX idx_application_access_app ON application_access(application_id);

-- Group Application Access
CREATE TABLE application_group_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, application_id)
);

CREATE INDEX idx_application_group_access_group ON application_group_access(group_id);

-- Default Applications (auto-grant to new users)
CREATE TABLE application_defaults (
    application_id UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

-- Insert default applications
INSERT INTO applications (key, name, description, is_active) VALUES
('core', 'Taskcore', 'Core issue tracking functionality', true),
('software', 'Taskcore Software', 'Agile boards, sprints, and software development tools', true),
('servicedesk', 'Taskcore Service Desk', 'IT service management and customer support', false);

-- Set core as default (all new users get access)
INSERT INTO application_defaults (application_id, is_default)
SELECT id, true FROM applications WHERE key = 'core';

-- Insert default community license
INSERT INTO licenses (license_key, license_type, max_users, licensed_to) VALUES
('COMMUNITY-FREE', 'community', 10, 'Community Edition');

-- Grant all existing users access to core
INSERT INTO application_access (user_id, application_id)
SELECT u.id, a.id FROM users u, applications a WHERE a.key = 'core'
ON CONFLICT DO NOTHING;
