package sprint

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("sprint not found")

type Status string

const (
	StatusPlanning Status = "planning"
	StatusActive   Status = "active"
	StatusClosed   Status = "closed"
)

type Sprint struct {
	ID        uuid.UUID  `db:"id"`
	ProjectID uuid.UUID  `db:"project_id"`
	Name      string     `db:"name"`
	Goal      *string    `db:"goal"`
	StartDate *time.Time `db:"start_date"`
	EndDate   *time.Time `db:"end_date"`
	Status    Status     `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}
