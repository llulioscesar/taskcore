package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/server"
	"golang.org/x/oauth2"
)

// OpenIDHandler handles OpenID Connect authentication
type OpenIDHandler struct {
	db      *sqlx.DB
	baseURL string
}

func NewOpenIDHandler(db *sqlx.DB, baseURL string) *OpenIDHandler {
	return &OpenIDHandler{
		db:      db,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *OpenIDHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/auth/providers", h.ListProviders)
	mux.HandleFunc("GET /api/v1/auth/openid/{provider}", h.Authorize)
	mux.HandleFunc("GET /api/v1/auth/openid/{provider}/callback", h.Callback)
}

type providerResponse struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (h *OpenIDHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := ListActiveProviders(r.Context(), h.db)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to list providers")
		return
	}

	resp := make([]providerResponse, len(providers))
	for i, p := range providers {
		resp[i] = providerResponse{Slug: p.Slug, Name: p.Name}
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *OpenIDHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("provider")

	provider, err := GetProviderBySlug(r.Context(), h.db, slug)
	if errors.Is(err, ErrProviderNotFound) {
		server.Error(w, http.StatusNotFound, "provider not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get provider")
		return
	}

	redirectURI := r.URL.Query().Get("redirect_uri")
	var redirectPtr *string
	if redirectURI != "" {
		redirectPtr = &redirectURI
	}

	state, err := CreateOIDCState(r.Context(), h.db, provider.ID, redirectPtr)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create state")
		return
	}

	oauth2Config, err := h.getOAuth2Config(r.Context(), provider)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to configure provider")
		return
	}

	authURL := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *OpenIDHandler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("provider")

	// Validate state
	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		server.Error(w, http.StatusBadRequest, "missing state parameter")
		return
	}

	oidcState, err := ValidateOIDCState(ctx, h.db, stateParam)
	if errors.Is(err, ErrStateNotFound) || errors.Is(err, ErrStateExpired) {
		server.Error(w, http.StatusBadRequest, "invalid or expired state")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to validate state")
		return
	}
	_ = DeleteOIDCState(ctx, h.db, stateParam)

	// Get provider
	provider, err := GetProviderBySlug(ctx, h.db, slug)
	if errors.Is(err, ErrProviderNotFound) {
		server.Error(w, http.StatusNotFound, "provider not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get provider")
		return
	}

	if provider.ID != oidcState.ProviderID {
		server.Error(w, http.StatusBadRequest, "state/provider mismatch")
		return
	}

	// Check for error from provider
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		if errDesc == "" {
			errDesc = errParam
		}
		server.Error(w, http.StatusUnauthorized, errDesc)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		server.Error(w, http.StatusBadRequest, "missing code parameter")
		return
	}

	oauth2Config, err := h.getOAuth2Config(ctx, provider)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to configure provider")
		return
	}

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		server.Error(w, http.StatusUnauthorized, "failed to exchange code")
		return
	}

	// Verify ID token
	oidcProvider, err := oidc.NewProvider(ctx, provider.Issuer)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create OIDC provider")
		return
	}

	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		server.Error(w, http.StatusUnauthorized, "no id_token in response")
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		server.Error(w, http.StatusUnauthorized, "failed to verify id_token")
		return
	}

	// Extract claims
	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to parse claims")
		return
	}

	rawClaims, _ := json.Marshal(claims)

	// Find or create user identity
	identity, err := GetIdentityBySubject(ctx, h.db, provider.ID, claims.Subject)
	if errors.Is(err, ErrIdentityNotFound) {
		// Identity doesn't exist - check if we should auto-create
		if !provider.AutoCreateUsers {
			server.Error(w, http.StatusForbidden, "user not registered with this provider")
			return
		}

		// Auto-create user and identity
		identity, err = h.createUserAndIdentity(ctx, provider, claims.Subject, claims.Email, claims.Name, claims.Picture, rawClaims)
		if err != nil {
			server.Error(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	} else if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to lookup identity")
		return
	} else {
		// Update identity with latest info
		var emailPtr, namePtr, picturePtr *string
		if claims.Email != "" {
			emailPtr = &claims.Email
		}
		if claims.Name != "" {
			namePtr = &claims.Name
		}
		if claims.Picture != "" {
			picturePtr = &claims.Picture
		}
		_ = UpdateIdentityLogin(ctx, h.db, identity.ID, emailPtr, namePtr, picturePtr, rawClaims)
	}

	// Get user info
	userInfo, err := getUserByID(ctx, h.db, identity.UserID)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if !userInfo.IsActive {
		server.Error(w, http.StatusUnauthorized, "account is disabled")
		return
	}

	// Create session
	session, err := CreateSession(ctx, h.db, identity.UserID, SessionDuration)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Return tokens (or redirect if redirect_uri was provided)
	if oidcState.RedirectURI != nil && *oidcState.RedirectURI != "" {
		redirectURL := *oidcState.RedirectURI + "?token=" + session.RefreshToken
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	server.JSON(w, http.StatusOK, map[string]any{
		"user":          toUserResponse(userInfo),
		"access_token":  session.RefreshToken,
		"refresh_token": session.RefreshToken,
		"expires_at":    session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *OpenIDHandler) getOAuth2Config(ctx context.Context, provider *IdentityProvider) (*oauth2.Config, error) {
	oidcProvider, err := oidc.NewProvider(ctx, provider.Issuer)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  h.baseURL + "/api/v1/auth/openid/" + provider.Slug + "/callback",
		Scopes:       provider.Scopes,
	}, nil
}

func (h *OpenIDHandler) createUserAndIdentity(ctx context.Context, provider *IdentityProvider, subject, email, name, picture string, rawClaims json.RawMessage) (*UserIdentity, error) {
	// This would need to create a user first
	// For now, return an error - the actual implementation depends on how users are created
	return nil, errors.New("auto-create users not implemented - link identity to existing user required")
}

// LinkIdentity links an existing user to an identity provider
func LinkIdentity(ctx context.Context, db *sqlx.DB, userID, providerID uuid.UUID, subject, email, name, picture string, rawClaims json.RawMessage) (*UserIdentity, error) {
	var emailPtr, namePtr, picturePtr *string
	if email != "" {
		emailPtr = &email
	}
	if name != "" {
		namePtr = &name
	}
	if picture != "" {
		picturePtr = &picture
	}

	identity := &UserIdentity{
		UserID:     userID,
		ProviderID: providerID,
		Subject:    subject,
		Email:      emailPtr,
		Name:       namePtr,
		Picture:    picturePtr,
		RawClaims:  rawClaims,
	}

	err := CreateIdentity(ctx, db, identity)
	return identity, err
}
