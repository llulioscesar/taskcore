package group

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, g *Group) error {
	query := `
		INSERT INTO groups (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query, g.Name, g.Description).
		Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)

	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Group, error) {
	g := &Group{}
	err := db.GetContext(ctx, g, `SELECT * FROM groups WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func GetByName(ctx context.Context, db *sqlx.DB, name string) (*Group, error) {
	g := &Group{}
	err := db.GetContext(ctx, g, `SELECT * FROM groups WHERE name = $1`, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func List(ctx context.Context, db *sqlx.DB) ([]*Group, error) {
	var groups []*Group
	err := db.SelectContext(ctx, &groups, `SELECT * FROM groups ORDER BY name`)
	return groups, err
}

func Update(ctx context.Context, db *sqlx.DB, g *Group) error {
	query := `UPDATE groups SET name = $2, description = $3 WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, g.ID, g.Name, g.Description).Scan(&g.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM groups WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func AddMember(ctx context.Context, db *sqlx.DB, groupID, userID uuid.UUID) error {
	query := `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, groupID, userID)
	return err
}

func RemoveMember(ctx context.Context, db *sqlx.DB, groupID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func GetMembers(ctx context.Context, db *sqlx.DB, groupID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT user_id FROM group_members WHERE group_id = $1`, groupID)
	return ids, err
}

func GetUserGroups(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Group, error) {
	var groups []*Group
	query := `
		SELECT g.* FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1 ORDER BY g.name`
	err := db.SelectContext(ctx, &groups, query, userID)
	return groups, err
}

func CountMembers(ctx context.Context, db *sqlx.DB, groupID uuid.UUID) (int, error) {
	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM group_members WHERE group_id = $1`, groupID)
	return count, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
