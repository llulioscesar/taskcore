package activity

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, l *Log) error {
	query := `
		INSERT INTO activity_logs (entity_type, entity_id, user_id, action, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query,
		l.EntityType, l.EntityID, l.UserID, l.Action, l.OldValue, l.NewValue,
	).Scan(&l.ID, &l.CreatedAt)
}

func ListByEntity(ctx context.Context, db *sqlx.DB, entityType EntityType, entityID uuid.UUID, limit int) ([]*Log, error) {
	var logs []*Log
	query := `
		SELECT * FROM activity_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3`
	err := db.SelectContext(ctx, &logs, query, entityType, entityID, limit)
	return logs, err
}

func ListByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID, limit int) ([]*Log, error) {
	var logs []*Log
	query := `
		SELECT * FROM activity_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
	err := db.SelectContext(ctx, &logs, query, userID, limit)
	return logs, err
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID, limit int) ([]*Log, error) {
	var logs []*Log
	query := `
		SELECT al.* FROM activity_logs al
		INNER JOIN issues i ON al.entity_type = 'issue' AND al.entity_id = i.id
		WHERE i.project_id = $1
		UNION ALL
		SELECT al.* FROM activity_logs al
		WHERE al.entity_type = 'project' AND al.entity_id = $1
		UNION ALL
		SELECT al.* FROM activity_logs al
		INNER JOIN sprints s ON al.entity_type = 'sprint' AND al.entity_id = s.id
		WHERE s.project_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
	err := db.SelectContext(ctx, &logs, query, projectID, limit)
	return logs, err
}

func DeleteOlderThan(ctx context.Context, db *sqlx.DB, days int) (int64, error) {
	result, err := db.ExecContext(ctx,
		`DELETE FROM activity_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
