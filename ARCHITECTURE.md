# How Threadlight Is Built

This document is a deeper look at how the code fits together. Read it after the README and the Guide to the Code (AGENTS.md). It explains the same things in more detail, from the ground up.

---

## The Foundation

Threadlight is a monolith. That is a technical word that means "one piece." The entire application — every feature, every API endpoint, every background task — is compiled into a single Go program. When you run it, it starts an HTTP server, connects to PostgreSQL and Redis, and begins serving requests.

There are good reasons for this:

- **Simple to deploy.** You build one binary, copy it to a server, and run it. There is nothing else to install, configure, or wire together.
- **Simple to develop.** You do not need to run five different services to test a change. You run one command.
- **Simple to debug.** A request enters through one door and leaves through the same door. There is no chain of microservices to trace through.

The tradeoff is that the codebase is large — about 25 service files, 26 handler files, 30 model files, and 13 worker files. But each file is small and focused on one thing.

---

## The Request Path

When a user does something in the browser, the request follows a fixed path through the code. Understanding this path is the key to understanding the whole system.

### Step 1: The Frontend Makes a Call

A Svelte component calls `client().createPost(data)`. The `client()` function returns an object that knows how to talk to the backend. It builds the URL, adds the user's authentication token, and sends an HTTP request.

### Step 2: The Adapter Translates

The Threadlight adapter (`adapter.ts`) receives the response from the backend. The backend returns data in its own format — flat JSON with fields like `author_id` and `community_id`. But the frontend UI components expect a different format — nested objects with `post.creator` and `post.community`. The adapter translates between the two.

This translation is the most common source of frontend bugs. If a page shows blank data or missing fields, the adapter is usually the problem.

### Step 3: The Router Matches the URL

The Go router (`router.go`) receives the HTTP request. It looks at the URL path and HTTP method, and it matches them to a handler function. If you send `POST /api/v1/posts`, it calls `postHandler.Create`.

### Step 4: The Handler Does Three Things

Every handler does the same three things in the same order:

1. **Read the input** — parse the JSON body, read URL parameters, get the user ID from the authentication token
2. **Call a service** — pass the input to a service method
3. **Return JSON** — send the service's response back to the browser

Handlers never touch the database directly. They are the middle layer that connects HTTP to business logic.

### Step 5: The Service Runs the Logic

Services contain all the SQL queries and business rules. A service method:
- Validates the input (is the user allowed to do this?)
- Runs SQL queries against PostgreSQL
- Returns a result or an error

Services use a library called `pgx` to talk to PostgreSQL. They write raw SQL — there is no ORM that hides the queries. This means you can always see exactly what the database is being asked to do.

### Step 6: The Response Travels Back

The service returns data to the handler. The handler sends it as JSON to the browser. The adapter translates it for the UI. The page renders.

---

## The Backend Layers

### api/router.go — The Map

This is a single file, about 390 lines, that lists every API endpoint in the system. It is organized into groups:

- `/auth` — registration, login, password reset, session
- `/posts` — creating, reading, updating, deleting posts
- `/tags` — tags, tag voting, post-tag association
- `/communities` — communities, membership, curators, forking
- `/moderation` — actions, reports, juries
- `/modlog` — public moderation decision reviews
- `/lists` — user lists, members, subscriptions, collaborators
- `/affinity` — affinity scores and similar users
- `/feed-plugins` — plugin marketplace
- `/search` — full-text search across posts, users, communities
- `/users` — profiles, settings, notifications
- `/blocks` — blocking users
- `/feeds` — custom feeds
- `/circles` — friend circles
- `/collections` — post collections
- `/credits` — economy
- `/admin` — configuration and invites
- `/health` — health check

Each group is about 5-15 lines. You can see the entire surface area of the API in one screen.

### api/handlers/ — 26 Files

Each handler file is named after its feature area. A handler file typically contains 2-10 handler functions. Each function is about 15-25 lines.

The pattern is always the same:

```
Read input → Call service → Return JSON
```

Error handling is also the same everywhere:

```
if err is "not found" → return 404
if err is "forbidden" → return 403
if err is "validation error" → return 400
if err is anything else → return 500 with the error message
```

### services/ — 25 Files

This is where the real work happens. Each service file contains one struct with methods. The struct holds a database connection pool. The methods take a context, do work, and return results.

Services handle:

- **Authorization** — checking that the user owns the resource they are trying to change
- **Validation** — checking that inputs are reasonable
- **SQL queries** — reading and writing to PostgreSQL
- **Error wrapping** — converting database errors into model errors

A typical service method looks like this in structure:

1. Query the database to check ownership
2. If the resource does not exist, return "not found"
3. If the user does not own it, return "forbidden"
4. Run the actual insert, update, or delete query
5. Return the result

### model/ — 30 Files

Every Go struct that maps to a database row is here. They have JSON tags that control how they appear in API responses. They have no methods and no logic.

There are also error sentinels — special variables that represent common error conditions:

