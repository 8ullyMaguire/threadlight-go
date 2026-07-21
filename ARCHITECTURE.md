# Polaris Architecture & Codebase Guide

Read this in order. Each section builds on the previous. By the end you'll know how everything
fits together and how to make changes.

---

## 1. What Is Polaris?

Polaris is a social network. Like Twitter or Reddit, but with no algorithm and no ads, where
moderation is done by juries of regular users instead of single moderators. Users can create
posts, build trust networks, form communities, add fact-check notes to posts, customise their
feeds, and vote on moderation decisions.

The whole thing is a **monolith**: one Go binary serves the REST API, and one Svelte single-page
app (SPA) runs in the browser. There's no microservice split, no GraphQL, no message queue — just
Go + PostgreSQL + Redis.

---

## 2. Directory Layout

```
polaris/
├── cmd/polaris/main.go        # Entry point: wires everything together
|── internal/                  # Go backend (not importable by external modules)
│   ├── api/
│   │   ├── router.go          # Gin routes + middleware setup
│   │   └── handlers/          # HTTP handlers (thin layer, calls services)
│   ├── db/
│   │   ├── db.go              # PostgreSQL + Redis connection, migration runner
│   │   └── migrations/        # SQL files run in order on startup (4 migrations)
│   ├── model/                 # Go structs + error types (shared everywhere)
│   ├── search/
│   │   └── parser.go          # Boolean query parser (recursive-descent AST)
│   ├── services/              # Business logic (all database queries live here)
│   └── worker/                # Background workers (scheduler + individual workers)
├── web/                       # Svelte frontend (separate Vite build)
│   ├── src/
│   │   ├── App.svelte         # Root component + SPA router
│   │   ├── app.css            # Global styles, CSS variables
│   │   ├── main.js            # Mounts App to #app
│   │   ├── lib/
│   │   │   ├── api/client.js  # Fetch wrapper, all API call functions
│   │   │   ├── stores/        # Svelte stores (auth, feed, notifications)
│   │   │   └── components/    # Shared UI components (Navbar, Footer, Sidebar)
│   │   └── routes/            # One .svelte per page (routed by svelte-spa-router)
│   └── dist/                  # Built frontend (gitignored)
└── docs/                      # mdbook documentation (served at /docs/)
```

**File count**: ~105 Go + Svelte + config files.

---

## 3. Request Flow (How a Click Reaches the Database)

```
Browser                  Go Server                   PostgreSQL
   │                        │                           │
   ├─ POST /api/v1/posts ──►│                           │
   │                        ├─ router.go matches route  │
   │                        ├─ middleware checks JWT    │
   │                        ├─ handler parses JSON      │
   │                        ├─ service validates+inserts│
   │                        │        └─ pg.Exec(...) ───►│
   │                        │        ◄── row ───────────┤
   │                        ├─ handler returns JSON     │
   │◄── 201 + post object ──┤                           │
```

Every request follows the same path:

1. **Gin router** (`internal/api/router.go`) matches the URL path + HTTP method
2. **Middleware** checks auth, logs the request, adds user ID to context
3. **Handler** (`internal/api/handlers/*.go`) parses the request, calls the service
4. **Service** (`internal/services/*.go`) does business logic + database queries
5. **Handler** sends JSON response back to the browser

The rule: **handlers never touch the database directly**, and **services never write HTTP
responses**. This keeps them independently testable.

---

## 4. The Router (How Routes Are Defined)

All routes live in a single function in `internal/api/router.go`:

```go
func NewRouter(pg *pgxpool.Pool, rdb *redis.Client, jwtSecret string) *gin.Engine
```

The structure is:

```go
r := gin.New()
r.Use(CORS middleware)

r.GET("/api/v1/health", healthHandler.Health)
r.POST("/api/v1/auth/register", authHandler.Register)
r.POST("/api/v1/auth/login", authHandler.Login)

api := r.Group("/api/v1")
api.Use(AuthMiddleware)  // everything below requires a valid JWT
{
    api.GET("/posts", postHandler.GetFeed)
    api.POST("/posts", postHandler.Create)
    api.GET("/posts/:id", postHandler.GetPost)
    api.PUT("/posts/:id", postHandler.Update)
    api.DELETE("/posts/:id", postHandler.Delete)
    // ... users, communities, moderation, search, etc.
}
```

