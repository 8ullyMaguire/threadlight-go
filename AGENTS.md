# Threadlight — A Guide to the Code

This document is for anyone who wants to understand how the code is organized and how to make changes to it. You do not need to be an expert. You just need to know the shape of the project.

---

## The Shape of Things

Threadlight is one program that does everything. When you run it, three things start:

- A **Go server** that handles API requests, talks to the database, and runs background tasks
- A **SvelteKit frontend** that runs in the browser and talks to the Go server
- A **PostgreSQL database** and a **Redis cache** that store data and sessions

That is the whole architecture. There are no hidden pieces.

The Go server listens on port 8080. The frontend development server listens on port 5173 (and proxies API calls to 8080). The database listens on port 5433. The Redis cache listens on port 6379. Everything runs on one machine.

---

## How a Request Travels

When you click something in the browser, this happens:

1. The **frontend page** calls `client().someMethod()`. The `client()` function returns an object that knows how to talk to the API.

2. The **adapter** (`adapter.ts`) builds the URL, adds your authentication token, and sends the request. It also translates the response from Threadlight's format into the shape the UI components expect.

3. The **router** (`router.go`) matches the URL to a handler function.

4. The **handler** reads the request data, calls a service method, and returns JSON.

5. The **service** runs SQL queries against the database, does any logic needed, and returns the result.

6. The JSON response travels back through the same path to the browser.

This is the same path for every feature. Once you understand it, you can trace any feature from click to database.

---

## The Backend (Go)

All the Go code lives in `internal/`. It is divided into five groups:

### api/ — The Doorway

`router.go` is a single file that lists every API endpoint. If you want to know what endpoints exist, read this file. There are about 110 of them.

`handlers/` contains one file per feature area. Each handler is a small function that parses the request, calls a service, and returns JSON. Handlers never touch the database directly.

### services/ — Where the Work Happens

This is the heart of the backend. Every database query lives here. Each feature area has its own service file:

- `auth_service.go` — registration, login, password reset
- `post_service.go` — creating, reading, updating, deleting posts
- `tag_service.go` — tags, tagging posts, voting on tags
- `community_service.go` — communities, joining, leaving, forking
- `moderation_service.go` — reports, moderation actions, juries
- `search_service.go` — full-text search with boolean operators and autocomplete
- `userlist_service.go` — user lists, members, subscriptions, collaborators
- `user_affinity_service.go` — affinity scores between users
- `feed_plugin_service.go` — WASM plugin marketplace

Every service method takes a `ctx` (context) as its first argument and returns either a result or an error.

### model/ — The Data Shapes

Every database row has a matching Go struct here. These structs are used by services to read data and by handlers to send JSON responses. They have no logic — just fields with tags that say what the JSON name should be.

### search/ — The Query Parser

A small recursive-descent parser that turns search strings like `(climate OR environment) AND NOT "fossil fuels" author:alice` into PostgreSQL text search queries. It is standalone and has its own tests.

### worker/ — The Background Tasks

Some work cannot happen during a request. Computing affinity scores between all users takes time. Checking trust level promotions does too. These tasks run on schedules in the background:

- affinity computation (every hour)
- algorithmic list refresh (every 30 minutes)
- trust level progression (every hour)
- trust decay (every hour)
- feed rebuild (every 5 minutes)
- post pruning (every hour)
- achievement checking (every 30 minutes)
- trending updates (every minute)
- active stats (every 5 minutes)

Each worker is one function in its own file. They are registered in `scheduler.go`.

---

## The Frontend (SvelteKit)

The frontend is a SvelteKit application. It was originally a client for Lemmy (another social platform), and it has been adapted to work with Threadlight. This adaptation means there is a translation layer between the API and the UI.

### api/ — The Translator

Three files work together:

- `client.svelte.ts` — the entry point. Calling `client()` returns an object with methods like `createPost()`, `getTags()`, etc. It now defaults to the Threadlight client.

- `threadlight/adapter.ts` — the Threadlight adapter. Each method here calls the Threadlight API and translates the response into the shape the UI expects. This is where most frontend bugs live, because the translation is not always complete.

- `threadlight/types.ts` — TypeScript interfaces that describe the data shapes. These need to match what the backend actually returns.

### routes/ — The Pages

Each page of the application has its own folder under `routes/`. A route folder typically has:

- `+page.svelte` — the HTML template and JavaScript logic
- `+page.ts` — a load function that fetches data before the page renders

Pages call `client().methodName()` to get data from the API.

### lib/feature/ — Reusable Pieces

Components that appear in multiple places live here. `Post.svelte` renders a single post. `Comment.svelte` renders a comment. `TagVote.svelte` shows upvote/downvote buttons for a tag.

### lib/ui/ — Design System

Buttons, text inputs, modals, toasts, select dropdowns, and other building blocks. These come from a library called `mono-svelte`.

---

## The Database

The schema is created by SQL files in `internal/db/migrations/`. They run in order when the server starts:

1. `001_initial.sql` — the core tables (users, posts, communities, tags)
2. `002_tag_votes.sql` — tag voting
3. `003_features.sql` — user lists, affinity scores, feed plugins
4. `004_mod_decision_reviews.sql` — public moderation decision reviews
5. `005_search.sql` — full-text search indexes and triggers

There are about 40 tables. The most important ones are:

- `users` — accounts, trust levels, credits
- `posts` — content, moods, content types
- `tags` and `post_tags` — topics that connect posts
- `communities` and `community_members` — shared spaces
- `moderation_actions` and `mod_decision_reviews` — enforcement and review
- `user_lists`, `list_members`, `list_subscriptions` — shared account lists
- `user_affinities` — connection scores between users
- `feed_plugins`, `feed_plugin_installs` — the plugin marketplace

---

## How to Make a Change

**To add a new backend feature:**

1. Add a migration file if you need new tables or columns
2. Add model structs if you need new data shapes
3. Add a service file with the business logic and SQL queries
4. Add a handler file with the HTTP endpoint
5. Register the route in `router.go`
6. Run `go build ./...` to check for errors

**To add a new frontend page:**

1. Add TypeScript interfaces to `types.ts` if you need new data shapes
2. Add API methods to `adapter.ts` that call the backend and translate responses
3. Create a route folder under `routes/` with a `+page.svelte` file
4. Call `client().yourMethod()` from the page to get data

**To fix something that is broken:**

1. Find which API endpoint the page calls (check the Network tab in browser dev tools)
2. Call that endpoint directly with `curl` to see what it returns
3. Fix the service if the SQL is wrong, or fix the adapter if the translation is wrong

---

## The Collective Power Idea

Every feature in Threadlight is a tool for collective action. User lists let communities protect themselves together. Moderation reviews let everyone hold moderators accountable. Affinity scores help people find each other through genuine connection, not algorithmic manipulation. Plugins let users control their own feed logic instead of being fed whatever makes a company money.

The code is open. The data is yours. The power belongs to the people who use it.
