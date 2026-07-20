# Threadlight — Compressed Codebase Reference
For LLM agents. ~10k tokens.

## Stack
**Backend**: Go 1.26, Gin (HTTP router), pgx v5 (PostgreSQL driver), go-redis, gorilla/websocket, zerolog (logging), golang-jwt (JWT), bcrypt.
**Frontend**: Svelte 4 + Vite SPA. Plain `fetch()` API client. Svelte writable stores (`svelte/store`). Vanilla CSS with custom properties. No SvelteKit, no TypeScript, no JSX, no React.
**Storage**: PostgreSQL 17 (~31 tables), Redis 7 (cache / pub/sub / rate limit / online presence).
**Workers**: 9 goroutine-based background tasks in `internal/worker/` started by scheduler.go.

## Directory Layout
```
threadlight/
├── cmd/polaris/main.go         # Entry point: env vars, DB connect, seed admin, HTTP server
├── internal/
│   ├── api/
│   │   ├── router.go           # ~95 route registrations, wires handlers to services
│   │   ├── handlers/           # 21 files (one per resource group)
│   │   ├── middleware/         # auth.go (JWT validation), cors.go (permissive CORS), logging.go
│   │   └── websocket/          # hub.go (connection manager), client.go, message types
│   ├── db/
│   │   ├── db.go               # NewPostgres(dsn), NewRedis(addr)
│   │   ├── migrations.go       # RunMigrations(pg, dir) → reads + executes .sql files
│   │   └── migrations/         # 001_initial.sql, 002_tag_votes.sql
│   ├── model/                  # 26 struct files + errors.go
│   ├── services/               # 21 files — all business logic, all SQL queries
│   └── worker/                 # scheduler.go + 9 worker implementations
│       ├── scheduler.go        # Job registry: name + interval + func
│       ├── feed_worker.go      # Rebuild materialized feed views (30s)
│       ├── trending.go         # Update trending topics (60s)
│       ├── prune_worker.go     # Remove old posts (1h)
│       ├── achievement_worker.go  # Check/award achievements (10m)
│       ├── trust_decay.go      # Decay trust connections (6h)
│       ├── community_notes_worker.go  # Promote voted notes (5m)
│       ├── credit.go           # Distribute rewards, expire bounties, archive content (1h)
│       ├── active_stats_worker.go    # Update user stats (15m)
│       └── scheduled_post_worker.go  # Publish scheduled posts (30s)
└── web/src/
    ├── App.svelte              # Shell: navbar + sidebar + custom-event router + footer
    ├── app.css                 # Global CSS: variables, dark mode, responsive grid
    ├── main.js                 # Mounts App to #app
    ├── lib/
    │   ├── api/client.js       # fetch wrapper + 25 endpoint functions
    │   ├── stores/             # auth.js, feed.js, notifications.js
    │   └── components/         # Navbar.svelte, TrendingSidebar.svelte, Footer.svelte
    └── routes/                 # 14 page components
```

## Request Flow
```
Browser → Vite dev proxy (or nginx) → Gin router → (optional auth middleware)
  → Handler (parse params/body) → Service (business logic + SQL) → PostgreSQL
  → Handler (marshal JSON) → HTTP response
```

## Conventions (Critical Reading)

**No ORM.** Raw SQL in services using pgx. `pgxpool.Pool` injected via struct field.

**No code comments.** This file + AGENTS.md + design/DESIGN.md are the docs.

**JWT auth.** Middleware extracts JWT from `Authorization: Bearer <token>`, validates, stores `user_id` with `c.Set("user_id", uid)`.
- Get it: `uid, _ := c.Get("user_id")` → returns `interface{}`, assert `uid.(int64)`.
- Short form: `c.GetInt64("user_id")` → returns `int64` directly.

**Two auth modes:**
- `authMw.RequireAuth()` → returns 401 if no valid JWT, sets `user_id` in context
- `authMw.OptionalAuth()` → passes through without JWT, sets `user_id` only if valid JWT present

**Error sentinels** (model/errors.go):
- `ErrNotFound` → 404 ("resource not found")
- `ErrForbidden` → 403
- `ErrUnauthorized` → 401
- `ErrConflict` → 409
- `ErrValidation` → 400

