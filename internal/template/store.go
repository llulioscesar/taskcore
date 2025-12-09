package template

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func Create(ctx context.Context, db *sqlx.DB, t *Template) error {
	query := `
		INSERT INTO issue_templates (project_id, issue_type_id, name, description, summary, content, priority, labels, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`
	err := db.QueryRowxContext(ctx, query,
		t.ProjectID, t.IssueTypeID, t.Name, t.Description, t.Summary,
		t.Content, t.Priority, pq.Array(t.Labels), t.IsDefault,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Template, error) {
	t := &Template{}
	err := db.QueryRowxContext(ctx,
		`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at FROM issue_templates WHERE id = $1`, id).Scan(
		&t.ID, &t.ProjectID, &t.IssueTypeID, &t.Name, &t.Description, &t.Summary,
		&t.Content, &t.Priority, pq.Array(&t.Labels), &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func ListGlobal(ctx context.Context, db *sqlx.DB) ([]*Template, error) {
	rows, err := db.QueryxContext(ctx,
		`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at FROM issue_templates WHERE project_id IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTemplates(rows)
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Template, error) {
	rows, err := db.QueryxContext(ctx,
		`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at FROM issue_templates WHERE project_id = $1 OR project_id IS NULL ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTemplates(rows)
}

func ListByIssueType(ctx context.Context, db *sqlx.DB, issueTypeID uuid.UUID) ([]*Template, error) {
	rows, err := db.QueryxContext(ctx,
		`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at FROM issue_templates WHERE issue_type_id = $1 ORDER BY name`, issueTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTemplates(rows)
}

func GetDefault(ctx context.Context, db *sqlx.DB, projectID *uuid.UUID, issueTypeID uuid.UUID) (*Template, error) {
	t := &Template{}
	var err error
	if projectID != nil {
		err = db.QueryRowxContext(ctx,
			`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at
			 FROM issue_templates WHERE (project_id = $1 OR project_id IS NULL) AND issue_type_id = $2 AND is_default = true
			 ORDER BY project_id NULLS LAST LIMIT 1`, projectID, issueTypeID).Scan(
			&t.ID, &t.ProjectID, &t.IssueTypeID, &t.Name, &t.Description, &t.Summary,
			&t.Content, &t.Priority, pq.Array(&t.Labels), &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
	} else {
		err = db.QueryRowxContext(ctx,
			`SELECT id, project_id, issue_type_id, name, description, summary, content, priority, labels, is_default, created_at, updated_at
			 FROM issue_templates WHERE project_id IS NULL AND issue_type_id = $1 AND is_default = true LIMIT 1`, issueTypeID).Scan(
			&t.ID, &t.ProjectID, &t.IssueTypeID, &t.Name, &t.Description, &t.Summary,
			&t.Content, &t.Priority, pq.Array(&t.Labels), &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func Update(ctx context.Context, db *sqlx.DB, t *Template) error {
	query := `
		UPDATE issue_templates SET name = $2, description = $3, summary = $4, content = $5,
		       priority = $6, labels = $7, is_default = $8
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query,
		t.ID, t.Name, t.Description, t.Summary, t.Content,
		t.Priority, pq.Array(t.Labels), t.IsDefault,
	).Scan(&t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM issue_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func SetDefault(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var t Template
	err = tx.QueryRowxContext(ctx,
		`SELECT project_id, issue_type_id FROM issue_templates WHERE id = $1`, id).Scan(&t.ProjectID, &t.IssueTypeID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if t.ProjectID != nil {
		_, err = tx.ExecContext(ctx,
			`UPDATE issue_templates SET is_default = false WHERE project_id = $1 AND issue_type_id = $2`,
			t.ProjectID, t.IssueTypeID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE issue_templates SET is_default = false WHERE project_id IS NULL AND issue_type_id = $1`,
			t.IssueTypeID)
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE issue_templates SET is_default = true WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Field Values

func SetFieldValue(ctx context.Context, db *sqlx.DB, fv *FieldValue) error {
	query := `
		INSERT INTO issue_template_fields (template_id, field_id, text_value, num_value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (template_id, field_id) DO UPDATE
		SET text_value = EXCLUDED.text_value, num_value = EXCLUDED.num_value`
	_, err := db.ExecContext(ctx, query, fv.TemplateID, fv.FieldID, fv.TextValue, fv.NumValue)
	return err
}

func ListFieldValues(ctx context.Context, db *sqlx.DB, templateID uuid.UUID) ([]*FieldValue, error) {
	var values []*FieldValue
	err := db.SelectContext(ctx, &values,
		`SELECT * FROM issue_template_fields WHERE template_id = $1`, templateID)
	return values, err
}

func DeleteFieldValue(ctx context.Context, db *sqlx.DB, templateID, fieldID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM issue_template_fields WHERE template_id = $1 AND field_id = $2`,
		templateID, fieldID)
	return err
}

func scanTemplates(rows *sqlx.Rows) ([]*Template, error) {
	var templates []*Template
	for rows.Next() {
		t := &Template{}
		err := rows.Scan(&t.ID, &t.ProjectID, &t.IssueTypeID, &t.Name, &t.Description,
			&t.Summary, &t.Content, &t.Priority, pq.Array(&t.Labels), &t.IsDefault,
			&t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
