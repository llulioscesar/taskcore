# Technical Architecture

## Technical goals

- Single binary, easy to deploy.
- Low memory and CPU footprint.
- Maintainable and extensible by domain.
- Modern UX without a heavy SPA.

---

## Current stack

**Backend**

- Go 1.26
- HTTP router: `net/http` stdlib (Go 1.22+ supports method routing and path params natively)
- Database: PostgreSQL
- SQL layer: `database/sql` + `sqlx`
- Migrations: `golang-migrate`

**Frontend**

- SvelteKit 2 + Svelte 5
- Vite
- Tailwind CSS 4
- Local shadcn-style components built on top of Bits UI primitives (no external shadcn-svelte library dependency)
- Paraglide for i18n (EN + ES)

The frontend is compiled and embedded into the Go binary at build time. The Go server serves the SvelteKit app and the API from a single process.

---

## Application style

- Modular monolith (no microservices in MVP).
- API REST mounted under `/api` (no version prefix in current routes).
- Issue belongs to a project and has a status; boards are views (filter + columns) over issues.

---

## Actual directory structure

```
cmd/
  server/
    main.go           # entrypoint: configures DB, router, starts server
    api.go            # newAPIHandler: assembles API sub-mux with all domain routes
    middleware.go     # withRequestID, withLogger, withRecover, withAuth
    static.go         # registerUI: serves embedded frontend assets

internal/
  authz/
    authz.go          # context helpers, RequireWorkspaceMembership, RequireWorkspaceAdmin
    store.go          # private SQL: membership checks, role queries
    authz_test.go
    store_integration_test.go

  sessions/
    sessions.go       # Create, Validate, Delete; SHA-256 token hashing
    store.go          # private SQL persistence
    sessions_test.go
    store_integration_test.go

  boards/
    boards.go         # types, errors, public API
    handler.go        # HTTP handlers + RegisterRoutes
    store.go          # private SQL persistence
    boards_test.go
    store_integration_test.go

  issues/
    issues.go
    handler.go
    store.go
    issues_test.go
    store_integration_test.go

  issuetypes/
    issuetypes.go
    handler.go
    store.go

  projects/
    projects.go
    handler.go
    store.go
    projects_test.go
    store_integration_test.go

  statuses/
    statuses.go
    handler.go
    store.go

  users/
    users.go
    handler.go
    store.go
    password.go       # Argon2id password hashing
    users_test.go
    store_integration_test.go

  workspaces/
    workspaces.go
    handler.go
    store.go
    workspaces_test.go
    store_integration_test.go

  respond/
    respond.go        # shared HTTP utilities (respond.JSON, respond.Error, respond.Decode)

  pgutil/
    pgutil.go         # shared PostgreSQL helpers

  testpg/
    testpg.go         # test helpers: Open, EnsureMigrated, SeedUser, SeedWorkspace, SeedProject

migrations/
  *.up.sql
  *.down.sql
  migrate.go          # embedded migrations, Up() function

ui/
  ui.go               # embedded frontend dist

front/
  src/
  package.json
  ...
```

---

## Handler pattern

Handlers are thin by design: parse the request, call the domain function, write the response.

Each domain package exposes a single `RegisterRoutes(mux *http.ServeMux, db *sqlx.DB)` function. `cmd/server/api.go` calls them in order. All handler functions are private (`handleCreate`, not `HandleCreate`).

Domain packages register paths **without** the `/api/` prefix. `cmd/server/main.go` mounts them on a sub-mux with `http.StripPrefix("/api", api)`, so the full public URL becomes `/api/projects/...`.

```go
func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
    mux.HandleFunc("POST /projects/{projectID}/issues", handleCreateIssue(db))
    mux.HandleFunc("GET /projects/{projectID}/issues/{issueID}", handleGetIssue(db))
}

func handleCreateIssue(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        projectID := r.PathValue("projectID")
        var p issues.CreateIssueParams
        if err := respond.Decode(r, &p); err != nil {
            respond.Error(w, http.StatusBadRequest, err)
            return
        }
        p.ProjectID = projectID
        issue, err := issues.CreateIssue(r.Context(), db, p)
        if err != nil {
            fail(w, err)
            return
        }
        respond.JSON(w, http.StatusCreated, issue)
    }
}
```

Each `handler.go` defines a local `fail(w, err)` function that maps domain sentinel errors to HTTP codes. Unknown errors → 500.

---

## Authentication and authorization

### Current state