**PGX scanning**: NULL columns MUST scan into pointer types: `*time.Time`, `*string`, `*int64`, `*float64`. Scanning a NULL into a non-pointer type will PANIC.

**Ownership checks**: Every mutation must verify the caller owns the resource. Pattern:
```go
func (s *PostService) UpdatePost(ctx, postID, userID int64, req) error {
    var authorID int64
    err := s.pg.QueryRow(ctx, `SELECT author_id FROM posts WHERE id = $1`, postID).Scan(&authorID)
    if err != nil { return model.ErrNotFound }
    if authorID != userID { return model.ErrForbidden }
}
```

**JSON responses**: Always `c.JSON(http.StatusXxx, gin.H{"key": val})` or `c.JSON(http.StatusXxx, modelStruct)`.
Errors: `c.JSON(http.StatusXxx, gin.H{"error": "message"})`.

## All Routes (router.go)

### Health
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | /health | healthHandler.Check | none |

### About
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | /api/v1/about | aboutHandler.Get | none |

### Auth
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| POST | /api/v1/auth/register | authHandler.Register | none |
| POST | /api/v1/auth/login | authHandler.Login | none |
| POST | /api/v1/auth/forgot | authHandler.ForgotPassword | none |
| POST | /api/v1/auth/reset | authHandler.ResetPassword | none |
| POST | /api/v1/auth/logout | authHandler.Logout | Required |
| GET | /api/v1/auth/session | authHandler.Session | Required |

### Posts
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/posts | postHandler.Create |
| GET | /api/v1/posts | postHandler.List |
| GET | /api/v1/posts/count | postHandler.GetCount |
| GET | /api/v1/posts/author/:author_id | postHandler.ListByAuthor |
| GET | /api/v1/posts/:id | postHandler.GetByID |
| PUT | /api/v1/posts/:id | postHandler.Update |
| DELETE | /api/v1/posts/:id | postHandler.Delete |
| POST | /api/v1/posts/:id/archive | postHandler.Archive |
| POST | /api/v1/posts/:id/remove | postHandler.ModRemove |

### Tags
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/tags | tagHandler.Create |
| GET | /api/v1/tags | tagHandler.List |
| GET | /api/v1/tags/:id | tagHandler.GetByID |
| PUT | /api/v1/tags/:id | tagHandler.Update |
| DELETE | /api/v1/tags/:id | tagHandler.Delete |
| POST | /api/v1/tags/:id/vote | tagHandler.VoteTag |
| POST | /api/v1/tags/posts/:post_id/tags/:tag_id | tagHandler.TagPost |
| DELETE | /api/v1/tags/posts/:post_id/tags/:tag_id | tagHandler.UntagPost |
| GET | /api/v1/tags/posts/:post_id/tags | tagHandler.GetPostTags |

### Feeds
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/feeds | feedHandler.GetFeed |
| POST | /api/v1/feeds | feedHandler.Create |
| GET | /api/v1/feeds/:id | feedHandler.GetByID |
| PUT | /api/v1/feeds/:id | feedHandler.Update |
| DELETE | /api/v1/feeds/:id | feedHandler.Delete |
| GET | /api/v1/feeds/:id/sources | feedHandler.GetSources |
| POST | /api/v1/feeds/:id/sources | feedHandler.AddSource |
| DELETE | /api/v1/feeds/sources/:source_id | feedHandler.RemoveSource |

### Search
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/search | searchHandler.Search |
| POST | /api/v1/search/advanced | searchHandler.AdvancedSearch |

### Trending
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/trending | trendingHandler.GetTopics |
| POST | /api/v1/trending/increment | trendingHandler.Increment |

### Interactions
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/interactions | interactionHandler.Create |
| GET | /api/v1/interactions/post/:postID | interactionHandler.GetByPost |
| DELETE | /api/v1/interactions/:id | interactionHandler.Remove |
| GET | /api/v1/interactions/check | interactionHandler.Check |

