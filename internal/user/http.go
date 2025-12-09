package user

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/auth"
	"github.com/start-codex/taskcode/internal/role"
	"github.com/start-codex/taskcode/internal/server"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/users", h.List)
	mux.HandleFunc("GET /api/v1/users/me", h.Me)
	mux.HandleFunc("GET /api/v1/users/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/users", h.Create)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.Delete)
}

type createRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type updateRequest struct {
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Avatar *string `json:"avatar"`
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    *string   `json:"avatar,omitempty"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt string    `json:"created_at"`
}

func ToResponse(u *User) Response {
	return Response{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Avatar:    u.Avatar,
		IsActive:  u.IsActive,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	u, err := getByID(r.Context(), h.db, userID)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	server.JSON(w, http.StatusOK, ToResponse(u))
}

// User CRUD handlers

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	canBrowse, err := hasBrowseOrManageUsers(r.Context(), h.db, currentUserID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to check permissions")
		return
	}
	if !canBrowse {
		server.Error(w, http.StatusForbidden, "permission denied: browse_users required")
		return
	}

	users, err := list(r.Context(), h.db)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	resp := make([]Response, len(users))
	for i, u := range users {
		resp[i] = ToResponse(u)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if currentUserID != id {
		canBrowse, err := hasBrowseOrManageUsers(r.Context(), h.db, currentUserID)
		if err != nil {
			server.Error(w, http.StatusInternalServerError, "failed to check permissions")
			return
		}
		if !canBrowse {
			server.Error(w, http.StatusForbidden, "permission denied: browse_users required")
			return
		}
	}

	u, err := getByID(r.Context(), h.db, id)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	server.JSON(w, http.StatusOK, ToResponse(u))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	if !auth.RequireGlobalPermission(r.Context(), h.db, w, currentUserID, role.GlobalPermManageUsers) {
		return
	}

	var req createRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		server.Error(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	if len(req.Password) < 8 {
		server.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		IsActive:     true,
	}

	err = create(r.Context(), h.db, u)
	if errors.Is(err, ErrEmailExists) {
		server.Error(w, http.StatusConflict, "email already exists")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	server.Created(w, "/api/v1/users/"+u.ID.String(), ToResponse(u))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	canEdit, err := canEditUser(r.Context(), h.db, currentUserID, id)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to check permissions")
		return
	}
	if !canEdit {
		server.Error(w, http.StatusForbidden, "permission denied: can only edit your own profile or need manage_users permission")
		return
	}

	var req updateRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := getByID(r.Context(), h.db, id)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	u.Avatar = req.Avatar

	if err := update(r.Context(), h.db, u); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	server.JSON(w, http.StatusOK, ToResponse(u))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := auth.RequireAuth(h.db, w, r)
	if !ok {
		return
	}

	if !auth.RequireGlobalPermission(r.Context(), h.db, w, currentUserID, role.GlobalPermManageUsers) {
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if currentUserID == id {
		server.Error(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	err = delete_(r.Context(), h.db, id)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	server.NoContent(w)
}
