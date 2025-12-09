-- Drop indexes
DROP INDEX IF EXISTS idx_perm_scheme_perms_permission;
DROP INDEX IF EXISTS idx_perm_scheme_perms_scheme_id;

-- Drop table
DROP TABLE IF EXISTS permission_scheme_permissions;

-- Drop ENUMs
DROP TYPE IF EXISTS project_role;
DROP TYPE IF EXISTS grantee_type;
DROP TYPE IF EXISTS permission_type;
