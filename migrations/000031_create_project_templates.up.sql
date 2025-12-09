-- Project templates (like Jira: Scrum, Kanban, Bug Tracking)
CREATE TABLE project_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'software',
    board_type VARCHAR(20) NOT NULL DEFAULT 'kanban',
    icon VARCHAR(50) NOT NULL DEFAULT 'project',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Template statuses (workflow configuration)
CREATE TABLE project_template_statuses (
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    category VARCHAR(20) NOT NULL, -- todo, in_progress, done
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (template_id, name)
);

-- Template transitions
CREATE TABLE project_template_transitions (
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (template_id, from_status, to_status)
);

-- Template issue types (which types are enabled)
CREATE TABLE project_template_issue_types (
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    issue_type_id UUID NOT NULL REFERENCES issue_types(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (template_id, issue_type_id)
);

-- Template board columns
CREATE TABLE project_template_columns (
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    status_name VARCHAR(50) NOT NULL,
    position INTEGER NOT NULL,
    min_limit INTEGER,
    max_limit INTEGER,
    PRIMARY KEY (template_id, position)
);

-- Add template_id to projects
ALTER TABLE projects ADD COLUMN template_id UUID REFERENCES project_templates(id) ON DELETE SET NULL;

-- Insert default templates
INSERT INTO project_templates (key, name, description, category, board_type, icon, is_default) VALUES
('scrum', 'Scrum', 'Scrum board for agile teams with sprints, backlog, and velocity tracking', 'software', 'scrum', 'sprint', true),
('kanban', 'Kanban', 'Kanban board for continuous flow with WIP limits', 'software', 'kanban', 'board', true),
('bug-tracking', 'Bug Tracking', 'Track and manage bugs with priority-based workflow', 'software', 'kanban', 'bug', true);

-- Scrum statuses
INSERT INTO project_template_statuses (template_id, name, category, position)
SELECT id, 'To Do', 'todo', 0 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'In Progress', 'in_progress', 1 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'In Review', 'in_progress', 2 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'Done', 'done', 3 FROM project_templates WHERE key = 'scrum';

-- Kanban statuses
INSERT INTO project_template_statuses (template_id, name, category, position)
SELECT id, 'Backlog', 'todo', 0 FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'Selected for Development', 'todo', 1 FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'In Progress', 'in_progress', 2 FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'Done', 'done', 3 FROM project_templates WHERE key = 'kanban';

-- Bug Tracking statuses
INSERT INTO project_template_statuses (template_id, name, category, position)
SELECT id, 'Open', 'todo', 0 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'In Progress', 'in_progress', 1 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'Resolved', 'done', 2 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'Closed', 'done', 3 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'Reopened', 'todo', 4 FROM project_templates WHERE key = 'bug-tracking';

-- Scrum columns
INSERT INTO project_template_columns (template_id, name, status_name, position)
SELECT id, 'To Do', 'To Do', 0 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'In Progress', 'In Progress', 1 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'In Review', 'In Review', 2 FROM project_templates WHERE key = 'scrum'
UNION ALL
SELECT id, 'Done', 'Done', 3 FROM project_templates WHERE key = 'scrum';

-- Kanban columns (with WIP limits)
INSERT INTO project_template_columns (template_id, name, status_name, position, max_limit)
SELECT id, 'Backlog', 'Backlog', 0, NULL FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'Selected', 'Selected for Development', 1, 5 FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'In Progress', 'In Progress', 2, 3 FROM project_templates WHERE key = 'kanban'
UNION ALL
SELECT id, 'Done', 'Done', 3, NULL FROM project_templates WHERE key = 'kanban';

-- Bug Tracking columns
INSERT INTO project_template_columns (template_id, name, status_name, position)
SELECT id, 'Open', 'Open', 0 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'In Progress', 'In Progress', 1 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'Resolved', 'Resolved', 2 FROM project_templates WHERE key = 'bug-tracking'
UNION ALL
SELECT id, 'Closed', 'Closed', 3 FROM project_templates WHERE key = 'bug-tracking';

-- Link issue types to templates
-- Scrum: Epic, Story, Task, Bug, Subtask
INSERT INTO project_template_issue_types (template_id, issue_type_id, is_default)
SELECT pt.id, it.id, it.name = 'Story'
FROM project_templates pt, issue_types it
WHERE pt.key = 'scrum' AND it.name IN ('Epic', 'Story', 'Task', 'Bug', 'Subtask');

-- Kanban: Story, Task, Bug, Subtask (no Epic)
INSERT INTO project_template_issue_types (template_id, issue_type_id, is_default)
SELECT pt.id, it.id, it.name = 'Task'
FROM project_templates pt, issue_types it
WHERE pt.key = 'kanban' AND it.name IN ('Story', 'Task', 'Bug', 'Subtask');

-- Bug Tracking: Bug only
INSERT INTO project_template_issue_types (template_id, issue_type_id, is_default)
SELECT pt.id, it.id, true
FROM project_templates pt, issue_types it
WHERE pt.key = 'bug-tracking' AND it.name = 'Bug';
