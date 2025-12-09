package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Application CRUD

func Create(ctx context.Context, db *sqlx.DB, a *Application) error {
	query := `
		INSERT INTO applications (key, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query, a.Key, a.Name, a.Description, a.IsActive).
		Scan(&a.ID, &a.CreatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Application, error) {
	a := &Application{}
	err := db.GetContext(ctx, a, `SELECT * FROM applications WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func GetByKey(ctx context.Context, db *sqlx.DB, key string) (*Application, error) {
	a := &Application{}
	err := db.GetContext(ctx, a, `SELECT * FROM applications WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Application, error) {
	var apps []*Application
	err := db.SelectContext(ctx, &apps, `SELECT * FROM applications ORDER BY name`)
	return apps, err
}

func ListActive(ctx context.Context, db *sqlx.DB) ([]*Application, error) {
	var apps []*Application
	err := db.SelectContext(ctx, &apps,
		`SELECT * FROM applications WHERE is_active = true ORDER BY name`)
	return apps, err
}

func Update(ctx context.Context, db *sqlx.DB, a *Application) error {
	result, err := db.ExecContext(ctx,
		`UPDATE applications SET name = $2, description = $3, is_active = $4 WHERE id = $1`,
		a.ID, a.Name, a.Description, a.IsActive)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func SetActive(ctx context.Context, db *sqlx.DB, id uuid.UUID, isActive bool) error {
	result, err := db.ExecContext(ctx,
		`UPDATE applications SET is_active = $2 WHERE id = $1`, id, isActive)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// User Access

func GrantUserAccess(ctx context.Context, db *sqlx.DB, a *UserAccess) error {
	query := `
		INSERT INTO application_user_access (user_id, application_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, application_id) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, a.UserID, a.ApplicationID, a.GrantedBy).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func RevokeUserAccess(ctx context.Context, db *sqlx.DB, userID, applicationID uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`DELETE FROM application_user_access WHERE user_id = $1 AND application_id = $2`,
		userID, applicationID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoAccess
	}
	return nil
}

func HasAccess(ctx context.Context, db *sqlx.DB, userID, applicationID uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM application_user_access
			WHERE user_id = $1 AND application_id = $2
			UNION ALL
			SELECT 1 FROM application_group_access aga
			INNER JOIN group_members gm ON aga.group_id = gm.group_id
			WHERE gm.user_id = $1 AND aga.application_id = $2
		)`
	err := db.GetContext(ctx, &exists, query, userID, applicationID)
	return exists, err
}

func HasAccessByKey(ctx context.Context, db *sqlx.DB, userID uuid.UUID, appKey string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM application_user_access aua
			INNER JOIN applications a ON aua.application_id = a.id
			WHERE aua.user_id = $1 AND a.key = $2 AND a.is_active = true
			UNION ALL
			SELECT 1 FROM application_group_access aga
			INNER JOIN applications a ON aga.application_id = a.id
			INNER JOIN group_members gm ON aga.group_id = gm.group_id
			WHERE gm.user_id = $1 AND a.key = $2 AND a.is_active = true
		)`
	err := db.GetContext(ctx, &exists, query, userID, appKey)
	return exists, err
}

func ListUserApplications(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Application, error) {
	var apps []*Application
	query := `
		SELECT DISTINCT a.* FROM applications a
		LEFT JOIN application_user_access aua ON a.id = aua.application_id AND aua.user_id = $1
		LEFT JOIN application_group_access aga ON a.id = aga.application_id
		LEFT JOIN group_members gm ON aga.group_id = gm.group_id AND gm.user_id = $1
		WHERE a.is_active = true AND (aua.user_id IS NOT NULL OR gm.user_id IS NOT NULL)
		ORDER BY a.name`
	err := db.SelectContext(ctx, &apps, query, userID)
	return apps, err
}

func ListUsersWithAccess(ctx context.Context, db *sqlx.DB, applicationID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	query := `
		SELECT DISTINCT user_id FROM (
			SELECT user_id FROM application_user_access WHERE application_id = $1
			UNION
			SELECT gm.user_id FROM application_group_access aga
			INNER JOIN group_members gm ON aga.group_id = gm.group_id
			WHERE aga.application_id = $1
		) u`
	err := db.SelectContext(ctx, &ids, query, applicationID)
	return ids, err
}

// Group Access

func GrantGroupAccess(ctx context.Context, db *sqlx.DB, a *GroupAccess) error {
	query := `
		INSERT INTO application_group_access (group_id, application_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (group_id, application_id) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, a.GroupID, a.ApplicationID, a.GrantedBy).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func RevokeGroupAccess(ctx context.Context, db *sqlx.DB, groupID, applicationID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM application_group_access WHERE group_id = $1 AND application_id = $2`,
		groupID, applicationID)
	return err
}

func ListGroupsWithAccess(ctx context.Context, db *sqlx.DB, applicationID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT group_id FROM application_group_access WHERE application_id = $1`, applicationID)
	return ids, err
}

// Grant access to all active applications for a user
func GrantAllAccess(ctx context.Context, db *sqlx.DB, userID uuid.UUID, grantedBy *uuid.UUID) error {
	query := `
		INSERT INTO application_user_access (user_id, application_id, granted_by)
		SELECT $1, id, $2 FROM applications WHERE is_active = true
		ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, userID, grantedBy)
	return err
}
