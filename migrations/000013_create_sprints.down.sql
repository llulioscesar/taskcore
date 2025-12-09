-- Drop trigger
DROP TRIGGER IF EXISTS update_sprints_updated_at ON sprints;

-- Drop indexes
DROP INDEX IF EXISTS idx_sprints_state;
DROP INDEX IF EXISTS idx_sprints_project_id;

-- Drop table
DROP TABLE IF EXISTS sprints;

-- Drop ENUM
DROP TYPE IF EXISTS sprint_state;
