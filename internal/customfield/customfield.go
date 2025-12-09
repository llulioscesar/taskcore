package customfield

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("custom field not found")
	ErrNameExists     = errors.New("custom field name already exists")
	ErrValueNotFound  = errors.New("custom field value not found")
	ErrOptionNotFound = errors.New("custom field option not found")
)

type FieldType string

const (
	TypeText     FieldType = "text"
	TypeNumber   FieldType = "number"
	TypeDate     FieldType = "date"
	TypeDateTime FieldType = "datetime"
	TypeSelect   FieldType = "select"
	TypeMulti    FieldType = "multi_select"
	TypeUser     FieldType = "user"
	TypeURL      FieldType = "url"
	TypeCheckbox FieldType = "checkbox"
)

type Field struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	FieldType   FieldType `db:"field_type"`
	IsRequired  bool      `db:"is_required"`
	IsGlobal    bool      `db:"is_global"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Option struct {
	ID       uuid.UUID `db:"id"`
	FieldID  uuid.UUID `db:"field_id"`
	Value    string    `db:"value"`
	Color    *string   `db:"color"`
	Position int       `db:"position"`
}

type ProjectField struct {
	ProjectID  uuid.UUID `db:"project_id"`
	FieldID    uuid.UUID `db:"field_id"`
	IsRequired bool      `db:"is_required"`
	CreatedAt  time.Time `db:"created_at"`
}

type Value struct {
	ID        uuid.UUID  `db:"id"`
	IssueID   uuid.UUID  `db:"issue_id"`
	FieldID   uuid.UUID  `db:"field_id"`
	TextValue *string    `db:"text_value"`
	NumValue  *float64   `db:"num_value"`
	DateValue *time.Time `db:"date_value"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type ValueOption struct {
	ValueID  uuid.UUID `db:"value_id"`
	OptionID uuid.UUID `db:"option_id"`
}