To add a new endpoint:
- Find the right group in `router.go`
- Add your route with the correct method and path
- Wire it to a handler method

---

## 5. Handlers (The Thin Layer)

Handlers are structs with method receivers. Each handler holds references to the services it needs:

```go
type PostHandler struct {
    postSvc *services.PostService
    userSvc *services.UserService
}
```

A handler method does three things:

1. **Read input** — `c.Param("id")`, `c.ShouldBindJSON(&req)`, `c.Get("user_id")`
2. **Call service** — `h.postSvc.SomeMethod(ctx, args...)`
3. **Return response** — `c.JSON(status, body)` or `c.JSON(status, gin.H{"error": ...})`

Error handling pattern:

```go
if err == model.ErrNotFound {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
}
if err == model.ErrForbidden {
    c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this post"})
    return
}
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
c.JSON(http.StatusOK, result)
```

The `user_id` is set in the context by the auth middleware. Get it with:

```go
uid, _ := c.Get("user_id")  // returns int64
```

---

## 6. Services (Where the Logic Lives)

Services are also struct-with-methods that hold a database connection pool:

```go
type PostService struct {
    pg *pgxpool.Pool
}
```

Every service method takes `context.Context` as the first parameter. This lets us pass request
cancellation and timeouts down to the database driver.

**Database queries use pgx directly** (no ORM). Raw SQL is written inline:

```go
func (s *PostService) GetPost(ctx context.Context, postID int64) (*model.Post, error) {
    post := &model.Post{}
    err := s.pg.QueryRow(ctx,
        `SELECT id, author_id, title, body, content_type, mood,
                is_educational, is_entertaining, is_nsfw, content_warning,
                interaction_count, cumulative_interactions, status,
                is_deleted, created_at, updated_at
         FROM posts WHERE id = $1 AND NOT is_deleted`,
        postID,
    ).Scan(&post.ID, &post.AuthorID, /* ... */)
    if err == pgx.ErrNoRows {
        return nil, model.ErrNotFound
    }
    return post, err
}
```

**Service pattern**: validate → query → return (nil, ErrXxx) or return (result, nil).

All errors the service returns should be from `internal/model/errors.go`:

```go
var ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "resource not found"}
var ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "forbidden"}
var ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized"}
var ErrConflict     = &AppError{Code: "CONFLICT", Message: "already exists"}
var ErrValidation   = &AppError{Code: "VALIDATION", Message: "validation failed"}
```

**IMPORTANT**: If a service method mutates a resource (edit/delete), it MUST check that the
requesting user owns that resource. Pattern:

```go
func (s *PostService) UpdatePost(ctx, postID, userID int64, req) error {
    var authorID int64
    err := s.pg.QueryRow(ctx, `SELECT author_id FROM posts WHERE id = $1`, postID).Scan(&authorID)
    if err != nil { return model.ErrNotFound }
    if authorID != userID { return model.ErrForbidden }
    // proceed with update...
}
```

---

## 7. Database Schema (Key Tables)

The full schema is in `internal/db/migrations/001_initial.sql` (~586 lines). Here are the most
important tables you need to understand:

### users
| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL PK | |
| username | VARCHAR(255) UNIQUE | |
| email | VARCHAR(255) UNIQUE | |
| password_hash | VARCHAR(255) | bcrypt hash |
| display_name | VARCHAR(255) | |
| bio | TEXT | |
| trust_score | FLOAT | 0.0 to 10.0 |
| trust_level | INT | derived from score |
| reputation | BIGINT | total trust-weighted interactions |
| onboarding_stage | INT | 1-4, gates features |
| is_admin | BOOLEAN | |
| invited_by | BIGINT FK(users) | |
| invite_code | VARCHAR(255) | |
| credits | BIGINT | economy |
| last_active_at | TIMESTAMPTZ | |
| is_local | BOOLEAN | |
| location_hash | VARCHAR(64) | for proximity |
| created_at | TIMESTAMPTZ | |

