# Threadlight

A social network where the people who use it decide how it works.

Most platforms are built like a castle. A few people at the top own everything. They decide what you see, who gets heard, and what is allowed. They make money from your attention, and they keep the power for themselves.

Threadlight is built like a town square. Everyone has a voice. Everyone helps take care of the space. The rules are transparent. The tools are shared. The platform works for the people on it, not for shareholders somewhere else.

There are no ads. There is no algorithm engineered to keep you angry and scrolling. Your feed is yours. Moderation is done by juries of regular users, not by a single moderator with unchecked power. Tags are owned by everyone, not by a corporation.

---

## What You Can Do

**Post and share.** Write text posts, add moods and content warnings, schedule them for later. Tag your posts with topics so others can find them.

**Find your people.** Tags connect posts by topic. Communities are spaces where people with shared interests gather. Affinity scores help you discover people who share your values — not through a secret algorithm, but through the things you actually do together.

**Trust matters.** You can vouch for people you know and trust. The platform uses these trust connections to help good-faith contributors earn more voice over time. Trust is built by the community, not granted by an admin.

**Moderate together.** When something breaks the rules, anyone can report it. A randomly selected jury of users reviews the report and decides what to do. Every moderation decision can be reviewed by anyone — you can vote "Fair" or "Unfair" on any action, and the most controversial decisions are surfaced publicly.

**Control your feed.** User lists let you follow or block groups of accounts at once. You can share lists with others. Algorithmic lists build themselves based on rules you choose. Plugins let you write — or install — your own feed-ranking logic, running safely in a sandbox.

---

## Quick Start

```bash
# Start the database and cache
docker compose up -d

# Run the backend
DATABASE_URL="postgres://polaris:polaris@localhost:5433/polaris?sslmode=disable" go run ./cmd/polaris

# Start the frontend (in another terminal)
cd web && npm run dev
```

Open **http://localhost:5173**.

---

## One Binary, Everything Included

The whole thing is a single program. One Go binary serves the API. One SvelteKit app runs in the browser. There is no microservice maze, no message queue, no GraphQL layer. Just Go talking to PostgreSQL and Redis.

The database has about 40 tables. The cache (Redis) is used for sessions, the affinity badge counters, and temporary data that expires on its own.

---

## What's Inside

The backend code lives in `internal/`. It is split into four layers:

**Handlers** receive HTTP requests. They check the data, call a service, and return JSON. They are thin — about 20 lines each.

**Services** contain all the business logic and all the SQL queries. This is where the actual work happens. Every service is a Go struct with a connection to the database.

**Models** are the data shapes — Go structs that match database rows and JSON responses. They have no logic, just fields.

**Workers** are background tasks that run on a schedule — computing affinity scores, refreshing algorithmic lists, promoting trust levels, pruning old posts.

The frontend lives in `web/`. It is a SvelteKit app that talks to the backend through a single translation layer. The UI components expect a certain data shape (from the old Lemmy client they were forked from), and the adapter file translates the Threadlight API responses into that shape.

---

## The Big Idea

Most social platforms are designed to extract value from you. Threadlight is designed to give you control. The code is open. The moderation is transparent. The features are tools for collective action — shared lists, collaborative curation, public decision reviews, user-owned feed logic.

It is not finished. It may never be finished, because a platform built for its community evolves with that community. But everything here is a step toward a different kind of social network — one where the people, not the platform, hold the power.
