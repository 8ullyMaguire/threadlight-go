# How to Start Making Changes

This page tells you everything you need to know to start changing the code. You do not need to be a Go expert or a Svelte expert. You just need to know where things are.

## What You Need

- Go 1.26 or later
- Docker (for PostgreSQL and Redis)
- Node.js (for the frontend)
- A terminal and a text editor

## Getting Started

```bash
# Start the database
git clone <repo-url> && cd threadlight
docker compose up -d

# Run the backend
DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go run ./cmd/polaris

# In another terminal, run the frontend
cd web && npm install && npm run dev
```

Open http://localhost:5173 in your browser. You should see the application.

## Running Tests

```bash
# All backend tests
DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go test ./... -count=1

# Frontend type check
cd web && npm run check

# Production build
cd web && npx vite build
```

## The Most Common Task: Adding an API Endpoint

Let us say you want to add `GET /api/v1/hello` that returns `{"message": "world"}`.

**Create the handler** in `internal/api/handlers/hello.go`:

A handler is a struct with a method. The method reads the request, does something, and returns JSON. That is all.

**Register the route** in `internal/api/router.go`:

Find the right group (or create one), add your route, and wire it to your handler. The pattern is `group.METHOD("/path", handler.Method)`.

**Rebuild and test:**

```bash
go build ./...  # check for errors
curl http://localhost:8080/api/v1/hello  # should return {"message":"world"}
```

That is the whole process for a simple endpoint. For an endpoint that needs a database query, you also add a service method. But the handler + route pattern is always the same.

## The Most Common Task: Adding a Frontend Page

Let us say you want to add a page at `/about` that shows some text.

**Create the route folder:**

```
web/src/routes/about/+page.svelte
```

**Write the page:**

A Svelte page is a single file that contains HTML, JavaScript, and CSS. The simplest page just renders some text.

**Open http://localhost:5173/about** to see it.

For a page that needs data from the API, add a `+page.ts` file that calls `client().someMethod()` and returns the data to the page.

## The Most Common Task: Fixing a Broken Page

1. Open the browser's developer tools (F12) and go to the Network tab.
2. Reload the page and look for the API call.
3. If the API call returns an error, the problem is in the backend. Call the same URL with `curl` to confirm.
4. If the API call returns data but the page shows nothing, the problem is in the adapter or the Svelte template. Check the adapter method that handles that API call.
5. If the data looks correct in the Network tab but the page does not render it, the Svelte template is probably referencing a field that does not exist.

## Common Pitfalls

**Backend returns 500 with "can't scan into dest[5]".** This means a SQL query is trying to read a NULL value into a Go type that does not accept NULL. The fix is to change the Go field to a pointer type (`*string` instead of `string`) or use `COALESCE` in the SQL.

**Frontend says "X is not a function".** The client is probably not a ThreadlightClient. Check that `DEFAULT_CLIENT_TYPE` is set to `threadlight` in `base.ts`.

**Svelte build fails with "Attributes need to be unique".** An attribute value contains a space but is not wrapped in quotes. Find the line and add quotes around the value.

**The page loads but shows blank data.** The adapter method is probably a stub that returns `[]` or `{}`. Check `adapter.ts` for the method and replace the stub with a real `#api()` call.

## Where to Go for Help

The codebase is designed to be readable. If you are unsure how something works:

- Search for a similar feature and read its code
- Read `router.go` to see all the endpoints
- Read `adapter.ts` to see how the frontend talks to the backend
- Read the migration files to see the database schema

Every feature follows the same patterns. Once you understand one, you understand them all.
