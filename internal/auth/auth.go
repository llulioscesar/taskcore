package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/server"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

// ExtractToken extracts Bearer token from Authorization header
func ExtractToken(r *http.Request) string {
	auth := r.Header.Get(AuthorizationHeader)
	if strings.HasPrefix(auth, BearerPrefix) {
		return strings.TrimPrefix(auth, BearerPrefix)
	}
	return ""
}

// Middleware validates the session token and adds user to context
func Middleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				server.Error(w, http.StatusUnauthorized, "missing authorization token")
				return
			}

			session, err := GetSessionByToken(r.Context(), db, token)
			if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
				server.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			if err != nil {
				server.Error(w, http.StatusInternalServerError, "failed to validate token")
				return
			}

			ctx := context.WithValue(r.Context(), server.CtxUserID, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth extracts user from token and returns user ID, or writes error response
func RequireAuth(db *sqlx.DB, w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	token := ExtractToken(r)
	if token == "" {
		server.Error(w, http.StatusUnauthorized, "missing authorization token")
		return uuid.Nil, false
	}

	session, err := GetSessionByToken(r.Context(), db, token)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
		server.Error(w, http.StatusUnauthorized, "invalid or expired token")
		return uuid.Nil, false
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to validate token")
		return uuid.Nil, false
	}

	return session.UserID, true
}
