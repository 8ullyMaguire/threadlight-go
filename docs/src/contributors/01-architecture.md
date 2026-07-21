# How the Code Is Organized

This page is for developers who want to understand how Threadlight is built and how to make changes.

## One Program, Everything Included

Threadlight is a single Go program that does everything. When it starts, it connects to PostgreSQL and Redis, begins serving API requests, and starts background workers. There is no separate queue, no microservice, no second server.

The frontend is a separate application — a SvelteKit app that runs in the browser and talks to the Go server through HTTP requests. The two are independent: you could replace either one without changing the other.

## The Request Path

Every action a user takes follows the same path through the code.

In the browser, a Svelte component calls `client().someMethod()`. This builds an HTTP request, adds the user's authentication token, and sends it to the server.

When the server receives the request, the router matches the URL to a handler. The handler reads the input, calls a service method, and returns JSON. The service method runs SQL queries against the database and returns the result.

The handler sends the JSON response back to the browser. A translation layer in the frontend converts the response into the format the UI components expect. The page updates.

This path is the same for every feature. To understand any feature, follow this path from the browser to the database and back.

## The Backend

All Go code is in the `internal/` directory.

**Router** (`router.go`) — A single file listing every API endpoint. About 110 routes in groups like auth, posts, tags, communities, search, lists, affinity, plugins, and moderation.

**Handlers** — One file per feature area. Handlers are thin: they parse the request, call a service, and return JSON. They never touch the database.

**Services** — The heart of the backend. Each service file contains all the business logic and SQL queries for one feature area. Services do validation, authorization, database queries, and error handling.

**Models** — Data structures that match database rows and JSON responses. No logic, just fields.

**Workers** — Background tasks that run on schedules: computing affinity scores, refreshing algorithmic lists, promoting trust levels, pruning old posts. Each worker is one function in one file.

**Search parser** — A small standalone parser that converts search strings like `"climate AND NOT fossil"` into PostgreSQL text search queries.

## The Frontend

The frontend is a SvelteKit application that was adapted from a Lemmy client. This means there is a translation layer between the Threadlight API and the UI components.

**API client** (`client.svelte.ts`) — The entry point for all API calls. Returns an object with methods like `createPost()`, `getTags()`, and `getLists()`.

**Adapter** (`adapter.ts`) — Translates Threadlight API responses into the format the UI expects. This is where most frontend issues live, because the translation is not always complete.

**Pages** (`routes/`) — One folder per page. Each page has a template (`.svelte` file) and optionally a data loader (`.ts` file).

**Components** (`lib/feature/`) — Reusable UI pieces: Post, Comment, TagVote, CommunityCard, and so on.

**Design system** (`lib/ui/`) — Buttons, text inputs, selects, modals, toasts, and layout components from the `mono-svelte` library.

## The Database

The schema is created by SQL files in `internal/db/migrations/` that run in order when the server starts. There are five migrations that create about 40 tables.

The most important tables are: users, posts, tags, communities, community_members, interactions, trust_connections, moderation_actions, user_lists, user_affinities, and feed_plugins.

## How to Make Changes

To add a new backend endpoint, create a service method, create a handler, and register the route in router.go. The pattern is the same for every feature.

To add a new frontend page, create a route folder, write a template, and call `client().yourMethod()` to get data from the API.

To fix something that is broken, find the API call in the browser's Network tab, call it directly with curl to see what the backend returns, then fix either the service (if the SQL is wrong) or the adapter (if the translation is wrong).

The code is designed to be changed by one person in one afternoon. There is no build pipeline, no code generation, no configuration file that needs updating in three places. Just Go, Svelte, and SQL.
