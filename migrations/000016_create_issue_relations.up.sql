-- Create ENUM for issue relation types
CREATE TYPE issue_relation_type AS ENUM ('blocks', 'is_blocked_by', 'relates_to', 'duplicates', 'is_duplicated_by');

-- Create issue_relations table
CREATE TABLE issue_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    relation_type issue_relation_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(source_issue_id, target_issue_id, relation_type)
);

-- Indexes
CREATE INDEX idx_issue_relations_source ON issue_relations(source_issue_id);
CREATE INDEX idx_issue_relations_target ON issue_relations(target_issue_id);
