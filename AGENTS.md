# Polaris — AI Agent Guide

## Overview
Social platform backend in Go (Gin/pgx/Redis), frontend in Svelte 5 + Vite (SPA). Go 1.26, PostgreSQL 17, Redis 7. Single binary monolith + background workers.

## How to Run

| Component | Command | URL |
|-----------|---------|-----|
| Infra | `docker compose up -d` | Postgres :5433 (host) → :5432, Redis :6379 |
| Backend | `go run ./cmd/polaris` | `http://localhost:8080` |
| Frontend | `cd web && npm run dev` | `http://localhost:5173` |
| Tests | `DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go test ./... -count=1` | — |

**Important**: The `.gitignore` is bare (only 3 swap-file entries). The first AI agent should add: `bin/`, `*.exe`, `*.test`, `.env`, `node_modules/`, `dist/`, `tmp/`, `*.log`, `.DS_Store`.

## Architecture Gaps (Known Issues for AI Agents)

### Critical (will fail if an agent doesn't fix first)
1. **No CORS middleware** — Gin router has no CORS config. Frontend on `:5173` calling backend on `:8080` will hit CORS errors. First backend agent must add `github.com/gin-contrib/cors` and configure it.
2. **No Vite proxy** — `web/vite.config.js` is bare with no proxy config. When the frontend runs on `:5173`, API calls to `http://localhost:8080` will cross origin. Either fix CORS on the Go side (preferred) or add `server.proxy` in vite.config.js.
3. **Workers incomplete** — `internal/worker/` has only `credit.go`, `trending.go`, `worker.go`. Missing: `scheduler.go` (main orchestration), `feed_worker.go`, `trust_decay.go`, `achievement_worker.go`, `scheduled_post_worker.go`, `community_notes_worker.go`, `active_stats_worker.go`, `prune_worker.go`. The AI that wrote them duplicated some files.
4. **main.go doesn't start workers** — `cmd/polaris/main.go` creates the HTTP server but never initializes or runs any background workers. A worker agent needs to add worker startup (likely a scheduler in `internal/worker/scheduler.go`).

### Moderate
5. **Handler code quality varies** — Some handler files from subagents don't match the exact method signatures expected by `router.go`. There will be compilation errors when adding missing handlers. Check `go build ./...` after any handler changes.
6. **Test coverage is thin** — 14 integration tests in `handlers/handlers_test.go` cover basic CRUD but not edge cases, error paths, or any service-level unit tests. Uses `TestMain` with migration and shared router.
7. **No Dockerfile for production** — The repo needs a multi-stage Dockerfile: build stage (golang:1.26-alpine) with `CGO_ENABLED=0`, runtime stage (scratch or alpine), exposes :8080, uses `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET` env vars.

## Architecture Conventions

### Backend Layers
```
router.go → handlers/*.go → services/*.go → db (pgx SQL) → PostgreSQL
                                       ↕
                                  model/*.go
```

- **Handlers** parse HTTP request, call service, return JSON response. All handlers in `internal/api/handlers/`. Use `gin.H` for simple responses.
- **Services** contain business logic, raw SQL via pgx. All services in `internal/services/`. Struct per service with `pg *pgxpool.Pool`.
- **Models** are pure data structs with JSON tags in `internal/model/`. Error sentinels in `errors.go`.
- **Middleware** in `internal/api/middleware/`. Currently only `auth.go` (JWT validation).

### Conventions to Follow
- **No ORMs** — raw SQL with pgx. All queries inline or in `internal/services/`.
- **No comments in code** — the design doc and AGENTS.md are the documentation.
- **JWT auth** — token in `Authorization: Bearer <token>` header. Set `user_id` in Gin context via middleware. Extract with `c.Get("user_id")`.
- **Error handling** — services return Go errors. Handlers check and return `c.JSON(status, gin.H{"error": msg})`.
- **Time fields** — use `*time.Time` for nullable timestamps (the DB has nullable TIMESTAMPTZ columns).
- **API paths** — all under `/api/v1/`. Router in `internal/api/router.go` (~244 lines, 18 handler groups).
- **PGX scanning** — use `.Scan(&field1, &field2, ...)` with pointers to model fields. `*time.Time` and `*string` for nullable DB columns.

