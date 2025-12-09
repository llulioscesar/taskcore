CREATE TABLE issue_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    issue_type_id UUID NOT NULL REFERENCES issue_types(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    summary VARCHAR(255) NOT NULL,
    content TEXT,
    priority VARCHAR(20),
    labels TEXT[] DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, name)
);

CREATE INDEX idx_issue_templates_project ON issue_templates(project_id);
CREATE INDEX idx_issue_templates_type ON issue_templates(issue_type_id);

CREATE TABLE issue_template_fields (
    template_id UUID NOT NULL REFERENCES issue_templates(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    text_value TEXT,
    num_value DOUBLE PRECISION,
    PRIMARY KEY (template_id, field_id)
);
