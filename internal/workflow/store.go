package workflow

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, w *Workflow) error {
	query := `
		INSERT INTO workflows (name, description, is_default)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query, w.Name, w.Description, w.IsDefault).
		Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Workflow, error) {
	w := &Workflow{}
	err := db.GetContext(ctx, w, `SELECT * FROM workflows WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return w, err
}

func GetDefault(ctx context.Context, db *sqlx.DB) (*Workflow, error) {
	w := &Workflow{}
	err := db.GetContext(ctx, w, `SELECT * FROM workflows WHERE is_default = true LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return w, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Workflow, error) {
	var workflows []*Workflow
	err := db.SelectContext(ctx, &workflows, `SELECT * FROM workflows ORDER BY is_default DESC, name`)
	return workflows, err
}

func Update(ctx context.Context, db *sqlx.DB, w *Workflow) error {
	query := `UPDATE workflows SET name = $2, description = $3, is_default = $4 WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, w.ID, w.Name, w.Description, w.IsDefault).Scan(&w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func CreateStatus(ctx context.Context, db *sqlx.DB, st *Status) error {
	query := `
		INSERT INTO workflow_statuses (workflow_id, name, category, position)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return db.QueryRowxContext(ctx, query, st.WorkflowID, st.Name, st.Category, st.Position).Scan(&st.ID)
}

func GetStatus(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Status, error) {
	st := &Status{}
	err := db.GetContext(ctx, st, `SELECT * FROM workflow_statuses WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStatusNotFound
	}
	return st, err
}

func ListStatuses(ctx context.Context, db *sqlx.DB, workflowID uuid.UUID) ([]*Status, error) {
	var statuses []*Status
	err := db.SelectContext(ctx, &statuses,
		`SELECT * FROM workflow_statuses WHERE workflow_id = $1 ORDER BY position`, workflowID)
	return statuses, err
}

func GetInitialStatus(ctx context.Context, db *sqlx.DB, workflowID uuid.UUID) (*Status, error) {
	st := &Status{}
	err := db.GetContext(ctx, st,
		`SELECT * FROM workflow_statuses WHERE workflow_id = $1 ORDER BY position LIMIT 1`, workflowID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStatusNotFound
	}
	return st, err
}

func UpdateStatus(ctx context.Context, db *sqlx.DB, st *Status) error {
	result, err := db.ExecContext(ctx,
		`UPDATE workflow_statuses SET name = $2, category = $3, position = $4 WHERE id = $1`,
		st.ID, st.Name, st.Category, st.Position)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrStatusNotFound
	}
	return nil
}

func DeleteStatus(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM workflow_statuses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrStatusNotFound
	}
	return nil
}

func CreateTransition(ctx context.Context, db *sqlx.DB, t *Transition) error {
	query := `
		INSERT INTO workflow_transitions (workflow_id, from_status_id, to_status_id, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workflow_id, from_status_id, to_status_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`
	return db.QueryRowxContext(ctx, query, t.WorkflowID, t.FromStatusID, t.ToStatusID, t.Name).Scan(&t.ID)
}

func ListTransitions(ctx context.Context, db *sqlx.DB, workflowID uuid.UUID) ([]*Transition, error) {
	var transitions []*Transition
	err := db.SelectContext(ctx, &transitions,
		`SELECT * FROM workflow_transitions WHERE workflow_id = $1 ORDER BY name`, workflowID)
	return transitions, err
}

func ListTransitionsFrom(ctx context.Context, db *sqlx.DB, workflowID, fromStatusID uuid.UUID) ([]*Transition, error) {
	var transitions []*Transition
	err := db.SelectContext(ctx, &transitions,
		`SELECT * FROM workflow_transitions WHERE workflow_id = $1 AND from_status_id = $2 ORDER BY name`,
		workflowID, fromStatusID)
	return transitions, err
}

func IsValidTransition(ctx context.Context, db *sqlx.DB, workflowID, fromStatusID, toStatusID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM workflow_transitions WHERE workflow_id = $1 AND from_status_id = $2 AND to_status_id = $3)`
	err := db.GetContext(ctx, &exists, query, workflowID, fromStatusID, toStatusID)
	return exists, err
}

func DeleteTransition(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM workflow_transitions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTransitionNotFound
	}
	return nil
}

// Conditions

func CreateCondition(ctx context.Context, db *sqlx.DB, c *Condition) error {
	query := `
		INSERT INTO workflow_transition_conditions (transition_id, type, config, position)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return db.QueryRowxContext(ctx, query, c.TransitionID, c.Type, c.Config, c.Position).Scan(&c.ID)
}

func ListConditions(ctx context.Context, db *sqlx.DB, transitionID uuid.UUID) ([]*Condition, error) {
	var conditions []*Condition
	err := db.SelectContext(ctx, &conditions,
		`SELECT * FROM workflow_transition_conditions WHERE transition_id = $1 ORDER BY position`, transitionID)
	return conditions, err
}