### Frontend
- **No SvelteKit** — pure Vite + Svelte SPA. Routing via `window.dispatchEvent(new CustomEvent('navigate', { detail: { page, params } }))`.
- **API client** — `web/src/lib/api/client.js` (plain fetch, no Axios). Base URL from `VITE_API_URL` env var or `http://localhost:8080`.
- **Stores** — `auth.js`, `feed.js`, `notifications.js` in `web/src/lib/stores/`. Svelte writable stores with `$store` auto-subscription.
- **No CSS framework** — vanilla CSS in `<style>` blocks. CSS vars: `--primary`, `--bg`, `--text`, `--muted`, `--border`, `--success`, `--danger`.

## File Layout Reference

```
polaris/
└── internal/
    ├── api/
    │   ├── router.go               # 244 lines — all routes, handler init, middleware setup
    │   ├── handlers/               # 18 files + test (auth, health, post, tag, feed, trending, 
    │   │                           #   interaction, trust, achievement, blocklist, circle, collection,
    │   │                           #   community, credit, filter, moderation, note, search)
    │   ├── middleware/
    │   │   ├── auth.go             # JWT validation, sets "user_id" in context
    │   │   └── ratelimit.go        # (may be missing — check before relying on it)
    │   └── websocket/
    │       ├── hub.go              # Gorilla WebSocket + Redis pub/sub
    │       └── handlers.go         # Message types, event routing
    ├── db/
    │   ├── db.go                   # NewPostgres(dsn), NewRedis(addr)
    │   └── migrations/001_initial.sql  # 410 lines — 31 tables, all indexes, idempotent
    ├── model/                      # 20 struct files + errors.go
    ├── services/                   # 18 files (auth, user/post/tag/feed/...)
    └── worker/                     # 3 files — INCOMPLETE, needs scheduler and all workers
```

## What AI Agents Should Do (Priority Order)

1. **Fix .gitignore** — add Go, Node, env, IDE artifacts
2. **Add CORS middleware** — gin-contrib/cors, allow origin `http://localhost:5173` in dev
3. **Complete worker system** — write `internal/worker/scheduler.go` with ticker-based orchestration, implement missing workers (feed rebuild, trust decay, trending, achievements, scheduled posts, community notes consensus, active stats, pruning)
4. **Start workers in main.go** — add scheduler initialization and graceful shutdown
5. **Add Vite proxy** — configure `vite.config.js` to proxy `/api` to `http://localhost:8080`
6. **Fix handler compilation errors** — run `go build ./...`, fix any signature mismatches between router.go and handlers
7. **Expand test coverage** — add service-level unit tests (mock/skip pgx), more integration tests for edge cases
8. **Add Dockerfile** — multi-stage build for production deployment
9. **Implement ActivityPub stubs** — webfinger, outbox/inbox endpoints (data model is ready with actor_id/is_local/public_key)
10. **Add image upload** — POST /uploads, store to disk or S3, thumbnail generation worker

## Quick Reference

```go
// Handler pattern
func (h *XHandler) Method(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var req model.SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
    result, err := h.svc.SomeMethod(c.Request.Context(), userID.(int64), req)
    if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(200, result)
}

// Service SQL pattern
func (s *XService) Get(ctx context.Context, id int64) (*model.X, error) {
    x := &model.X{}
    err := s.pg.QueryRow(ctx, `SELECT ... FROM x WHERE id = $1`, id).Scan(...)
    if err != nil { return nil, model.ErrNotFound }
    return x, nil
}

// Frontend navigation
window.dispatchEvent(new CustomEvent('navigate', { detail: { page: 'profile', params: { id: 42 } } }));
```

### Key URLs
| Resource | URL |
|----------|-----|
| API health | `GET http://localhost:8080/health` |
| Register | `POST http://localhost:8080/api/v1/auth/register` |
| Login | `POST http://localhost:8080/api/v1/auth/login` |
| Feed | `GET http://localhost:8080/api/v1/feeds` |
| All routes | See `internal/api/router.go` lines 58–243 |

### Design Documents
| File | What It Covers |
|------|---------------|
| `design.md` | Full 27-section platform design (features, schema, API, trust, moderation, economy) |
| `AI_01_PROJECT_OVERVIEW.md` | Directory structure, stack, run commands |
| `AI_02_DATABASE_SCHEMA.md` | All 31 tables with columns, indexes, purpose |
| `AI_03_API_REFERENCE.md` | All ~90 endpoints with methods and descriptions |
| `AI_04_FRONTEND_REFERENCE.md` | Svelte component tree, stores, routing pattern |
| `AI_05_BUILD_PLAN.md` | 20-iteration roadmap with current completion status |