### posts
| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL PK | |
| author_id | BIGINT FK(users) | |
| title | TEXT | |
| body | TEXT | |
| content_type | INT | 0=text, 1=link, 2=image, 3=video |
| mood | INT | 0=none, 1=curious, 2=excited, 3=thoughtful, etc. |
| is_educational | BOOLEAN | |
| is_entertaining | BOOLEAN | |
| is_nsfw | BOOLEAN | |
| content_warning | VARCHAR(255) | |
| interaction_count | INT | |
| cumulative_interactions | BIGINT | |
| status | INT | |
| scheduled_at | TIMESTAMPTZ | for scheduled posting |
| archived_at | TIMESTAMPTZ | soft delete |
| is_deleted | BOOLEAN | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### communities
| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL PK | |
| name | VARCHAR(255) | |
| slug | VARCHAR(255) UNIQUE | |
| description | TEXT | |
| rules | TEXT | |
| owner_id | BIGINT FK(users) | |
| is_invite_only | BOOLEAN | |
| slow_boot_days | INT | gradual growth period |
| forked_from | BIGINT FK(communities) | |
| created_at | TIMESTAMPTZ | |

### trust_connections
| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL PK | |
| truster_id | BIGINT FK(users) | the person doing the trusting |
| trustee_id | BIGINT FK(users) | the person being trusted |
| weight | FLOAT | |
| signature | TEXT | cryptographic signature |
| created_at | TIMESTAMPTZ | |
| expires_at | TIMESTAMPTZ | |

### community_notes + community_note_votes
Fact-check notes on posts. Users vote notes up/down; highly-voted notes get "verified" status.

### reports + jury_votes + moderation_actions
Report → Jury Vote → Moderation Action. Users get randomly selected as jurors.

### site_config
Single-row table with instance settings (name, description, contact email, etc.).

### M:join tables
- `community_members` — user ↔ community membership
- `post_tags` — post ↔ tag many-to-many
- `feed_sources` — sources in a custom feed
- `user_achievements` — earned achievements

### Migrations
Migrations are plain SQL files in `internal/db/migrations/`. They run in filename order
on every server startup. Current migrations:

1. `001_initial.sql` — core schema (users, posts, communities, tags, etc.)
2. `002_tag_votes.sql` — tag voting
3. `003_features.sql` — user lists, user_affinities, feed_plugins, collaborative lists
4. `004_mod_decision_reviews.sql` — public moderation decision reviews

New columns are added with `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` to make migrations
idempotent. To add a new migration:

1. Create `internal/db/migrations/005_your_feature.sql`
2. Write `ALTER TABLE ... ADD COLUMN ...` or `CREATE TABLE ...` statements
3. Restart the server — it runs automatically

---

## 8. Authentication Flow

```
Register ──► POST /api/v1/auth/register
               ├─ validates email + username uniqueness
               ├─ hashes password with bcrypt
               ├─ inserts into users table
               └─ returns 201

Login    ──► POST /api/v1/auth/login
               ├─ looks up user by email
               ├─ compares bcrypt hash
               ├─ generates JWT (claims: user_id, exp, iat)
               └─ returns { token, user }

Every subsequent request:
  Header: Authorization: Bearer <token>
    │
    ▼
  Auth middleware (internal/api/middleware/):
    ├─ validates JWT signature + expiry
    ├─ sets c.Set("user_id", userId)
    └─ calls next handler
```

JWT tokens last 24 hours by default. The secret is configured via `JWT_SECRET` env var.
If you change `JWT_SECRET`, all existing tokens become invalid.

---

## 9. Frontend Architecture

### Tech Stack
- **Svelte 3** (reactive UI, not SvelteKit)
- **svelte-spa-router** (hash-based routing, no SSR)
- **Vite** (build tool, dev server with HMR)
- **Plain CSS** with CSS variables (no Tailwind, no SCSS)

### Routing
Routes are defined in `web/src/App.svelte` as a map:

