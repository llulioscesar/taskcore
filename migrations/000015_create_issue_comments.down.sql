-- Drop trigger
DROP TRIGGER IF EXISTS update_issue_comments_updated_at ON issue_comments;

-- Drop indexes
DROP INDEX IF EXISTS idx_issue_comments_author_id;
DROP INDEX IF EXISTS idx_issue_comments_issue_id;

-- Drop table
DROP TABLE IF EXISTS issue_comments;
