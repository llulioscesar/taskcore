package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/auth"
	"github.com/start-codex/taskcode/internal/role"
)

// canEditUser checks if currentUser can edit targetUser (self or has manage permission)
func canEditUser(ctx context.Context, db *sqlx.DB, currentUserID, targetUserID uuid.UUID) (bool, error) {
	if currentUserID == targetUserID {
		return true, nil
	}
	return auth.HasGlobalPermission(ctx, db, currentUserID, role.GlobalPermManageUsers)
}

// hasBrowseOrManageUsers checks if user can browse users
func hasBrowseOrManageUsers(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (bool, error) {
	hasBrowse, err := auth.HasGlobalPermission(ctx, db, userID, role.GlobalPermBrowseUsers)
	if err != nil {
		return false, err
	}
	if hasBrowse {
		return true, nil
	}
	return auth.HasGlobalPermission(ctx, db, userID, role.GlobalPermManageUsers)
}
