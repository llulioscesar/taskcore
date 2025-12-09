package issue

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("issue not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelNameExists    = errors.New("label name already exists")
	ErrTypeNotFound       = errors.New("issue type not found")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrWorkLogNotFound    = errors.New("work log not found")
)

type Priority string

const (
	PriorityHighest Priority = "highest"
	PriorityHigh    Priority = "high"
	PriorityMedium  Priority = "medium"
	PriorityLow     Priority = "low"
	PriorityLowest  Priority = "lowest"
)

type Issue struct {
	ID          uuid.UUID  `db:"id"`
	ProjectID   uuid.UUID  `db:"project_id"`
	IssueTypeID uuid.UUID  `db:"issue_type_id"`
	StatusID    uuid.UUID  `db:"status_id"`
	SprintID    *uuid.UUID `db:"sprint_id"`
	EpicID      *uuid.UUID `db:"epic_id"`
	ParentID    *uuid.UUID `db:"parent_id"`
	Key         string     `db:"key"`
	Summary     string     `db:"summary"`
	Description *string    `db:"description"`
	Priority    Priority   `db:"priority"`
	ReporterID  uuid.UUID  `db:"reporter_id"`
	AssigneeID  *uuid.UUID `db:"assignee_id"`
	StoryPoints *int       `db:"story_points"`
	DueDate     *time.Time `db:"due_date"`
	Color       *string    `db:"color"`
	StartDate   *time.Time `db:"start_date"`
	TargetDate  *time.Time `db:"target_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `db:"id"`
	IssueID   uuid.UUID `db:"issue_id"`
	AuthorID  uuid.UUID `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RelationType string

const (
	RelationBlocks     RelationType = "blocks"
	RelationDuplicates RelationType = "duplicates"
	RelationRelatesTo  RelationType = "relates_to"
)

type Relation struct {
	ID            uuid.UUID    `db:"id"`
	SourceIssueID uuid.UUID    `db:"source_issue_id"`
	TargetIssueID uuid.UUID    `db:"target_issue_id"`
	RelationType  RelationType `db:"relation_type"`
	CreatedAt     time.Time    `db:"created_at"`
}

type Label struct {
	ID        uuid.UUID `db:"id"`
	ProjectID uuid.UUID `db:"project_id"`
	Name      string    `db:"name"`
	Color     string    `db:"color"`
	CreatedAt time.Time `db:"created_at"`
}

type HierarchyLevel int

const (
	HierarchyEpic    HierarchyLevel = 0
	HierarchyStandard HierarchyLevel = 1
	HierarchySubtask HierarchyLevel = 2
)

type Type struct {
	ID             uuid.UUID      `db:"id"`
	Name           string         `db:"name"`
	Description    *string        `db:"description"`
	Icon           string         `db:"icon"`
	Color          string         `db:"color"`
	HierarchyLevel HierarchyLevel `db:"hierarchy_level"`
	IsSubtask      bool           `db:"is_subtask"`
}

type Watcher struct {
	IssueID   uuid.UUID `db:"issue_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Voter struct {
	IssueID   uuid.UUID `db:"issue_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Attachment struct {
	ID        uuid.UUID `db:"id"`
	IssueID   uuid.UUID `db:"issue_id"`
	UserID    uuid.UUID `db:"user_id"`
	Filename  string    `db:"filename"`
	Path      string    `db:"path"`
	Size      int64     `db:"size"`
	MimeType  string    `db:"mime_type"`
	CreatedAt time.Time `db:"created_at"`
}

type WorkLog struct {
	ID          uuid.UUID `db:"id"`
	IssueID     uuid.UUID `db:"issue_id"`
	UserID      uuid.UUID `db:"user_id"`
	TimeSpent   int       `db:"time_spent"`
	Description *string   `db:"description"`
	LoggedAt    time.Time `db:"logged_at"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
