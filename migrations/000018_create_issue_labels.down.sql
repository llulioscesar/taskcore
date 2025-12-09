-- Drop indexes
DROP INDEX IF EXISTS idx_issue_labels_label_id;
DROP INDEX IF EXISTS idx_issue_labels_issue_id;

-- Drop table
DROP TABLE IF EXISTS issue_labels;