- Server-side sessions with secure cookies: `HttpOnly`, `Secure` (conditional), `SameSite=Strict`, 7-day TTL.
- Session tokens are SHA-256 hashed before database storage; raw tokens never persisted.
- `POST /api/auth/login` — authenticates credentials, creates session, sets `session_id` cookie.
- `GET /api/auth/me` — always returns 200; `{ authenticated: true, user }` or `{ authenticated: false }`.
- `POST /api/auth/logout` — idempotent, deletes session, clears cookie, returns 204.
- `withAuth` middleware wraps the `/api` sub-mux. Allowlisted routes: `POST /users`, `POST /auth/login`, `GET /auth/me`, `POST /auth/logout`.
- `internal/authz` provides context helpers (`WithUserID`, `UserIDFromContext`) and authorization functions:
  - `RequireWorkspaceMembership` — verifies user is a workspace member.
  - `RequireWorkspaceAdmin` — verifies user has `admin` or `owner` role.
  - `RequireProjectMembership` — resolves project → workspace, then checks membership.
  - `RequireBoardAccess` — resolves board → project → workspace, then checks membership.
  - `RequireColumnAccess` — resolves column → board → project → workspace, then checks membership.
- Identity derived from session context — no client-controlled `owner_id`, `reporter_id`, or `user_id` in API contracts.
- Workspace admin/owner role required for administrative routes (archive workspace, manage members, create projects, configure workflow).
- Frontend auth uses in-memory store with `/auth/me` validation on page load; no `localStorage` for auth state.

### Planned (Phase 1.5)

- Transactional email: SMTP configuration, email templates, delivery pipeline.
- Password reset and change-password flows with expiring tokens.
- User invitations: invite by email, pending invite lifecycle, role assignment on acceptance.
- Federated identity: OpenID Connect (OIDC), external identity providers, JIT provisioning.
- Instance bootstrap: first-install flow, global administrator creation, system-wide configuration.

---

## Design rules

- One package per domain, not per technical layer.
- Do not create `internal/domain`, `internal/app`, or `internal/store` globals.
- Do not use OOP subdirectory patterns inside a domain (`repository/`, `service/`, `manager/`).
- Domain and persistence coexist in the same package:
  - `<domain>.go` — types, errors, validation, public API.
  - `store.go` — private SQL functions and persistence details.
- Prefer free functions with explicit dependencies (e.g. `func MoveIssue(ctx, db, p)`).
- Explicit SQL, testable against real PostgreSQL in integration tests.
- Interfaces only when there is a concrete need (not preventive).

---

## Observability

- Structured JSON logs via `log/slog`.
- Request ID per request (middleware in `cmd/server/middleware.go`).
- Static asset paths (`/_app/`) excluded from request logging to reduce noise.
- Minimal metrics target: latency, error rate, throughput.

---

## Security

- Input validation in the handler (before calling the domain function) and again inside the domain (`Validate()` as second line of defense).
- Server-side sessions with `HttpOnly`, `Secure`, `SameSite=Strict` cookies.
- Workspace and project membership enforced per handler using context user.
- Admin/owner role enforcement on all administrative and workflow configuration routes.
- CSRF protection for state-changing endpoints — planned.
- Password hashing: Argon2id (memory=64MB, iterations=3, parallelism=4, salt=16B, key=32B).

---

## Target architecture — documentation-led planning

In the documentation-led planning phase (Phase 3), the architecture extends to support project documentation alongside execution:

- Documentation pages belong to projects, stored in the same database.
- Pages and work items share explicit link records — no implicit coupling.
- Planning workflows (backlog refinement, sprint planning, reviews) reference documented context directly.
- The initial implementation is **user-driven and manual**: users create links, not the system.

This means no separate "wiki service" or external documentation product. Documentation is a first-class domain in the same monolith.

---

## Future architecture — AI assistant and MCP

In Phase 6, the architecture extends to support a workflow-oriented AI assistant. See [docs/06-ai-assistant.md](06-ai-assistant.md) for full details.

- The assistant runs within the same monolith, not as a separate service.
- AI provider is configurable per workspace or instance (OpenAI, Anthropic, Google, Ollama, or any compatible API).
- The assistant executes operations under the authenticated user's session and permissions — no AI superuser.
- MCP (Model Context Protocol) provides a connector layer for reading and acting on external systems.
- Without a configured provider, AI features are unavailable but Tookly works normally.

---

## Planned additions

- Sprint and backlog endpoints (`internal/sprints/`).
- Project templates API (extend `internal/projects/` or add `internal/templates/`).
- Notification domain (`internal/notifications/`).
- Project documentation pages domain (`internal/pages/`).
- AI assistant domain and MCP connector layer.
- Transactional email pipeline.
- Instance configuration and administration.
