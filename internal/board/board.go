package board

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("board not found")
	ErrColumnNotFound = errors.New("board column not found")
)

type BoardType string

const (
	TypeKanban BoardType = "kanban"
	TypeScrum  BoardType = "scrum"
)

type Board struct {
	ID        uuid.UUID `db:"id"`
	ProjectID uuid.UUID `db:"project_id"`
	Name      string    `db:"name"`
	Type      BoardType `db:"type"`
	FilterJQL *string   `db:"filter_jql"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Column struct {
	ID       uuid.UUID  `db:"id"`
	BoardID  uuid.UUID  `db:"board_id"`
	Name     string     `db:"name"`
	StatusID *uuid.UUID `db:"status_id"`
	Position int        `db:"position"`
	MinLimit *int       `db:"min_limit"`
	MaxLimit *int       `db:"max_limit"`
}

type Swimlane string

const (
	SwimlaneNone      Swimlane = "none"
	SwimlaneAssignee  Swimlane = "assignee"
	SwimlanePriority  Swimlane = "priority"
	SwimlaneIssueType Swimlane = "issue_type"
	SwimlaneEpic      Swimlane = "epic"
)

type Config struct {
	BoardID       uuid.UUID `db:"board_id"`
	Swimlane      Swimlane  `db:"swimlane"`
	CardFields    []string  `db:"card_fields"`
	ShowDaysInCol bool      `db:"show_days_in_col"`
	ShowEpicAsBar bool      `db:"show_epic_as_bar"`
}
