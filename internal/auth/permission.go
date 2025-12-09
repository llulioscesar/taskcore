package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/permission"
	"github.com/start-codex/taskcode/internal/role"
	"github.com/start-codex/taskcode/internal/server"
)

// RequireGlobalPermission checks if user has a specific global permission
// Returns false and writes error response if permission denied
func RequireGlobalPermission(ctx context.Context, db *sqlx.DB, w http.ResponseWriter, userID uuid.UUID, perm role.GlobalPermission) bool {
	hasPermission, err := role.HasGlobalPermission(ctx, db, userID, perm)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to check permissions")
		return false
	}
	if !hasPermission {
		server.Error(w, http.StatusForbidden, "permission denied")
		return false
	}
	return true
}

// RequireProjectPermission checks if user has a specific permission on a project
// Returns false and writes error response if permission denied
func RequireProjectPermission(ctx context.Context, db *sqlx.DB, w http.ResponseWriter, userID, projectID uuid.UUID, perm permission.Permission) bool {
	hasPermission, err := permission.HasProjectPermission(ctx, db, userID, projectID, perm)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to check permissions")
		return false
	}
	if !hasPermission {
		server.Error(w, http.StatusForbidden, "permission denied")
		return false
	}
	return true
}

// HasGlobalPermission checks if user has a specific global permission (without writing response)
func HasGlobalPermission(ctx context.Context, db *sqlx.DB, userID uuid.UUID, perm role.GlobalPermission) (bool, error) {
	return role.HasGlobalPermission(ctx, db, userID, perm)
}

// HasProjectPermission checks if user has a specific permission on a project (without writing response)
func HasProjectPermission(ctx context.Context, db *sqlx.DB, userID, projectID uuid.UUID, perm permission.Permission) (bool, error) {
	return permission.HasProjectPermission(ctx, db, userID, projectID, perm)
}

// IsSystemAdmin checks if user is a system administrator
func IsSystemAdmin(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (bool, error) {
	return role.IsSystemAdmin(ctx, db, userID)
}