```js
import AboutPage from './routes/AboutPage.svelte'
import FeedPage from './routes/FeedPage.svelte'
// ...

const routes = {
  '/':           FeedPage,
  '/login':      LoginPage,
  '/register':   RegisterPage,
  '/about':      AboutPage,
  '/search':     SearchPage,
  '/profile/:id': ProfilePage,
  '/settings':   SettingsPage,
  // ...
}
```

The router component replaces itself with the matching page component. The navbar, sidebar,
and footer are *outside* the router — they render on every page.

### Global State (Stores)
Svelte stores hold shared state:

| Store | File | What it holds |
|-------|------|---------------|
| auth | `stores/auth.js` | currentUser, token, isAuthenticated |
| feed | `stores/feed.js` | posts array, feedType, page, loading |
| notifications | `stores/notifications.js` | notifications array, unreadCount |

Stores fetch data through the API client and update reactively. Components that import a store
automatically re-render when the store changes.

### API Client
All API calls go through a single file: `web/src/lib/api/client.js`.

The client:
1. Reads the JWT token from the auth store
2. Attaches it as `Authorization: Bearer <token>`
3. Makes the fetch request
4. Parses the JSON response
5. Returns the data (or throws on error)

Pattern for calling the API from a component:

```js
import { api } from '../lib/api/client.js'
import { currentUser } from '../lib/stores/auth.js'

async function submit() {
  const post = await api.createPost({ title, body, mood: 2 })
  posts = [post, ...posts]
}
```

### Component Tree
```
App.svelte
├── Navbar.svelte          (always visible, top)
├── Router                 (changes by route)
│   ├── FeedPage.svelte
│   ├── LoginPage.svelte
│   ├── ProfilePage.svelte
│   ├── ... (16 pages)
├── TrendingSidebar.svelte (always visible, right)
└── Footer.svelte          (always visible, bottom)
```

### CSS Conventions
- CSS variables in `app.css` for theming (colors, spacing, border-radius, shadows)
- Dark mode via `@media (prefers-color-scheme: dark)`
- Grid layout: `sidebar | main | right-sidebar`
- Each component imports `app.css` and uses the CSS variables
- No CSS modules, no scoped styles in components — it's all global CSS

---

## 10. Background Workers

Workers are goroutines that run on a schedule. They're started in `cmd/polaris/main.go`:

```go
scheduler := worker.NewScheduler(pg, rdb, logger)
scheduler.Start()
defer scheduler.Stop()
```

The scheduler (`internal/worker/scheduler.go`) is a list of (name, interval, function) tuples:

```go
jobs := []struct{
    Name     string
    Interval time.Duration
    Run      func(context.Context, *pgxpool.Pool)
}{
    {"affinity_computation",  1 * time.Hour,     RunAffinityComputation},
    {"affinity_trust_dists",  1 * time.Hour,     RunAffinityTrustDistances},
    {"algorithmic_lists",     30 * time.Minute,   RunAlgorithmicListRefresh},
    {"feed_refresh",          30 * time.Second,   RebuildFeeds},
    {"trending",              60 * time.Second,   UpdateTrending},
    {"trust_decay",           6 * time.Hour,      DecayTrust},
    {"prune",                 1 * time.Hour,      RunPrune},
    {"achievements",          10 * time.Minute,   CheckAchievements},
    {"notes_promote",         5 * time.Minute,    PromoteNotes},
    {"credits",               1 * time.Hour,      DistributeCredits},
    {"active_stats",          15 * time.Minute,   UpdateActiveStats},
    {"scheduled_posts",       30 * time.Second,   PublishScheduled},
}
```

Each job runs in its own goroutine (controlled by a worker pool). If a job panics, it's caught
and logged — the scheduler keeps running.

To add a new worker:
1. Create `internal/worker/my_worker.go` with a function `func MyWorker(ctx context.Context, pg *pgxpool.Pool)`
2. Add it to the jobs list in `scheduler.go`
3. Done — it starts on next server restart

---

## 11. Configuration (Environment Variables)

