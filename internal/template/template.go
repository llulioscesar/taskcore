package template

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("template not found")
	ErrNameExists = errors.New("template name already exists")
)

type Template struct {
	ID          uuid.UUID  `db:"id"`
	ProjectID   *uuid.UUID `db:"project_id"`
	IssueTypeID uuid.UUID  `db:"issue_type_id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	Summary     string     `db:"summary"`
	Content     *string    `db:"content"`
	Priority    *string    `db:"priority"`
	Labels      []string   `db:"labels"`
	IsDefault   bool       `db:"is_default"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type FieldValue struct {
	TemplateID uuid.UUID `db:"template_id"`
	FieldID    uuid.UUID `db:"field_id"`
	TextValue  *string   `db:"text_value"`
	NumValue   *float64  `db:"num_value"`
}
