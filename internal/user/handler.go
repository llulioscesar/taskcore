package user

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
	mux.HandleFunc("GET /api/v1/users", h.List)
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

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    *string   `json:"avatar"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt string    `json:"created_at"`
}

func toResponse(u *User) userResponse {
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := List(r.Context(), h.db)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	resp := make([]userResponse, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := GetByID(r.Context(), h.db, id)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(u))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		server.Error(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	// TODO: hash password properly with bcrypt
	u := &User{
		Email:        req.Email,
		PasswordHash: req.Password, // placeholder
		Name:         req.Name,
		IsActive:     true,
	}

	err := Create(r.Context(), h.db, u)
	if errors.Is(err, ErrEmailExists) {
		server.Error(w, http.StatusConflict, "email already exists")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	server.Created(w, "/api/v1/users/"+u.ID.String(), toResponse(u))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req updateRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := GetByID(r.Context(), h.db, id)
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

	if err := Update(r.Context(), h.db, u); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(u))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = Delete(r.Context(), h.db, id)
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
