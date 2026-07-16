# Polaris — AI Agent Guide

## Overview
Social platform backend in Go (Gin/pgx/Redis), frontend in Svelte + Vite (SPA). Go 1.26, PostgreSQL 17, Redis 7. Single binary monolith + 11 background workers.

**21 service files, 23 handler files, 26 model files, 11 worker files, 14 Svelte page components, ~95 API endpoints.**

## How to Run

| Component | Command | URL |
|-----------|---------|-----|
| Infra | `docker compose up -d` | Postgres :5432, Redis :6379 |
| Backend | `go run ./cmd/polaris` | `http://localhost:8080` |
| Frontend | `cd web && npm run dev` | `http://localhost:5173` |
| Tests | `DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1` | — |

**The Vite dev server has `server.proxy` configured — `/api` calls forward to `:8080`. The Go server has CORS middleware allowing all origins. Both issues from previous iterations are fixed.**

## Architecture

### Backend Layers
```
router.go → 23 handlers/*.go → 21 services/*.go → pgx SQL → PostgreSQL
                                                  ↕ Redis (cache, pub/sub, rate limit)
                                       ↕
                                 26 model/*.go
                                 11 worker/*.go (goroutine background tasks)
```

### Key Conventions
- **No ORMs** — raw SQL with pgx. Queries in services. No query builders.
- **No comments in code** — design doc and AGENTS.md are the documentation.
- **JWT auth** — `Authorization: Bearer <token>`. Middleware sets `user_id` in Gin context. Extract with `c.Get("user_id")`.
- **Error handling** — services return Go errors. Sentinels in `model/errors.go` (`ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`).
- **Time fields** — use `*time.Time` for nullable timestamps throughout (all TIMESTAMPTZ columns are nullable).
- **API paths** — all under `/api/v1/` except `/health`, `/metrics`, `/pages/:slug`, `/.well-known/webfinger`.
- **CORS** — `internal/api/middleware/cors.go` allows all origins in dev. Add restrictions for production.
- **Workers** — `internal/worker/scheduler.go` starts all 11 workers via goroutines. Workers log their runs. Actual implementations are stubs.
- **PGX scanning** — NULL columns must scan into `*time.Time`, `*string`, `*int64`. Use `.Scan(&field1, &field2, ...)`.

### Frontend
- **No framework routing** — `window.dispatchEvent(new CustomEvent('navigate', { detail: { page, params } }))` in App.svelte
- **API client** — `web/src/lib/api/client.js` (plain fetch). Base URL from VITE_API_URL env var or proxied through Vite.
- **Stores** — svelte writable stores in `web/src/lib/stores/`. Auto-subscribe with `$storeName`.
- **CSS** — vanilla CSS with CSS custom properties: `--primary`, `--bg`, `--text`, `--muted`, `--border`, `--success`, `--danger`.

## Current State

### Fixed Since v1
- ✅ **CORS middleware** — gin-contrib/cors alternative, allows all origins
- ✅ **Vite proxy** — `vite.config.js` proxies `/api` to `http://localhost:8080`
- ✅ **Worker scheduler** — `internal/worker/scheduler.go` starts all 11 workers
- ✅ **Worker startup in main.go** — scheduler initialized and started in `cmd/polaris/main.go`
- ✅ **Updated .gitignore** — covers Go, Node, env, IDE artifacts
- ✅ **Email verification** — `users.email_verified`, `users.email_verify_token`, endpoints, flows in LoginPage
- ✅ **Password reset** — `users.password_reset_token`, endpoints, UI in LoginPage
- ✅ **Account deletion** — soft delete with `is_deleted`/`deleted_at`
- ✅ **User blocking** — `blocked_users` table, BlockService, BlockHandler, UI in ProfilePage
- ✅ **Post editing** — 15-min window enforcement, `edited_at` tracking
- ✅ **Post locking** — mod disables comments, `locked` flag
- ✅ **Post sticky** — community-level pin, `sticky`/`sticky_at`
- ✅ **Post hiding** — per-user hide, `hidden_posts` table
- ✅ **Post moving** — mod moves between communities
- ✅ **Reports** — `post_reports` table with categories, ReportService/Handler, UI in ModCenter
- ✅ **Notifications** — `user_notifications` table, NotificationService/Handler, UI in NotificationsPage
- ✅ **Mod notes** — `mod_notes` table, private per-user notes
- ✅ **Custom pages** — `custom_pages` table, PageService/Handler, served at `/pages/:slug`
- ✅ **Custom emoji** — `custom_emoji` table, EmojiService/Handler, admin UI
- ✅ **OAuth stubs** — `oauth_states` table, provider config ready
- ✅ **Theme support** — `users.theme`, light/dark toggle in Settings

### Still Needed
- **Production Dockerfile** — multi-stage: golang:1.26-alpine build → scratch runtime, expose :8080
- **Image upload** — POST /api/v1/uploads, disk/S3 storage, thumbnail worker
- **ActivityPub federation** — WebFinger, outbox/inbox, HTTP Signatures, Follow/Like/Announce
- **Email delivery** — SMTP config, `email_queue` table, worker sends via sendmail/SMTP
- **WebSocket frontend** — connect WebSocket client for real-time feed + notifications
- **Seed data** — migration or script for demo users, posts, communities
- **Test coverage** — currently 14 integration tests; need service unit tests, E2E with Playwright

### File Layout
```
internal/
├── api/                   # router.go (244 lines, ~95 endpoint registrations)
│   ├── handlers/          # 23 files: auth, post, tag, feed, search, collection, community,
│   │                      #   circle, note, moderation, blocklist, filter, credit, achievement,
│   │                      #   onboarding, admin, block, report, notification, mod_note, emoji, page
│   ├── middleware/        # auth.go, cors.go
│   └── websocket/         # hub.go, handlers.go (gorilla + Redis pub/sub)
├── db/                    # db.go + migrations/001_initial.sql (31 tables)
├── model/                 # 26 files + errors.go
├── services/              # 21 files
└── worker/                # scheduler.go + 10 worker files
```

### Key Design Documents
| File | Purpose |
|------|---------|
| `design.md` | Full 27-section platform design |
| `AI_01_PROJECT_OVERVIEW.md` | Directory structure, run commands, test status |
| `AI_02_DATABASE_SCHEMA.md` | All 31+ tables with columns and indexes |
| `AI_03_API_REFERENCE.md` | All API endpoints with descriptions |
| `AI_04_FRONTEND_REFERENCE.md` | Component tree, stores, routing |
| `AI_05_BUILD_PLAN.md` | Build roadmap and current completion status |
| `TODO.md` | Feature backlog: Tier 1–4 + incomplete items |