### Trust
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/trust | trustHandler.List |
| POST | /api/v1/trust | trustHandler.Create |
| GET | /api/v1/trust/incoming | trustHandler.Incoming |
| GET | /api/v1/trust/outgoing | trustHandler.Outgoing |
| GET | /api/v1/trust/connections/:id | trustHandler.GetByID |
| PUT | /api/v1/trust/connections/:id | trustHandler.Update |
| DELETE | /api/v1/trust/connections/:id | trustHandler.Delete |

### Achievements
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | /api/v1/achievements | achievementHandler.List | Required |
| GET | /api/v1/achievements/user/:user_id | achievementHandler.GetUserAchievements | Required |
| POST | /api/v1/achievements/unlock | achievementHandler.Unlock | Required |

### Blocklists
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/blocklists | blocklistHandler.Create |
| GET | /api/v1/blocklists | blocklistHandler.List |
| GET | /api/v1/blocklists/:id | blocklistHandler.GetByID |
| POST | /api/v1/blocklists/:id/subscribe | blocklistHandler.Subscribe |
| POST | /api/v1/blocklists/:id/unsubscribe | blocklistHandler.Unsubscribe |
| POST | /api/v1/blocklists/:id/entries | blocklistHandler.AddEntry |
| DELETE | /api/v1/blocklists/entries/:id | blocklistHandler.RemoveEntry |

### Circles
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/circles | circleHandler.Create |
| GET | /api/v1/circles | circleHandler.List |
| GET | /api/v1/circles/suggestions | circleHandler.Suggestions |
| GET | /api/v1/circles/:id | circleHandler.GetByID |
| GET | /api/v1/circles/:id/members | circleHandler.Members |
| PUT | /api/v1/circles/:id | circleHandler.Update |
| POST | /api/v1/circles/:id/join | circleHandler.Join |
| POST | /api/v1/circles/:id/leave | circleHandler.Leave |
| POST | /api/v1/circles/:id/suggest | circleHandler.Suggest |

### Collections
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/collections | collectionHandler.Create |
| GET | /api/v1/collections | collectionHandler.List |
| GET | /api/v1/collections/:id | collectionHandler.GetByID |
| GET | /api/v1/collections/:id/posts | collectionHandler.GetPosts |
| PUT | /api/v1/collections/:id | collectionHandler.Update |
| DELETE | /api/v1/collections/:id | collectionHandler.Delete |
| POST | /api/v1/collections/:id/posts | collectionHandler.AddPost |
| DELETE | /api/v1/collections/:id/posts/:post_id | collectionHandler.RemovePost |

### Communities
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/communities | communityHandler.Create |
| GET | /api/v1/communities | communityHandler.List |
| GET | /api/v1/communities/:slug | communityHandler.GetBySlug |
| GET | /api/v1/communities/:slug/members | communityHandler.Members |
| GET | /api/v1/communities/:slug/trust | communityHandler.Trust |
| GET | /api/v1/communities/:slug/trust/score | communityHandler.TrustScore |
| GET | /api/v1/communities/:slug/join/check | communityHandler.JoinCheck |
| PUT | /api/v1/communities/:slug | communityHandler.Update |
| DELETE | /api/v1/communities/:slug | communityHandler.Delete |
| POST | /api/v1/communities/:slug/join | communityHandler.Join |
| POST | /api/v1/communities/:slug/leave | communityHandler.Leave |
| POST | /api/v1/communities/:slug/curators | communityHandler.AddCurator |
| DELETE | /api/v1/communities/:slug/curators/:user_id | communityHandler.RemoveCurator |
| POST | /api/v1/communities/:slug/fork | communityHandler.Fork |
| POST | /api/v1/communities/:slug/trust | communityHandler.AddTrust |

### Credits
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/credits | creditHandler.GetBalance |
| GET | /api/v1/credits/bounties | bountyHandler.List |
| POST | /api/v1/credits/bounties | bountyHandler.Create |
| GET | /api/v1/credits/bounties/:id | bountyHandler.GetByID |
| POST | /api/v1/credits/bounties/:id/award | bountyHandler.Award |

### Filters
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/filters | filterHandler.Create |
| GET | /api/v1/filters | filterHandler.List |
| PUT | /api/v1/filters/:id | filterHandler.Update |
| DELETE | /api/v1/filters/:id | filterHandler.Delete |

