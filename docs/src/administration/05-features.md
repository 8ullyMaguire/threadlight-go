# Features (Migration 003)

This page documents the new feature tables introduced in migration 003 — user lists, user affinity, and feed plugins — and how they are configured on your server.

## What Migration 003 Adds

Migration 003 creates three groups of tables that power the feature system. It runs automatically when the server starts, just like all other migrations. You do not need to do anything manually.

### User Lists

Four tables support the user list system:

| Table | Purpose |
|-------|---------|
| `user_lists` | List metadata — name, description, type (follow or block), visibility, and algorithmic rules |
| `list_members` | Accounts in each list, with who added them and when |
| `list_subscriptions` | Users who subscribe to a shared list and whether it is active |
| `list_collaborators` | People invited to help manage a list, with editor or viewer roles |

A list can be a simple curated group you build by hand (manual list) or one that fills itself automatically based on criteria rules (algorithmic list). See the [User Lists](../users/07-user-lists.md) guide for how users work with them.

### User Affinity

The `user_affinities` table stores computed similarity scores between user pairs:

| Column | Purpose |
|--------|---------|
| `user_a_id`, `user_b_id` | The two users in the pair |
| `affinity_score` | Composite score (floating point) |
| `recency_factor` | Decay based on how recently both users were active |
| `breakdown` | JSONB with component scores (tag_overlap, co_community, reaction_agreement, trust_distance) |
| `computed_at` | When the score was last calculated |

Redshift is not used. The affinity system uses a dual approach:

- **Redis counters** (Option A) — lightweight, always-fresh counters for the badge display. O(1) per lookup. See the [Affinity Redis Keys](#redis-keys-for-affinityaccumulator) section below.
- **Batch computation** (Option C) — full pairwise scores computed by the affinity worker and cached in the database. Used for the affinity page.

See the [Affinity](../users/08-affinity.md) user guide for how this looks to users.

### Feed Plugins (Marketplace)

Five tables support the WASM plugin marketplace:

| Table | Purpose |
|-------|---------|
| `feed_plugins` | Plugin metadata and the compiled WASM binary |
| `feed_plugin_installs` | Which users have installed which plugins |
| `feed_plugin_reviews` | User ratings and reviews for plugins |
| `user_follows` | Standalone follow records (used by the list apply action) |

Plugins are WASM modules that can rank or filter posts in a user's feed. See the [Plugin API](../contributors/04-plugin-api.md) guide for developers.

## Applying the Migration

Nothing special is required. Threadlight applies all pending migrations on startup:

```bash
# Restart the server to apply any new migrations
sudo systemctl restart polaris
```

Check the server logs to confirm the migration ran:

```bash
journalctl -u polaris | grep "003_features"
```

You should see something like "applying migration 003_features.sql" in the output.

## Worker Configuration

Two background workers support the feature system. They are defined in `internal/worker/scheduler.go` and start automatically when the server runs.

### Affinity Computation

- **Interval:** Every 1 hour
- **What it does:** Iterates all active user pairs, computes the four affinity components (tag overlap, co-community, reaction agreement, trust distance), and upserts the results into `user_affinities`.
- **Scope:** Users active within the last 30 days (`last_active_at >= 30 days ago`).
- **Log prefix:** `affinity:`

The worker only processes users who have been active recently to avoid wasting resources on dormant accounts.

### Algorithmic List Refresh

- **Interval:** Every 30 minutes
- **What it does:** Finds all algorithmic lists that are due for refresh (their `last_refreshed_at + refresh_interval < NOW()`), evaluates the criteria query against the user table, and replaces the list members atomically.
- **Scope:** Only lists where `is_algorithmic = true`.
- **Log prefix:** `algorithmic_list:`

## Redis Requirements for the AffinityAccumulator

The AffinityAccumulator uses Redis to maintain real-time counters for the badge display (Option A). It also caches computed affinity results (Option C).

### Key Scheme

| Key pattern | Type | Purpose | TTL |
|-------------|------|---------|-----|
| `aff:follow:{userID}` | Hash | Follow counters keyed by target user ID | 72h |
| `aff:community:{userID}` | Hash | Shared community counters | 72h |
| `aff:reaction:{userID}` | Hash | Reaction agreement delta scores | 72h |
| `aff:trust:{userID}` | Hash | Trust path depth as `1/depth` | 24h (recomputed) |
| `aff:cached:{userID}` | String | JSON blob of computed affinities (Option C) | 24h |
| `aff:lock:{userID}` | String | 30s lock to prevent concurrent on-demand compute | 30s |

### Counter Details

- **Follow counters:** `HIncrBy` with ±1 deltas. A mutual follow gives a score of 2.
- **Community counters:** Incremented when a user joins a community; the counter is bumped for all other active members of that community.
- **Reaction counters:** `HIncrByFloat` with +1.0 for same reaction type, -0.5 for different types.
- **Trust counters:** Pre-computed via recursive CTE in PostgreSQL, stored as `1/depth` (max depth 5).

### Badge Query (Option A)

The `GetPairAffinity` method reads four hash fields (follow, community, reaction, trust), normalizes each to [0, 1], and computes:

```
score = follow * 1.0 + community * 1.5 + reaction * 2.0 + trust * 3.0
```

This runs in O(1) and is suitable for showing a badge on every profile page load without noticeable latency. The badge is always visible because the counters are always fresh.

### Affinity Page (Option C)

When a user visits their full affinity page, the server checks `aff:cached:{userID}` in Redis. If present (and within the 24-hour TTL), it returns the cached result. If absent, it computes the scores on demand, stores them in both PostgreSQL and Redis, and sets the cache for 24 hours.

A 30-second lock (`aff:lock:{userID}`) prevents multiple requests from triggering concurrent computation for the same user.

## Dependencies

These features require:

- **PostgreSQL** — all feature data is stored in the main database
- **Redis** — required for affinity counters and caching (affinity badge and affinity page)

Both dependencies must be running and accessible before the server starts, otherwise the workers will log connection errors.
