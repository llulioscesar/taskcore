-- Add hierarchy level to issue_types
-- 0 = Epic (top level, cannot have parent)
-- 1 = Story/Bug/Task (can belong to Epic)
-- 2 = Subtask (must have parent Story/Bug/Task)
ALTER TABLE issue_types ADD COLUMN hierarchy_level SMALLINT NOT NULL DEFAULT 1;

UPDATE issue_types SET hierarchy_level = 0 WHERE name = 'Epic';
UPDATE issue_types SET hierarchy_level = 1 WHERE name IN ('Story', 'Task', 'Bug');
UPDATE issue_types SET hierarchy_level = 2 WHERE name = 'Subtask';

-- Add epic_id to issues (separate from parent_id)
-- epic_id links Story/Bug/Task to an Epic
-- parent_id links Subtask to Story/Bug/Task
ALTER TABLE issues ADD COLUMN epic_id UUID REFERENCES issues(id) ON DELETE SET NULL;

CREATE INDEX idx_issues_epic ON issues(epic_id) WHERE epic_id IS NOT NULL;

-- Add epic-specific fields
ALTER TABLE issues ADD COLUMN color VARCHAR(7);
ALTER TABLE issues ADD COLUMN start_date DATE;
ALTER TABLE issues ADD COLUMN target_date DATE;
