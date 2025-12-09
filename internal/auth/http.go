package auth

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/server"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("PUT /api/v1/auth/password", h.ChangePassword)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    *string   `json:"avatar,omitempty"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt string    `json:"created_at"`
}

type loginResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func toUserResponse(u *UserInfo) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Avatar:    u.Avatar,
		IsActive:  u.IsActive,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		server.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	u, err := getUserByEmail(r.Context(), h.db, req.Email)
	if errors.Is(err, ErrUserNotFound) {
		server.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to authenticate")
		return
	}

	if !u.IsActive {
		server.Error(w, http.StatusUnauthorized, "account is disabled")
		return
	}

	if !CheckPassword(req.Password, u.PasswordHash) {
		server.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	session, err := CreateSession(r.Context(), h.db, u.ID, SessionDuration)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	server.JSON(w, http.StatusOK, loginResponse{
		User:         toUserResponse(u),
		AccessToken:  session.RefreshToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := ExtractToken(r)
	if token == "" {
		server.Error(w, http.StatusUnauthorized, "missing authorization token")
		return
	}

	session, err := GetSessionByToken(r.Context(), h.db, token)
	if err != nil {
		server.NoContent(w)
		return
	}

	_ = DeleteSession(r.Context(), h.db, session.ID)
	server.NoContent(w)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		server.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	oldSession, err := GetSessionByToken(r.Context(), h.db, req.RefreshToken)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
		server.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to validate token")
		return
	}

	u, err := getUserByID(r.Context(), h.db, oldSession.UserID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if !u.IsActive {
		server.Error(w, http.StatusUnauthorized, "account is disabled")
		return
	}

	_ = DeleteSession(r.Context(), h.db, oldSession.ID)

	newSession, err := CreateSession(r.Context(), h.db, u.ID, SessionDuration)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	server.JSON(w, http.StatusOK, loginResponse{
		User:         toUserResponse(u),
		AccessToken:  newSession.RefreshToken,
		RefreshToken: newSession.RefreshToken,
		ExpiresAt:    newSession.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	var req changePasswordRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentPassword == "" {
		server.Error(w, http.StatusBadRequest, "current_password is required")
		return
	}

	if req.NewPassword == "" {
		server.Error(w, http.StatusBadRequest, "new_password is required")
		return
	}

	if len(req.NewPassword) < 8 {
		server.Error(w, http.StatusBadRequest, "new password must be at least 8 characters")
		return
	}

	u, err := getUserByID(r.Context(), h.db, userID)
	if errors.Is(err, ErrUserNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if !CheckPassword(req.CurrentPassword, u.PasswordHash) {
		server.Error(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := updatePassword(r.Context(), h.db, userID, hash); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	server.NoContent(w)
}