### Notes
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/notes | noteHandler.Create |
| GET | /api/v1/notes/:id | noteHandler.GetByID |
| PUT | /api/v1/notes/:id | noteHandler.Update |
| GET | /api/v1/notes/post/:postID | noteHandler.GetByPost |
| POST | /api/v1/notes/:id/vote | noteHandler.Vote |

### Moderation
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/moderation/actions | moderationHandler.CreateAction |
| GET | /api/v1/moderation/actions | moderationHandler.ListActions |
| GET | /api/v1/moderation/actions/:id | moderationHandler.GetAction |
| POST | /api/v1/moderation/actions/:id/jurors | moderationHandler.AddJuror |
| POST | /api/v1/moderation/actions/:id/resolve | moderationHandler.ResolveAction |
| POST | /api/v1/moderation/jurors/:panel_id/vote | moderationHandler.JuryVote |

### Reports
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/reports | reportHandler.Create |
| GET | /api/v1/reports | reportHandler.List |
| GET | /api/v1/reports/:id | reportHandler.GetByID |
| PUT | /api/v1/reports/:id/resolve | reportHandler.Resolve |

### Notifications
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/notifications | userHandler.GetNotifications |
| PUT | /api/v1/notifications/:id/read | userHandler.MarkNotificationRead |
| PUT | /api/v1/notifications/read-all | userHandler.MarkAllRead |
| GET | /api/v1/notifications/unread-count | userHandler.UnreadCount |

### Users
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | /api/v1/users/:username | userHandler.GetProfileByUsername | Optional |
| GET | /api/v1/users/@me | userHandler.GetProfile | Required |
| PUT | /api/v1/users/@me | userHandler.UpdateProfile | Required |
| POST | /api/v1/users/@me/avatar | userHandler.UploadAvatar | Required |
| POST | /api/v1/users/@me/banner | userHandler.UploadBanner | Required |
| GET | /api/v1/users/@me/notifications | userHandler.GetNotifications | Required |
| PUT | /api/v1/users/@me/notifications | userHandler.UpdateNotificationPreferences | Required |
| PUT | /api/v1/users/@me/privacy | userHandler.UpdatePrivacySettings | Required |
| GET | /api/v1/users/@me/blocks | userHandler.ListBlockedUsers | Required |
| POST | /api/v1/users/@me/trust/:user_id | userHandler.TrustUser | Required |
| POST | /api/v1/users/:id/block | userHandler.BlockUser | Required |
| DELETE | /api/v1/users/:id/block | userHandler.UnblockUser | Required |

### Blocks
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/blocks | blockHandler.BlockUser |
| GET | /api/v1/blocks | blockHandler.List |
| DELETE | /api/v1/blocks/:id | blockHandler.Delete |
| GET | /api/v1/blocks/check/:user_id | blockHandler.Check |

### Admin
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| GET | /api/v1/admin/config | configHandler.GetConfig |
| PUT | /api/v1/admin/config | configHandler.UpdateConfig |
| POST | /api/v1/admin/invites | configHandler.CreateInvite |
| GET | /api/v1/admin/invites | configHandler.ListInvites |
| GET | /api/v1/admin/invites/limit | configHandler.GetInviteLimit |

### Invites
All require auth.
| Method | Path | Handler |
|--------|------|---------|
| POST | /api/v1/invites | inviteHandler.Create |
| GET | /api/v1/invites | inviteHandler.List |
| GET | /api/v1/invites/limit | inviteHandler.GetLimit |

## Services (21 files in internal/services/)

### AuthService (auth.go)
- `Register(ctx, req) → (model.User, string, error)` — email/username uniqueness check, bcrypt hash, INSERT user, generate JWT
- `Login(ctx, emailOrUsername, password) → (model.User, string, error)` — lookup by email/username, bcrypt compare, JWT
- `ValidateToken(tokenStr) → (int64, error)` — parse/validate JWT, return user_id
- `ForgotPassword(ctx, email) → (string, error)` — generate reset token, store with expiry
- `ResetPassword(ctx, token, newPassword) → error` — validate token within 1hr, bcrypt hash

