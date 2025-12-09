ALTER TABLE projects DROP COLUMN IF EXISTS template_id;
DROP TABLE IF EXISTS project_template_columns;
DROP TABLE IF EXISTS project_template_issue_types;
DROP TABLE IF EXISTS project_template_transitions;
DROP TABLE IF EXISTS project_template_statuses;
DROP TABLE IF EXISTS project_templates;
