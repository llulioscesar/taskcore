-- Create workflow_statuses table
CREATE TABLE workflow_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    category VARCHAR(20) NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_statuses_category_check CHECK (category IN ('todo', 'in_progress', 'done'))
);

-- Index
CREATE INDEX idx_workflow_statuses_workflow_id ON workflow_statuses(workflow_id);
