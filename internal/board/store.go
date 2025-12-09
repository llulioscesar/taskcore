package board

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Board CRUD

func Create(ctx context.Context, db *sqlx.DB, b *Board) error {
	query := `
		INSERT INTO boards (project_id, name, type, filter_jql)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query,
		b.ProjectID, b.Name, b.Type, b.FilterJQL,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Board, error) {
	b := &Board{}
	err := db.GetContext(ctx, b, `SELECT * FROM boards WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return b, err
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Board, error) {
	var boards []*Board
	err := db.SelectContext(ctx, &boards,
		`SELECT * FROM boards WHERE project_id = $1 ORDER BY name`, projectID)
	return boards, err
}

func Update(ctx context.Context, db *sqlx.DB, b *Board) error {
	query := `
		UPDATE boards SET name = $2, filter_jql = $3
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query, b.ID, b.Name, b.FilterJQL).Scan(&b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM boards WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Columns

func CreateColumn(ctx context.Context, db *sqlx.DB, c *Column) error {
	query := `
		INSERT INTO board_columns (board_id, name, status_id, position, min_limit, max_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	return db.QueryRowxContext(ctx, query,
		c.BoardID, c.Name, c.StatusID, c.Position, c.MinLimit, c.MaxLimit,
	).Scan(&c.ID)
}

func ListColumns(ctx context.Context, db *sqlx.DB, boardID uuid.UUID) ([]*Column, error) {
	var columns []*Column
	err := db.SelectContext(ctx, &columns,
		`SELECT * FROM board_columns WHERE board_id = $1 ORDER BY position`, boardID)
	return columns, err
}

func UpdateColumn(ctx context.Context, db *sqlx.DB, c *Column) error {
	result, err := db.ExecContext(ctx,
		`UPDATE board_columns SET name = $2, status_id = $3, position = $4, min_limit = $5, max_limit = $6 WHERE id = $1`,
		c.ID, c.Name, c.StatusID, c.Position, c.MinLimit, c.MaxLimit)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrColumnNotFound
	}
	return nil
}

func DeleteColumn(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM board_columns WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrColumnNotFound
	}
	return nil
}

func ReorderColumns(ctx context.Context, db *sqlx.DB, boardID uuid.UUID, columnIDs []uuid.UUID) error {
	query := `
		UPDATE board_columns SET position = c.pos
		FROM (SELECT unnest($2::uuid[]) AS id, generate_series(0, $3) AS pos) c
		WHERE board_columns.id = c.id AND board_columns.board_id = $1`
	_, err := db.ExecContext(ctx, query, boardID, pq.Array(columnIDs), len(columnIDs)-1)
	return err
}

// Config

func GetConfig(ctx context.Context, db *sqlx.DB, boardID uuid.UUID) (*Config, error) {
	c := &Config{}
	err := db.GetContext(ctx, c, `SELECT * FROM board_configs WHERE board_id = $1`, boardID)
	if errors.Is(err, sql.ErrNoRows) {
		return &Config{BoardID: boardID, Swimlane: SwimlaneNone}, nil
	}
	return c, err
}

func SetConfig(ctx context.Context, db *sqlx.DB, c *Config) error {
	query := `
		INSERT INTO board_configs (board_id, swimlane, card_fields, show_days_in_col, show_epic_as_bar)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (board_id) DO UPDATE
		SET swimlane = EXCLUDED.swimlane, card_fields = EXCLUDED.card_fields,
		    show_days_in_col = EXCLUDED.show_days_in_col, show_epic_as_bar = EXCLUDED.show_epic_as_bar`
	_, err := db.ExecContext(ctx, query,
		c.BoardID, c.Swimlane, pq.Array(c.CardFields), c.ShowDaysInCol, c.ShowEpicAsBar)
	return err
}
