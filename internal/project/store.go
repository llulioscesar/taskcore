package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/store"
)

func Create(ctx context.Context, db *sqlx.DB, p *Project) error {
	query := `
		INSERT INTO projects (template_id, key, name, description, lead_id, default_assignee_id, permission_scheme_id, workflow_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, issue_counter, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query,
		p.TemplateID, p.Key, p.Name, p.Description, p.LeadID,
		p.DefaultAssigneeID, p.PermissionSchemeID, p.WorkflowID,
	).Scan(&p.ID, &p.IssueCounter, &p.CreatedAt, &p.UpdatedAt)

	if err != nil && store.IsUniqueViolation(err) {
		return ErrKeyExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Project, error) {
	p := &Project{}
	err := db.GetContext(ctx, p, `SELECT * FROM projects WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func GetByKey(ctx context.Context, db *sqlx.DB, key string) (*Project, error) {
	p := &Project{}
	err := db.GetContext(ctx, p, `SELECT * FROM projects WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Project, error) {
	var projects []*Project
	err := db.SelectContext(ctx, &projects, `SELECT * FROM projects ORDER BY name`)
	return projects, err
}

func ListByUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Project, error) {
	query := `
		SELECT DISTINCT p.* FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		LEFT JOIN group_members gm ON pm.group_id = gm.group_id
		WHERE p.lead_id = $1 OR pm.user_id = $1 OR gm.user_id = $1
		ORDER BY p.name`

	var projects []*Project
	err := db.SelectContext(ctx, &projects, query, userID)
	return projects, err
}

func Update(ctx context.Context, db *sqlx.DB, p *Project) error {
	query := `
		UPDATE projects
		SET key = $2, name = $3, description = $4, lead_id = $5,
		    default_assignee_id = $6, permission_scheme_id = $7, workflow_id = $8
		WHERE id = $1 RETURNING updated_at`

	err := db.QueryRowxContext(ctx, query,
		p.ID, p.Key, p.Name, p.Description, p.LeadID,
		p.DefaultAssigneeID, p.PermissionSchemeID, p.WorkflowID,
	).Scan(&p.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && store.IsUniqueViolation(err) {
		return ErrKeyExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func AddUserMember(ctx context.Context, db *sqlx.DB, projectID, userID uuid.UUID, role Role) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) WHERE user_id IS NOT NULL DO UPDATE SET role = EXCLUDED.role`
	_, err := db.ExecContext(ctx, query, projectID, userID, role)
	return err
}

func AddGroupMember(ctx context.Context, db *sqlx.DB, projectID, groupID uuid.UUID, role Role) error {
	query := `
		INSERT INTO project_members (project_id, group_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (project_id, group_id) WHERE group_id IS NOT NULL DO UPDATE SET role = EXCLUDED.role`
	_, err := db.ExecContext(ctx, query, projectID, groupID, role)
	return err
}

func RemoveUserMember(ctx context.Context, db *sqlx.DB, projectID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID)
	return err
}

func RemoveGroupMember(ctx context.Context, db *sqlx.DB, projectID, groupID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id = $1 AND group_id = $2`, projectID, groupID)
	return err
}

func ListMembers(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Member, error) {
	var members []*Member
	err := db.SelectContext(ctx, &members,
		`SELECT * FROM project_members WHERE project_id = $1 ORDER BY created_at`, projectID)
	return members, err
}

func GetUserRole(ctx context.Context, db *sqlx.DB, projectID, userID uuid.UUID) (Role, error) {
	var role Role
	err := db.GetContext(ctx, &role,
		`SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID)
	if err == nil {
		return role, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	query := `
		SELECT pm.role FROM project_members pm
		INNER JOIN group_members gm ON pm.group_id = gm.group_id
		WHERE pm.project_id = $1 AND gm.user_id = $2
		ORDER BY CASE pm.role WHEN 'admin' THEN 1 WHEN 'developer' THEN 2 ELSE 3 END
		LIMIT 1`

	err = db.GetContext(ctx, &role, query, projectID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func IsMember(ctx context.Context, db *sqlx.DB, projectID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM project_members pm
			LEFT JOIN group_members gm ON pm.group_id = gm.group_id
			WHERE pm.project_id = $1 AND (pm.user_id = $2 OR gm.user_id = $2)
		)`
	err := db.GetContext(ctx, &exists, query, projectID, userID)
	return exists, err
}

func NextIssueKey(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) (string, error) {
	var result struct {
		Key     string `db:"key"`
		Counter int    `db:"issue_counter"`
	}
	query := `
		UPDATE projects SET issue_counter = issue_counter + 1
		WHERE id = $1
		RETURNING key, issue_counter`
	err := db.QueryRowxContext(ctx, query, projectID).Scan(&result.Key, &result.Counter)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return result.Key + "-" + itoa(result.Counter), nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	idx := len(b)
	for i > 0 {
		idx--
		b[idx] = byte('0' + i%10)
		i /= 10
	}
	return string(b[idx:])
}
