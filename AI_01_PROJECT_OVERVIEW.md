# Polaris — Project Overview

## What It Is
A social media platform built for positive connection: trust-based moderation, collaborative curation (booru-style tagging), democratic self-governance (90% jury bans), and interest+proximity matching. No outrage amplification.

## Stack
- **Backend**: Go 1.26, Gin framework, pgx v5 (PostgreSQL), go-redis, gorilla/websocket, golang-jwt, bcrypt, zerolog
- **Frontend**: Svelte 4 + Vite (SPA, no SvelteKit), vanilla CSS
- **Infra**: Single binary deploy, Docker Compose for dev (PostgreSQL 17, Redis 7), Prometheus metrics at /metrics

## Directories
```
polaris/
├── cmd/polaris/          # Entry point
├── internal/
│   ├── api/
│   │   ├── router.go           # All routes (18 handler groups, ~90 endpoints)
│   │   ├── handlers/           # 18 handler files (auth, posts, tags, feed, search,
│   │   │                       #   collections, communities, circles, notes, moderation,
│   │   │                       #   blocklist, filters, credits, achievements, onboarding, admin)
│   │   ├── middleware/         # auth.go (JWT), ratelimit.go
│   │   └── websocket/         # hub.go + handlers.go (gorilla/websocket + Redis pub/sub)
│   ├── db/
│   │   ├── db.go               # Postgres pool + Redis client creation
│   │   └── migrations/        # 001_initial.sql (31 tables, all indexes)
│   ├── model/                  # 20 struct files with JSON tags
│   ├── services/               # 18 service files (auth, user, post, tag, feed, custom_feed,
│   │                           #   search, collection, community, circle, note, trending,
│   │                           #   moderation, blocklist, filter, credit, achievement, onboarding)
│   └── worker/                 # 8 workers (feed rebuild, trust decay, trending, achievements,
│                               #   scheduled posts, community notes consensus, active stats, pruning)
├── web/                        # Svelte frontend
│   ├── src/
│   │   ├── App.svelte          # Routing hub: auth guard → onboarding → main app
│   │   ├── lib/
│   │   │   ├── api/client.js   # Fetch-based API client with JWT token management
│   │   │   ├── stores/         # auth.js, feed.js, notifications.js
│   │   │   ├── components/     # Navbar, TrendingSidebar, FeedItem, MoodPicker, CommunityCard
│   │   │   └── utils/          # (empty, ready for helpers)
│   │   └── routes/             # 12 route pages
│   ├── package.json            # Vite + Svelte, no other deps
│   └── vite.config.js
├── docker-compose.yml          # postgres:17 + redis:7
├── Makefile                    # dev, build, run, test, migrate
├── go.mod / go.sum
└── *.md                        # Design doc + 5 AI context files
```

## Test Status
- Backend: 14 integration tests pass (auth, posts, tags, feed, collections, communities, search, blocklist, middleware)
- Frontend: builds cleanly (`npm run build`)
- `go vet ./...` — clean, `go build ./...` — clean

## How to Run
```bash
# Terminal 1: Start infrastructure
docker compose up -d

# Terminal 2: Start backend
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run ./cmd/polaris

# Terminal 3: Start frontend
cd web && npm run dev

# Visit http://localhost:5173
```

## Testing
```bash
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1
```
