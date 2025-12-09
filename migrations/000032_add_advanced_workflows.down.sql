DROP TABLE IF EXISTS workflow_scheme_mappings;
DROP TABLE IF EXISTS workflow_schemes;
DROP TABLE IF EXISTS workflow_transition_post_functions;
DROP TABLE IF EXISTS workflow_transition_validators;
DROP TABLE IF EXISTS workflow_transition_conditions;
ALTER TABLE workflow_transitions DROP COLUMN IF EXISTS screen_id;
DROP TABLE IF EXISTS workflow_screen_fields;
DROP TABLE IF EXISTS workflow_screens;
