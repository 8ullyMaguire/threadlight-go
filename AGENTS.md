# Threadlight — AI Agent Guide

## Overview
Go/Gin backend + SvelteKit/Vite SPA. Single binary monolith. PostgreSQL 17 + Redis 7.
25 services, 26 handlers, 30+ models, 13 workers, 30+ Svelte pages, ~110 API endpoints.
4 database migrations, auto-run on startup.

## Run Commands

| Component | Command | URL |
|-----------|---------|-----|
| Infra | `docker compose up -d` | PG :5433, Redis :6379 |
| Backend | `go run ./cmd/polaris` | `:8080` |
| Frontend | `cd web && npm run dev` | `:5173` |
| Tests | `DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go test ./... -count=1` | — |

Vite proxies `/api` → `:8080`. Go CORS allows all origins.

## Architecture

```
router.go → 26 handlers/*.go → 25 services/*.go → pgx SQL → PostgreSQL
                      ↑                        ↑↓ Redis
                      |                 30+ model/*.go
               search/parser.go         13 worker/*.go
```

### Conventions
- **No ORM** — raw SQL in services. `pgxpool.Pool` injected via struct fields.
- **Parser** — `internal/search/parser.go` is a standalone recursive-descent parser for the Boolean query language (used by SearchPosts).
- **JWT auth** — middleware sets `user_id` in Gin context. Get it: `c.GetInt64("user_id")`.
- **Error sentinels** — `model.ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrValidation`.
- **Nullable fields** — use `*T` for nullable DB columns (scan with `&field`).
- **API paths** — `/api/v1/` except `/health`.
- **Workers** — `scheduler.go` starts 13 goroutines. Each worker gets its own goroutine.
- **Frontend** — SvelteKit with client-side rendering. API client at `web/src/lib/api/client.svelte.ts`.
- **Affinity accumulator** — `AffinityAccumulator` writes event-driven Redis counters on user actions (follow, join, react). O(1) badge lookup. 24h TTL cache for full affinity page.
- **CSS** — Custom CSS variables: `--primary`, `--bg`, `--text`, `--muted`, `--border`, `--radius`.

## Backend File Layout

```
internal/
├── api/
│   ├── router.go                        # ~110 route registrations
│   └── handlers/ (26 files)             # auth, post, tag, feed, search, collection,
│                                         # community, circle, note, moderation, blocklist,
│                                         # filter, credit, achievement, user, block, report,
│                                         # notification, mod_note, userlist, affinity,
│                                         # feedplugin, mod_decision, config, health, about
├── search/
│   └── parser.go                        # Boolean query parser (recursive descent)
├── db/                                   # db.go + migrations/ (4 files)
├── model/ (30+ files including errors.go)
├── services/ (25 files)                 # all business logic + SQL
└── worker/ (scheduler.go + 13 workers)
```

## Key Services

| Service | DB table(s) | Key methods |
|---------|------------|-------------|
| AuthService | users | Register, Login, ForgotPassword, ResetPassword |
| PostService | posts | CreatePost, UpdatePost, GetFeedPosts, ArchivePost |
| TagService | tags, post_tags, tag_votes | CreateTag, TagPost, VoteTag, SearchTags |
| FeedService | — | GetFeed, GetCustomFeed (algorithmic ranking) |
| SearchService | posts (tsvector) | SearchPosts (Boolean + fielded + facets), SearchUsers (trigram), SearchCommunities, GetSuggestions |
| CommunityService | communities, community_members, curators | Create, Join, Leave, Fork |
| ModerationService | moderation_actions, reports, jury_votes | CreateAction, Report, Vote, Jury |
| BlockService | blocked_users | BlockUser, IsBlocked, Unblock |
| CreditService | credit_transactions | Award, Spend, GetBalance, Transfer |
| AchievementService | achievements, user_achievements | CheckAndAward, GetUserAchievements |
| NotificationService | notifications | Send, MarkRead, GetUnread |
| UserListService | user_lists, list_members, list_subscriptions, list_collaborators | CreateUserList, AddMember, Subscribe, InviteCollaborator, ApplyListAction |
| AlgorithmicListService | user_lists (algorithmic) | CreateAlgorithmicList, EvaluateAlgorithmicList |
| UserAffinityService | user_affinities | ComputeAffinity (batch), GetUserAffinities, GetSimilarUsers |
| AffinityAccumulator | Redis hashes | RecordFollow, RecordCommunityJoin, RecordReaction, GetPairAffinity (O(1) badge), GetCachedAffinities (24h TTL) |
| FeedPluginService | feed_plugins, feed_plugin_installs, feed_plugin_reviews | CreatePlugin, Install, Execute, Review |
| ModDecisionReviewsService | mod_decision_reviews | CastVote, GetControversial, GetActionReviews |