func UpdateCondition(ctx context.Context, db *sqlx.DB, c *Condition) error {
	_, err := db.ExecContext(ctx,
		`UPDATE workflow_transition_conditions SET type = $2, config = $3, position = $4 WHERE id = $1`,
		c.ID, c.Type, c.Config, c.Position)
	return err
}

func DeleteCondition(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM workflow_transition_conditions WHERE id = $1`, id)
	return err
}

// Validators

func CreateValidator(ctx context.Context, db *sqlx.DB, v *Validator) error {
	query := `
		INSERT INTO workflow_transition_validators (transition_id, type, config, error_message, position)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return db.QueryRowxContext(ctx, query, v.TransitionID, v.Type, v.Config, v.ErrorMessage, v.Position).Scan(&v.ID)
}

func ListValidators(ctx context.Context, db *sqlx.DB, transitionID uuid.UUID) ([]*Validator, error) {
	var validators []*Validator
	err := db.SelectContext(ctx, &validators,
		`SELECT * FROM workflow_transition_validators WHERE transition_id = $1 ORDER BY position`, transitionID)
	return validators, err
}

func UpdateValidator(ctx context.Context, db *sqlx.DB, v *Validator) error {
	_, err := db.ExecContext(ctx,
		`UPDATE workflow_transition_validators SET type = $2, config = $3, error_message = $4, position = $5 WHERE id = $1`,
		v.ID, v.Type, v.Config, v.ErrorMessage, v.Position)
	return err
}

func DeleteValidator(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM workflow_transition_validators WHERE id = $1`, id)
	return err
}

// Post Functions

func CreatePostFunction(ctx context.Context, db *sqlx.DB, pf *PostFunction) error {
	query := `
		INSERT INTO workflow_transition_post_functions (transition_id, type, config, position)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return db.QueryRowxContext(ctx, query, pf.TransitionID, pf.Type, pf.Config, pf.Position).Scan(&pf.ID)
}

func ListPostFunctions(ctx context.Context, db *sqlx.DB, transitionID uuid.UUID) ([]*PostFunction, error) {
	var functions []*PostFunction
	err := db.SelectContext(ctx, &functions,
		`SELECT * FROM workflow_transition_post_functions WHERE transition_id = $1 ORDER BY position`, transitionID)
	return functions, err
}

func UpdatePostFunction(ctx context.Context, db *sqlx.DB, pf *PostFunction) error {
	_, err := db.ExecContext(ctx,
		`UPDATE workflow_transition_post_functions SET type = $2, config = $3, position = $4 WHERE id = $1`,
		pf.ID, pf.Type, pf.Config, pf.Position)
	return err
}

func DeletePostFunction(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM workflow_transition_post_functions WHERE id = $1`, id)
	return err
}

// Screens

func CreateScreen(ctx context.Context, db *sqlx.DB, s *Screen) error {
	query := `INSERT INTO workflow_screens (name, description) VALUES ($1, $2) RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query, s.Name, s.Description).Scan(&s.ID, &s.CreatedAt)
}

func GetScreen(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Screen, error) {
	s := &Screen{}
	err := db.GetContext(ctx, s, `SELECT * FROM workflow_screens WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrScreenNotFound
	}
	return s, err
}

func ListScreens(ctx context.Context, db *sqlx.DB) ([]*Screen, error) {
	var screens []*Screen
	err := db.SelectContext(ctx, &screens, `SELECT * FROM workflow_screens ORDER BY name`)
	return screens, err
}

func UpdateScreen(ctx context.Context, db *sqlx.DB, s *Screen) error {
	result, err := db.ExecContext(ctx,
		`UPDATE workflow_screens SET name = $2, description = $3 WHERE id = $1`,
		s.ID, s.Name, s.Description)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrScreenNotFound
	}
	return nil
}

func DeleteScreen(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM workflow_screens WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrScreenNotFound
	}
	return nil
}

// Screen Fields

func AddScreenField(ctx context.Context, db *sqlx.DB, sf *ScreenField) error {
	query := `
		INSERT INTO workflow_screen_fields (screen_id, field_type, field_name, field_id, is_required, position)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return db.QueryRowxContext(ctx, query,
		sf.ScreenID, sf.FieldType, sf.FieldName, sf.FieldID, sf.IsRequired, sf.Position).Scan(&sf.ID)
}

func ListScreenFields(ctx context.Context, db *sqlx.DB, screenID uuid.UUID) ([]*ScreenField, error) {
	var fields []*ScreenField
	err := db.SelectContext(ctx, &fields,
		`SELECT * FROM workflow_screen_fields WHERE screen_id = $1 ORDER BY position`, screenID)
	return fields, err
}

func UpdateScreenField(ctx context.Context, db *sqlx.DB, sf *ScreenField) error {
	_, err := db.ExecContext(ctx,
		`UPDATE workflow_screen_fields SET field_type = $2, field_name = $3, field_id = $4, is_required = $5, position = $6 WHERE id = $1`,
		sf.ID, sf.FieldType, sf.FieldName, sf.FieldID, sf.IsRequired, sf.Position)
	return err
}

func DeleteScreenField(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM workflow_screen_fields WHERE id = $1`, id)
	return err
}

// Workflow Schemes

func CreateScheme(ctx context.Context, db *sqlx.DB, s *Scheme) error {
	query := `
		INSERT INTO workflow_schemes (name, description, is_default)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query, s.Name, s.Description, s.IsDefault).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func GetScheme(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Scheme, error) {
	s := &Scheme{}
	err := db.GetContext(ctx, s, `SELECT * FROM workflow_schemes WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSchemeNotFound
	}
	return s, err
}

func GetDefaultScheme(ctx context.Context, db *sqlx.DB) (*Scheme, error) {
	s := &Scheme{}
	err := db.GetContext(ctx, s, `SELECT * FROM workflow_schemes WHERE is_default = true LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSchemeNotFound
	}
	return s, err
}

func ListSchemes(ctx context.Context, db *sqlx.DB) ([]*Scheme, error) {
	var schemes []*Scheme
	err := db.SelectContext(ctx, &schemes, `SELECT * FROM workflow_schemes ORDER BY is_default DESC, name`)
	return schemes, err
}

func UpdateScheme(ctx context.Context, db *sqlx.DB, s *Scheme) error {
	query := `UPDATE workflow_schemes SET name = $2, description = $3, is_default = $4 WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, s.ID, s.Name, s.Description, s.IsDefault).Scan(&s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSchemeNotFound
	}
	return err
}

func DeleteScheme(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM workflow_schemes WHERE id = $1 AND is_default = false`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrSchemeNotFound
	}
	return nil
}

