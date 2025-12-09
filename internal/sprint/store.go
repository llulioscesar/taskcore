package sprint

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, sp *Sprint) error {
	query := `
		INSERT INTO sprints (project_id, name, goal, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return db.QueryRowxContext(ctx, query,
		sp.ProjectID, sp.Name, sp.Goal, sp.StartDate, sp.EndDate, sp.Status,
	).Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Sprint, error) {
	sp := &Sprint{}
	err := db.GetContext(ctx, sp, `SELECT * FROM sprints WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return sp, err
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	query := `
		SELECT * FROM sprints WHERE project_id = $1
		ORDER BY CASE status WHEN 'active' THEN 1 WHEN 'planning' THEN 2 ELSE 3 END, start_date DESC NULLS LAST`
	err := db.SelectContext(ctx, &sprints, query, projectID)
	return sprints, err
}

func GetActive(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) (*Sprint, error) {
	sp := &Sprint{}
	err := db.GetContext(ctx, sp,
		`SELECT * FROM sprints WHERE project_id = $1 AND status = 'active' LIMIT 1`, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return sp, err
}

func Update(ctx context.Context, db *sqlx.DB, sp *Sprint) error {
	query := `
		UPDATE sprints SET name = $2, goal = $3, start_date = $4, end_date = $5, status = $6
		WHERE id = $1 RETURNING updated_at`

	err := db.QueryRowxContext(ctx, query, sp.ID, sp.Name, sp.Goal, sp.StartDate, sp.EndDate, sp.Status).
		Scan(&sp.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func Start(ctx context.Context, db *sqlx.DB, id uuid.UUID, startDate, endDate time.Time) error {
	result, err := db.ExecContext(ctx,
		`UPDATE sprints SET status = 'active', start_date = $2, end_date = $3 WHERE id = $1 AND status = 'planning'`,
		id, startDate, endDate)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func Close(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`UPDATE sprints SET status = 'closed', end_date = NOW() WHERE id = $1 AND status = 'active'`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM sprints WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func GetIssueCount(ctx context.Context, db *sqlx.DB, sprintID uuid.UUID) (int, error) {
	var count int
	err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM issues WHERE sprint_id = $1`, sprintID)
	return count, err
}

func GetCompletedIssueCount(ctx context.Context, db *sqlx.DB, sprintID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM issues i
		INNER JOIN workflow_statuses ws ON i.status_id = ws.id
		WHERE i.sprint_id = $1 AND ws.category = 'done'`
	err := db.GetContext(ctx, &count, query, sprintID)
	return count, err
}

func GetTotalStoryPoints(ctx context.Context, db *sqlx.DB, sprintID uuid.UUID) (int, error) {
	var total int
	err := db.GetContext(ctx, &total, `SELECT COALESCE(SUM(story_points), 0) FROM issues WHERE sprint_id = $1`, sprintID)
	return total, err
}
