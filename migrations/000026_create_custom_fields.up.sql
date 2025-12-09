CREATE TABLE custom_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    field_type VARCHAR(20) NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_global BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE custom_field_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    value VARCHAR(255) NOT NULL,
    color VARCHAR(7),
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_custom_field_options_field ON custom_field_options(field_id);

CREATE TABLE project_custom_fields (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, field_id)
);

CREATE TABLE custom_field_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    text_value TEXT,
    num_value DOUBLE PRECISION,
    date_value TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (issue_id, field_id)
);

CREATE INDEX idx_custom_field_values_issue ON custom_field_values(issue_id);

CREATE TABLE custom_field_value_options (
    value_id UUID NOT NULL REFERENCES custom_field_values(id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES custom_field_options(id) ON DELETE CASCADE,
    PRIMARY KEY (value_id, option_id)
);