// Scheme Mappings

func SetSchemeMapping(ctx context.Context, db *sqlx.DB, m *SchemeMapping) error {
	query := `
		INSERT INTO workflow_scheme_mappings (scheme_id, issue_type_id, workflow_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (scheme_id, issue_type_id) DO UPDATE SET workflow_id = EXCLUDED.workflow_id`
	_, err := db.ExecContext(ctx, query, m.SchemeID, m.IssueTypeID, m.WorkflowID)
	return err
}

func ListSchemeMappings(ctx context.Context, db *sqlx.DB, schemeID uuid.UUID) ([]*SchemeMapping, error) {
	var mappings []*SchemeMapping
	err := db.SelectContext(ctx, &mappings,
		`SELECT * FROM workflow_scheme_mappings WHERE scheme_id = $1`, schemeID)
	return mappings, err
}

func GetWorkflowForIssueType(ctx context.Context, db *sqlx.DB, schemeID, issueTypeID uuid.UUID) (*Workflow, error) {
	w := &Workflow{}
	query := `
		SELECT w.* FROM workflows w
		INNER JOIN workflow_scheme_mappings m ON w.id = m.workflow_id
		WHERE m.scheme_id = $1 AND (m.issue_type_id = $2 OR m.issue_type_id IS NULL)
		ORDER BY m.issue_type_id NULLS LAST
		LIMIT 1`
	err := db.GetContext(ctx, w, query, schemeID, issueTypeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return w, err
}

func DeleteSchemeMapping(ctx context.Context, db *sqlx.DB, schemeID uuid.UUID, issueTypeID *uuid.UUID) error {
	if issueTypeID == nil {
		_, err := db.ExecContext(ctx,
			`DELETE FROM workflow_scheme_mappings WHERE scheme_id = $1 AND issue_type_id IS NULL`, schemeID)
		return err
	}
	_, err := db.ExecContext(ctx,
		`DELETE FROM workflow_scheme_mappings WHERE scheme_id = $1 AND issue_type_id = $2`, schemeID, issueTypeID)
	return err
}

// Full transition with all configs

type FullTransition struct {
	Transition    *Transition
	Conditions    []*Condition
	Validators    []*Validator
	PostFunctions []*PostFunction
	Screen        *Screen
	ScreenFields  []*ScreenField
}

func GetFullTransition(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*FullTransition, error) {
	t := &Transition{}
	err := db.GetContext(ctx, t, `SELECT * FROM workflow_transitions WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTransitionNotFound
	}
	if err != nil {
		return nil, err
	}

	conditions, _ := ListConditions(ctx, db, id)
	validators, _ := ListValidators(ctx, db, id)
	postFunctions, _ := ListPostFunctions(ctx, db, id)

	var screen *Screen
	var screenFields []*ScreenField
	if t.ScreenID != nil {
		screen, _ = GetScreen(ctx, db, *t.ScreenID)
		if screen != nil {
			screenFields, _ = ListScreenFields(ctx, db, screen.ID)
		}
	}

	return &FullTransition{
		Transition:    t,
		Conditions:    conditions,
		Validators:    validators,
		PostFunctions: postFunctions,
		Screen:        screen,
		ScreenFields:  screenFields,
	}, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