### PostService (post.go)
- `CreatePost(ctx, post) → (*model.PostResponse, error)` — INSERT posts, RETURNING full row
- `GetPost(ctx, id, uid) → (*model.Post, error)` — SELECT with NOT is_deleted
- `UpdatePost(ctx, postID, userID, req) → (*model.PostResponse, error)` — UPDATE with COALESCE, checks author_id and 15-min window
- `GetFeedPosts(ctx, uid, filter) → ([]*model.Post, int, error)` — JOIN feeds + communities + tags, paginated CTE
- `ArchivePost(ctx, postID, userID) → error` — sets archived_at + is_deleted
- `VotePost(ctx, postID, userID, vote) → error` — upsert interaction, update interaction_count
- `GetAuthorPosts(ctx, authorID, page, limit) → ([]*model.Post, error)` — list by author
- `GetPostCount(ctx) → (int64, error)` — count active posts

### TagService (tag.go)
- `CreateTag(ctx, name, desc, cat, userID) → (*model.Tag, error)`
- `ListTags(ctx) → ([]*model.Tag, error)`
- `GetTagByID(ctx, id) → (*model.Tag, error)`
- `UpdateTag(ctx, id, name, desc, cat) → (*model.Tag, error)`
- `DeleteTag(ctx, id) → error`
- `TagPost(ctx, postID, tagID, userID) → error` — INSERT INTO post_tags ON CONFLICT DO NOTHING
- `UntagPost(ctx, postID, tagID) → error` — DELETE FROM post_tags
- `GetPostTags(ctx, postID) → ([]*model.PostTag, error)`
- `VoteTag(ctx, tagID, userID, vote) → error` — INSERT/UPDATE tag_votes

### FeedService (feed.go)
- `GetFeed(ctx, uid, type, page, limit) → ([]*model.Post, int, error)`
- `GetCustomFeed(ctx, feedID, uid) → ([]*model.Post, error)`
- `CreateFeed(ctx, uid, name, desc) → (*model.CustomFeed, error)`
- `UpdateFeed(ctx, feedID, uid, name, desc) → error`
- `DeleteFeed(ctx, feedID, uid) → error`
- `AddSource(ctx, feedID, source) → error`
- `GetSources(ctx, feedID) → ([]*model.FeedSource, error)`
- `RemoveSource(ctx, sourceID) → error`

### SearchService (search.go)
- `Search(ctx, query, searchType, limit, offset) → ([]*model.Post, error)` — parses AND/OR/NOT, fielded search (author:, community:, tag:, after:, before:), PostgreSQL tsquery, ranked results

### UserService (user.go)
- `GetProfile(ctx, userID) → (*model.User, error)`
- `GetProfileByUsername(ctx, username) → (*model.User, error)`
- `UpdateProfile(ctx, userID, updates) → (*model.User, error)` — COALESCE for display_name, bio, avatar_url, banner_url, theme, hide_read_posts, etc.
- `GetNotifications(ctx, userID, page, limit) → ([]*model.Notification, int, error)`
- `MarkNotificationRead(ctx, userID, notifID) → error`
- `MarkAllNotificationsRead(ctx, userID) → error`
- `GetUnreadCount(ctx, userID) → (int64, error)`
- `BlockUser(ctx, userID, blockedID) → error` — checks self-block, user existence, INSERT blocked_users
- `UnblockUser(ctx, userID, blockedID) → error`
- `ListBlockedUsers(ctx, userID) → ([]model.BlockedUser, error)`

### BlockService (block.go)
- `BlockUser(ctx, blockerID, blockedID) → (*model.BlockedUser, error)` — checks self-block and user existence
- `ListBlocked(ctx, userID) → ([]*model.BlockedUser, error)`
- `Unblock(ctx, blockID, userID) → error`
- `CheckBlocked(ctx, userID, targetID) → (bool, error)`

### Other services (same pattern): CommunityService, CollectionService, CircleService, NoteService, FilterService, ReportService, ModerationService, NotificationService, AchievementService, CreditService, TrendService, ConfigService, TrustService, BountyService, InviteService.

