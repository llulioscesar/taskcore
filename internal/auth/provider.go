package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrProviderNotFound = errors.New("identity provider not found")
	ErrProviderExists   = errors.New("identity provider slug already exists")
)

type IdentityProvider struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	Name            string         `db:"name" json:"name"`
	Slug            string         `db:"slug" json:"slug"`
	Issuer          string         `db:"issuer" json:"issuer"`
	ClientID        string         `db:"client_id" json:"client_id"`
	ClientSecret    string         `db:"client_secret" json:"-"`
	Scopes          pq.StringArray `db:"scopes" json:"scopes"`
	AutoCreateUsers bool           `db:"auto_create_users" json:"auto_create_users"`
	IsActive        bool           `db:"is_active" json:"is_active"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

func CreateProvider(ctx context.Context, db *sqlx.DB, p *IdentityProvider) error {
	query := `
		INSERT INTO identity_providers (name, slug, issuer, client_id, client_secret, scopes, auto_create_users, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := db.QueryRowxContext(ctx, query,
		p.Name, p.Slug, p.Issuer, p.ClientID, p.ClientSecret, p.Scopes, p.AutoCreateUsers, p.IsActive,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil && isUniqueViolation(err) {
		return ErrProviderExists
	}
	return err
}

func GetProviderByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*IdentityProvider, error) {
	p := &IdentityProvider{}
	err := db.GetContext(ctx, p, `SELECT * FROM identity_providers WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProviderNotFound
	}
	return p, err
}

func GetProviderBySlug(ctx context.Context, db *sqlx.DB, slug string) (*IdentityProvider, error) {
	p := &IdentityProvider{}
	err := db.GetContext(ctx, p, `SELECT * FROM identity_providers WHERE slug = $1 AND is_active = true`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProviderNotFound
	}
	return p, err
}

func ListProviders(ctx context.Context, db *sqlx.DB) ([]*IdentityProvider, error) {
	var providers []*IdentityProvider
	err := db.SelectContext(ctx, &providers,
		`SELECT * FROM identity_providers ORDER BY name`)
	return providers, err
}

func ListActiveProviders(ctx context.Context, db *sqlx.DB) ([]*IdentityProvider, error) {
	var providers []*IdentityProvider
	err := db.SelectContext(ctx, &providers,
		`SELECT * FROM identity_providers WHERE is_active = true ORDER BY name`)
	return providers, err
}

func UpdateProvider(ctx context.Context, db *sqlx.DB, p *IdentityProvider) error {
	query := `
		UPDATE identity_providers
		SET name = $2, slug = $3, issuer = $4, client_id = $5, client_secret = $6,
		    scopes = $7, auto_create_users = $8, is_active = $9, updated_at = NOW()
		WHERE id = $1 RETURNING updated_at`

	err := db.QueryRowxContext(ctx, query,
		p.ID, p.Name, p.Slug, p.Issuer, p.ClientID, p.ClientSecret,
		p.Scopes, p.AutoCreateUsers, p.IsActive,
	).Scan(&p.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrProviderNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return ErrProviderExists
	}
	return err
}

func DeleteProvider(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM identity_providers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
