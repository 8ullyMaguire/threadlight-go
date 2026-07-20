# Threadlight

A social platform built for positive connection.

Threadlight has no ads, no outrage algorithms, and no billionaires pulling the strings.
Your feed is yours. Moderation is shared. Tags are owned by everyone.

## Quick Start

```bash
# Start PostgreSQL + Redis
docker compose up -d

# Run the backend
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run ./cmd/polaris

# Start the frontend (another terminal)
cd web && npm run dev
```

Open **http://localhost:5173**.

## Stack

| Layer | What |
|-------|------|
| Backend | Go 1.26, Gin, pgx v5, go-redis |
| Frontend | Svelte 4 + Vite SPA |
| Database | PostgreSQL 17 (31 tables) |
| Cache | Redis 7 |
| Workers | 11 goroutine-based background workers |
| API | ~95 endpoints under `/api/v1` |

## Project Layout

```
cmd/polaris/          # Entry point — wires everything together
internal/
  api/                # router.go + 23 handlers + middleware + websocket
  services/           # 21 business logic files (all SQL lives here)
  model/              # 26 data structs + errors
  db/                 # postgres + redis connections + async migrations
  worker/             # scheduler + 11 background workers
web/src/              # Svelte SPA
  lib/api/            # Fetch-based API client
  lib/stores/         # Auth, feed, notifications (svelte writable stores)
  lib/components/     # Navbar, TrendingSidebar, Footer
  routes/             # 14 page components
design/               # DESIGN.md, TODO.md, COMPRESSED_CODEBASE.md
docs/                 # mdbook documentation
```

## Design Docs

All in `design/`:
- `DESIGN.md` — platform philosophy and system design
- `TODO.md` — feature backlog
- `COMPRESSED_CODEBASE.md` — ultra-condensed context for LLM agents
