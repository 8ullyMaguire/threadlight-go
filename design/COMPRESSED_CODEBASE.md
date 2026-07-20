# Threadlight — Compressed Codebase Reference
For LLM agents. ~10k tokens.

## Stack
**Backend**: Go 1.26, Gin (HTTP router), pgx v5 (PostgreSQL driver), go-redis, gorilla/websocket, zerolog (logging), golang-jwt (JWT), bcrypt.
**Frontend**: Svelte 4 + Vite SPA. Plain `fetch()` API client. Svelte writable stores (`svelte/store`). Vanilla CSS with custom properties. No SvelteKit, no TypeScript, no JSX, no React.
**Storage**: PostgreSQL 17 (~31 tables), Redis 7 (cache / pub/sub / rate limit / online presence).
**Workers**: 11 goroutine-based background tasks in `internal/worker/` started by scheduler.go.

## Directory Layout
```
threadlight/
├── cmd/polaris/main.go         # Entry point: env vars, DB connect, seed admin, HTTP server
├── internal/
│   ├── api/
│   │   ├── router.go           # ~95 route registrations, wires handlers to services
│   │   ├── handlers/           # 23 files (one per resource group)
│   │   ├── middleware/         # auth.go (JWT validation), cors.go (permissive CORS)
│   │   └── websocket/          # hub.go (connection manager), handlers.go (events)
│   ├── db/
│   │   ├── db.go               # NewPostgres(dsn), NewRedis(addr)
│   │   ├── migrations.go       # RunMigrations(pg, dir) → reads + executes .sql files
│   │   └── migrations/         # 001_initial.sql, 002_tag_votes.sql
│   ├── model/                  # 26 struct files + errors.go
│   ├── services/               # 21 files — all business logic, all SQL queries
│   └── worker/                 # scheduler.go + 10 worker implementations
│       ├── scheduler.go        # Job registry: name + interval + func
│       ├── prune_worker.go     # Remove old posts (env configurable)
│       ├── achievement_worker.go
│       ├── trending_worker.go
│       ├── feed_refresh_worker.go
│       ├── notes_promote_worker.go
│       ├── credit_worker.go
│       ├── active_stats_worker.go
│       ├── scheduled_posts_worker.go
│       ├── email_worker.go
│       └── stale_sessions_worker.go
└── web/src/
    ├── App.svelte              # Shell: navbar + sidebar + custom-event router + footer
    ├── app.css                 # Global CSS: variables, dark mode, responsive grid
    ├── main.js                 # Mounts App to #app
    ├── lib/
    │   ├── api/client.js       # fetch wrapper + 25 endpoint functions
    │   ├── stores/             # auth.js, feed.js, notifications.js
    │   └── components/         # Navbar.svelte, TrendingSidebar.svelte, Footer.svelte
    └── routes/                 # 14 page components
        FeedPage, ProfilePage, SettingsPage, LoginPage, RegisterPage, SearchPage,
        CommunitiesPage, CollectionsPage, NotificationsPage, LeaderboardPage,
        AdminPage, AboutPage, OnboardingWizard, ModCenter, BlockedUsersPage
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

**Error sentinels** (model/errors.go):
- `ErrNotFound` → 404
- `ErrForbidden` → 403
- `ErrUnauthorized` → 401
- `ErrConflict` → 409
- `ErrValidation` → 400

**PGX scanning**: NULL columns MUST scan into pointer types: `*time.Time`, `*string`, `*int64`, `*float64`. Scanning a NULL into a non-pointer type (string, int64, time.Time) will PANIC at runtime with "can't scan into ...: cannot scan NULL into ...".

**Service pattern**: Every method takes `ctx context.Context` as first arg. Returns `(result, error)` or just `error`. Use sentinels for known error states.

**Ownership checks**: Every mutation must verify the caller owns the resource. Pattern:
```go
func (s *PostService) UpdatePost(ctx, postID, userID int64, req) error {
    var authorID int64
    err := s.pg.QueryRow(ctx, `SELECT author_id FROM posts WHERE id = $1`, postID).Scan(&authorID)
    if err != nil { return model.ErrNotFound }
    if authorID != userID { return model.ErrForbidden }
    // proceed with update
}
```

**Nullable struct fields**: Use `*string` for nullable text, `*time.Time` for nullable timestamps, `*int64` for nullable integers. Tag with `json:"name,omitempty"`.

**Middleware ordering**: `router.Use(middleware.Logging(logger))` after `NewRouter()`. Route groups use `authMw.RequireAuth()` or `authMw.OptionalAuth()`.

**Time fields**: All PostgreSQL timestamps are `TIMESTAMPTZ`. Go `time.Time` for NOT NULL, `*time.Time` for nullable. Default `DEFAULT NOW()` in SQL.

**JSON responses**: Always `c.JSON(http.StatusXxx, gin.H{"key": val})` or `c.JSON(http.StatusXxx, modelStruct)`.

**Workers**: scheduler.go starts goroutines for each worker. Panics caught with `recover()`. Workers stop via `ctx.Done()` when scheduler stops.

## Handlers (23 files in internal/api/handlers/)

### AuthHandler
| Method | Route | Description |
|--------|-------|-------------|
| Register | POST /api/v1/auth/register | Create account. Returns JWT + user |
| Login | POST /api/v1/auth/login | Email or username + password → JWT. Checks by email first, falls back to username |
| Logout | POST /api/v1/auth/logout | Invalidate session (auth required) |
| Session | GET /api/v1/auth/session | Returns current user_id from JWT |
| ForgotPassword | POST /api/v1/auth/forgot | Accepts email, generates reset token, stores in DB, returns token (TODO: email sending) |
| ResetPassword | POST /api/v1/auth/reset | Accepts token + new_password, validates expiry (<1hr), updates hash |

### PostHandler
| Method | Route | Description |
|--------|-------|-------------|
| GetFeed | GET /api/v1/posts | Paginated feed, ?type=hot/new/top/trusted, ?mood=, ?community_id= |
| Create | POST /api/v1/posts | Create post with title, body, content_type, mood, is_educational, is_entertaining, is_nsfw, content_warning |
| GetPost | GET /api/v1/posts/:id | Single post by ID |
| Update | PUT /api/v1/posts/:id | Update within 15 minutes of creation |
| Delete | DELETE /api/v1/posts/:id | Soft delete (archived_at + is_deleted). Own posts only |
| Archive | POST /api/v1/posts/:id/archive | Same as Delete |
| ModRemove | POST /api/v1/posts/:id/remove | Mod-only: trust_level >= 3 can remove any post |
| VotePost | POST /api/v1/posts/:id/vote | Vote on post |

### TagHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/tags | All tags |
| Create | POST /api/v1/tags | Create tag (name, description, category) |
| GetByID | GET /api/v1/tags/:id | Single tag |
| Delete | DELETE /api/v1/tags/:id | Delete tag |
| TagPost | POST /api/v1/tags/posts/:post_id/tags/:tag_id | Associate tag with post |
| UntagPost | DELETE /api/v1/tags/posts/:post_id/tags/:tag_id | Remove tag from post |
| GetPostTags | GET /api/v1/tags/posts/:post_id/tags | Get post's tags |
| VoteTag | POST /api/v1/tags/:id/vote | Vote on tag (body: {"vote": 1 or -1}) → returns total score |

### FeedHandler
| Method | Route | Description |
|--------|-------|-------------|
| GetFeed | GET /api/v1/feeds | Chronological feed. ?type, ?page, ?limit |
| GetCustomFeed | GET /api/v1/feeds/:id | Saved custom feed |
| Create | POST /api/v1/feeds | Create custom feed |
| Delete | DELETE /api/v1/feeds/:id | Delete custom feed |

### SearchHandler
| Method | Route | Description |
|--------|-------|-------------|
| Search | GET /api/v1/search | Full-text search. Supports: `q` (query with AND/OR/NOT), `type` (posts/users/communities/tags), fielded `author:`, `community:`, `tag:`, date `after:`, `before:` |
| AdvancedSearch | POST /api/v1/search/advanced | Same but POST with JSON body |

### CommunityHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/communities | Public communities |
| Create | POST /api/v1/communities | Create with name, description, rules, is_invite_only, slow_boot_days |
| GetByID | GET /api/v1/communities/:id | Single community |
| Join | POST /api/v1/communities/:id/join | Join (checks invite-only, slow boot) |
| Leave | POST /api/v1/communities/:id/leave | Leave community |

### CollectionHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/collections | User's collections |
| Create | POST /api/v1/collections | Create collection (name, description, visibility) |
| AddPost | POST /api/v1/collections/:id/posts | Add post to collection |
| RemovePost | DELETE /api/v1/collections/:id/posts/:post_id | Remove post from collection |

### CircleHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/circles | User's circles |
| Create | POST /api/v1/circles | Create circle (name, topic, location_hash) |
| Join | POST /api/v1/circles/:id/join | Join circle |
| Leave | POST /api/v1/circles/:id/leave | Leave circle |

### NoteHandler (Community Notes)
| Method | Route | Description |
|--------|-------|-------------|
| Create | POST /api/v1/posts/:id/notes | Add community note to post |
| GetNotes | GET /api/v1/posts/:id/notes | Get notes for post |
| VoteNote | POST /api/v1/notes/:id/vote | Vote on note (up/down) |

### ModerationHandler
| Method | Route | Description |
|--------|-------|-------------|
| ApplyAction | POST /api/v1/moderation/action | Create mod action (warn/suspend/ban). Requires target_user_id, action_type, duration (for suspend), reason |
| CreateReport | POST /api/v1/moderation/report | Report content. target_type (post/user/community), target_id, reason |
| VoteOnJury | POST /api/v1/moderation/jury/:id/vote | Vote on a jury/due-process case |
| GetJuryCases | GET /api/v1/moderation/jury | List open jury cases |
| GetModLog | GET /api/v1/moderation/log | Public moderation log |

### BlockHandler
| Method | Route | Description |
|--------|-------|-------------|
| BlockUser | POST /api/v1/users/:id/block | Block user |
| UnblockUser | DELETE /api/v1/users/:id/block | Unblock user |
| GetBlocked | GET /api/v1/users/me/blocked | List blocked users |

### BlocklistHandler (Shared Blocklist)
| Method | Route | Description |
|--------|-------|-------------|
| Get | GET /api/v1/blocklist | Export blocklist as JSON |
| Subscribe | POST /api/v1/blocklist/subscribe | Subscribe to remote blocklist URL |

### FilterHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/filters | User's content filters |
| Create | POST /api/v1/filters | Create filter (type: keyword/regex/domain, pattern) |
| Delete | DELETE /api/v1/filters/:id | Delete filter |

### NotificationHandler
| Method | Route | Description |
|--------|-------|-------------|
| List | GET /api/v1/notifications | User's notifications |
| MarkRead | PUT /api/v1/notifications/:id/read | Mark notification read |
| MarkAllRead | PUT /api/v1/notifications/read-all | Mark all read |

### AdminHandler
| Method | Route | Description |
|--------|-------|-------------|
| GetStats | GET /api/v1/admin/stats | Instance statistics (user count, post count, etc.) |
| SeedTags | POST /api/v1/admin/seed/tags | Seed default tags |

### OnboardingHandler
| Method | Route | Description |
|--------|-------|-------------|
| PickTags | POST /api/v1/onboarding/tags | Save selected tags for user |
| GetSuggested | GET /api/v1/onboarding/suggested | Get suggested users (active in selected tags + top by trust) |
| AdvanceStage | POST /api/v1/onboarding/stage | Increment onboarding_stage |

### UserHandler
| Method | Route | Description |
|--------|-------|-------------|
| Me | GET /api/v1/users/me | Current user |
| UpdateProfile | PUT /api/v1/users/me | Update display_name, bio, avatar_url, theme, etc. |
| GetUser | GET /api/v1/users/:id | Public profile |
| TrustUser | POST /api/v1/users/me/trust/:id | Trust/untrust user |

## Services (21 files in internal/services/)

### AuthService
- `Register(ctx, req) → (model.User, error)` — email/username uniqueness check, bcrypt hash, INSERT with onboarding_stage=4, RETURNING full user
- `Login(ctx, emailOrUsername, password) → (model.User, string, error)` — looks up by email, falls back to username; bcrypt compare; generates JWT (24h expiry with user_id claim)
- `ForgotPassword(ctx, email) → (string, error)` — generates 32-byte random hex token, stores with timestamp, returns token
- `ResetPassword(ctx, token, newPassword) → error` — validates token within 1hr, bcrypt hashes new password, clears token
- `generateToken(userID int64) → (string, error)` — JWT with user_id, exp, iat claims

### PostService
- `CreatePost(ctx, post) → (*model.PostResponse, error)` — INSERT posts, RETURNING full row
- `GetPost(ctx, id, uid) → (*model.Post, error)` — SELECT with NOT is_deleted
- `UpdatePost(ctx, postID, userID, req) → (*model.PostResponse, error)` — UPDATE with COALESCE for each field, checks author_id
- `GetFeedPosts(ctx, uid, filter) → ([]*model.Post, int, error)` — JOIN against feed_sources + communities + tags. Paginated with CTE
- `ArchivePost(ctx, postID, userID) → error` — sets archived_at + is_deleted. Checks author_id
- `VotePost(ctx, postID, userID, vote) → error` — upsert in interactions table, update interaction_count

### TagService
- `CreateTag(ctx, name, description, category, userID) → (*model.Tag, error)` — INSERT name, description, category
- `ListTags(ctx) → ([]*model.Tag, error)` — SELECT all
- `GetTagByID(ctx, id) → (*model.Tag, error)`
- `DeleteTag(ctx, id) → error`
- `TagPost(ctx, postID, tagID, userID) → error` — INSERT INTO post_tags
- `UntagPost(ctx, postID, tagID) → error` — DELETE FROM post_tags
- `GetPostTags(ctx, postID) → ([]*model.PostTag, error)`
- `SearchTags(ctx, query, limit) → ([]*model.Tag, error)` — ILIKE search
- `VoteTag(ctx, tagID, userID, vote) → error` — INSERT/UPDATE tag_votes
- `GetTagVoteCount(ctx, tagID) → (int, error)` — SUM(vote) FROM tag_votes

### FeedService
- `GetFeed(ctx, uid, type, page, limit) → ([]*model.Post, int, error)` — chronological feed with pagination
- `GetCustomFeed(ctx, feedID, uid) → ([]*model.Post, error)` — saved feed with sources

### SearchService
- `Search(ctx, query, searchType, limit, offset) → ([]*model.Post, error)` — Parses query for AND/OR/NOT, author:, community:, tag:, after:, before: fields. Builds PostgreSQL tsquery. Returns ranked results.

### UserService
- `GetProfile(ctx, userID) → (*model.User, error)`
- `GetProfileByUsername(ctx, username) → (*model.User, error)`
- `UpdateProfile(ctx, userID, updates) → (*model.User, error)` — COALESCE updates for bio, display_name, avatar_url, banner_url, theme, hide_read_posts
- `GetUserByID(ctx, userID) → (*model.User, error)` — full user row including trust_level
- `GetUserTrustLevel(ctx, userID) → (int16, error)` — SELECT trust_level FROM users

### ModerationService
- `CreateAction(ctx, moderatorID, targetUserID, targetPostID, actionType, duration, reason) → error` — INSERT into moderation_actions
- `CreateReport(ctx, reporterID, reportedID, targetType, targetID, reason) → error` — INSERT into reports
- `VoteOnCase(ctx, jurorID, caseID, vote) → error` — INSERT into jury_votes
- `GetJuryCases(ctx) → ([]*Report, error)` — SELECT open reports

### CommunityService
- `Create(ctx, req, ownerID) → (*model.Community, error)` — INSERT with slow_boot_days, is_invite_only
- `Join(ctx, communityID, userID) → error` — Checks invite-only, slow boot limits
- `Leave(ctx, communityID, userID) → error` — DELETE FROM community_members
- `GetCommunity(ctx, id) → (*model.Community, error)`

### BlockService
- `BlockUser(ctx, userID, blockedID) → error`
- `IsBlocked(ctx, userID, targetID) → (bool, error)`
- `UnblockUser(ctx, userID, blockedID) → error`
- `GetBlockedUsers(ctx, userID) → ([]int64, error)`

### CreditService
- `Award(ctx, userID, amount, reason) → error`
- `Spend(ctx, userID, amount, reason) → error`
- `GetBalance(ctx, userID) → (int64, error)`
- `Transfer(ctx, fromID, toID, amount) → error`

### NotificationService
- `Send(ctx, userID, notifType, sourceUserID, refType, refID, body) → error`
- `GetNotifications(ctx, userID, page, limit) → ([]*Notification, int, error)`
- `MarkRead(ctx, notificationID) → error`

### AchievementService
- `CheckAndAward(ctx, userID) → ([]*Achievement, error)` — checks criteria: first_post, first_like, 100_posts, trust_level_up, etc.
- `GetUserAchievements(ctx, userID) → ([]*UserAchievement, error)`

### Other services (same pattern): CollectionService, CircleService, NoteService, FilterService, ReportService, TrendingService, AdminService, OnboardingService.

## Database Summary (~31 tables)

### users (25 columns)
id (PK bigserial), username (varchar UNIQUE), email (varchar UNIQUE), password_hash (varchar), display_name (*varchar), bio (*text), avatar_url (*varchar), banner_url (*varchar), bio_html (*text), trust_level (int2 DEFAULT 0), trust_score (float8 DEFAULT 0), reputation (int8 DEFAULT 0), credits (int8 DEFAULT 0), onboarding_stage (int2 DEFAULT 0), is_active (bool), is_admin (bool DEFAULT false), email_verified (bool DEFAULT false), email_verify_token (*varchar), password_reset_token (*varchar), password_reset_sent_at (*timestamptz), actor_id (*varchar), public_key (*text), private_key (*text), invited_by (*int8 FK self), invite_code (*varchar), theme (*varchar), proximity_opt_out (bool), hide_read_posts (bool), location_hash (*varchar), last_active_at (*timestamptz), is_local (bool DEFAULT true), created_at (timestamptz), updated_at (timestamptz).
Indexes: username, email, actor_id.

### posts (25 columns)
id (PK bigserial), author_id (FK users NOT NULL), title (text), body (text), content_type (int2 DEFAULT 0), mood (int2 DEFAULT 0), is_educational (bool), is_entertaining (bool), is_nsfw (bool), content_warning (*varchar), interaction_count (int4 DEFAULT 0), cumulative_interactions (int8 DEFAULT 0), status (int2 DEFAULT 0), reply_to_id (*int8 FK self), scheduled_at (*timestamptz), locked (bool DEFAULT false), sticky (bool DEFAULT false), archived_at (*timestamptz), is_deleted (bool DEFAULT false), edited_at (*timestamptz), likes (int4 DEFAULT 0), created_at (timestamptz), updated_at (timestamptz).
Indexes: author_id, created_at DESC, (reply_to_id) WHERE reply_to_id IS NOT NULL.

### communities (11 columns)
id (PK bigserial), name (varchar), slug (varchar UNIQUE), description (text), rules (text), owner_id (FK users), is_invite_only (bool DEFAULT false), slow_boot_days (int4 DEFAULT 0), forked_from (*int8 FK self), created_at (timestamptz).
Indexes: slug, owner_id.

### tags (8 columns)
id (PK bigserial), name (varchar), description (text), category (varchar), is_wiki (bool DEFAULT false), post_count (int4 DEFAULT 0), created_by (FK users), created_at (timestamptz).

### post_tags (4 columns)
post_id (FK posts), tag_id (FK tags), tagged_by (FK users), created_at. PK: (post_id, tag_id).

### tag_votes (5 columns)
id (PK bigserial), tag_id (FK tags), user_id (FK users), vote (int2 CHECK -1/1), created_at. UNIQUE(tag_id, user_id). Indexes: tag_id, user_id.

### moderation_actions (11 columns)
id (PK bigserial), moderator_id (FK users), target_user_id (FK users), target_post_id (*int8 FK posts), action_type (int2: 1=warn, 2=suspend, 3=ban), duration (*interval), reason (text), is_public (bool), created_at.

### reports (10 columns)
id (PK bigserial), reporter_id (FK users), reported_id (FK users), target_type (text), target_id (int8), reason (text), status (text: open/reviewed/resolved), reviewed_by (*int8 FK users), created_at, resolved_at (*timestamptz).

### jury_votes (5 columns)
id (PK bigserial), case_id (FK reports), juror_id (FK users), vote (text: dismiss/warning/suspend/ban), created_at.

### interactions (5 columns)
id (PK bigserial), user_id (FK users), post_id (FK posts), interaction_type (text: like/repost/bookmark/reaction), metadata (*jsonb), created_at. UNIQUE(user_id, post_id, interaction_type).

### credit_transactions (10 columns)
id (PK bigserial), from_user_id (*int8 FK), to_user_id (*int8 FK), amount (int8), transaction_type (text: award/spend/gift/penalty/stake/profit), reference_type (*text), reference_id (*text), description (text), signature (*text), created_at.

### Other tables: vouches, trust_connections, community_members, community_notes, community_note_votes, collections, collection_posts, circles, circle_members, hidden_posts, user_notifications, user_achievements, achievements, blocked_users, filters, blocklist, blocklist_subscriptions, custom_pages, custom_emoji, registration_queue, site_config, feed_sources, saved_search, user_settings.

## Workers (scheduler.go)
| Name | Interval | Function |
|------|----------|----------|
| feed_refresh | 30s | Rebuild materialized feed views |
| trending | 60s | Update trending posts + tags (period-based) |
| trust_decay | 6h | Decay old trust connections |
| prune | 1h | Remove posts older than PRUNE_AGE_DAYS with less than PRUNE_MIN_INTERACTIONS. Env overrides: PRUNE_AGE_DAYS, PRUNE_MIN_INTERACTIONS, PRUNE_DISABLE |
| achievements | 10m | Check and award achievements |
| notes_promote | 5m | Promote highly-voted community notes |
| credits | 1h | Distribute activity rewards |
| active_stats | 15m | Update user active stats |
| scheduled_posts | 30s | Publish posts where scheduled_at <= NOW() |
| email_queue | 1m | Process email verification queue (stub) |
| stale_sessions | 1h | Clean expired sessions |

## Frontend Quick Reference

### API Client (web/src/lib/api/client.js)
- `request(method, path, body?) → Promise<data>` — base fetch wrapper. Reads token from localStorage('threadlight_token'). Sets Authorization header. Parses JSON.
- 25 exported functions: `login/register/forgot/reset`, `getPosts/createPost/getPost/deletePost`, `search`, `getProfile/updateProfile`, `joinCommunity/leaveCommunity`, `getTags/createTag/voteTag`, `getFeed`, `getNotifications`, etc.

### Stores (web/src/lib/stores/)
- **auth.js**: `currentUser` (writable store), `token`, `isAuthenticated` (derived). `login(email, password)`, `register(username, email, password)`, `logout()`, `loadUserFromToken()`.
- **feed.js**: `posts`, `feedType`, `page`, `loading`, `error`, `totalPosts`, `loadMore()`. Pagination code exists but never called by UI.
- **notifications.js**: `notifications`, `unreadCount`, `fetchNotifs()`, `markRead(id)`, `startPolling(interval)`, `stopPolling()`.

### Known Issues
1. **Frontend error handling**: Many pages have `catch(e) { /* ignore */ }` — users never see errors or loading states.
2. **No pagination UI**: FeedPage shows initial batch but no "load more" button despite backend pagination support.
3. **SettingsPage monolith**: 672-line single component handling profile, filters, credits, and account. Should be 4 components.
4. **Hardcoded sample data**: LeaderboardPage, ModCenter use fake arrays instead of calling API.
5. **XSS in notifications**: `{@html notif.body}` on line 88 of NotificationsPage.svelte. Replace with `{notif.body}`.
6. **Svelte runes mode mismatch**: `svelte.config.js` has `runes: true` but no component uses `$state/$derived`. All use legacy `export let`, `$:`, `on:click`, `createEventDispatcher`. Only `Counter.svelte` uses `$state()` and isn't wired into the app. Fix: remove `runes: true`.
7. **Test parallelism**: services_test.go TestMain cleanup deletes users with `email LIKE '%@test.com'`. Handler tests also use @test.com. Running `go test ./...` hits FK violations because services test deletes users that handler tests reference. Fix: order cleanup or run sequentially.
8. **Migration runner not wired**: `internal/db/migrations.go` exists, `main.go` has db import but doesn't call `db.RunMigrations()`. Add after postgres connect in main.go.
9. **Onboarding backend works but frontend calls non-existent endpoints**: The old Polaris frontend called routes that don't exist in Threadlight (POST /onboarding/tags, etc.). The Threadlight backend DOES have these routes now — but verify frontend/backend compatibility if porting between versions.
