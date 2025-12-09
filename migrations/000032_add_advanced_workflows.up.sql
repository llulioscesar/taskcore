-- Workflow Screens (fields to show during transition)
CREATE TABLE workflow_screens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow_screen_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    screen_id UUID NOT NULL REFERENCES workflow_screens(id) ON DELETE CASCADE,
    field_type VARCHAR(20) NOT NULL DEFAULT 'standard', -- standard, custom
    field_name VARCHAR(100) NOT NULL,
    field_id UUID REFERENCES custom_fields(id) ON DELETE CASCADE, -- for custom fields
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_screen_fields_screen ON workflow_screen_fields(screen_id);

-- Add screen_id to transitions
ALTER TABLE workflow_transitions ADD COLUMN screen_id UUID REFERENCES workflow_screens(id) ON DELETE SET NULL;

-- Transition Conditions (who can execute)
CREATE TABLE workflow_transition_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transition_id UUID NOT NULL REFERENCES workflow_transitions(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_transition_conditions ON workflow_transition_conditions(transition_id);

-- Transition Validators (validate before transition)
CREATE TABLE workflow_transition_validators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transition_id UUID NOT NULL REFERENCES workflow_transitions(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    error_message TEXT,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_transition_validators ON workflow_transition_validators(transition_id);

-- Transition Post Functions (actions after transition)
CREATE TABLE workflow_transition_post_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transition_id UUID NOT NULL REFERENCES workflow_transitions(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_transition_post_functions ON workflow_transition_post_functions(transition_id);

-- Workflow Schemes (map issue types to workflows)
CREATE TABLE workflow_schemes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow_scheme_mappings (
    scheme_id UUID NOT NULL REFERENCES workflow_schemes(id) ON DELETE CASCADE,
    issue_type_id UUID REFERENCES issue_types(id) ON DELETE CASCADE, -- NULL = default for scheme
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    UNIQUE (scheme_id, issue_type_id)
);

CREATE INDEX idx_workflow_scheme_mappings_scheme ON workflow_scheme_mappings(scheme_id);

-- Insert default workflow scheme
INSERT INTO workflow_schemes (name, description, is_default) VALUES
('Default Workflow Scheme', 'Default scheme using the standard workflow for all issue types', true);

-- Map default workflow to default scheme
INSERT INTO workflow_scheme_mappings (scheme_id, issue_type_id, workflow_id)
SELECT ws.id, NULL, w.id
FROM workflow_schemes ws, workflows w
WHERE ws.is_default = true AND w.is_default = true;

-- Insert default screens
INSERT INTO workflow_screens (name, description) VALUES
('Resolve Issue Screen', 'Screen shown when resolving an issue'),
('Close Issue Screen', 'Screen shown when closing an issue'),
('Reopen Issue Screen', 'Screen shown when reopening an issue');

-- Add resolution field to resolve screen
INSERT INTO workflow_screen_fields (screen_id, field_type, field_name, is_required, position)
SELECT id, 'standard', 'resolution', true, 0
FROM workflow_screens WHERE name = 'Resolve Issue Screen';

-- Add comment field to close screen
INSERT INTO workflow_screen_fields (screen_id, field_type, field_name, is_required, position)
SELECT id, 'standard', 'comment', false, 0
FROM workflow_screens WHERE name = 'Close Issue Screen';