| Variable | Required | Default | What it does |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | (none, panics) | PostgreSQL connection string |
| `JWT_SECRET` | Yes (prod) | dev-secret | Secret for JWT signing |
| `LISTEN_ADDR` | No | `:8080` | HTTP listen address |
| `REDIS_ADDR` | No | `localhost:6379` | Redis address |
| `ADMIN_EMAIL` | No | — | Auto-seed admin email |
| `ADMIN_PASSWORD` | No | — | Auto-seed admin password |
| `INSTANCE_NAME` | No | `Polaris` | Name on About page |
| `INSTANCE_DESCRIPTION` | No | — | Description on About page |
| `DOCS_DIR` | No | `./docs/book` | Path to mdbook output |

---

## 12. Models (The Shared Types)

Every database row has a corresponding Go struct in `internal/model/`. These structs are used
by services (to scan query results) and by handlers (to marshal to JSON). They have NO methods
and NO validation — they're pure data.

```go
type Post struct {
    ID                     int64     `json:"id"`
    AuthorID               int64     `json:"author_id"`
    Title                  string    `json:"title"`
    Body                   string    `json:"body"`
    ContentType            int       `json:"content_type"`
    Mood                   int       `json:"mood"`
    IsEducational          bool      `json:"is_educational"`
    IsEntertaining         bool      `json:"is_entertaining"`
    IsNsfw                 bool      `json:"is_nsfw"`
    ContentWarning         string    `json:"content_warning"`
    InteractionCount       int       `json:"interaction_count"`
    CumulativeInteractions int64     `json:"cumulative_interactions"`
    Status                 int       `json:"status"`
    ScheduledAt            *time.Time `json:"scheduled_at,omitempty"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
    ArchivedAt             *time.Time `json:"archived_at,omitempty"`
    IsDeleted              bool      `json:"is_deleted"`
}
```

There are also "request" structs for API input and "response" structs for API output, following
the same pattern (`model.CreatePostRequest`, `model.PostResponse`).

---

## 13. How to Make Common Changes

### Add a new field to an existing resource

**Example**: Add a `featured_image` field to posts.

1. **Database**: Create a new migration file:
   ```sql
   -- internal/db/migrations/002_featured_image.sql
   ALTER TABLE posts ADD COLUMN IF NOT EXISTS featured_image VARCHAR(512);
   ```

2. **Model**: Add the field to `model/post.go`:
   ```go
   FeaturedImage string `json:"featured_image"`
   ```

3. **Service**: Update the SQL queries in `internal/services/post.go`:
   - Add `featured_image` to the INSERT in `CreatePost`
   - Add `featured_image` to the UPDATE in `UpdatePost`
   - Add `featured_image` to the SELECT in `GetPost` / `GetFeedPosts` / etc.

4. **Handler**: If the field comes from user input, add it to the request struct:
   ```go
   type CreatePostRequest struct {
       Title         string `json:"title"`
       Body          string `json:"body"`
       FeaturedImage string `json:"featured_image"`
       // ...
   }
   ```

5. **Frontend**: Add the UI element to the post creation form and the post display.

### Add a new API endpoint

**Example**: Add `GET /api/v1/posts/trending` to get trending posts.

1. **Service**: Add a method to the appropriate service:
   ```go
   // internal/services/post.go
   func (s *PostService) GetTrendingPosts(ctx context.Context, limit int) ([]model.Post, error) {
       rows, _ := s.pg.Query(ctx, `SELECT ... FROM posts ORDER BY ... LIMIT $1`, limit)
       defer rows.Close()
       var posts []model.Post
       for rows.Next() {
           var p model.Post
           rows.Scan(&p.ID, &p.AuthorID, ...)
           posts = append(posts, p)
       }
       return posts, nil
   }
   ```

2. **Handler**: Add a method to the handler:
   ```go
   // internal/api/handlers/post.go
   func (h *PostHandler) GetTrending(c *gin.Context) {
       posts, err := h.postSvc.GetTrendingPosts(c.Request.Context(), 20)
       if err != nil {
           c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
           return
       }
       c.JSON(http.StatusOK, posts)
   }
   ```

3. **Router**: Add the route:
   ```go
   // internal/api/router.go
   api.GET("/posts/trending", postHandler.GetTrending)
   ```

4. **Frontend API client**: Add a function:
   ```js
   // web/src/lib/api/client.js
   getTrendingPosts: () => request('/api/v1/posts/trending'),
   ```

5. **Frontend page/component**: Add the UI in the appropriate Svelte component.

### Add a new database table

**Example**: Add a `bookmarks` table.

1. **Migration**:
   ```sql
   CREATE TABLE IF NOT EXISTS bookmarks (
       id BIGSERIAL PRIMARY KEY,
       user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
       post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       UNIQUE(user_id, post_id)
   );
   ```

2. **Model**: Create `internal/model/bookmark.go`:
   ```go
   package model
   type Bookmark struct {
       ID        int64     `json:"id"`
       UserID    int64     `json:"user_id"`
       PostID    int64     `json:"post_id"`
       CreatedAt time.Time `json:"created_at"`
   }
   ```

3. **Service**: Create `internal/services/bookmark.go` with CRUD methods.

4. **Handler**: Create `internal/api/handlers/bookmark.go` with HTTP handlers.

5. **Router**: Add routes.

6. **Frontend**: Add API functions and UI.

### Modify a background worker

1. Edit the worker function in `internal/worker/`
2. If you need to change the interval, edit `scheduler.go`
3. Restart the server to pick up the change

### Add a new page to the frontend

1. Create `web/src/routes/MyNewPage.svelte`
2. Import it in `web/src/App.svelte` and add a route entry
3. Add a link in `Navbar.svelte` if it should be navigable from the nav

---

## 14. Key Conventions

### Go (Backend)
- **No ORM** — use raw SQL with pgx
- **No global variables** — dependencies are injected through struct fields
- **Error handling**: always check errors; use model.ErrXxx for known failure modes
- **Ownership checks**: every mutation (edit/delete) must verify the user owns the resource
- **Context passing**: every service method takes `ctx` as first arg
- **File naming**: lower_snake_case.go for all Go files
- **SQL naming**: lowercase with underscores (snake_case) for columns and tables

### Svelte (Frontend)
- **One component per file**, named PascalCase.svelte
- **Stores** for global state, props for local state
- **No TypeScript** — the frontend is plain JavaScript
- **No CSS framework** — use CSS variables from app.css
- **Hash-based routing** via svelte-spa-router
- **API calls** go through `api/client.js`, never `fetch()` directly in components

### Database
- **Migrations are idempotent** — use `IF NOT EXISTS` and `IF EXISTS`
- **Soft deletes** using `is_deleted` or `archived_at` columns (no hard DELETE)
- **Foreign keys** use `ON DELETE CASCADE` where it makes sense
- **Timestamps** are `TIMESTAMPTZ` (timezone-aware)
- **Indexes** are created in the migration alongside the tables

### Testing
- **Go tests** use `testing.T` and run against a real PostgreSQL database
- **Test cleanup** happens in `TestMain` via `DELETE FROM ... WHERE email LIKE '%@test.com'`
- **Frontend tests** use Vitest + jsdom
- **E2E tests** use Playwright

---

## 15. Quick Reference: Important Files

### Backend (by what they do)

| What you want | File |
|--------------|------|
| Server startup + env vars | `cmd/polaris/main.go` |
| Route definitions | `internal/api/router.go` |
| Auth middleware | `internal/api/middleware/` |
| Database connection | `internal/db/db.go` |
| SQL schema | `internal/db/migrations/001_initial.sql` |
| All Go structs | `internal/model/*.go` |
| All business logic | `internal/services/*.go` |
| All HTTP handlers | `internal/api/handlers/*.go` |
| Background workers | `internal/worker/*.go` |
| Error types | `internal/model/errors.go` |

### Frontend (by what they do)

| What you want | File |
|--------------|------|
| App shell + routes | `web/src/App.svelte` |
| Global CSS | `web/src/app.css` |
| API client | `web/src/lib/api/client.js` |
| Auth store | `web/src/lib/stores/auth.js` |
| Feed store | `web/src/lib/stores/feed.js` |
| Notifications store | `web/src/lib/stores/notifications.js` |
| Navbar | `web/src/lib/components/Navbar.svelte` |
| Trending sidebar | `web/src/lib/components/TrendingSidebar.svelte` |
| Footer | `web/src/lib/components/Footer.svelte` |
| Feed page | `web/src/routes/FeedPage.svelte` |
| Profile page | `web/src/routes/ProfilePage.svelte` |
| About page | `web/src/routes/AboutPage.svelte` |
| Admin page | `web/src/routes/AdminPage.svelte` |
| Mod center | `web/src/routes/ModCenter.svelte` |
| Settings page | `web/src/routes/SettingsPage.svelte` |
| Communities page | `web/src/routes/CommunitiesPage.svelte` |
| Login page | `web/src/routes/LoginPage.svelte` |
| Register page | `web/src/routes/RegisterPage.svelte` |
| Search page | `web/src/routes/SearchPage.svelte` |
| Notifications page | `web/src/routes/NotificationsPage.svelte` |
| Onboarding wizard | `web/src/routes/OnboardingWizard.svelte` |
| Collections page | `web/src/routes/CollectionsPage.svelte` |
| Leaderboard page | `web/src/routes/LeaderboardPage.svelte` |
| Blocked users page | `web/src/routes/BlockedUsersPage.svelte` |

---

## 16. The Makefile (Useful Commands)

```
make dev-backend    — go run ./cmd/polaris
make dev-frontend   — cd web && pnpm dev
make build          — go build + pnpm build
make test           — go test ./..., pnpm test
make migrate        — not needed (runs on startup)
make docs           — mdbook build docs
```

---

## 17. Quick Start for Development

```bash
# Terminal 1: Database (Docker)
docker compose up -d postgres redis

# Terminal 2: Backend
cd polaris
export DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable"
export JWT_SECRET="dev"
go run ./cmd/polaris
# Starts on :8080

# Terminal 3: Frontend (hot-reload)
cd polaris/web
pnpm install
pnpm dev
# Starts on :5173, proxies /api to :8080
```

Open `http://localhost:5173` in your browser. The Vite dev server proxies `/api/*` requests
to the Go backend on `:8080`.

---

## 18. Ownership Checklist (Security)

Every time you add a mutation endpoint, make sure you check:

- [ ] Does this endpoint need authentication? (use `api.Use(AuthMiddleware)`)
- [ ] Does the user own the resource they're modifying? (check `author_id == uid`)
- [ ] Is there a `SELECT` before the mutation to verify existence + ownership?
- [ ] Does the handler pass `uid` from context to the service?

This applies to: Update, Delete, Archive, and any status-change operations.

---

## 19. File Dependency Graph

```
cmd/polaris/main.go
    ├── internal/api/router.go
    │   ├── internal/api/handlers/*.go
    │   │   └── internal/services/*.go
    │   │       └── internal/model/*.go  (types + errors)
    │   ├── internal/db/db.go
    │   │   ├── internal/db/migrations/*.sql
    │   │   └── github.com/jackc/pgx/v5/pgxpool
    │   └── internal/api/middleware/*.go
    ├── internal/worker/*.go
    │   └── internal/model/*.go
    └── web/dist/  (built Svelte app, served as static files)
```

Every layer depends only on the layers below it. Services know about models but not about
handlers. Handlers know about services. The router wires them together.

---

## 20. Adding a feature end-to-end: The mental model

When you want to add something new, think through these steps in order:

1. **What data do I need to store?** → Add to the migration SQL + model struct
2. **What queries do I need?** → Add methods to the relevant service
3. **What HTTP endpoints do I need?** → Add handlers + route
4. **What does the user see?** → Add frontend API calls + Svelte page/component

Each step depends on the previous one. If you get stuck, trace back up this chain to find
what's missing.

---

That's the whole codebase. Read this document top to bottom, then look at the actual files
in the order: `main.go → router.go → handlers/*.go → services/*.go → model/*.go → web/src/*`.
You'll be making changes in no time.