## Database Summary (~31 tables)

### users (25 columns)
id (PK bigserial), username (varchar UNIQUE), email (varchar UNIQUE), password_hash (varchar), display_name (*varchar), bio (*text), avatar_url (*varchar), banner_url (*varchar), bio_html (*text), trust_level (int2 DEFAULT 0), trust_score (float8 DEFAULT 0), reputation (int8 DEFAULT 0), credits (int8 DEFAULT 0), onboarding_stage (int2 DEFAULT 0), is_active (bool), is_admin (bool DEFAULT false), email_verified (bool DEFAULT false), email_verify_token (*varchar), password_reset_token (*varchar), password_reset_sent_at (*timestamptz), actor_id (*varchar), public_key (*text), private_key (*text), invited_by (*int8 FK self), invite_code (*varchar), theme (*varchar), proximity_opt_out (bool), hide_read_posts (bool), location_hash (*varchar), last_active_at (*timestamptz), is_local (bool DEFAULT true), created_at (timestamptz), updated_at (timestamptz).

### posts (25 columns)
id (PK bigserial), author_id (FK users NOT NULL), title (text), body (text), content_type (int2 DEFAULT 0), mood (int2 DEFAULT 0), is_educational (bool), is_entertaining (bool), is_nsfw (bool), content_warning (*varchar), interaction_count (int8 DEFAULT 0), cumulative_interactions (int8 DEFAULT 0), reply_to_id (*int8 FK self), status (int2 DEFAULT 0), scheduled_at (*timestamptz), locked (bool DEFAULT false), sticky (bool DEFAULT false), archived_at (*timestamptz), is_deleted (bool DEFAULT false), edited_at (*timestamptz), likes (int4 DEFAULT 0), created_at (timestamptz), updated_at (timestamptz).

### Other tables: communities, community_members, community_forks, curators, tags, post_tags, tag_votes, custom_feeds, feed_sources, collections, collection_posts, circles, circle_members, community_notes, community_note_votes, blocked_users, blocklists, blocklist_entries, content_filters, reports, moderation_actions, jury_panels, user_notifications, achievements, user_achievements, credit_transactions, bounties, bounty_claims, daily_rewards, trust_connections, hidden_posts, site_config, custom_pages, custom_emoji, registration_queue, oauth_states, mod_notes, saved_searches, user_settings, invite_codes.

## Workers (scheduler.go)
| Name | Interval | Function |
|------|----------|----------|
| feed_rebuild | 30s | RunFeedRebuildWorker — refresh materialized feed views |
| trending | 60s | RunTrendingWorker — update trending scores |
| trust_decay | 6h | RunTrustDecayWorker — decay old trust connections |
| prune | 1h | RunPruneWorker — soft-delete old posts with low engagement |
| achievements | 10m | RunAchievementWorker — check/award achievements |
| community_notes | 5m | RunCommunityNotesWorker — promote voted notes |
| credit | 1h | RunCreditDistributor — activity rewards; bounty/archive expiry |
| active_stats | 15m | RunActiveStatsWorker — update user activity stats |
| scheduled | 30s | RunScheduledPostWorker — publish posts where scheduled_at ≤ NOW() |

## Known Issues
1. **Frontend error handling**: Many pages have `catch(e) { /* ignore */ }` — users never see errors or loading states.
2. **No pagination UI**: FeedPage shows initial batch but no "load more" button despite backend pagination support.
3. **SettingsPage monolith**: 672-line single component handling profile, filters, credits, and account.
4. **Hardcoded sample data**: LeaderboardPage, ModCenter use fake arrays instead of calling API.
5. **XSS in notifications**: `{@html notif.body}` on NotificationsPage.svelte. Replace with `{notif.body}`.
6. **Svelte runes mode mismatch**: `svelte.config.js` has `runes: true` but most components use legacy syntax.
7. **Test parallelism**: services_test.go and handlers_test.go both operate in the same DB with no cleanup between them.
8. **Migration runner not wired**: `internal/db/migrations.go` exists, `main.go` imports db but doesn't call `db.RunMigrations()`.
