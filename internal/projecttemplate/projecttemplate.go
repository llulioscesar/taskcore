package projecttemplate

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("project template not found")
	ErrKeyExists  = errors.New("project template key already exists")
)

type Category string

const (
	CategorySoftware   Category = "software"
	CategoryBusiness   Category = "business"
	CategoryMarketing  Category = "marketing"
	CategoryHR         Category = "hr"
	CategoryOperations Category = "operations"
)

type BoardType string

const (
	BoardKanban BoardType = "kanban"
	BoardScrum  BoardType = "scrum"
)

type Template struct {
	ID          uuid.UUID `db:"id"`
	Key         string    `db:"key"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Category    Category  `db:"category"`
	BoardType   BoardType `db:"board_type"`
	Icon        string    `db:"icon"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// WorkflowTemplate defines the workflow for a project template
type WorkflowTemplate struct {
	TemplateID  uuid.UUID `db:"template_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
}

// StatusTemplate defines statuses for the workflow
type StatusTemplate struct {
	TemplateID uuid.UUID `db:"template_id"`
	Name       string    `db:"name"`
	Category   string    `db:"category"` // todo, in_progress, done
	Position   int       `db:"position"`
}

// TransitionTemplate defines transitions between statuses
type TransitionTemplate struct {
	TemplateID uuid.UUID `db:"template_id"`
	FromStatus string    `db:"from_status"`
	ToStatus   string    `db:"to_status"`
	Name       string    `db:"name"`
}

// IssueTypeTemplate defines which issue types are enabled
type IssueTypeTemplate struct {
	TemplateID  uuid.UUID `db:"template_id"`
	IssueTypeID uuid.UUID `db:"issue_type_id"`
	IsDefault   bool      `db:"is_default"`
}

// ColumnTemplate defines board columns
type ColumnTemplate struct {
	TemplateID uuid.UUID `db:"template_id"`
	Name       string    `db:"name"`
	StatusName string    `db:"status_name"` // maps to StatusTemplate.Name
	Position   int       `db:"position"`
	MinLimit   *int      `db:"min_limit"`
	MaxLimit   *int      `db:"max_limit"`
}

// Default templates
var (
	ScrumTemplate = Template{
		Key:         "scrum",
		Name:        "Scrum",
		Description: ptr("Scrum board for agile teams with sprints, backlog, and velocity tracking"),
		Category:    CategorySoftware,
		BoardType:   BoardScrum,
		Icon:        "sprint",
		IsDefault:   true,
	}

	KanbanTemplate = Template{
		Key:         "kanban",
		Name:        "Kanban",
		Description: ptr("Kanban board for continuous flow with WIP limits"),
		Category:    CategorySoftware,
		BoardType:   BoardKanban,
		Icon:        "board",
		IsDefault:   true,
	}

	BugTrackingTemplate = Template{
		Key:         "bug-tracking",
		Name:        "Bug Tracking",
		Description: ptr("Track and manage bugs with priority-based workflow"),
		Category:    CategorySoftware,
		BoardType:   BoardKanban,
		Icon:        "bug",
		IsDefault:   true,
	}
)

// Default statuses for Scrum
var ScrumStatuses = []StatusTemplate{
	{Name: "To Do", Category: "todo", Position: 0},
	{Name: "In Progress", Category: "in_progress", Position: 1},
	{Name: "In Review", Category: "in_progress", Position: 2},
	{Name: "Done", Category: "done", Position: 3},
}

// Default statuses for Kanban
var KanbanStatuses = []StatusTemplate{
	{Name: "Backlog", Category: "todo", Position: 0},
	{Name: "Selected for Development", Category: "todo", Position: 1},
	{Name: "In Progress", Category: "in_progress", Position: 2},
	{Name: "Done", Category: "done", Position: 3},
}

// Default statuses for Bug Tracking
var BugTrackingStatuses = []StatusTemplate{
	{Name: "Open", Category: "todo", Position: 0},
	{Name: "In Progress", Category: "in_progress", Position: 1},
	{Name: "Resolved", Category: "done", Position: 2},
	{Name: "Closed", Category: "done", Position: 3},
	{Name: "Reopened", Category: "todo", Position: 4},
}

func ptr(s string) *string {
	return &s
}