- `ErrNotFound` — the resource does not exist
- `ErrForbidden` — the user does not have permission
- `ErrUnauthorized` — the user is not logged in
- `ErrConflict` — a duplicate or conflict
- `ErrValidation` — the input did not pass validation

These are used everywhere in services and checked everywhere in handlers.

### search/parser.go — One File

This is a standalone recursive-descent parser that converts search strings into PostgreSQL text search queries. It handles:

- Boolean operators: AND, OR, NOT
- Parentheses for grouping
- Double-quoted phrases
- Field filters: `author:`, `tag:`, `community:`, `title:`, `body:`, `mood:`, `is:`, `before:`, `after:`

It is tested separately with its own test file (44 test cases).

### worker/ — 13 Files

Background tasks run on schedules defined in `scheduler.go`. Each worker is an exported function in its own file. The scheduler starts them in goroutines when the server starts.

Workers are used for tasks that:

- Take too long to run during a request (computing affinity scores for all users)
- Need to run on a schedule (checking trust level promotions)
- Clean up old data (pruning old posts)

---

## The Frontend Layers

### api/client.svelte.ts — The Entry Point

The `client()` function returns an object that has methods for every API endpoint. It reads the user's authentication token from the profile store and attaches it to every request.

By default, it returns a Threadlight client. This was changed from the original default (Lemmy) to match the backend.

### api/threadlight/adapter.ts — The Translator

This is the most important frontend file to understand. It has about 1800 lines. Each method:

1. Calls the Threadlight API using `this.#api(method, path, params)`
2. Translates the response from Threadlight's flat JSON format into the nested format the UI expects
3. Returns the translated data

For example, the backend returns a post with `author_id` and `community_id`. The UI expects `post.creator.id` and `post.community.id`. The adapter creates those nested objects.

Some adapter methods are stubs — they return empty data because the corresponding Threadlight API endpoint does not exist yet, or the translation has not been written.

### api/threadlight/types.ts — The Interfaces

TypeScript interfaces that describe the data shapes. They are used by the adapter methods and by the UI components.

### routes/ — The Pages

Each folder under `routes/` is a page. The `+page.svelte` file is the template. The `+page.ts` file is the data loader that runs before the page renders.

Pages use Svelte 5 runes:

- `$state()` — reactive variable that triggers re-renders when changed
- `$derived()` — computed value that updates automatically
- `$effect()` — code that runs when dependencies change

### lib/feature/ — Components

Reusable UI pieces grouped by feature:

- `post/` — Post, PostMeta, PostVote, PostActions
- `comment/` — Comment, CommentVote, CommentForm
- `community/` — CommunityCard, CommunityHeader, CommunityForm
- `user/` — UserLink, UserItem, UserNote
- `moderation/` — moderation UI components

### lib/ui/ — Design System

Buttons, text inputs, selects, switches, modals, toasts, badges, avatars, and layout components. These are from the `mono-svelte` library and are used throughout the application.

---

## The Database

### How Migrations Work

When the server starts, it reads all `.sql` files from `internal/db/migrations/` in alphabetical order and runs each one against the database. Every statement uses `IF NOT EXISTS` so that re-running a migration is safe.

There are currently 5 migrations that create about 40 tables.

### The Core Tables

**users** — Every account. Stores username, email, password hash, trust level, trust score, reputation, credits, and settings.

**posts** — Every post. Stores title, body, content type, mood, educational/entertaining/NSFW flags, interaction counts, and moderation status.

**tags** and **post_tags** — Tags are topics. Posts can have many tags. Anyone can tag any post.

**communities** and **community_members** — Communities are shared spaces. Members have roles and statuses.

**interactions** — Every reaction to a post (like, insightful, helpful, etc.). Used for computing affinity and feed ranking.

**trust_connections** — One user vouching for another. Forms a trust graph that powers trust scores and affinity.

### The Feature Tables

**user_lists**, **list_members**, **list_subscriptions**, **list_collaborators** — Shared lists of accounts. Users can follow or block everyone on a list at once. Lists can be collaborative (multiple editors) or algorithmic (auto-populated by rules).

**user_affinities** — Computed connection scores between users. Based on shared tags, shared communities, agreement on reactions, and trust distance.

**feed_plugins**, **feed_plugin_installs**, **feed_plugin_reviews** — WASM plugins that users can install to customize their feed logic.

**mod_decision_reviews** — Public votes on moderation actions. Any user can vote "Fair" or "Unfair" on any moderation decision.

---

## Running the Tests

The backend tests need a running PostgreSQL database. The test suite connects to it, creates test data, runs queries, and cleans up after itself.

```bash
# Start the database (PostgreSQL + Redis)
docker compose up -d

# Run all backend tests
DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go test ./... -count=1
```

The frontend type-checking can be run without a database:

```bash
cd web
npm run check
```

---

## A Note About Names

The project has two names. **Threadlight** is the name of the social network — what users see and what the documentation calls it. **Polaris** is the internal code name, used in package names (`github.com/opencode-ai/polaris`) and database connection strings. If you see "Polaris" in the code, it is the same thing as Threadlight.
