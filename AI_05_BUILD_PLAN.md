# Polaris — Build Plan & Current Status

## Overall Progress (~26 iterations, ~85% complete)

| Iteration | Name | Status | What's Built |
|-----------|------|--------|-------------|
| 1 | Project scaffold + auth | ✅ Done | Go module, Gin, Postgres+Redis, JWT, health check, Docker Compose, Makefile |
| 2 | User profiles + trust bootstrap | ✅ Done | User CRUD, trust connections, invite system, trust score |
| 3 | Post CRUD + content warnings | ✅ Done | Post CRUD, content_warning, scheduling, drafted/published/scheduled |
| 4 | Tags + booru-style tagging | ✅ Done | Tag CRUD, post_tag with attribution, conflict resolution |
| 5 | Feed v1 + interactions | ✅ Done | Chronological feed, likes/comments/shares, feed_items table |
| 6 | Feed v2 (algorithmic) | ✅ Done | Scoring, personalization, Redis candidate pool, diversity guarantee |
| 7 | Custom feeds + sources | ✅ Done | CustomFeed CRUD, feed_sources (tag/user/keyword), sharing |
| 8 | Advanced search | ✅ Done | Full-text search (GIN), boolean parser, saved searches |
| 9 | Collections + bookmarks | ✅ Done | Collection CRUD, 4 visibility layers, default bookmark |
| 10 | Communities | ✅ Done | CRUD, curators, slow-boot, scope enforcement, forking, active stats |
| 11 | Auto-circles | ✅ Done | Interest+proximity circles, suggestions, accept/dismiss |
| 12 | Community Notes | ✅ Done | Notes, weighted consensus, diversity requirement, display |
| 13 | Trending + pruning | ✅ Done | Trending detection, display, content pruning worker |
| 14 | WebSocket + real-time | ✅ Done | Hub + handlers + Redis pub/sub |
| 15 | Moderation + jury | ✅ Done | Flagging, timeouts, jury selection, mod log, circuit breaker |
| 16 | Blocklist + filters | ✅ Done | Blocklist CRUD + export, per-user content filters |
| 17 | Credit system + economy | ✅ Done | Ledger, daily rewards, gifting, bounties |
| 18 | Achievements + leaderboards | ✅ Done | Definitions, progress tracking, multi-period leaderboards |
| 19 | Onboarding + RSS + UX | ✅ Done | 4-stage wizard, RSS/Atom, keyboard shortcuts, mental health |
| 20 | ActivityPub + test + deploy | ⚠️ Partial | Actor IDs, webfinger stubs, 14 integration tests |
| 21 | CORS + infra fixes | ✅ Done | CORS middleware, Vite proxy, .gitignore, worker scheduler in main.go |
| 22 | Email verification + password reset | ✅ Done | Token flows, endpoints, UI forms |
| 23 | User blocking + account deletion | ✅ Done | Blocked_users table, block/unblock UI, soft delete |
| 24 | Post editing + locking + stickying | ✅ Done | 15-min edit window, lock/sticky/hide/move |
| 25 | Reports + mod notes + notifications | ✅ Done | Report categories, mod notes, notification history |
| 26 | Custom pages + emoji + themes + OAuth | ✅ Done | PageService, EmojiService, light/dark theme, OAuth stubs |

## What Remains
1. **Image upload** — POST /api/v1/uploads, disk/S3 storage, thumbnail worker
2. **ActivityPub federation** — actual outbox/inbox, HTTP Signatures, Follow/Like/Announce
3. **Email delivery** — SMTP config, email_queue worker sends via sendmail/SMTP
4. **WebSocket frontend** — connect WebSocket client for real-time feed + notifications
5. **Seed data** — migration or script with demo users, posts, communities
6. **Test coverage** — currently 14 integration tests; need service unit tests, E2E Playwright tests
7. **Production Dockerfile** — multi-stage build (golang:1.26-alpine → scratch)

## Key Architecture Decisions
- **No ORM** — Raw SQL with pgx
- **No microservices** — Single binary + goroutine workers
- **No real blockchain** — Cryptographic ledger in PostgreSQL
- **No SvelteKit** — Pure SPA with custom event routing
- **No Kubernetes** — Single server, Docker Compose for dev

## Critical Files
| File | Purpose |
|------|---------|
| `internal/api/router.go` | All ~95 routes |
| `internal/db/migrations/001_initial.sql` | Full schema (31 tables) |
| `internal/services/*.go` | Business logic (21 files) |
| `internal/api/handlers/*.go` | HTTP handlers (23 files) |
| `web/src/lib/api/client.js` | Frontend API client |
| `web/src/App.svelte` | Frontend routing + auth guard |
