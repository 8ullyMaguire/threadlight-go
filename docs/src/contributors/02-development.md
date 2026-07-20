# Local Development

This page explains how to set up a development environment for working on Threadlight.

## Prerequisites

- Go 1.23+
- Node.js 20+ and pnpm
- PostgreSQL 16+
- Redis 7+

## Quick Start

### 1. Start the Database Services

If you have Docker, this is the easiest way:

```bash
docker run -d --name polaris-pg -e POSTGRES_USER=polaris -e POSTGRES_PASSWORD=polaris -e POSTGRES_DB=polaris -p 5432:5432 postgres:16
docker run -d --name polaris-redis -p 6379:6379 redis:7
```

### 2. Clone and Build

```bash
git clone https://github.com/anomalyco/polaris.git
cd polaris

# Build the frontend
cd web
pnpm install
pnpm dev &  # starts Vite dev server on :5173

# In another terminal, build and run the backend
cd ..
go run ./cmd/polaris
```

### 3. Configure

The defaults in the code work with the Docker setup above:

- Database: `postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable`
- Redis: `localhost:6379`
- Server: `:8080`

Set any environment variables to override:

```bash
export DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable"
export REDIS_ADDR="localhost:6379"
```

### 4. Open the App

- Frontend (Vite dev server): http://localhost:5173
- Backend API: http://localhost:8080

The Vite dev server automatically proxies API requests to the Go backend.

## Development Workflow

### Backend Development

After changing Go code, just restart the server:

```bash
go run ./cmd/polaris
```

The server will automatically run database migrations on startup.

### Frontend Development

The Vite dev server supports hot module replacement (HMR), so changes to Svelte files appear instantly in the browser.

### Running Tests

```bash
# Run all tests
DATABASE_URL="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go test ./...

# Run specific package tests
go test ./internal/api/handlers/...

# Run with verbose output
go test -v ./internal/services/...
```

### Code Checks

```bash
# Build check
go build ./...

# Vet check
go vet ./...
```

## Code Style

- Follow standard Go conventions (gofmt)
- Use meaningful variable names
- Write tests for new functionality
- Keep functions small and focused
- Prefer readability over cleverness

## Database Migrations

Migrations are SQL files in `internal/db/migrations/`. They're run in filename order on server startup.

To create a new migration:

1. Create a new file like `002_descriptive_name.sql`
2. Write your `ALTER TABLE` or `CREATE TABLE` statements
3. Restart the server to apply it

Migrations should be idempotent when possible (use `IF NOT EXISTS`, `IF EXISTS`, etc.).
