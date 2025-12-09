package license

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Applications

func CreateApplication(ctx context.Context, db *sqlx.DB, a *Application) error {
	query := `
		INSERT INTO applications (key, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query, a.Key, a.Name, a.Description, a.IsActive).
		Scan(&a.ID, &a.CreatedAt)
}

func GetApplication(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Application, error) {
	a := &Application{}
	err := db.GetContext(ctx, a, `SELECT * FROM applications WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	return a, err
}

func GetApplicationByKey(ctx context.Context, db *sqlx.DB, key string) (*Application, error) {
	a := &Application{}
	err := db.GetContext(ctx, a, `SELECT * FROM applications WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	return a, err
}

func ListApplications(ctx context.Context, db *sqlx.DB) ([]*Application, error) {
	var apps []*Application
	err := db.SelectContext(ctx, &apps, `SELECT * FROM applications ORDER BY name`)
	return apps, err
}

func ListActiveApplications(ctx context.Context, db *sqlx.DB) ([]*Application, error) {
	var apps []*Application
	err := db.SelectContext(ctx, &apps,
		`SELECT * FROM applications WHERE is_active = true ORDER BY name`)
	return apps, err
}

func UpdateApplication(ctx context.Context, db *sqlx.DB, a *Application) error {
	result, err := db.ExecContext(ctx,
		`UPDATE applications SET name = $2, description = $3, is_active = $4 WHERE id = $1`,
		a.ID, a.Name, a.Description, a.IsActive)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

// License

func GetLicense(ctx context.Context, db *sqlx.DB) (*License, error) {
	l := &License{}
	err := db.GetContext(ctx, l, `SELECT * FROM licenses ORDER BY created_at DESC LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return l, err
}

func SetLicense(ctx context.Context, db *sqlx.DB, l *License) error {
	query := `
		INSERT INTO licenses (license_key, license_type, max_users, expires_at, licensed_to, support_expires)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query,
		l.LicenseKey, l.LicenseType, l.MaxUsers, l.ExpiresAt, l.LicensedTo, l.SupportExpires,
	).Scan(&l.ID, &l.CreatedAt, &l.UpdatedAt)
}

func GetLicensedUserCount(ctx context.Context, db *sqlx.DB) (int, error) {
	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(DISTINCT user_id) FROM application_access`)
	return count, err
}

func CanAddUser(ctx context.Context, db *sqlx.DB) (bool, error) {
	license, err := GetLicense(ctx, db)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return true, nil // No license = community mode, allow
		}
		return false, err
	}

	if license.IsExpired() {
		return false, ErrLicenseExpired
	}

	if license.IsUnlimited() {
		return true, nil
	}

	count, err := GetLicensedUserCount(ctx, db)
	if err != nil {
		return false, err
	}

	return count < license.MaxUsers, nil
}

// User Application Access

func GrantAccess(ctx context.Context, db *sqlx.DB, a *ApplicationAccess) error {
	query := `
		INSERT INTO application_access (user_id, application_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, application_id) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, a.UserID, a.ApplicationID, a.GrantedBy).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // already has access
	}
	return err
}

func RevokeAccess(ctx context.Context, db *sqlx.DB, userID, applicationID uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`DELETE FROM application_access WHERE user_id = $1 AND application_id = $2`,
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
			-- Direct user access
			SELECT 1 FROM application_access
			WHERE user_id = $1 AND application_id = $2
			UNION ALL
			-- Group access
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
			SELECT 1 FROM application_access aa
			INNER JOIN applications a ON aa.application_id = a.id
			WHERE aa.user_id = $1 AND a.key = $2 AND a.is_active = true
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
		LEFT JOIN application_access aa ON a.id = aa.application_id AND aa.user_id = $1
		LEFT JOIN application_group_access aga ON a.id = aga.application_id
		LEFT JOIN group_members gm ON aga.group_id = gm.group_id AND gm.user_id = $1
		WHERE a.is_active = true AND (aa.user_id IS NOT NULL OR gm.user_id IS NOT NULL)
		ORDER BY a.name`
	err := db.SelectContext(ctx, &apps, query, userID)
	return apps, err
}

func ListUsersWithAccess(ctx context.Context, db *sqlx.DB, applicationID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	query := `
		SELECT DISTINCT user_id FROM (
			SELECT user_id FROM application_access WHERE application_id = $1
			UNION
			SELECT gm.user_id FROM application_group_access aga
			INNER JOIN group_members gm ON aga.group_id = gm.group_id
			WHERE aga.application_id = $1
		) u`
	err := db.SelectContext(ctx, &ids, query, applicationID)
	return ids, err
}

// Group Application Access

func GrantGroupAccess(ctx context.Context, db *sqlx.DB, a *ApplicationGroupAccess) error {
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

// Default Access (auto-grant to new users)

func SetDefaultAccess(ctx context.Context, db *sqlx.DB, applicationID uuid.UUID, isDefault bool) error {
	query := `
		INSERT INTO application_defaults (application_id, is_default)
		VALUES ($1, $2)
		ON CONFLICT (application_id) DO UPDATE SET is_default = EXCLUDED.is_default`
	_, err := db.ExecContext(ctx, query, applicationID, isDefault)
	return err
}

func ListDefaultApplications(ctx context.Context, db *sqlx.DB) ([]*Application, error) {
	var apps []*Application
	query := `
		SELECT a.* FROM applications a
		INNER JOIN application_defaults ad ON a.id = ad.application_id
		WHERE ad.is_default = true AND a.is_active = true`
	err := db.SelectContext(ctx, &apps, query)
	return apps, err
}

func GrantDefaultAccessToUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID, grantedBy *uuid.UUID) error {
	query := `
		INSERT INTO application_access (user_id, application_id, granted_by)
		SELECT $1, a.id, $2 FROM applications a
		INNER JOIN application_defaults ad ON a.id = ad.application_id
		WHERE ad.is_default = true AND a.is_active = true
		ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, userID, grantedBy)
	return err
}
