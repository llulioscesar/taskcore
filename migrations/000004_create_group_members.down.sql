-- Drop indexes
DROP INDEX IF EXISTS idx_group_members_user_id;
DROP INDEX IF EXISTS idx_group_members_group_id;

-- Drop table
DROP TABLE IF EXISTS group_members;
