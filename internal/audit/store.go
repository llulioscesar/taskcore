package audit

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Log(ctx context.Context, db *sqlx.DB, e *Entry) error {
	query := `
		INSERT INTO audit_log (user_id, action, resource_type, resource_id, resource_name,
			project_id, old_value, new_value, ip_address, user_agent, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query,
		e.UserID, e.Action, e.ResourceType, e.ResourceID, e.ResourceName,
		e.ProjectID, e.OldValue, e.NewValue, e.IPAddress, e.UserAgent, e.Details).
		Scan(&e.ID, &e.CreatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Entry, error) {
	e := &Entry{}
	err := db.GetContext(ctx, e, `SELECT * FROM audit_log WHERE id = $1`, id)
	return e, err
}

func Search(ctx context.Context, db *sqlx.DB, p SearchParams) ([]*Entry, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if p.UserID != nil {
		conditions = append(conditions, "user_id = $"+string(rune('0'+argNum)))
		args = append(args, *p.UserID)
		argNum++
	}
	if p.Action != nil {
		conditions = append(conditions, "action = $"+string(rune('0'+argNum)))
		args = append(args, *p.Action)
		argNum++
	}
	if p.ResourceType != nil {
		conditions = append(conditions, "resource_type = $"+string(rune('0'+argNum)))
		args = append(args, *p.ResourceType)
		argNum++
	}
	if p.ResourceID != nil {
		conditions = append(conditions, "resource_id = $"+string(rune('0'+argNum)))
		args = append(args, *p.ResourceID)
		argNum++
	}
	if p.ProjectID != nil {
		conditions = append(conditions, "project_id = $"+string(rune('0'+argNum)))
		args = append(args, *p.ProjectID)
		argNum++
	}
	if p.FromDate != nil {
		conditions = append(conditions, "created_at >= $"+string(rune('0'+argNum)))
		args = append(args, *p.FromDate)
		argNum++
	}
	if p.ToDate != nil {
		conditions = append(conditions, "created_at <= $"+string(rune('0'+argNum)))
		args = append(args, *p.ToDate)
		argNum++
	}
	if p.IPAddress != nil {
		conditions = append(conditions, "ip_address = $"+string(rune('0'+argNum)))
		args = append(args, *p.IPAddress)
		argNum++
	}

	query := "SELECT * FROM audit_log"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if p.Limit > 0 {
		query += " LIMIT $" + string(rune('0'+argNum))
		args = append(args, p.Limit)
		argNum++
	}
	if p.Offset > 0 {
		query += " OFFSET $" + string(rune('0'+argNum))
		args = append(args, p.Offset)
	}

	var entries []*Entry
	err := db.SelectContext(ctx, &entries, query, args...)
	return entries, err
}

func ListByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID, limit, offset int) ([]*Entry, error) {
	var entries []*Entry
	err := db.SelectContext(ctx, &entries,
		`SELECT * FROM audit_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	return entries, err
}

func ListByResource(ctx context.Context, db *sqlx.DB, resourceType ResourceType, resourceID uuid.UUID, limit int) ([]*Entry, error) {
	var entries []*Entry
	err := db.SelectContext(ctx, &entries,
		`SELECT * FROM audit_log WHERE resource_type = $1 AND resource_id = $2
		 ORDER BY created_at DESC LIMIT $3`,
		resourceType, resourceID, limit)
	return entries, err
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID, limit, offset int) ([]*Entry, error) {
	var entries []*Entry
	err := db.SelectContext(ctx, &entries,
		`SELECT * FROM audit_log WHERE project_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, limit, offset)
	return entries, err
}

func CountByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (int64, error) {
	var count int64
	err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM audit_log WHERE user_id = $1`, userID)
	return count, err
}

// DeleteOlderThan removes audit entries older than specified days (for cleanup)
func DeleteOlderThan(ctx context.Context, db *sqlx.DB, days int) (int64, error) {
	result, err := db.ExecContext(ctx,
		`DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL '1 day' * $1`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
