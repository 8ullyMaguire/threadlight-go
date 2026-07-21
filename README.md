# Threadlight

A social platform built for positive connection, transparent moderation, and user-owned discovery.

No ads, no outrage algorithms, no billionaires. Your feed is yours. Moderation is shared.

## Quick Start

```bash
# Start PostgreSQL + Redis
docker compose up -d

# Run the backend
DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go run ./cmd/polaris

# Start the frontend (another terminal)
cd web && npm run dev
```

Open **http://localhost:5173**.

## Stack

| Layer | What |
|-------|------|
| Backend | Go 1.26, Gin, pgx v5, go-redis |
| Frontend | SvelteKit + Vite |
| Database | PostgreSQL 17 (~40 tables) |
| Cache | Redis 7 |
| Workers | 13 goroutine-based background workers |
| API | ~110 endpoints under `/api/v1` |

## Features

- **Auth** — email/password registration, login, JWT sessions, password reset, email verification
- **Posts** — text posts with Markdown, content warnings, scheduling, moods, content types
- **Tags** — collaborative tagging, tag voting (upvote/downvote), tag search
- **Search** — Boolean query language (AND/OR/NOT/parentheses/phrases), field filters (`author:`, `tag:`, `community:`), facet counts, trigram autocomplete
- **Communities** — join/leave, slow-boot, probation, tenure-voting, curator system, forking
- **Moderation** — M1 labels, admin actions, jury system, **public decision reviews** (Fair/Unfair voting with controversy scoring)
- **Trust** — trust connections, trust levels, trust decay, community trust graphs
- **User Lists** — create/manage follow/block lists, algorithmic auto-populated lists, collaborative lists with editor/viewer roles
- **Affinity** — real-time event-driven affinity counters (Redis, O(1) badge lookup), on-demand affinity page with 24h cache
- **Feed Plugins** — WASM-sandboxed custom feed algorithms, plugin marketplace with install/review
- **Economy** — credits, achievements, bounties, daily rewards, leaderboards
- **Social** — circles (manual groups), collections, community notes (fact-checking), saved searches
- **Safety** — per-user blocking, shared blocklists, content filters, NSFW gating

## Project Layout

```
cmd/polaris/          # Entry point — wires everything together
internal/
  api/                # router.go + 26 handlers + middleware + websocket
  search/             # Boolean query parser (recursive-descent AST builder)
  services/           # 25 business logic files (all SQL lives here)
  model/              # 30+ data structs + errors
  db/                 # postgres + redis connections + async migrations (4 migrations)
  worker/             # scheduler + 13 background workers
web/src/              # SvelteKit SPA
  lib/api/            # Fetch-based API client (Threadlight adapter)
  lib/feature/        # Feature components (Post, Comment, TagVote, etc.)
  routes/             # 30+ page components
design/               # DESIGN.md, TODO.md, FEATURE_INVENTORY.md
docs/                 # mdbook documentation
```

## Design Docs

All in `design/`:
- `FEATURE_INVENTORY.md` — complete feature inventory with priority order
- `TODO.md` — feature backlog
- `DESIGN.md` — platform philosophy and system design