## Key Handlers

| Handler | Endpoints |
|---------|-----------|
| AuthHandler | POST register, login, forgot, reset, logout; GET session |
| PostHandler | GET/POST/PUT/DELETE posts; POST archive, remove (mod) |
| TagHandler | GET/POST/DELETE tags; POST /:id/vote; POST tags/posts/:post_id/tags/:tag_id |
| SearchHandler | GET /search/posts, /search/users, /search/communities, /search/suggest |
| CommunityHandler | GET/POST communities; POST /:id/join, /:id/leave, /:id/fork |
| ModerationHandler | POST/GET actions; POST /:id/jurors, /:id/resolve; POST /jurors/:panel_id/vote |
| UserListHandler | CRUD lists; POST /:id/members, /:id/subscribe, /:id/collaborators |
| AffinityHandler | GET /affinity, /affinity/similar |
| FeedPluginHandler | POST/GET plugins; POST /:id/install, /:id/review, /:id/execute |
| ModDecisionHandler | POST /modlog/:id/review; GET /modlog/:id/reviews, /modlog/:id/my-review, /modlog/controversial |

## Database (~40 tables)

**Core**: users, posts, communities, tags, post_tags, tag_votes
**Trust**: trust_connections, trust_score_log, vouches
**Moderation**: moderation_actions, reports, jury_votes, mod_notes, **mod_decision_reviews**
**Social**: community_members, community_notes, community_note_votes
**Interaction**: interactions, hidden_posts
**Discovery**: collections, collection_posts, circles, circle_members
**Economy**: credit_transactions, achievements, user_achievements, bounties
**Safety**: blocked_users, filters, blocklist, blocklist_entries, **user_follows**, **user_lists**, list_members, list_subscriptions, list_collaborators, **user_affinities**
**Feeds**: custom_feeds, feed_sources, **feed_plugins**, feed_plugin_installs, feed_plugin_reviews
**Notifications**: user_notifications, email_queue
**Admin**: site_config, custom_pages, custom_emoji, registration_queue, oauth_states

## Current State

**Implemented (latest additions in bold):**

Auth, posts, tags with voting, full-text search with Boolean/fielded search + trigram autocomplete, communities (join/leave/slow-boot/probation/tenure-voting/fork), moderation (M1 labels + admin actions + jury), **public mod decision reviews with controversy scoring**, collections, circles, blocking, content filters, credits, achievements, notifications, onboarding, custom feeds, trending, migration runner, Dockerfile, password reset, email verification, account deletion, **user lists (CRUD + collaborative + algorithmic)**, **user affinity engine (event-driven Redis accumulator + on-demand cache)**, **WASM feed plugin marketplace**, **tag voting frontend**, **handler integration tests for new features**

**Missing from docs but not implemented**: M2 meta-moderation (jury system exists but isn't M2-style), Positive Algorithm (feed is chronological), "Why this post?" button, full reaction set (only generic like/unlike), tag autocomplete (frontend), auto-generated circles (manual only), credit hash-chain ledger (simple transactions), ActivityPub (keys exist but no inbox/outbox), WebSocket (hub exists but not wired to feed), image upload (model exists, service exists, no endpoint yet), trust level auto-progression cron, SMTP-sent password reset

**Known issues**: Test cleanup FKs when run in parallel (services + handlers share DB), some pre-existing TypeScript warnings (85+) in frontend from Lemmy API type gaps
