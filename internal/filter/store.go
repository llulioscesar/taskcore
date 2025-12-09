package filter

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, f *Filter) error {
	query := `
		INSERT INTO filters (owner_id, name, description, jql, is_public, is_favorite)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	err := db.QueryRowxContext(ctx, query,
		f.OwnerID, f.Name, f.Description, f.JQL, f.IsPublic, f.IsFavorite,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Filter, error) {
	f := &Filter{}
	err := db.GetContext(ctx, f, `SELECT * FROM filters WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return f, err
}

func ListByOwner(ctx context.Context, db *sqlx.DB, ownerID uuid.UUID) ([]*Filter, error) {
	var filters []*Filter
	err := db.SelectContext(ctx, &filters,
		`SELECT * FROM filters WHERE owner_id = $1 ORDER BY is_favorite DESC, name`, ownerID)
	return filters, err
}

func ListFavorites(ctx context.Context, db *sqlx.DB, ownerID uuid.UUID) ([]*Filter, error) {
	var filters []*Filter
	err := db.SelectContext(ctx, &filters,
		`SELECT * FROM filters WHERE owner_id = $1 AND is_favorite = true ORDER BY name`, ownerID)
	return filters, err
}

func ListSharedWithUser(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Filter, error) {
	var filters []*Filter
	query := `
		SELECT DISTINCT f.* FROM filters f
		LEFT JOIN filter_shares fs ON f.id = fs.filter_id
		LEFT JOIN group_members gm ON fs.group_id = gm.group_id
		LEFT JOIN project_members pm ON fs.project_id = pm.project_id
		WHERE f.is_public = true
		   OR fs.user_id = $1
		   OR gm.user_id = $1
		   OR pm.user_id = $1
		ORDER BY f.name`
	err := db.SelectContext(ctx, &filters, query, userID)
	return filters, err
}

func ListPublic(ctx context.Context, db *sqlx.DB) ([]*Filter, error) {
	var filters []*Filter
	err := db.SelectContext(ctx, &filters,
		`SELECT * FROM filters WHERE is_public = true ORDER BY name`)
	return filters, err
}

func Update(ctx context.Context, db *sqlx.DB, f *Filter) error {
	query := `
		UPDATE filters SET name = $2, description = $3, jql = $4, is_public = $5, is_favorite = $6
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query,
		f.ID, f.Name, f.Description, f.JQL, f.IsPublic, f.IsFavorite,
	).Scan(&f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrNameExists
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM filters WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func SetFavorite(ctx context.Context, db *sqlx.DB, id uuid.UUID, isFavorite bool) error {
	result, err := db.ExecContext(ctx,
		`UPDATE filters SET is_favorite = $2 WHERE id = $1`, id, isFavorite)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Shares

func AddShare(ctx context.Context, db *sqlx.DB, s *Share) error {
	query := `
		INSERT INTO filter_shares (filter_id, user_id, group_id, project_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, s.FilterID, s.UserID, s.GroupID, s.ProjectID)
	return err
}

func RemoveShare(ctx context.Context, db *sqlx.DB, filterID uuid.UUID, userID, groupID, projectID *uuid.UUID) error {
	query := `
		DELETE FROM filter_shares
		WHERE filter_id = $1
		  AND (user_id = $2 OR (user_id IS NULL AND $2 IS NULL))
		  AND (group_id = $3 OR (group_id IS NULL AND $3 IS NULL))
		  AND (project_id = $4 OR (project_id IS NULL AND $4 IS NULL))`
	_, err := db.ExecContext(ctx, query, filterID, userID, groupID, projectID)
	return err
}

func ListShares(ctx context.Context, db *sqlx.DB, filterID uuid.UUID) ([]*Share, error) {
	var shares []*Share
	err := db.SelectContext(ctx, &shares,
		`SELECT * FROM filter_shares WHERE filter_id = $1`, filterID)
	return shares, err
}

// Subscriptions

func Subscribe(ctx context.Context, db *sqlx.DB, s *Subscription) error {
	query := `
		INSERT INTO filter_subscriptions (filter_id, user_id, schedule)
		VALUES ($1, $2, $3)
		ON CONFLICT (filter_id, user_id) DO UPDATE SET schedule = EXCLUDED.schedule
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query, s.FilterID, s.UserID, s.Schedule).Scan(&s.ID, &s.CreatedAt)
}

func Unsubscribe(ctx context.Context, db *sqlx.DB, filterID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM filter_subscriptions WHERE filter_id = $1 AND user_id = $2`, filterID, userID)
	return err
}

func ListSubscriptions(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Subscription, error) {
	var subs []*Subscription
	err := db.SelectContext(ctx, &subs,
		`SELECT * FROM filter_subscriptions WHERE user_id = $1`, userID)
	return subs, err
}

func ListSubscribers(ctx context.Context, db *sqlx.DB, filterID uuid.UUID) ([]*Subscription, error) {
	var subs []*Subscription
	err := db.SelectContext(ctx, &subs,
		`SELECT * FROM filter_subscriptions WHERE filter_id = $1`, filterID)
	return subs, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
