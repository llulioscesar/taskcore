package role

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/store"
)

// Project Roles CRUD

func Create(ctx context.Context, db *sqlx.DB, r *Role) error {
	query := `
		INSERT INTO project_roles (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`
	err := db.QueryRowxContext(ctx, query, r.Name, r.Description).
		Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
	if err != nil && store.IsUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Role, error) {
	r := &Role{}
	err := db.GetContext(ctx, r, `SELECT * FROM project_roles WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

func GetByName(ctx context.Context, db *sqlx.DB, name string) (*Role, error) {
	r := &Role{}
	err := db.GetContext(ctx, r, `SELECT * FROM project_roles WHERE name = $1`, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Role, error) {
	var roles []*Role
	err := db.SelectContext(ctx, &roles, `SELECT * FROM project_roles ORDER BY name`)
	return roles, err
}

func Update(ctx context.Context, db *sqlx.DB, r *Role) error {
	query := `UPDATE project_roles SET name = $2, description = $3 WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, r.ID, r.Name, r.Description).Scan(&r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && store.IsUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM project_roles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Role Actors (users/groups in project role)

func AddActor(ctx context.Context, db *sqlx.DB, a *RoleActor) error {
	query := `
		INSERT INTO project_role_actors (project_id, role_id, actor_type, actor_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, role_id, actor_type, actor_id) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, a.ProjectID, a.RoleID, a.ActorType, a.ActorID).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // already exists
	}
	return err
}

func RemoveActor(ctx context.Context, db *sqlx.DB, projectID, roleID uuid.UUID, actorType ActorType, actorID uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`DELETE FROM project_role_actors WHERE project_id = $1 AND role_id = $2 AND actor_type = $3 AND actor_id = $4`,
		projectID, roleID, actorType, actorID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrActorNotFound
	}
	return nil
}

func ListActors(ctx context.Context, db *sqlx.DB, projectID, roleID uuid.UUID) ([]*RoleActor, error) {
	var actors []*RoleActor
	err := db.SelectContext(ctx, &actors,
		`SELECT * FROM project_role_actors WHERE project_id = $1 AND role_id = $2 ORDER BY created_at`,
		projectID, roleID)
	return actors, err
}

func ListUserActors(ctx context.Context, db *sqlx.DB, projectID, roleID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT actor_id FROM project_role_actors WHERE project_id = $1 AND role_id = $2 AND actor_type = 'user'`,
		projectID, roleID)
	return ids, err
}

func ListProjectRolesForUser(ctx context.Context, db *sqlx.DB, projectID, userID uuid.UUID) ([]*Role, error) {
	var roles []*Role
	query := `
		SELECT DISTINCT r.* FROM project_roles r
		INNER JOIN project_role_actors pra ON r.id = pra.role_id
		LEFT JOIN group_members gm ON pra.actor_type = 'group' AND pra.actor_id = gm.group_id
		WHERE pra.project_id = $1 AND (
			(pra.actor_type = 'user' AND pra.actor_id = $2) OR
			(pra.actor_type = 'group' AND gm.user_id = $2)
		)
		ORDER BY r.name`
	err := db.SelectContext(ctx, &roles, query, projectID, userID)
	return roles, err
}

func IsUserInRole(ctx context.Context, db *sqlx.DB, projectID, roleID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM project_role_actors pra
			LEFT JOIN group_members gm ON pra.actor_type = 'group' AND pra.actor_id = gm.group_id
			WHERE pra.project_id = $1 AND pra.role_id = $2 AND (
				(pra.actor_type = 'user' AND pra.actor_id = $3) OR
				(pra.actor_type = 'group' AND gm.user_id = $3)
			)
		)`
	err := db.GetContext(ctx, &exists, query, projectID, roleID, userID)
	return exists, err
}

// Default Role Actors

func AddDefaultActor(ctx context.Context, db *sqlx.DB, a *DefaultRoleActor) error {
	query := `
		INSERT INTO default_role_actors (role_id, actor_type, actor_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (role_id, actor_type, actor_id) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, a.RoleID, a.ActorType, a.ActorID).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func RemoveDefaultActor(ctx context.Context, db *sqlx.DB, roleID uuid.UUID, actorType ActorType, actorID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM default_role_actors WHERE role_id = $1 AND actor_type = $2 AND actor_id = $3`,
		roleID, actorType, actorID)
	return err
}

func ListDefaultActors(ctx context.Context, db *sqlx.DB, roleID uuid.UUID) ([]*DefaultRoleActor, error) {
	var actors []*DefaultRoleActor
	err := db.SelectContext(ctx, &actors,
		`SELECT * FROM default_role_actors WHERE role_id = $1 ORDER BY created_at`, roleID)
	return actors, err
}

func CopyDefaultActorsToProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) error {
	query := `
		INSERT INTO project_role_actors (project_id, role_id, actor_type, actor_id)
		SELECT $1, role_id, actor_type, actor_id FROM default_role_actors
		ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, projectID)
	return err
}

