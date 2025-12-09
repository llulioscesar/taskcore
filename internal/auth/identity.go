package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrIdentityNotFound = errors.New("user identity not found")
	ErrStateNotFound    = errors.New("OIDC state not found")
	ErrStateExpired     = errors.New("OIDC state expired")
)

type UserIdentity struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	UserID      uuid.UUID       `db:"user_id" json:"user_id"`
	ProviderID  uuid.UUID       `db:"provider_id" json:"provider_id"`
	Subject     string          `db:"subject" json:"subject"`
	Email       *string         `db:"email" json:"email,omitempty"`
	Name        *string         `db:"name" json:"name,omitempty"`
	Picture     *string         `db:"picture" json:"picture,omitempty"`
	RawClaims   json.RawMessage `db:"raw_claims" json:"raw_claims,omitempty"`
	LastLoginAt *time.Time      `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

type OIDCState struct {
	State       string    `db:"state"`
	ProviderID  uuid.UUID `db:"provider_id"`
	RedirectURI *string   `db:"redirect_uri"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

// UserIdentity CRUD

func CreateIdentity(ctx context.Context, db *sqlx.DB, i *UserIdentity) error {
	query := `
		INSERT INTO user_identities (user_id, provider_id, subject, email, name, picture, raw_claims)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return db.QueryRowxContext(ctx, query,
		i.UserID, i.ProviderID, i.Subject, i.Email, i.Name, i.Picture, i.RawClaims,
	).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
}

func GetIdentityBySubject(ctx context.Context, db *sqlx.DB, providerID uuid.UUID, subject string) (*UserIdentity, error) {
	i := &UserIdentity{}
	err := db.GetContext(ctx, i,
		`SELECT * FROM user_identities WHERE provider_id = $1 AND subject = $2`,
		providerID, subject)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIdentityNotFound
	}
	return i, err
}

func GetIdentityByUserAndProvider(ctx context.Context, db *sqlx.DB, userID, providerID uuid.UUID) (*UserIdentity, error) {
	i := &UserIdentity{}
	err := db.GetContext(ctx, i,
		`SELECT * FROM user_identities WHERE user_id = $1 AND provider_id = $2`,
		userID, providerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIdentityNotFound
	}
	return i, err
}

func ListUserIdentities(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*UserIdentity, error) {
	var identities []*UserIdentity
	err := db.SelectContext(ctx, &identities,
		`SELECT * FROM user_identities WHERE user_id = $1 ORDER BY created_at`, userID)
	return identities, err
}

func UpdateIdentityLogin(ctx context.Context, db *sqlx.DB, id uuid.UUID, email, name, picture *string, rawClaims json.RawMessage) error {
	_, err := db.ExecContext(ctx, `
		UPDATE user_identities
		SET email = $2, name = $3, picture = $4, raw_claims = $5, last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1`,
		id, email, name, picture, rawClaims)
	return err
}

func DeleteIdentity(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM user_identities WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrIdentityNotFound
	}
	return nil
}

func DeleteUserIdentities(ctx context.Context, db *sqlx.DB, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM user_identities WHERE user_id = $1`, userID)
	return err
}

// OIDC State Management

func CreateOIDCState(ctx context.Context, db *sqlx.DB, providerID uuid.UUID, redirectURI *string) (string, error) {
	state, err := generateState()
	if err != nil {
		return "", err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO oidc_states (state, provider_id, redirect_uri)
		VALUES ($1, $2, $3)`,
		state, providerID, redirectURI)

	return state, err
}

func ValidateOIDCState(ctx context.Context, db *sqlx.DB, state string) (*OIDCState, error) {
	s := &OIDCState{}
	err := db.GetContext(ctx, s, `SELECT * FROM oidc_states WHERE state = $1`, state)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStateNotFound
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(s.ExpiresAt) {
		_ = DeleteOIDCState(ctx, db, state)
		return nil, ErrStateExpired
	}
	return s, nil
}

func DeleteOIDCState(ctx context.Context, db *sqlx.DB, state string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM oidc_states WHERE state = $1`, state)
	return err
}

func CleanupExpiredStates(ctx context.Context, db *sqlx.DB) (int64, error) {
	result, err := db.ExecContext(ctx, `DELETE FROM oidc_states WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
