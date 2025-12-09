package workflow

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("workflow not found")
	ErrNameExists         = errors.New("workflow name already exists")
	ErrStatusNotFound     = errors.New("status not found")
	ErrTransitionNotFound = errors.New("transition not found")
	ErrSchemeNotFound     = errors.New("workflow scheme not found")
	ErrScreenNotFound     = errors.New("screen not found")
	ErrConditionFailed    = errors.New("transition condition not met")
	ErrValidationFailed   = errors.New("transition validation failed")
)

type Workflow struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type StatusCategory string

const (
	CategoryTodo       StatusCategory = "todo"
	CategoryInProgress StatusCategory = "in_progress"
	CategoryDone       StatusCategory = "done"
)

type Status struct {
	ID         uuid.UUID      `db:"id"`
	WorkflowID uuid.UUID      `db:"workflow_id"`
	Name       string         `db:"name"`
	Category   StatusCategory `db:"category"`
	Position   int            `db:"position"`
}

type Transition struct {
	ID           uuid.UUID  `db:"id"`
	WorkflowID   uuid.UUID  `db:"workflow_id"`
	FromStatusID uuid.UUID  `db:"from_status_id"`
	ToStatusID   uuid.UUID  `db:"to_status_id"`
	Name         string     `db:"name"`
	ScreenID     *uuid.UUID `db:"screen_id"`
}

// Condition types for transitions
type ConditionType string

const (
	ConditionUserInRole       ConditionType = "user_in_role"
	ConditionUserInGroup      ConditionType = "user_in_group"
	ConditionUserHasPermission ConditionType = "user_has_permission"
	ConditionOnlyReporter     ConditionType = "only_reporter"
	ConditionOnlyAssignee     ConditionType = "only_assignee"
	ConditionSubtasksResolved ConditionType = "subtasks_resolved"
	ConditionFieldValue       ConditionType = "field_value"
)

type Condition struct {
	ID           uuid.UUID     `db:"id"`
	TransitionID uuid.UUID     `db:"transition_id"`
	Type         ConditionType `db:"type"`
	Config       string        `db:"config"` // JSON config for the condition
	Position     int           `db:"position"`
}

// Validator types for transitions
type ValidatorType string

const (
	ValidatorFieldRequired    ValidatorType = "field_required"
	ValidatorFieldNotEmpty    ValidatorType = "field_not_empty"
	ValidatorRegex            ValidatorType = "regex"
	ValidatorPreviousStatus   ValidatorType = "previous_status"
	ValidatorParentStatus     ValidatorType = "parent_status"
	ValidatorPermission       ValidatorType = "permission"
	ValidatorCommentRequired  ValidatorType = "comment_required"
	ValidatorResolutionSet    ValidatorType = "resolution_set"
)

type Validator struct {
	ID           uuid.UUID     `db:"id"`
	TransitionID uuid.UUID     `db:"transition_id"`
	Type         ValidatorType `db:"type"`
	Config       string        `db:"config"` // JSON config for the validator
	ErrorMessage *string       `db:"error_message"`
	Position     int           `db:"position"`
}

// Post-function types for transitions
type PostFunctionType string

const (
	PostFuncUpdateField       PostFunctionType = "update_field"
	PostFuncAssignToReporter  PostFunctionType = "assign_to_reporter"
	PostFuncAssignToLead      PostFunctionType = "assign_to_lead"
	PostFuncClearField        PostFunctionType = "clear_field"
	PostFuncCopyFieldValue    PostFunctionType = "copy_field_value"
	PostFuncAddComment        PostFunctionType = "add_comment"
	PostFuncSendNotification  PostFunctionType = "send_notification"
	PostFuncTriggerWebhook    PostFunctionType = "trigger_webhook"
	PostFuncUpdateParent      PostFunctionType = "update_parent"
)

type PostFunction struct {
	ID           uuid.UUID        `db:"id"`
	TransitionID uuid.UUID        `db:"transition_id"`
	Type         PostFunctionType `db:"type"`
	Config       string           `db:"config"` // JSON config for the function
	Position     int              `db:"position"`
}

// Screen for transition (which fields to show/require)
type Screen struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type ScreenFieldType string

const (
	ScreenFieldStandard ScreenFieldType = "standard"
	ScreenFieldCustom   ScreenFieldType = "custom"
)

type ScreenField struct {
	ID         uuid.UUID       `db:"id"`
	ScreenID   uuid.UUID       `db:"screen_id"`
	FieldType  ScreenFieldType `db:"field_type"`
	FieldName  string          `db:"field_name"`
	FieldID    *uuid.UUID      `db:"field_id"` // For custom fields
	IsRequired bool            `db:"is_required"`
	Position   int             `db:"position"`
}

// Workflow Scheme - maps issue types to workflows
type Scheme struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type SchemeMapping struct {
	SchemeID    uuid.UUID  `db:"scheme_id"`
	IssueTypeID *uuid.UUID `db:"issue_type_id"` // NULL = default workflow for scheme
	WorkflowID  uuid.UUID  `db:"workflow_id"`
}