// Global Roles

func GetGlobalRole(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*GlobalRole, error) {
	r := &GlobalRole{}
	err := db.GetContext(ctx, r, `SELECT * FROM global_roles WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGlobalNotFound
	}
	return r, err
}

func GetGlobalRoleByName(ctx context.Context, db *sqlx.DB, name GlobalRoleType) (*GlobalRole, error) {
	r := &GlobalRole{}
	err := db.GetContext(ctx, r, `SELECT * FROM global_roles WHERE name = $1`, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGlobalNotFound
	}
	return r, err
}

func ListGlobalRoles(ctx context.Context, db *sqlx.DB) ([]*GlobalRole, error) {
	var roles []*GlobalRole
	err := db.SelectContext(ctx, &roles, `SELECT * FROM global_roles ORDER BY name`)
	return roles, err
}

func AddGlobalRoleMember(ctx context.Context, db *sqlx.DB, globalRoleID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO global_role_members (global_role_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		globalRoleID, userID)
	return err
}

func RemoveGlobalRoleMember(ctx context.Context, db *sqlx.DB, globalRoleID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM global_role_members WHERE global_role_id = $1 AND user_id = $2`,
		globalRoleID, userID)
	return err
}

func ListGlobalRoleMembers(ctx context.Context, db *sqlx.DB, globalRoleID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT user_id FROM global_role_members WHERE global_role_id = $1`, globalRoleID)
	return ids, err
}

func GetUserGlobalRoles(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*GlobalRole, error) {
	var roles []*GlobalRole
	query := `
		SELECT gr.* FROM global_roles gr
		INNER JOIN global_role_members grm ON gr.id = grm.global_role_id
		WHERE grm.user_id = $1`
	err := db.SelectContext(ctx, &roles, query, userID)
	return roles, err
}

func IsSystemAdmin(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM global_role_members grm
			INNER JOIN global_roles gr ON grm.global_role_id = gr.id
			WHERE grm.user_id = $1 AND gr.name = 'system_admin'
		)`
	err := db.GetContext(ctx, &exists, query, userID)
	return exists, err
}

// Global Permissions

func GrantGlobalPermission(ctx context.Context, db *sqlx.DB, g *GlobalPermissionGrant) error {
	query := `
		INSERT INTO global_permission_grants (permission, global_role_id, group_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
		RETURNING id`
	return db.QueryRowxContext(ctx, query, g.Permission, g.GlobalRoleID, g.GroupID).Scan(&g.ID)
}

func RevokeGlobalPermission(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM global_permission_grants WHERE id = $1`, id)
	return err
}

func ListGlobalPermissionGrants(ctx context.Context, db *sqlx.DB, permission GlobalPermission) ([]*GlobalPermissionGrant, error) {
	var grants []*GlobalPermissionGrant
	err := db.SelectContext(ctx, &grants,
		`SELECT * FROM global_permission_grants WHERE permission = $1`, permission)
	return grants, err
}

func HasGlobalPermission(ctx context.Context, db *sqlx.DB, userID uuid.UUID, permission GlobalPermission) (bool, error) {
	// System admins have all permissions
	isAdmin, err := IsSystemAdmin(ctx, db, userID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM global_permission_grants gpg
			LEFT JOIN global_role_members grm ON gpg.global_role_id = grm.global_role_id
			LEFT JOIN group_members gm ON gpg.group_id = gm.group_id
			WHERE gpg.permission = $1 AND (
				grm.user_id = $2 OR
				gm.user_id = $2 OR
				(gpg.global_role_id IS NULL AND gpg.group_id IS NULL) -- granted to anyone
			)
		)`
	err = db.GetContext(ctx, &exists, query, permission, userID)
	return exists, err
}
