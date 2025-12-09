package project

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
	mux.HandleFunc("GET /api/v1/projects", h.List)
	mux.HandleFunc("GET /api/v1/projects/{key}", h.Get)
	mux.HandleFunc("POST /api/v1/projects", h.Create)
	mux.HandleFunc("PUT /api/v1/projects/{key}", h.Update)
	mux.HandleFunc("DELETE /api/v1/projects/{key}", h.Delete)

	// Members
	mux.HandleFunc("GET /api/v1/projects/{key}/members", h.ListMembers)
	mux.HandleFunc("POST /api/v1/projects/{key}/members", h.AddMember)
	mux.HandleFunc("DELETE /api/v1/projects/{key}/members/{memberID}", h.RemoveMember)
}

type createRequest struct {
	Key                string     `json:"key"`
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	LeadID             uuid.UUID  `json:"lead_id"`
	TemplateID         *uuid.UUID `json:"template_id"`
	PermissionSchemeID uuid.UUID  `json:"permission_scheme_id"`
	WorkflowID         uuid.UUID  `json:"workflow_id"`
}

type updateRequest struct {
	Name               *string    `json:"name"`
	Description        *string    `json:"description"`
	LeadID             *uuid.UUID `json:"lead_id"`
	DefaultAssigneeID  *uuid.UUID `json:"default_assignee_id"`
	PermissionSchemeID *uuid.UUID `json:"permission_scheme_id"`
	WorkflowID         *uuid.UUID `json:"workflow_id"`
}

type addMemberRequest struct {
	UserID  *uuid.UUID `json:"user_id"`
	GroupID *uuid.UUID `json:"group_id"`
	Role    Role       `json:"role"`
}

type projectResponse struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	LeadID      uuid.UUID  `json:"lead_id"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type memberResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	GroupID   *uuid.UUID `json:"group_id,omitempty"`
	Role      Role       `json:"role"`
	CreatedAt string     `json:"created_at"`
}

func toResponse(p *Project) projectResponse {
	return projectResponse{
		ID:          p.ID,
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		LeadID:      p.LeadID,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toMemberResponse(m *Member) memberResponse {
	return memberResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		GroupID:   m.GroupID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// If user_id query param, filter by user
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			server.Error(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		projects, err := ListByUser(ctx, h.db, userID)
		if err != nil {
			server.Error(w, http.StatusInternalServerError, "failed to list projects")
			return
		}
		resp := make([]projectResponse, len(projects))
		for i, p := range projects {
			resp[i] = toResponse(p)
		}
		server.JSON(w, http.StatusOK, resp)
		return
	}

	projects, err := List(ctx, h.db)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	resp := make([]projectResponse, len(projects))
	for i, p := range projects {
		resp[i] = toResponse(p)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(p))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Key == "" || req.Name == "" {
		server.Error(w, http.StatusBadRequest, "key and name are required")
		return
	}

	p := &Project{
		TemplateID:         req.TemplateID,
		Key:                req.Key,
		Name:               req.Name,
		Description:        req.Description,
		LeadID:             req.LeadID,
		PermissionSchemeID: req.PermissionSchemeID,
		WorkflowID:         req.WorkflowID,
	}

	err := Create(r.Context(), h.db, p)
	if errors.Is(err, ErrKeyExists) {
		server.Error(w, http.StatusConflict, "project key already exists")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	server.Created(w, "/api/v1/projects/"+p.Key, toResponse(p))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req updateRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = req.Description
	}
	if req.LeadID != nil {
		p.LeadID = *req.LeadID
	}
	if req.DefaultAssigneeID != nil {
		p.DefaultAssigneeID = req.DefaultAssigneeID
	}
	if req.PermissionSchemeID != nil {
		p.PermissionSchemeID = *req.PermissionSchemeID
	}
	if req.WorkflowID != nil {
		p.WorkflowID = *req.WorkflowID
	}

	if err := Update(r.Context(), h.db, p); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(p))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	if err := Delete(r.Context(), h.db, p.ID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	server.NoContent(w)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	members, err := ListMembers(r.Context(), h.db, p.ID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	resp := make([]memberResponse, len(members))
	for i, m := range members {
		resp[i] = toMemberResponse(m)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req addMemberRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == nil && req.GroupID == nil {
		server.Error(w, http.StatusBadRequest, "user_id or group_id is required")
		return
	}

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	if req.UserID != nil {
		err = AddUserMember(r.Context(), h.db, p.ID, *req.UserID, req.Role)
	} else {
		err = AddGroupMember(r.Context(), h.db, p.ID, *req.GroupID, req.Role)
	}

	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	server.NoContent(w)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	memberIDStr := r.PathValue("memberID")

	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid member id")
		return
	}

	p, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	// Try user first, then group
	_ = RemoveUserMember(r.Context(), h.db, p.ID, memberID)
	_ = RemoveGroupMember(r.Context(), h.db, p.ID, memberID)

	server.NoContent(w)
}
