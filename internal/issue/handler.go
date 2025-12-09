package issue

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
	// Issues
	mux.HandleFunc("GET /api/v1/issues/{key}", h.Get)
	mux.HandleFunc("POST /api/v1/issues", h.Create)
	mux.HandleFunc("PUT /api/v1/issues/{key}", h.Update)
	mux.HandleFunc("DELETE /api/v1/issues/{key}", h.Delete)

	// By project
	mux.HandleFunc("GET /api/v1/projects/{projectKey}/issues", h.ListByProject)
	mux.HandleFunc("GET /api/v1/projects/{projectKey}/backlog", h.ListBacklog)
	mux.HandleFunc("GET /api/v1/projects/{projectKey}/epics", h.ListEpics)

	// Comments
	mux.HandleFunc("GET /api/v1/issues/{key}/comments", h.ListComments)
	mux.HandleFunc("POST /api/v1/issues/{key}/comments", h.CreateComment)
	mux.HandleFunc("PUT /api/v1/issues/{key}/comments/{commentID}", h.UpdateComment)
	mux.HandleFunc("DELETE /api/v1/issues/{key}/comments/{commentID}", h.DeleteComment)

	// Labels
	mux.HandleFunc("GET /api/v1/issues/{key}/labels", h.ListIssueLabels)
	mux.HandleFunc("POST /api/v1/issues/{key}/labels/{labelID}", h.AddLabel)
	mux.HandleFunc("DELETE /api/v1/issues/{key}/labels/{labelID}", h.RemoveLabel)

	// Watchers
	mux.HandleFunc("GET /api/v1/issues/{key}/watchers", h.ListWatchers)
	mux.HandleFunc("POST /api/v1/issues/{key}/watch", h.Watch)
	mux.HandleFunc("DELETE /api/v1/issues/{key}/watch", h.Unwatch)

	// Transitions
	mux.HandleFunc("POST /api/v1/issues/{key}/transitions", h.Transition)
}

type createRequest struct {
	ProjectID   uuid.UUID  `json:"project_id"`
	IssueTypeID uuid.UUID  `json:"issue_type_id"`
	Summary     string     `json:"summary"`
	Description *string    `json:"description"`
	Priority    Priority   `json:"priority"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	SprintID    *uuid.UUID `json:"sprint_id"`
	EpicID      *uuid.UUID `json:"epic_id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	StoryPoints *int       `json:"story_points"`
}

