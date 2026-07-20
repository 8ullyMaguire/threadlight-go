# Threadlight — AI Agent Guide

## Overview
Go/Gin backend + Svelte/Vite SPA. Single binary monolith. PostgreSQL 17 + Redis 7.
21 services, 23 handlers, 26 models, 11 workers, 14 Svelte pages, ~95 API endpoints.

## Run Commands

| Component | Command | URL |
|-----------|---------|-----|
| Infra | `docker compose up -d` | PG :5432, Redis :6379 |
| Backend | `go run ./cmd/polaris` | `:8080` |
| Frontend | `cd web && npm run dev` | `:5173` |
| Tests | `DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./... -count=1` | — |

Vite proxies `/api` → `:8080`. Go CORS allows all origins.

## Architecture

```
router.go → 23 handlers/*.go → 21 services/*.go → pgx SQL → PostgreSQL
                                                     ↑↓ Redis
                                         26 model/*.go
                                         11 worker/*.go
```

### Conventions
- **No ORM** — raw SQL in services. `pgxpool.Pool` injected via struct fields.
- **No code comments** — this file + design/ are the docs.
- **JWT auth** — middleware sets `user_id` in Gin context. Get it: `c.Get("user_id")`. Returns `interface{}`, assert to `int64`.
- **Error sentinels** — `model.ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`.
- **Nullable time fields** — use `*time.Time` for all TIMESTAMPTZ columns. Scan into `*time.Time`.
- **Nullable string/int fields** — use `*string`, `*int64` for columns that can be NULL.
- **API paths** — `/api/v1/` except `/health`, `/metrics`, `/pages/:slug`, `/.well-known/webfinger`.
- **Workers** — `scheduler.go` starts 11 goroutines. Each worker gets its own goroutine. Panics are caught and logged.
- **Frontend routing** — Custom event-based: `window.dispatchEvent(new CustomEvent('navigate', { detail: { page, params } }))`.
- **API client** — `web/src/lib/api/client.js` — plain `fetch()`. Base URL from `VITE_API_URL` env var or proxy.
- **Svelte stores** — writable stores in `lib/stores/`. Subscribe with `$storeName`.
- **CSS** — Vanilla CSS with custom properties: `--primary`, `--bg`, `--text`, `--muted`, `--border`, `--radius`, `--success`, `--danger`.

## Backend File Layout

```
internal/
├── api/
│   ├── router.go                          # ~95 route registrations
│   ├── handlers/ (23 files)               # auth, post, tag, feed, search, collection,
│   │                                      # community, circle, note, moderation, blocklist,
│   │                                      # filter, credit, achievement, onboarding, admin,
│   │                                      # block, report, notification, mod_note, emoji, page
│   ├── middleware/                        # auth.go, cors.go
│   └── websocket/                         # hub.go, handlers.go
├── db/                                    # db.go + migrations/
├── model/ (26 files + errors.go)
├── services/ (21 files)
└── worker/ (scheduler.go + 10 workers)
```

## Key Services

| Service | DB table(s) | Key methods |
|---------|------------|-------------|
| AuthService | users | Register, Login, ForgotPassword, ResetPassword |
| PostService | posts | CreatePost, UpdatePost, GetFeedPosts, ArchivePost, VotePost |
| TagService | tags, post_tags, tag_votes | CreateTag, TagPost, VoteTag, SearchTags |
| FeedService | — | GetFeed, GetCustomFeed (algorithmic ranking) |
| SearchService | posts (tsvector) | Search (boolean + fielded) |
| CommunityService | communities, community_members | Create, Join, Leave, SlowBootCheck |
| ModerationService | moderation_actions, reports, jury_votes | CreateAction, Report, Vote, Jury |
| BlockService | blocked_users | BlockUser, IsBlocked, Unblock, GetBlockedUsers |
| CreditService | credit_transactions | Award, Spend, GetBalance, Transfer |
| AchievementService | achievements, user_achievements | CheckAndAward, GetUserAchievements |
| NotificationService | notifications | Send, MarkRead, GetUnread |
| CollectionService | collections, collection_posts | Create, AddPost, RemovePost |
| CircleService | circles, circle_members | Create, Join, Suggest |
| NoteService | community_notes, community_note_votes | CreateNote, VoteNote |
| FilterService | filters | Create, Evaluate, Delete |
| ReportService | reports | Create, Review, Resolve |

## Key Handlers

| Handler | Endpoints |
|---------|-----------|
| AuthHandler | POST register, login, forgot, reset, logout |
| PostHandler | GET/POST/PUT/DELETE posts, POST /:id/archive, /:id/remove (mod), /:id/vote |
| TagHandler | GET/POST/DELETE tags, POST /:id/vote |
| FeedHandler | GET feed, custom feeds |
| SearchHandler | GET /search (q, type, boolean operators) |
| CommunityHandler | GET/POST communities, POST /:id/join, /:id/leave |
| ModerationHandler | POST /moderation/action, report, vote, /jury |
| AdminHandler | GET /admin/stats, POST /admin/seed |
| OnboardingHandler | POST /onboarding/tags, /stage, GET /onboarding/suggested |
| NotificationHandler | GET /notifications, PUT /notifications/:id/read |
| ModNoteHandler | POST /mod-notes, GET /mod-notes/:user_id |

## Database (~31 tables)

**Core**: users, posts, communities, tags, post_tags, tag_votes
**Trust**: trust_connections, trust_score_log, vouches
**Moderation**: moderation_actions, reports, jury_votes, mod_notes
**Social**: community_members, community_notes, community_note_votes
**Interaction**: interactions, reactions, hidden_posts
**Discovery**: tags, post_tags, tag_votes, collections, collection_posts, circles, circle_members
**Economy**: credit_transactions, achievements, user_achievements
**Safety**: blocked_users, filters, blocklist, blocklist_subscriptions
**Notifications**: user_notifications
**Admin**: site_config, custom_pages, custom_emoji, registration_queue

## Current State

**Implemented**: auth, posts, tags with voting, search with boolean/fielded, communities (join/leave/slow-boot/probation/tenure-voting), moderation (M1 labels + admin actions + jury), collections, circles, blocking, content filters, credits, achievements, notifications, onboarding, custom feeds, trending, migration runner, Dockerfile, password reset, email verification, account deletion

**Missing from docs but not implemented**: M2 meta-moderation (jury system exists but isn't M2-style), Positive Algorithm (feed is chronological), "Why this post?" button, full reaction set (only generic like/unlike), tag autocomplete, auto-generated circles (manual only), credit hash-chain ledger (simple transactions), ActivityPub (keys exist but no inbox/outbox), WebSocket (hub exists but not wired to feed), image upload (model exists, service exists, no endpoint yet)

**Known issues**: Silent error catching in frontend (some catch blocks log but don't show UI), no pagination UI on feed (backend supports it), test cleanup FKs when run in parallel (services + handlers share DB), migration runner not wired into main.go startup (called from db.go but main.go doesn't call it — fix: add `db.RunMigrations(pg, "internal/db/migrations")` in main after postgres connect)
