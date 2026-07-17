# Polaris

A social platform built for positive connection — trust-based moderation, collaborative curation, and democratic self-governance.

## Quick Start

```bash
# 1. Start infrastructure
docker compose up -d

# 2. Run migrations (first time only)
cat internal/db/migrations/001_initial.sql | docker compose exec -T postgres psql -U polaris

# 3. Start backend
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run ./cmd/polaris

# 4. Start frontend (another terminal)
cd web && npm run dev
```

Visit **http://localhost:5173** — login/register, then explore.

## Features

**Content**: Create posts with mood, content warnings, scheduled publishing, polls, cross-posting, and multiple licenses (CC0/CC-BY/etc.)  
**Collaborative Curation**: Booru-style tagging where any trusted user can tag any post. Tag wikis, conflict resolution voting.  
**Communities**: Topic-based public spaces with curators, slow-boot probation, scope-enforced tags, and forking.  
**Auto-Circles**: System-generated interest + proximity groups ("Photography in Berlin"). Accept or dismiss.  
**Custom Feeds**: Build weighted feed sources from tags, users, and keywords with include/exclude rules. Shareable.  
**Algorithmic Feed**: Personalized by mood, tag affinity, trust network, and content quality. Diversity guarantee prevents filter bubbles.  
**Advanced Search**: Full boolean (AND/OR/NOT), field search (title:, author:, tag:), date ranges, saved searches with notifications.  
**Collections**: Bookmark collections with 4 visibility layers (private, public read, public edit, anybody edit). Default bookmark.  
**Community Notes**: Twitter/X-style consensus-driven misinformation warnings. Weighted voting with cross-tag diversity requirement.  
**Trending Topics**: Auto-detected from post content (not hashtags). Anti-abuse filters (min 3 users, 2 tag categories).  
**Moderation**: Trust-score-gated moderation. Timeouts (threshold 3+). Bans require 90% jury consensus. Public append-only mod log.  
**Post Management**: Edit (15-min window), lock (mod disables comments), sticky (pin to community), hide (per-user), move between communities.  
**User Controls**: Block users (full content + notification suppression), content filters (keyword/regex/domain/user), report content with categories.  
**Notifications**: In-app notification history, configurable per type, WebSocket push, email digests (planned).  
**Credit Economy**: Daily rewards, gifting, bounty Q&A system. 1% platform tax on transfers. Stripe integration planned.  
**Achievements & Leaderboards**: Multi-period leaderboards (daily/weekly/monthly/yearly/all-time) for posting, tagging, gifting, trust, achievements.  
**Onboarding Wizard**: 4-stage guided first-run (pick tags → follow users → first post → done).  
**RSS/Atom Everywhere**: Every feed, tag, user, community, collection has RSS/Atom. ETag-cached.  
**ActivityPub-Ready**: Actor IDs, ed25519 keys, WebFinger stub, outbox/inbox stubs. Federation-ready data model.  
**OAuth Login**: Google, Discord, Mastodon OAuth providers (configurable).  
**Custom Pages**: Editable About, Terms, Privacy pages served at `/about`, `/terms`, `/privacy`.  
**Custom Emoji**: Upload and manage custom emoji; `:emoji:` replacement in post bodies.  
**Light/Dark Theme**: User-selectable theme persisted in profile. CSS variable-based.  
**Security**: JWT auth, bcrypt passwords, CORS headers, rate limiting, shared blocklist, email verification.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, Gin, pgx v5, go-redis, gorilla/websocket |
| Frontend | Svelte 4 + Vite (SPA) |
| Database | PostgreSQL 17 (31 tables) |
| Cache | Redis 7 |
| Workers | 11 goroutine-based background workers |
| API | ~95 endpoints under /api/v1 |

## Project Structure

```
polaris/
├── cmd/polaris/              # Entry point + worker startup
├── internal/
│   ├── api/                  # Router (57 route registrations), 23 handlers, 2 middleware, WebSocket
│   ├── db/                   # pgx pool + Redis client + 410-line migration
│   ├── model/                # 26 data struct files
│   ├── services/             # 21 business logic files
│   └── worker/               # 11 background workers (scheduled, feed, trending, etc.)
├── web/                      # Svelte frontend
│   ├── src/
│   │   ├── lib/api/          # Fetch-based API client
│   │   ├── lib/stores/       # Auth, feed, notifications
│   │   ├── lib/components/   # Navbar, TrendingSidebar
│   │   └── routes/           # 14 page components
│   └── package.json
├── design/                    # Design docs (DESIGN.md + 5 AI reference files)
├── docker-compose.yml         # PostgreSQL 17 + Redis 7
├── Makefile
├── TODO.md                    # Feature backlog (Tiers 1–4, ~60 items)
└── AGENTS.md                  # AI agent guide with architecture gaps
```

## Testing

```bash
# Backend (requires running PostgreSQL)
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1

# Frontend build check
cd web && npm run build
```

## Design Docs

All in the `design/` directory:

- [DESIGN.md](design/DESIGN.md) — condensed design overview
- [AI_01_ARCH.md](design/AI_01_ARCH.md) — architecture patterns
- [AI_02_MODELS.md](design/AI_02_MODELS.md) — data model reference
- [AI_03_API.md](design/AI_03_API.md) — API endpoint reference
- [AI_04_DB.md](design/AI_04_DB.md) — database schema reference
- [AI_05_FRONTEND.md](design/AI_05_FRONTEND.md) — frontend patterns
