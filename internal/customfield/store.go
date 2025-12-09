package customfield

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Field CRUD

func CreateField(ctx context.Context, db *sqlx.DB, f *Field) error {
	query := `
		INSERT INTO custom_fields (name, description, field_type, is_required, is_global)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	err := db.QueryRowxContext(ctx, query,
		f.Name, f.Description, f.FieldType, f.IsRequired, f.IsGlobal,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetField(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Field, error) {
	f := &Field{}
	err := db.GetContext(ctx, f, `SELECT * FROM custom_fields WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return f, err
}

func ListFields(ctx context.Context, db *sqlx.DB) ([]*Field, error) {
	var fields []*Field
	err := db.SelectContext(ctx, &fields, `SELECT * FROM custom_fields ORDER BY name`)
	return fields, err
}

func ListGlobalFields(ctx context.Context, db *sqlx.DB) ([]*Field, error) {
	var fields []*Field
	err := db.SelectContext(ctx, &fields,
		`SELECT * FROM custom_fields WHERE is_global = true ORDER BY name`)
	return fields, err
}

func UpdateField(ctx context.Context, db *sqlx.DB, f *Field) error {
	query := `
		UPDATE custom_fields SET name = $2, description = $3, is_required = $4
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query,
		f.ID, f.Name, f.Description, f.IsRequired,
	).Scan(&f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func DeleteField(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM custom_fields WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Options

func CreateOption(ctx context.Context, db *sqlx.DB, o *Option) error {
	query := `
		INSERT INTO custom_field_options (field_id, value, color, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	return db.QueryRowxContext(ctx, query, o.FieldID, o.Value, o.Color, o.Position).Scan(&o.ID)
}

func ListOptions(ctx context.Context, db *sqlx.DB, fieldID uuid.UUID) ([]*Option, error) {
	var options []*Option
	err := db.SelectContext(ctx, &options,
		`SELECT * FROM custom_field_options WHERE field_id = $1 ORDER BY position`, fieldID)
	return options, err
}

func UpdateOption(ctx context.Context, db *sqlx.DB, o *Option) error {
	result, err := db.ExecContext(ctx,
		`UPDATE custom_field_options SET value = $2, color = $3, position = $4 WHERE id = $1`,
		o.ID, o.Value, o.Color, o.Position)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOptionNotFound
	}
	return nil
}

func DeleteOption(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM custom_field_options WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOptionNotFound
	}
	return nil
}

// Project Fields

func AddFieldToProject(ctx context.Context, db *sqlx.DB, projectID, fieldID uuid.UUID, isRequired bool) error {
	query := `
		INSERT INTO project_custom_fields (project_id, field_id, is_required)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, field_id) DO UPDATE SET is_required = EXCLUDED.is_required`
	_, err := db.ExecContext(ctx, query, projectID, fieldID, isRequired)
	return err
}

func RemoveFieldFromProject(ctx context.Context, db *sqlx.DB, projectID, fieldID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_custom_fields WHERE project_id = $1 AND field_id = $2`,
		projectID, fieldID)
	return err
}

func ListProjectFields(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Field, error) {
	var fields []*Field
	query := `
		SELECT cf.* FROM custom_fields cf
		INNER JOIN project_custom_fields pcf ON cf.id = pcf.field_id
		WHERE pcf.project_id = $1 OR cf.is_global = true
		ORDER BY cf.name`
	err := db.SelectContext(ctx, &fields, query, projectID)
	return fields, err
}

// Values

func SetValue(ctx context.Context, db *sqlx.DB, v *Value) error {
	query := `
		INSERT INTO custom_field_values (issue_id, field_id, text_value, num_value, date_value)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (issue_id, field_id) DO UPDATE
		SET text_value = EXCLUDED.text_value, num_value = EXCLUDED.num_value,
		    date_value = EXCLUDED.date_value, updated_at = NOW()
		RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query,
		v.IssueID, v.FieldID, v.TextValue, v.NumValue, v.DateValue,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func GetValue(ctx context.Context, db *sqlx.DB, issueID, fieldID uuid.UUID) (*Value, error) {
	v := &Value{}
	err := db.GetContext(ctx, v,
		`SELECT * FROM custom_field_values WHERE issue_id = $1 AND field_id = $2`,
		issueID, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrValueNotFound
	}
	return v, err
}

func ListIssueValues(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Value, error) {
	var values []*Value
	err := db.SelectContext(ctx, &values,
		`SELECT * FROM custom_field_values WHERE issue_id = $1`, issueID)
	return values, err
}

func DeleteValue(ctx context.Context, db *sqlx.DB, issueID, fieldID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM custom_field_values WHERE issue_id = $1 AND field_id = $2`,
		issueID, fieldID)
	return err
}

// Multi-select options

func SetValueOptions(ctx context.Context, db *sqlx.DB, valueID uuid.UUID, optionIDs []uuid.UUID) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM custom_field_value_options WHERE value_id = $1`, valueID)
	if err != nil {
		return err
	}

	if len(optionIDs) > 0 {
		query := `INSERT INTO custom_field_value_options (value_id, option_id) VALUES ($1, unnest($2::uuid[]))`
		_, err = tx.ExecContext(ctx, query, valueID, optionIDs)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ListValueOptions(ctx context.Context, db *sqlx.DB, valueID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT option_id FROM custom_field_value_options WHERE value_id = $1`, valueID)
	return ids, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
