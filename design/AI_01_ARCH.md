# ARCH PATTERNS

## Imports & Dependencies
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"  // no ORM, raw SQL
    "github.com/redis/go-redis/v9"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "github.com/gorilla/websocket"
)
```
JWT secret from env `JWT_SECRET`, default `"dev-secret-change-in-production"`.

## Service Pattern (e.g. services/post.go)
```go
type PostService struct { pg *pgxpool.Pool }
func NewPostService(pg *pgxpool.Pool) *PostService
func (s *PostService) Create(ctx, authorID int64, req CreatePostRequest) (*Post, error)
func (s *PostService) Get(ctx, id int64) (*Post, error)
```
SQL scanning pattern for nullable columns:
```go
var p Post
err := pg.QueryRow(ctx, "SELECT id, title, body, edited_at, locked FROM posts WHERE id=$1", id).
    Scan(&p.ID, &p.Title, &p.Body, &p.EditedAt, &p.Locked)
// *time.Time, *string, *int64 for nullable DB columns
if err != nil { return nil, model.ErrNotFound }  // sentinel errors
```
Lists use `pg.Query(ctx, sql, args...).Scan(...)` + `defer rows.Close()` + `for rows.Next()`.

## Handler Pattern (e.g. handlers/post.go)
```go
type PostHandler struct { svc *services.PostService }
func NewPostHandler(svc) *PostHandler

func (h *PostHandler) CreatePost(c *gin.Context) {
    userID, _ := c.Get("user_id")  // int64 set by auth middleware
    var req model.CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()}); return
    }
    post, err := h.svc.Create(c.Request.Context(), userID.(int64), req)
    if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(201, post)
}
```

## Middleware
`middleware/auth.go`: extracts Bearer token, validates JWT, sets `c.Set("user_id", userID)`.  
`middleware/cors.go`: sets ACAO to request Origin, handles OPTIONS preflight.

## Router Registration Pattern
```go
func NewRouter(pg *pgxpool.Pool, rdb *redis.Client, jwt string) *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), middleware.CORS("*"))

    authSvc := services.NewAuthService(pg, rdb, jwt)
    authMw := middleware.NewAuthMiddleware(authSvc)
    h := handlers.NewAuthHandler(authSvc)

    apiV1 := r.Group("/api/v1")
    {
        auth := apiV1.Group("/auth"); { auth.POST("/register", h.Register)    auth.POST("/login", h.Login)
            auth.POST("/logout", authMw.RequireAuth(), h.Logout)              auth.GET("/session", authMw.RequireAuth(), h.Session) }
        posts := apiV1.Group("/posts", authMw.RequireAuth()); {
            posts.POST("", h.CreatePost) posts.GET("", h.ListPosts)  posts.GET("/:id", h.GetPost)
            posts.PATCH("/:id", h.EditPost) posts.DELETE("/:id", h.DeletePost)
            posts.POST("/:id/lock", h.LockPost) posts.POST("/:id/sticky", h.StickyPost) ... }
    }
    r.GET("/health", healthHandler.Health)
    return r
}
```
~95 route registrations across ~20 groups.

## Worker Pattern
```go
// internal/worker/scheduler.go — started in main.go:
sched := worker.NewScheduler(pg, rdb)
sched.Start(ctx)
defer sched.Stop()

// Each worker is a goroutine with time.Ticker:
func (s *Scheduler) Start(ctx) {
    go func() { for { select { case <-ticker.C: RunFeedRebuild(ctx, pg); case <-ctx.Done(): return } } }()
}
```
11 workers: FeedRebuild(5m), TrustDecay(1h), Trending(15m), Achievements(30m), ScheduledPosts(60s), NotesConsensus(5m), ActiveStats(30m), Prune(24h). All stubs currently.

## Test Pattern
```go
// handlers/handlers_test.go — TestMain spins up DB connection, runs migration, creates router.
func TestMain(m *testing.M) { dsn=os.Getenv("DATABASE_URL"); testPG=db.NewPostgres(dsn); migrate(testPG); testRouter=api.NewRouter(testPG, testRDB, jwt); os.Exit(m.Run()) }
func request(method, path, body io.Reader, token string) *httptest.ResponseRecorder {
    req := httptest.NewRequest(method, path, body); req.Header.Set("Authorization","Bearer "+token); w:=httptest.NewRecorder(); testRouter.ServeHTTP(w,req); return w }
func jsonBody(v any) io.Reader { b,_:=json.Marshal(v); return bytes.NewReader(b) }
```

## Frontend Patterns
```js
// api/client.js — plain fetch class:
class ApiClient { token=null; setToken(t){this.token=t;localStorage.setItem('polaris_token',t)}
  async request(method,path,body=null){ const headers={'Content-Type':'application/json'}
    if(this.token) headers['Authorization']=`Bearer ${this.token}`
    const res=await fetch(`http://localhost:8080${path}`,{method,headers,body:body?JSON.stringify(body):null})
    if(!res.ok) throw new Error((await res.json()).error||'failed'); return res.json() }
  get(p){return this.request('GET',p)} post(p,b){return this.request('POST',p,b)} ... }

// Navigation — custom events on window:
window.dispatchEvent(new CustomEvent('navigate',{detail:{page:'feed',params:{}}}))
// App.svelte listens and swaps page component via page variable + {#if} blocks.

// Stores — svelte writable:
import { writable } from 'svelte/store'
export const currentUser = writable(null)
export const isAuthenticated = writable(false)
// In templates: {$currentUser?.username}
```

## CSS Variables
```css
--primary:#4f46e5; --bg:#fff; --text:#1f2937; --muted:#9ca3af; --border:#e5e7eb; --success:#10b981; --danger:#ef4444; --radius:8px
```

## Key Gotchas
- All TIMESTAMPTZ columns are nullable → scan into `*time.Time`, not `time.Time`
- User ID from context is `interface{}`, cast: `userID.(int64)`
- CORS middleware allows `*` in dev; restrict for prod
- Workers are stubs — real implementations need SQL in each worker file
- `model/errors.go` has sentinels: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrValidation`
