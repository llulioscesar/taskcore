-- Drop trigger
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;

-- Drop indexes
DROP INDEX IF EXISTS idx_projects_permission_scheme_id;
DROP INDEX IF EXISTS idx_projects_is_archived;
DROP INDEX IF EXISTS idx_projects_key;

-- Drop table
DROP TABLE IF EXISTS projects;
