package projecttemplate

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Template CRUD

func Create(ctx context.Context, db *sqlx.DB, t *Template) error {
	query := `
		INSERT INTO project_templates (key, name, description, category, board_type, icon, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	err := db.QueryRowxContext(ctx, query,
		t.Key, t.Name, t.Description, t.Category, t.BoardType, t.Icon, t.IsDefault,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrKeyExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Template, error) {
	t := &Template{}
	err := db.GetContext(ctx, t, `SELECT * FROM project_templates WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func GetByKey(ctx context.Context, db *sqlx.DB, key string) (*Template, error) {
	t := &Template{}
	err := db.GetContext(ctx, t, `SELECT * FROM project_templates WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Template, error) {
	var templates []*Template
	err := db.SelectContext(ctx, &templates,
		`SELECT * FROM project_templates ORDER BY category, name`)
	return templates, err
}

func ListByCategory(ctx context.Context, db *sqlx.DB, category Category) ([]*Template, error) {
	var templates []*Template
	err := db.SelectContext(ctx, &templates,
		`SELECT * FROM project_templates WHERE category = $1 ORDER BY name`, category)
	return templates, err
}

func ListDefault(ctx context.Context, db *sqlx.DB) ([]*Template, error) {
	var templates []*Template
	err := db.SelectContext(ctx, &templates,
		`SELECT * FROM project_templates WHERE is_default = true ORDER BY category, name`)
	return templates, err
}

func Update(ctx context.Context, db *sqlx.DB, t *Template) error {
	query := `
		UPDATE project_templates SET name = $2, description = $3, category = $4,
		       board_type = $5, icon = $6
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query,
		t.ID, t.Name, t.Description, t.Category, t.BoardType, t.Icon,
	).Scan(&t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM project_templates WHERE id = $1 AND is_default = false`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Status Templates

func AddStatus(ctx context.Context, db *sqlx.DB, s *StatusTemplate) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO project_template_statuses (template_id, name, category, position)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (template_id, name) DO UPDATE SET category = EXCLUDED.category, position = EXCLUDED.position`,
		s.TemplateID, s.Name, s.Category, s.Position)
	return err
}

func ListStatuses(ctx context.Context, db *sqlx.DB, templateID uuid.UUID) ([]*StatusTemplate, error) {
	var statuses []*StatusTemplate
	err := db.SelectContext(ctx, &statuses,
		`SELECT * FROM project_template_statuses WHERE template_id = $1 ORDER BY position`,
		templateID)
	return statuses, err
}

func RemoveStatus(ctx context.Context, db *sqlx.DB, templateID uuid.UUID, name string) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_template_statuses WHERE template_id = $1 AND name = $2`,
		templateID, name)
	return err
}

// Transition Templates

func AddTransition(ctx context.Context, db *sqlx.DB, t *TransitionTemplate) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO project_template_transitions (template_id, from_status, to_status, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (template_id, from_status, to_status) DO UPDATE SET name = EXCLUDED.name`,
		t.TemplateID, t.FromStatus, t.ToStatus, t.Name)
	return err
}

func ListTransitions(ctx context.Context, db *sqlx.DB, templateID uuid.UUID) ([]*TransitionTemplate, error) {
	var transitions []*TransitionTemplate
	err := db.SelectContext(ctx, &transitions,
		`SELECT * FROM project_template_transitions WHERE template_id = $1`,
		templateID)
	return transitions, err
}

// Issue Type Templates

func AddIssueType(ctx context.Context, db *sqlx.DB, it *IssueTypeTemplate) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO project_template_issue_types (template_id, issue_type_id, is_default)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (template_id, issue_type_id) DO UPDATE SET is_default = EXCLUDED.is_default`,
		it.TemplateID, it.IssueTypeID, it.IsDefault)
	return err
}

func ListIssueTypes(ctx context.Context, db *sqlx.DB, templateID uuid.UUID) ([]*IssueTypeTemplate, error) {
	var types []*IssueTypeTemplate
	err := db.SelectContext(ctx, &types,
		`SELECT * FROM project_template_issue_types WHERE template_id = $1`,
		templateID)
	return types, err
}

func RemoveIssueType(ctx context.Context, db *sqlx.DB, templateID, issueTypeID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_template_issue_types WHERE template_id = $1 AND issue_type_id = $2`,
		templateID, issueTypeID)
	return err
}

// Column Templates

func AddColumn(ctx context.Context, db *sqlx.DB, c *ColumnTemplate) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO project_template_columns (template_id, name, status_name, position, min_limit, max_limit)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (template_id, position) DO UPDATE
		 SET name = EXCLUDED.name, status_name = EXCLUDED.status_name,
		     min_limit = EXCLUDED.min_limit, max_limit = EXCLUDED.max_limit`,
		c.TemplateID, c.Name, c.StatusName, c.Position, c.MinLimit, c.MaxLimit)
	return err
}

func ListColumns(ctx context.Context, db *sqlx.DB, templateID uuid.UUID) ([]*ColumnTemplate, error) {
	var columns []*ColumnTemplate
	err := db.SelectContext(ctx, &columns,
		`SELECT * FROM project_template_columns WHERE template_id = $1 ORDER BY position`,
		templateID)
	return columns, err
}

func RemoveColumn(ctx context.Context, db *sqlx.DB, templateID uuid.UUID, position int) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_template_columns WHERE template_id = $1 AND position = $2`,
		templateID, position)
	return err
}

// Full template with all configs

type FullTemplate struct {
	Template    *Template
	Statuses    []*StatusTemplate
	Transitions []*TransitionTemplate
	IssueTypes  []*IssueTypeTemplate
	Columns     []*ColumnTemplate
}

func GetFull(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*FullTemplate, error) {
	t, err := GetByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	statuses, err := ListStatuses(ctx, db, id)
	if err != nil {
		return nil, err
	}

	transitions, err := ListTransitions(ctx, db, id)
	if err != nil {
		return nil, err
	}

	issueTypes, err := ListIssueTypes(ctx, db, id)
	if err != nil {
		return nil, err
	}

	columns, err := ListColumns(ctx, db, id)
	if err != nil {
		return nil, err
	}

	return &FullTemplate{
		Template:    t,
		Statuses:    statuses,
		Transitions: transitions,
		IssueTypes:  issueTypes,
		Columns:     columns,
	}, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
