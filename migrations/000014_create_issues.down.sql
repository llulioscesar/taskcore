-- Drop trigger
DROP TRIGGER IF EXISTS update_issues_updated_at ON issues;

-- Drop indexes
DROP INDEX IF EXISTS idx_issues_issue_type_id;
DROP INDEX IF EXISTS idx_issues_sprint_id;
DROP INDEX IF EXISTS idx_issues_parent_issue_id;
DROP INDEX IF EXISTS idx_issues_status_id;
DROP INDEX IF EXISTS idx_issues_reporter_id;
DROP INDEX IF EXISTS idx_issues_assignee_id;
DROP INDEX IF EXISTS idx_issues_key;
DROP INDEX IF EXISTS idx_issues_project_id;

-- Drop table
DROP TABLE IF EXISTS issues;
