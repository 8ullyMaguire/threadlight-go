# Architecture Overview

This page explains how Threadlight is built, for developers who want to understand or contribute to the code.

## High-Level Structure

Threadlight has three main parts:

```
polaris/
├── cmd/polaris/        # The server binary (entry point)
├── internal/           # Go code (the backend)
│   ├── api/            # HTTP handlers and API routes
│   ├── db/             # Database connection and migrations
│   ├── services/       # Business logic
│   └── worker/         # Background tasks
└── web/                # Svelte frontend (Svelte + Vite)
```

## Backend (Go)

The backend is written in Go and uses:

- **Gin** for HTTP routing
- **pgx** for PostgreSQL access
- **go-redis** for Redis access
- **zerolog** for logging

### Request Flow

1. A request comes in to `cmd/polaris/main.go`
2. It's routed by `internal/api/router.go` to the correct handler
3. Handlers in `internal/api/handlers/` process the request
4. They call services in `internal/services/` for business logic
5. Services interact with the database directly (using pgx)
6. The response goes back through the handler and out

### Key Files

| File | Purpose |
|------|---------|
| `cmd/polaris/main.go` | Server startup, environment variables, seeding |
| `internal/api/router.go` | Route definitions, middleware setup |
| `internal/api/handlers/*.go` | HTTP handlers (one per resource) |
| `internal/services/*.go` | Business logic layer |
| `internal/db/migrations/*.sql` | Database schema migrations |
| `internal/worker/*.go` | Background workers (content pruning, etc.) |

## Frontend (Svelte)

The frontend is a single-page application (SPA) built with Svelte and Vite.

- **Routing**: Handled client-side by `svelte-spa-router`
- **State**: Managed with Svelte stores
- **API calls**: Made directly to the Go backend

### Key Frontend Files

| File | Purpose |
|------|---------|
| `web/src/App.svelte` | Main app with router setup |
| `web/src/lib/` | Reusable components |
| `web/src/routes/` | Page components |
| `web/src/lib/api.js` | API client for backend communication |

## Database

Threadlight uses PostgreSQL with the following main tables:

- **users** &mdash; user accounts, profiles, trust scores
- **posts** &mdash; all posts and their content
- **communities** &mdash; community groups
- **community_notes** &mdash; fact-checks and context on posts
- **trust_connections** &mdash; who trusts whom
- **tags** &mdash; hashtags and topic labels

Migrations are applied automatically on startup from `internal/db/migrations/`.

## Background Workers

Threadlight runs background workers that handle periodic tasks:

- **Content Pruning**: Removes old, low-quality posts hourly
- More workers can be added in `internal/worker/scheduler.go`

## Data Flow Example: Creating a Post

1. User types a post and clicks "Post" in the browser
2. The frontend sends a POST request to `/api/v1/posts`
3. The router directs it to the `CreatePost` handler
4. The handler validates the input and calls `CreatePost` in the services layer
5. The service inserts the post into the database
6. The handler returns the created post as JSON
7. The frontend updates the feed to show the new post
