-- Drop indexes
DROP INDEX IF EXISTS idx_issue_relations_target;
DROP INDEX IF EXISTS idx_issue_relations_source;

-- Drop table
DROP TABLE IF EXISTS issue_relations;

-- Drop ENUM
DROP TYPE IF EXISTS issue_relation_type;