type updateRequest struct {
	Summary     *string    `json:"summary"`
	Description *string    `json:"description"`
	Priority    *Priority  `json:"priority"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	SprintID    *uuid.UUID `json:"sprint_id"`
	EpicID      *uuid.UUID `json:"epic_id"`
	StoryPoints *int       `json:"story_points"`
}

type transitionRequest struct {
	StatusID uuid.UUID `json:"status_id"`
}

type commentRequest struct {
	Content string `json:"content"`
}

type issueResponse struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Summary     string     `json:"summary"`
	Description *string    `json:"description,omitempty"`
	Priority    Priority   `json:"priority"`
	StatusID    uuid.UUID  `json:"status_id"`
	IssueTypeID uuid.UUID  `json:"issue_type_id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	ReporterID  uuid.UUID  `json:"reporter_id"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty"`
	EpicID      *uuid.UUID `json:"epic_id,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	StoryPoints *int       `json:"story_points,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type commentResponse struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func toResponse(i *Issue) issueResponse {
	return issueResponse{
		ID:          i.ID,
		Key:         i.Key,
		Summary:     i.Summary,
		Description: i.Description,
		Priority:    i.Priority,
		StatusID:    i.StatusID,
		IssueTypeID: i.IssueTypeID,
		ProjectID:   i.ProjectID,
		ReporterID:  i.ReporterID,
		AssigneeID:  i.AssigneeID,
		SprintID:    i.SprintID,
		EpicID:      i.EpicID,
		ParentID:    i.ParentID,
		StoryPoints: i.StoryPoints,
		CreatedAt:   i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toCommentResponse(c *Comment) commentResponse {
	return commentResponse{
		ID:        c.ID,
		AuthorID:  c.AuthorID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(issue))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Summary == "" {
		server.Error(w, http.StatusBadRequest, "summary is required")
		return
	}

	// Get reporter from context (authenticated user)
	reporterID, ok := server.GetUserID(r.Context())
	if !ok {
		server.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// TODO: Get next issue key from project
	// key, _ := project.NextIssueKey(r.Context(), h.db, req.ProjectID)

	issue := &Issue{
		ProjectID:   req.ProjectID,
		IssueTypeID: req.IssueTypeID,
		Summary:     req.Summary,
		Description: req.Description,
		Priority:    req.Priority,
		ReporterID:  reporterID,
		AssigneeID:  req.AssigneeID,
		SprintID:    req.SprintID,
		EpicID:      req.EpicID,
		ParentID:    req.ParentID,
		StoryPoints: req.StoryPoints,
		// TODO: Get default status from workflow
	}

	if err := Create(r.Context(), h.db, issue); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create issue")
		return
	}

	server.Created(w, "/api/v1/issues/"+issue.Key, toResponse(issue))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req updateRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	if req.Summary != nil {
		issue.Summary = *req.Summary
	}
	if req.Description != nil {
		issue.Description = req.Description
	}
	if req.Priority != nil {
		issue.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		issue.AssigneeID = req.AssigneeID
	}
	if req.SprintID != nil {
		issue.SprintID = req.SprintID
	}
	if req.EpicID != nil {
		issue.EpicID = req.EpicID
	}
	if req.StoryPoints != nil {
		issue.StoryPoints = req.StoryPoints
	}

	if err := Update(r.Context(), h.db, issue); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update issue")
		return
	}

	server.JSON(w, http.StatusOK, toResponse(issue))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	if err := Delete(r.Context(), h.db, issue.ID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to delete issue")
		return
	}

	server.NoContent(w)
}

func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectKey := r.PathValue("projectKey")

	// TODO: Get project by key to get ID
	// For now, try parsing as UUID
	projectID, err := uuid.Parse(projectKey)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid project key")
		return
	}

	issues, err := ListByProject(r.Context(), h.db, projectID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list issues")
		return
	}

	resp := make([]issueResponse, len(issues))
	for i, issue := range issues {
		resp[i] = toResponse(issue)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) ListBacklog(w http.ResponseWriter, r *http.Request) {
	projectKey := r.PathValue("projectKey")

	projectID, err := uuid.Parse(projectKey)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid project key")
		return
	}

	issues, err := ListBacklog(r.Context(), h.db, projectID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list backlog")
		return
	}

	resp := make([]issueResponse, len(issues))
	for i, issue := range issues {
		resp[i] = toResponse(issue)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) ListEpics(w http.ResponseWriter, r *http.Request) {
	projectKey := r.PathValue("projectKey")

	projectID, err := uuid.Parse(projectKey)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid project key")
		return
	}

	issues, err := ListEpics(r.Context(), h.db, projectID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list epics")
		return
	}

	resp := make([]issueResponse, len(issues))
	for i, issue := range issues {
		resp[i] = toResponse(issue)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Transition(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req transitionRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	// TODO: Validate transition is allowed by workflow

	if err := UpdateStatus(r.Context(), h.db, issue.ID, req.StatusID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to transition issue")
		return
	}

	issue.StatusID = req.StatusID
	server.JSON(w, http.StatusOK, toResponse(issue))
}

// Comments

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	comments, err := ListComments(r.Context(), h.db, issue.ID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	resp := make([]commentResponse, len(comments))
	for i, c := range comments {
		resp[i] = toCommentResponse(c)
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req commentRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Content == "" {
		server.Error(w, http.StatusBadRequest, "content is required")
		return
	}

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	authorID, ok := server.GetUserID(r.Context())
	if !ok {
		server.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	comment := &Comment{
		IssueID:  issue.ID,
		AuthorID: authorID,
		Content:  req.Content,
	}

	if err := CreateComment(r.Context(), h.db, comment); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	server.Created(w, "", toCommentResponse(comment))
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := uuid.Parse(r.PathValue("commentID"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	var req commentRequest
	if err := server.Decode(r, &req); err != nil {
		server.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := GetComment(r.Context(), h.db, commentID)
	if errors.Is(err, ErrCommentNotFound) {
		server.Error(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get comment")
		return
	}

	comment.Content = req.Content

	if err := UpdateComment(r.Context(), h.db, comment); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	server.JSON(w, http.StatusOK, toCommentResponse(comment))
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := uuid.Parse(r.PathValue("commentID"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	if err := DeleteComment(r.Context(), h.db, commentID); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			server.Error(w, http.StatusNotFound, "comment not found")
			return
		}
		server.Error(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	server.NoContent(w)
}

// Labels

func (h *Handler) ListIssueLabels(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	labels, err := ListIssueLabels(r.Context(), h.db, issue.ID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list labels")
		return
	}

	server.JSON(w, http.StatusOK, labels)
}

func (h *Handler) AddLabel(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	labelID, err := uuid.Parse(r.PathValue("labelID"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid label id")
		return
	}

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	if err := AddLabel(r.Context(), h.db, issue.ID, labelID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to add label")
		return
	}

	server.NoContent(w)
}

func (h *Handler) RemoveLabel(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	labelID, err := uuid.Parse(r.PathValue("labelID"))
	if err != nil {
		server.Error(w, http.StatusBadRequest, "invalid label id")
		return
	}

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	if err := RemoveLabel(r.Context(), h.db, issue.ID, labelID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to remove label")
		return
	}

	server.NoContent(w)
}

// Watchers

func (h *Handler) ListWatchers(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	watchers, err := ListWatchers(r.Context(), h.db, issue.ID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list watchers")
		return
	}

	server.JSON(w, http.StatusOK, watchers)
}

func (h *Handler) Watch(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	userID, ok := server.GetUserID(r.Context())
	if !ok {
		server.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := AddWatcher(r.Context(), h.db, issue.ID, userID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to watch issue")
		return
	}

	server.NoContent(w)
}

func (h *Handler) Unwatch(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	issue, err := GetByKey(r.Context(), h.db, key)
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, "issue not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get issue")
		return
	}

	userID, ok := server.GetUserID(r.Context())
	if !ok {
		server.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := RemoveWatcher(r.Context(), h.db, issue.ID, userID); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to unwatch issue")
		return
	}

	server.NoContent(w)
}
