# Polaris — Build Plan & Current Status

## Overall Progress (~20 iterations, ~60% complete)

| Iteration | Name | Status | What's Built |
|-----------|------|--------|-------------|
| 1 | Project scaffold + auth | ✅ Done | Go module, Gin router, Postgres+Redis connections, JWT auth, health check, Docker Compose, Makefile |
| 2 | User profiles + trust bootstrap | ✅ Done | User CRUD, trust connections, invite system, trust score computation |
| 3 | Post CRUD + content warnings | ✅ Done | Post CRUD, content_warning, scheduled_at + status, scheduled post worker |
| 4 | Tags + booru-style tagging | ✅ Done | Tag CRUD, post_tag mapping with attribution, tag voting/conflict resolution |
| 5 | Feed v1 + interactions | ✅ Done | Chronological feed, interaction recording (like/comment/share), feed_items table |
| 6 | Feed v2 (algorithmic) | ✅ Done | Scoring function, personalized feed, Redis candidate pool, diversity guarantee |
| 7 | Custom feeds + sources | ✅ Done | CustomFeed CRUD, feed_sources (tag/user/keyword), include/exclude, sharing |
| 8 | Advanced search | ✅ Done | Full-text search (GIN), boolean query parser, saved searches |
| 9 | Collections + bookmarks | ✅ Done | Collection CRUD, 4 visibility layers, default bookmark |
| 10 | Communities | ✅ Done | Community CRUD, curators, slow-boot, scope enforcement, forking, active stats worker |
| 11 | Auto-circles | ✅ Done | Interest+proximity circle generation, suggestions API, accept/dismiss |
| 12 | Community Notes | ✅ Done | Note CRUD, weighted consensus voting, diversity requirement, display |
| 13 | Trending + pruning | ✅ Done | Trending detection worker, trending display, content pruning worker |
| 14 | WebSocket + real-time | ⚠️ Partial | Hub + handlers written, not fully tested with frontend |
| 15 | Moderation + jury | ✅ Done | Flagging, timeouts, jury selection, voting, mod log, circuit breaker, shared blocklist |
| 16 | Blocklist + filters | ✅ Done | Blocklist CRUD + export, per-user content filters (keyword/regex/domain/user) |
| 17 | Credit system + economy | ✅ Done | Ledger, daily rewards, gifting, bounty system |
| 18 | Achievements + leaderboards | ✅ Done | Achievement definitions, progress tracking, leaderboards with periods |
| 19 | Onboarding + RSS + UX | ✅ Done | Onboarding wizard (4 stages), keyboard shortcuts, mental health features |
| 20 | ActivityPub + test + deploy | ⚠️ Partial | Actor ID mapping, webfinger stubs, 14 integration tests pass |

## What Remains
1. **WebSocket integration** — Frontend needs WebSocket client connecting to /ws/v1/feed for real-time feed updates
2. **ActivityPub federation** — Actual outbox/inbox endpoints, webfinger, HTTP Signatures
3. **Image/video upload** — S3/MinIO integration, thumbnail generation worker
4. **More tests** — Currently 14 integration tests; need unit tests for services, E2E tests with Playwright
5. **Mobile responsive polish** — CSS media queries work but need thorough testing
6. **Production deployment** — Dockerfile for Go binary, nginx for static files + reverse proxy, backup scripts

## How to Develop
```bash
# Start infrastructure
docker compose up -d

# Run backend tests (must connect to test DB)
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1

# Start backend
go run ./cmd/polaris

# Start frontend (another terminal)
cd web && npm run dev

# Run frontend build
cd web && npm run build
```

## Key Architecture Decisions
- **No ORM** — Raw SQL with pgx (compile-time checked via sqlc in future)
- **No microservices** — Single binary monolith + background workers
- **No real blockchain** — Cryptographic ledger in PostgreSQL (SHA-256 chain)
- **No SvelteKit** — Pure SPA, Vite dev server proxies API calls
- **No Kubernetes** — Single server deployment, Docker Compose for dev

## Critical Files for Refactoring
| File | Purpose |
|------|---------|
| `internal/api/router.go` | All routes, handler creation, middleware setup |
| `internal/db/migrations/001_initial.sql` | Complete database schema |
| `internal/services/*.go` | Business logic layer (18 files) |
| `internal/api/handlers/*.go` | HTTP handler layer (18 files) |
| `web/src/lib/api/client.js` | Frontend API client (all endpoints) |
| `web/src/App.svelte` | Frontend routing + auth guard |
| `web/src/lib/stores/auth.js` | Auth state management |
