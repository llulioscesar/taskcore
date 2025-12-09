package permission

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/store"
)

func CreateScheme(ctx context.Context, db *sqlx.DB, ps *Scheme) error {
	query := `
		INSERT INTO permission_schemes (name, description, is_default)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query, ps.Name, ps.Description, ps.IsDefault).
		Scan(&ps.ID, &ps.CreatedAt, &ps.UpdatedAt)
	if err != nil && store.IsUniqueViolation(err) {
		return ErrSchemeNameExists
	}
	return err
}

func GetScheme(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Scheme, error) {
	ps := &Scheme{}
	err := db.GetContext(ctx, ps, `SELECT * FROM permission_schemes WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSchemeNotFound
	}
	return ps, err
}

func GetDefaultScheme(ctx context.Context, db *sqlx.DB) (*Scheme, error) {
	ps := &Scheme{}
	err := db.GetContext(ctx, ps, `SELECT * FROM permission_schemes WHERE is_default = true LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSchemeNotFound
	}
	return ps, err
}

func ListSchemes(ctx context.Context, db *sqlx.DB) ([]*Scheme, error) {
	var schemes []*Scheme
	err := db.SelectContext(ctx, &schemes, `SELECT * FROM permission_schemes ORDER BY is_default DESC, name`)
	return schemes, err
}

func UpdateScheme(ctx context.Context, db *sqlx.DB, ps *Scheme) error {
	query := `UPDATE permission_schemes SET name = $2, description = $3, is_default = $4 WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, ps.ID, ps.Name, ps.Description, ps.IsDefault).Scan(&ps.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSchemeNotFound
	}
	if err != nil && store.IsUniqueViolation(err) {
		return ErrSchemeNameExists
	}
	return err
}

func DeleteScheme(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM permission_schemes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrSchemeNotFound
	}
	return nil
}

func Grant(ctx context.Context, db *sqlx.DB, sp *SchemePermission) error {
	query := `
		INSERT INTO permission_scheme_permissions (scheme_id, permission, grantee_type, grantee_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (scheme_id, permission, grantee_type, COALESCE(grantee_id, '00000000-0000-0000-0000-000000000000'))
		DO NOTHING RETURNING id`

	err := db.QueryRowxContext(ctx, query, sp.SchemeID, sp.Permission, sp.GranteeType, sp.GranteeID).Scan(&sp.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func Revoke(ctx context.Context, db *sqlx.DB, schemeID uuid.UUID, perm Permission, granteeType GranteeType, granteeID *uuid.UUID) error {
	var query string
	var args []any

	if granteeID == nil {
		query = `DELETE FROM permission_scheme_permissions WHERE scheme_id = $1 AND permission = $2 AND grantee_type = $3 AND grantee_id IS NULL`
		args = []any{schemeID, perm, granteeType}
	} else {
		query = `DELETE FROM permission_scheme_permissions WHERE scheme_id = $1 AND permission = $2 AND grantee_type = $3 AND grantee_id = $4`
		args = []any{schemeID, perm, granteeType, granteeID}
	}

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func ListByScheme(ctx context.Context, db *sqlx.DB, schemeID uuid.UUID) ([]*SchemePermission, error) {
	var perms []*SchemePermission
	err := db.SelectContext(ctx, &perms,
		`SELECT * FROM permission_scheme_permissions WHERE scheme_id = $1 ORDER BY permission, grantee_type`, schemeID)
	return perms, err
}

func HasProjectPermission(ctx context.Context, db *sqlx.DB, userID, projectID uuid.UUID, perm Permission) (bool, error) {
	query := `
		SELECT EXISTS (
			-- System admins have all permissions
			SELECT 1 FROM global_role_members grm
			INNER JOIN global_roles gr ON grm.global_role_id = gr.id
			WHERE grm.user_id = $1 AND gr.name = 'system_admin'
			UNION ALL
			-- Legacy: users.is_admin flag
			SELECT 1 FROM users WHERE id = $1 AND is_admin = true AND is_active = true
			UNION ALL
			-- Direct user grant
			SELECT 1 FROM permission_scheme_permissions psp
			INNER JOIN projects p ON p.permission_scheme_id = psp.scheme_id
			WHERE p.id = $2 AND psp.permission = $3
			  AND psp.grantee_type = 'user' AND psp.grantee_id = $1
			UNION ALL
			-- Group grant
			SELECT 1 FROM permission_scheme_permissions psp
			INNER JOIN projects p ON p.permission_scheme_id = psp.scheme_id
			INNER JOIN group_members gm ON psp.grantee_id = gm.group_id
			WHERE p.id = $2 AND psp.permission = $3
			  AND psp.grantee_type = 'group' AND gm.user_id = $1
			UNION ALL
			-- Project role grant (dynamic roles)
			SELECT 1 FROM permission_scheme_permissions psp
			INNER JOIN projects p ON p.permission_scheme_id = psp.scheme_id
			INNER JOIN project_role_actors pra ON pra.project_id = p.id AND pra.role_id = psp.grantee_id
			LEFT JOIN group_members gm ON pra.actor_type = 'group' AND pra.actor_id = gm.group_id
			WHERE p.id = $2 AND psp.permission = $3
			  AND psp.grantee_type = 'project_role'
			  AND ((pra.actor_type = 'user' AND pra.actor_id = $1) OR
			       (pra.actor_type = 'group' AND gm.user_id = $1))
			UNION ALL
			-- Anyone grant
			SELECT 1 FROM permission_scheme_permissions psp
			INNER JOIN projects p ON p.permission_scheme_id = psp.scheme_id
			WHERE p.id = $2 AND psp.permission = $3 AND psp.grantee_type = 'anyone'
		)`

	var has bool
	err := db.GetContext(ctx, &has, query, userID, projectID, perm)
	return has, err
}

func ListUserPermissions(ctx context.Context, db *sqlx.DB, userID, projectID uuid.UUID) ([]Permission, error) {
	query := `
		SELECT DISTINCT psp.permission FROM permission_scheme_permissions psp
		INNER JOIN projects p ON p.permission_scheme_id = psp.scheme_id
		LEFT JOIN group_members gm_group ON psp.grantee_type = 'group' AND psp.grantee_id = gm_group.group_id
		LEFT JOIN project_role_actors pra ON psp.grantee_type = 'project_role' AND pra.project_id = p.id AND pra.role_id = psp.grantee_id
		LEFT JOIN group_members gm_role ON pra.actor_type = 'group' AND pra.actor_id = gm_role.group_id
		WHERE p.id = $2 AND (
			(psp.grantee_type = 'user' AND psp.grantee_id = $1) OR
			(psp.grantee_type = 'group' AND gm_group.user_id = $1) OR
			(psp.grantee_type = 'project_role' AND (
				(pra.actor_type = 'user' AND pra.actor_id = $1) OR
				(pra.actor_type = 'group' AND gm_role.user_id = $1)
			)) OR
			(psp.grantee_type = 'anyone')
		)
		ORDER BY psp.permission`

	var perms []Permission
	err := db.SelectContext(ctx, &perms, query, userID, projectID)
	return perms, err
}

func IsAdmin(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (bool, error) {
	var isAdmin bool
	err := db.GetContext(ctx, &isAdmin, `SELECT is_admin FROM users WHERE id = $1 AND is_active = true`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}
