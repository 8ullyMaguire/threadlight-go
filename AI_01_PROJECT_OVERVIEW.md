# Polaris — Project Overview

## What It Is
A social media platform for positive connection: trust-based moderation, collaborative curation (booru-style tagging), democratic self-governance (90% jury bans), and interest+proximity matching. No outrage amplification. 26 model types, 21 services, 23 handlers, 11 workers, 14 frontend pages, ~95 API endpoints.

## Stack
- **Backend**: Go, Gin, pgx v5 (PostgreSQL), go-redis, gorilla/websocket, golang-jwt, zerolog
- **Frontend**: Svelte 4 + Vite (SPA), vanilla CSS
- **Infrastructure**: Single binary, Docker Compose (PostgreSQL 17, Redis 7), Prometheus at /metrics

## Directories
```
polaris/
├── cmd/polaris/              # main.go — HTTP server + worker scheduler
├── internal/
│   ├── api/
│   │   ├── router.go         # ~95 routes across 18+ handler groups
│   │   ├── handlers/         # 23 files: auth, health, post, tag, feed, search, collection,
│   │   │                     #   community, circle, note, moderation, blocklist, filter, credit,
│   │   │                     #   achievement, onboarding, admin, block, report, notification,
│   │   │                     #   mod_note, emoji, page
│   │   ├── middleware/       # auth.go (JWT), cors.go
│   │   └── websocket/       # hub.go + handlers.go
│   ├── db/
│   │   ├── db.go             # NewPostgres(dsn), NewRedis(addr)
│   │   └── migrations/      # 001_initial.sql (31 tables, idempotent, 410 lines)
│   ├── model/                # 26 struct files + errors.go
│   ├── services/             # 21 files: auth, user, post, tag, feed, custom_feed, search,
│   │                         #   collection, community, circle, note, trending, moderation,
│   │                         #   blocklist, filter, credit, achievement, onboarding, block,
│   │                         #   report, notification, mod_note, emoji, page
│   └── worker/               # scheduler.go + 10 workers (feed, trust_decay, trending,
│                             #   achievements, scheduled_posts, community_notes, active_stats, prune)
├── web/                      # Svelte frontend
│   ├── src/
│   │   ├── App.svelte        # Auth guard → onboarding → main app + 14 page routes
│   │   ├── lib/
│   │   │   ├── api/client.js # Full API client (~95 methods)
│   │   │   ├── stores/       # auth.js, feed.js, notifications.js
│   │   │   └── components/   # Navbar, TrendingSidebar
│   │   └── routes/           # 14 pages: login, register, feed, search, communities,
│   │                         #   collections, profile, notifications, blocked,
│   │                         #   moderation, settings, leaderboard, admin, onboarding
│   └── package.json
├── docker-compose.yml
├── Makefile
├── design.md
├── TODO.md
├── AGENTS.md
├── AI_*.md (5 files)
└── README.md
```

## Test Status
- Backend: 14 integration tests pass (auth, posts, tags, feed, collections, communities, search, blocklist, middleware)
- Frontend: builds clean (`npm run build`)
- `go vet ./...` — clean, `go build ./...` — clean

## How to Run
```bash
docker compose up -d
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run ./cmd/polaris
cd web && npm run dev  # → http://localhost:5173
```

## Testing
```bash
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1
```
