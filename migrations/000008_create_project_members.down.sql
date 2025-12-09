-- Drop indexes
DROP INDEX IF EXISTS idx_project_members_group_id;
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_project_members_project_id;

-- Drop table
DROP TABLE IF EXISTS project_members;
