# Polaris — Platform Design Document

> **Tagline:** Bringing people together.
> **Core Philosophy:** Positive connection through democratic self-governance, trust-based moderation, and collaborative curation.

## Platform Summary

Polaris is a social platform designed to bring people together through shared interests, trust-based moderation, and democratic self-governance. It strives to be genuinely **addictive but positive** — replacing outrage with connection and algorithmic amplification with collaborative curation.

At its core, Polaris is:
- **Influencers and interests coexist** — you can follow people you admire, join topic-based groups, or both; the feed blends network + interest graph
- **Self-governing** — moderation power flows from trust, not from a central authority; bans require 90% jury consensus
- **Lightweight yet rich** — a single Go binary powers thousands of concurrent users; no microservices, no Kubernetes, no cloud lock-in
- **Privacy-respecting by design** — proximity features use geohashes, never exact locations; no phone number or KYC required
- **Community-owned** — transparent mod logs, public health dashboards, and portable identities ensure the platform serves its users, not the other way around

Everything below describes how that vision is built.

---

# System Design

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                        Client                            │
│            Svelte SPA + Service Worker                    │
│   (offline-capable, optimistic UI, PWA)                  │
└──────────────┬──────────────────────────────┬────────────┘
               │ REST / WebSocket              │
               ▼                               ▼
┌─────────────────────────┐    ┌──────────────────────────┐
│    Go / Gin API Server   │    │  WebSocket Hub (Go)      │
│  (REST, SSE, admin)      │    │  (real-time: live feed,   │
│                          │    │   chat, presence)         │
└────────┬────────┬───────┘    └──────────┬───────────────┘
         │        │                       │
         ▼        ▼                       ▼
┌──────────┐ ┌──────────┐ ┌──────────────────────────────┐
│ PostgreSQL│ │  Redis    │ │  Background Workers (Go)     │
│ (primary) │ │ (cache,   │ │  - Feed generation           │
│           │ │  queues,  │ │  - Content pruning           │
│           │ │  presence)│ │  - Trust decay               │
│           │ │           │ │  - Digest engine             │
│           │ │           │ │  - Notification delivery     │
└──────────┘ └──────────┘ └──────────────────────────────┘
```

### Why this stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Backend   | Go     | Single binary deploy, excellent concurrency for WebSocket, great perf on low-cost VPS |
| API Layer | Gin    | Most widely used Go framework; massive ecosystem, middleware support, loved by AI codegen |
| Frontend  | Svelte | Smallest bundle size of major frameworks, minimal boilerplate, reactive by design, excellent for AI generation |
| Database  | PostgreSQL | Best-in-class relational DB with JSONB, full-text search, recursive CTEs, and extensions (pgvector for similarity) |
| Cache     | Redis   | Pub/sub for real-time, sorted sets for feeds, rate limiting, session store, presence tracking |

---

## 2. Database Schema (PostgreSQL)

### 2.1 Core Tables

```sql
-- Users with trust system (Discourse-inspired, extended)
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(30) UNIQUE NOT NULL,
    display_name    VARCHAR(100),
    bio             TEXT,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255),
    trust_level     SMALLINT DEFAULT 0 NOT NULL,  -- 0=New, 1=Basic, 2=Member, 3=Regular, 4=Leader, 5=Elder
    trust_score     REAL DEFAULT 0 NOT NULL,       -- continuous float for fine-grained ranking
    reputation      BIGINT DEFAULT 0 NOT NULL,     -- sum of gifts/votes received
    invited_by      BIGINT REFERENCES users(id),  -- vouch chain
    invite_code     VARCHAR(32) UNIQUE,            -- personal invite link
    credits         BIGINT DEFAULT 0 NOT NULL,     -- blockchain credits for bounties/gifts
    is_active       BOOLEAN DEFAULT TRUE,
    last_active_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    -- Web of trust: each user can sign others' public keys
    public_key      TEXT,                          -- ed25519 public key
    -- Privacy
    proximity_opt_out BOOLEAN DEFAULT FALSE,
    location_hash   VARCHAR(64)                    -- hash of approx lat/lng grid (privacy-safe)
);

-- Posts ("content items")
CREATE TABLE posts (
    id              BIGSERIAL PRIMARY KEY,
    author_id       BIGINT NOT NULL REFERENCES users(id),
    title           VARCHAR(200),
    body            TEXT,                           -- markdown or rich text
    content_type    SMALLINT NOT NULL DEFAULT 0,    -- 0=text, 1=image, 2=video, 3=link, 4=poll
    mood            SMALLINT NOT NULL DEFAULT 0,    -- 0=general, 1=educational, 2=entertaining, 3=social, 4=creative
    is_educational  BOOLEAN DEFAULT FALSE,
    is_entertaining BOOLEAN DEFAULT FALSE,
    is_nsfw         BOOLEAN DEFAULT FALSE,
    interaction_count BIGINT DEFAULT 0,             -- sum of likes, comments, shares in last 30 days
    cumulative_interactions BIGINT DEFAULT 0,       -- lifetime
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ,                   -- pruned/composted timestamp
    is_deleted      BOOLEAN DEFAULT FALSE
);

-- Tags (collaborative booru-style taxonomy)
CREATE TABLE tags (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) UNIQUE NOT NULL,
    description     TEXT,
    category        VARCHAR(50),                    -- e.g., "topic", "medium", "mood", "fandom"
    is_wiki         BOOLEAN DEFAULT FALSE,          -- tag has a wiki page
    created_by      BIGINT REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Post-Tag mapping (booru-style collaborative tagging)
CREATE TABLE post_tags (
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id          INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    tagged_by       BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (post_id, tag_id, tagged_by)
);

-- Interactions (likes, comments, shares, gifts)
CREATE TABLE interactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    interaction_type SMALLINT NOT NULL,             -- 0=like, 1=comment, 2=share, 3=gift_credit, 4=bounty_answer
    metadata        JSONB,                          -- e.g., comment text, gift amount
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, post_id, interaction_type)     -- one like per user
);

-- Trust connections (web of trust)
CREATE TABLE trust_connections (
    id              BIGSERIAL PRIMARY KEY,
    truster_id      BIGINT NOT NULL REFERENCES users(id),
    trustee_id      BIGINT NOT NULL REFERENCES users(id),
    weight          REAL DEFAULT 1.0 NOT NULL,      -- how much the truster vouches for trustee
    signature       TEXT,                           -- cryptographic signature of the trust statement
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    UNIQUE (truster_id, trustee_id)
);

-- Moderation log (fully transparent, append-only)
CREATE TABLE moderation_actions (
    id              BIGSERIAL PRIMARY KEY,
    action_type     SMALLINT NOT NULL,              -- 0=warning, 1=timeout, 2=suspension, 3=ban, 4=content_remove, 5=tag_cleanup
    target_user_id  BIGINT REFERENCES users(id),
    target_post_id  BIGINT REFERENCES posts(id),
    moderator_id    BIGINT REFERENCES users(id),
    reason          TEXT NOT NULL,
    duration        INTERVAL,                       -- for timeouts/suspensions
    is_jury_decision BOOLEAN DEFAULT FALSE,
    jury_yes        INTEGER DEFAULT 0,              -- count of jury votes
    jury_no         INTEGER DEFAULT 0,
    jury_total      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (id)                                     -- append-only, no UPDATEs
);

-- Juries (randomly selected panels)
CREATE TABLE jury_panels (
    id              BIGSERIAL PRIMARY KEY,
    target_action_id BIGINT REFERENCES moderation_actions(id),
    juror_id        BIGINT REFERENCES users(id),
    vote            BOOLEAN,                        -- true = uphold, false = overturn
    reason          TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (target_action_id, juror_id)
);

-- Circles / Groups
CREATE TABLE circles (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    visibility      SMALLINT DEFAULT 0,             -- 0=public, 1=unlisted, 2=private
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE circle_members (
    circle_id       BIGINT NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            SMALLINT DEFAULT 0,             -- 0=member, 1=moderator, 2=admin
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (circle_id, user_id)
);

-- Feeds (materialized view-like cache)
CREATE TABLE feed_items (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    score           REAL NOT NULL,                  -- algorithm score for ordering
    reason          TEXT,                           -- why this was recommended
    seen            BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),      -- when it was added to feed
    PRIMARY KEY (user_id, post_id)
) PARTITION BY HASH (user_id);

-- Proximity pings (opt-out, approximate)
CREATE TABLE proximity_interests (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    grid_cell       VARCHAR(20) NOT NULL,           -- geohash at ~5km resolution
    tag_id          INTEGER NOT NULL REFERENCES tags(id),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, grid_cell, tag_id)
);

-- Bounties
CREATE TABLE bounties (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    creator_id      BIGINT NOT NULL REFERENCES users(id),
    total_amount    BIGINT NOT NULL,                -- in credits
    status          SMALLINT DEFAULT 0,             -- 0=open, 1=closed, 2=refunded
    best_answer_id  BIGINT REFERENCES interactions(id),
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Dailies / Blockchain credit system
CREATE TABLE daily_rewards (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    date            DATE NOT NULL,
    amount          BIGINT NOT NULL,
    claimed         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, date)
);
```

### 2.2 Indexes

```sql
-- Feed query performance
CREATE INDEX idx_posts_content_type_created ON posts(content_type, created_at DESC);
CREATE INDEX idx_posts_author_created ON posts(author_id, created_at DESC);
CREATE INDEX idx_posts_interaction_count ON posts(interaction_count DESC) WHERE archived_at IS NULL;

-- Tag search
CREATE INDEX idx_post_tags_tag ON post_tags(tag_id);
CREATE INDEX idx_post_tags_post ON post_tags(post_id);

-- Feed generation
CREATE INDEX idx_feed_items_user_score ON feed_items(user_id, score DESC) WHERE seen = FALSE;
CREATE INDEX idx_feed_items_created ON feed_items(user_id, created_at DESC);

-- Mod log
CREATE INDEX idx_mod_actions_target_user ON moderation_actions(target_user_id);
CREATE INDEX idx_mod_actions_created ON moderation_actions(created_at DESC);

-- Trust
CREATE INDEX idx_trust_trustee ON trust_connections(trustee_id);
CREATE INDEX idx_trust_truster ON trust_connections(truster_id);

-- Proximity
CREATE INDEX idx_proximity_grid_tag ON proximity_interests(grid_cell, tag_id);

-- Full-text search on posts
CREATE INDEX idx_posts_fts ON posts USING GIN(to_tsvector('english', body));
```

---

## 3. API Design (Gin Routes)

### 3.1 Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register with invite code; creates user, issues session |
| POST | `/api/v1/auth/login` | Email/password or public-key challenge |
| POST | `/api/v1/auth/logout` | Invalidate session |
| POST | `/api/v1/auth/refresh` | Refresh session token |
| GET  | `/api/v1/auth/session` | Current session info |

### 3.2 Users & Trust

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/users/:id` | Public profile |
| PATCH | `/api/v1/users/:id` | Update own profile (privacy settings, bio) |
| POST | `/api/v1/users/:id/vouch` | Vouch for a user (extends trust connection) |
| GET  | `/api/v1/users/:id/trust` | Trust graph (1 hop) |
| GET  | `/api/v1/users/:id/reputation` | Reputation history |
| POST | `/api/v1/proximity/ping` | Report current location + interests (geohashed) |

### 3.3 Content

| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/v1/posts` | Create post |
| GET    | `/api/v1/posts/:id` | Get post with tags, interactions |
| PATCH  | `/api/v1/posts/:id` | Edit post (own only) |
| DELETE | `/api/v1/posts/:id` | Soft delete |
| GET    | `/api/v1/posts/:id/tags` | Tags on a post |
| POST   | `/api/v1/posts/:id/tags` | Add tag (booru-style, any trusted user) |
| DELETE | `/api/v1/posts/:id/tags/:tagId` | Remove tag (mods or tag creator) |
| POST   | `/api/v1/posts/:id/interact` | Like/comment/share/gift |
| GET    | `/api/v1/posts/:id/related` | Similar posts via tag overlap |

### 3.4 Feeds

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/feed` | Main algorithmic feed (query: ?mood=educational&limit=20) |
| GET  | `/api/v1/feed/hot` | Trending globally |
| GET  | `/api/v1/feed/tag/:tagId` | Feed filtered by tag |
| GET  | `/api/v1/feed/circle/:circleId` | Circle feed |
| GET  | `/api/v1/feed/proximity` | "People near you also like..." |

### 3.5 Moderation

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/mod/flag` | Flag content |
| GET   | `/api/v1/mod/log` | Public moderation log (paginated, filterable) |
| POST  | `/api/v1/mod/timeout` | Issue timeout (trust level 3+) |
| POST  | `/api/v1/mod/ban` | Propose ban (requires jury) |
| POST  | `/api/v1/mod/jury/:actionId/vote` | Jury vote on action |
| GET   | `/api/v1/mod/queue` | Current flag queue (trust level 3+) |

### 3.6 Circles

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/circles` | Create circle |
| GET   | `/api/v1/circles` | List public circles |
| GET   | `/api/v1/circles/:id` | Circle detail |
| POST  | `/api/v1/circles/:id/join` | Join circle |
| POST  | `/api/v1/circles/:id/leave` | Leave circle |

### 3.7 Bounties & Economy

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/bounties` | Create bounty on a post |
| POST  | `/api/v1/bounties/:id/award` | Award bounty to answer |
| GET   | `/api/v1/bounties/open` | Open bounties |
| POST  | `/api/v1/credits/claim-daily` | Claim daily reward |
| GET   | `/api/v1/leaderboard/:type` | Leaderboard (type=posting, tagging, gifts, trust) |

### 3.8 WebSocket

| Path | Description |
|------|-------------|
| `/ws/v1/feed` | Real-time feed updates (new posts, interactions) |
| `/ws/v1/chat/:circleId` | Circle chat |
| `/ws/v1/presence` | Who's online in your circles |
| `/ws/v1/notifications` | Live notification stream |

---

## 4. Feed Algorithm (the core of addictiveness)

### 4.1 Scoring Function

For each candidate post `p` and user `u`:

```
score(p, u) = 
    recency_score * 0.3 +
    engagement_velocity_score * 0.2 +
    tag_affinity_score(p, u) * 0.25 +
    trust_network_score(p, u) * 0.15 +
    proximity_score(p, u) * 0.05 +
    content_quality_score(p) * 0.05
```

**recency_score** = `1 / (1 + hours_since_posted / 24)` — decay over time

**engagement_velocity_score** = interactions per hour in last 24h, log-normalized to [0,1]. This is what makes it "addictive" — surfacing content that's gaining traction right now, but dampened so pure viral garbage doesn't dominate.

**tag_affinity_score** = cosine similarity between user's tag preference vector (derived from their past interactions) and the post's tag vector. Computed via pgvector or a Redis sorted set per user.

**trust_network_score** = average trust weight of users who interacted with this post. If someone you trust deeply liked it, it gets a boost. If someone you don't trust liked it, neutral or negative.

**proximity_score** = number of people in your geohash cell + 2 ring cells who have interacted with this tag. Only applies if `proximity_opt_out = false`.

**content_quality_score** = mix of: post length (too short = low effort), has images?, spelling score, is it derivative/remix?, how many tags does it have?

### 4.2 Feed Generation Pipeline

1. **Candidate pool** — Redis sorted set per user of candidate post IDs (filled by background worker every 5 min)
2. **Scoring** — worker iterates candidates, computes score, writes to `feed_items` table
3. **Serving** — API queries `feed_items WHERE user_id = $1 AND seen = FALSE ORDER BY score DESC`
4. **Real-time boost** — when someone in your trust network interacts with a post, the post gets a temporary score bump (WebSocket push to feed)

### 4.3 Diversity Guarantee

No more than 3 posts from the same tag cluster in the first 20 results. No more than 2 posts from the same author in the first 20. This prevents filter bubbles.

### 4.4 Content Pruning

A background cron runs daily:

```sql
UPDATE posts SET archived_at = NOW()
WHERE archived_at IS NULL
  AND created_at < NOW() - INTERVAL '6 months'
  AND content_type IN (2, 3)  -- entertaining: video, link
  AND interaction_count = 0;
```

Configurable per admin via `site_config` table.

---

## 5. Trust System

### 5.1 Trust Levels

There are no defined "levels" in the way Discourse has them (e.g., Level 1 is 'Basic', Level 2 is 'Member'). The system is a pure **merit-based trust score** with granular thresholds.

| Threshold | Minimum Requirements | Privileges Gated |
|-----------|---------------------|-----------------|
| 0 | Account created | Read, post with approval, no DMs |
| 1 | 5 posts, 3 days old, 1 like received | Post without approval, tag content |
| 2 | 30 days old, 50 posts, 10 likes received, 5 trust connections | Create circles, flag content, vote in polls |
| 3 | 90 days old, 200 posts, 50 likes, trust score > 0.5 | Issue timeouts, moderate own circle, award bounties |
| 4 | 1 year old, 200 likes, trust score > 2.0 | Mod queue access, issue suspensions, veto flags |
| 5 | Appointed by 80% of threshold-4+ vote, trust score > 5.0 | Vote on permanent bans, site configuration |

**No fixed "Level 2" — everything is score-driven.** If the community decides someone at trust score 1.8 is trustworthy enough to moderate, they can vouch for them. The thresholds are recommendations, not rigid walls.

### 5.2 Trust Score

Each user has a single **trust_score** (float). It is computed as:

```
trust_score =
    reputation * 0.5 +
    avg(weight of incoming trust connections) * 0.3 +
    time_spent_active * 0.1 +
    diversity_of_interactions * 0.1
```

Where:
- **reputation**: sum of credits gifted to this user, log-normalized
- **incoming trust connections**: weighted average of how much others trust this user
- **time_spent_active**: hours spent reading/interacting, capped at some reasonable maximum
- **diversity_of_interactions**: how many different tag clusters the user has interacted with (prevents single-topic echo chambers)

### 5.3 Vouch Chain

When a new user joins, they must be vouched for by someone at threshold 3+. That vouch is a **directed trust edge** from the inviter to the invitee.

**Responsibility**: For 90 days after the invite, if the invitee's content is flagged 3+ times, the inviter's trust_score takes a hit proportional to the severity. This makes invite chains self-policing.

### 5.4 Decay

trust_score decays by 5% per week of inactivity after 30 days. This is additive — 6 months of inactivity reduces trust score to near zero. Trust score never goes below 0.

### 5.5 Web of Trust Graph

Every user has an ed25519 keypair. Trust connections are signed statements: "I trust user X at weight Y."

The trust graph is **personalized** — when showing content to user A, we favor posts that scored highly with users who are close in A's trust graph (weighted by path distance).

**Sybil resistance**: One user can vouch for at most 5 invites per 90 days. A new account must reach threshold 3 before it can vouch for others.

### 5.6 Revocation

Users can revoke a trust connection at any time. When they do, the edge is removed, and any content scored using that edge is re-scored on the next feed rebuild. Revocations are public (logged in a dedicated table) to prevent abuse.

---

## 6. Moderation System

### 6.1 Soft Actions (Timeouts)

- Any user at threshold 3+ can issue a warning or timeout to a user they do not share a trust connection with
- **Timeout duration**: max 24h for users at threshold 3+, max 72h for threshold 4+
- Users can timeout a user they share a trust connection with, but this costs the issuer 10% of their own trust_score — a steep price that makes it reserved for serious cases
- All timeouts are logged to the public mod log automatically

### 6.2 Hard Actions (Suspension & Ban)

| Action | Who Can Propose | Who Decides |
|--------|-----------------|-------------|
| Warning | threshold 3+ | Single moderator, logged |
| Timeout | threshold 3+ | Single moderator, logged |
| Suspension (7-30 days) | threshold 4+ | 60% jury (threshold 2+, n=12) |
| Permanent ban | threshold 5 | 90% jury (threshold 2+, n=20) |

### 6.3 Jury Selection

1. A random pool is selected from active (last 7 days) users at the required threshold
2. Pool is balanced by:
   - Timezone (±3 hours of the accused's active hours)
   - Trust score distribution (mix of high and mid-score users)
   - No direct trust connections to the accused or accuser
3. Jury votes blind (cannot see other votes) for 48 hours
4. Votes are public after the decision
5. Appeals go to a new pool of threshold 4+ users (n=24 for bans, n=12 for suspensions)

### 6.4 Collaborative Curation (Boorus)

Any user at threshold 1+ can add tags to any post. Tags are **not author-controlled** — they're community-controlled. This prevents authors from hiding problematic content behind misleading tags.

Tags can have wiki pages (created by any threshold 3+ user). Wiki pages are collaboratively edited (Wikipedia-style) with full revision history.

**Tag conflict**: If multiple users disagree on a tag for a post, a vote is triggered among users at threshold 2+ who have tagged > 10 posts anywhere. Winner is tag with most votes after 24 hours.

---

## 7. Web of Trust & Proof of Humanity

Each user generates an ed25519 keypair on registration. The public key is their identity anchor.

**Web of trust graph**: 
- `trust_connections` table stores directed weighted edges.
- When you vouch for someone, you cryptographically sign their public key.
- Trust score = PageRank-like algorithm on the trust graph (personalized per user).
- Sybil resistance: new accounts must be vouched for by a level 3+ user. One user can vouch for at most 5 accounts in 90 days.

**Location privacy**: 
- Location is hashed to a geohash at ~5km resolution.
- `proximity_interests` stores only `(geohash, tag_id)` pairs.
- "X people near you also like Y" query: count distinct users in geohash + 2 ring cells who have interacted with tag Y.
- EXACT LOCATION IS NEVER STORED OR EXPOSED. The service worker on the client side broadcasts a coarse geohash; the client never sends raw coordinates.

**Proof of humanity**: 
- Optional: challenge-response where user must post original content (photo with a specific hand gesture + current day's code from the site).
- Trust level 3+ users can verify others by video call.
- Verified users get a `is_verified` flag visible on their profile.
- No KYC, no ID documents, no phone number requirement.

---

## 8. Blockchain Credits System

**Not a real blockchain.** Uses a ledger table with cryptographic hashing for auditability, but runs on PostgreSQL. This keeps the site lightweight.

```sql
CREATE TABLE credit_transactions (
    id              BIGSERIAL PRIMARY KEY,
    from_user       BIGINT REFERENCES users(id),
    to_user         BIGINT REFERENCES users(id),
    amount          BIGINT NOT NULL,
    transaction_type SMALLINT NOT NULL,  -- 0=daily, 1=gift, 2=bounty, 3=bounty_award
    reference_id    BIGINT,              -- post_id or bounty_id
    hash            VARCHAR(64),         -- sha256 of previous transaction to form chain
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

**Daily reward**: 10 credits per day (configurable). Claimed via a single click. Streak bonus: +1 per consecutive day, reset on miss.

**Purchasing credits**: Users buy credits with real money via Stripe. Exchange rate: 100 credits = $1 USD (configurable). Minimum purchase: 500 credits ($5).

**Spending**:
- Gift/tip another user: any amount
- Subscribe to an influencer's premium tier: monthly recurring credit cost set by the influencer
- Unlock a pay-per-view post: one-time credit cost set by the author
- Set bounty on a question: minimum 50 credits
- Boost a post in feed: 100 credits (post gets a one-time score multiplier)
- Create a circle: 25 credits (anti-spam)
- Change username: 100 credits (anti-squatting)

**1% platform tax**: Every on-platform credit transfer (gifts, subscriptions, bounty awards) incurs a 1% fee that is burned (removed from circulation). This is the platform's revenue. Purchases of credits are not taxed — only transfers between users.

**Cash-out**: Influencers can convert credits back to real money via Stripe Connect (bank transfer). Requirements:
- Minimum cash-out: 10,000 credits ($100)
- Must complete KYC (identity verification) before first cash-out
- Platform pays out at the same rate: 100 credits = $1 (the 1% tax on transfers ensures the platform's cut is taken before cash-out)
- Cash-out is limited to once per week
- Credits earned from daily rewards cannot be cashed out — only credits purchased with real money or received as gifts/tips are withdrawable. This prevents Sybil farming of the daily reward.

---

## 9. Real-Time Architecture

```
Client ──WebSocket──► WebSocket Hub (Go)
                          │
                          ├── Redis Pub/Sub: user:${id}:notifications
                          ├── Redis Pub/Sub: circle:${id}:chat
                          ├── Redis Pub/Sub: feed:boosts
                          │
                          ▼
                   Background Workers
                          │
                       PostgreSQL
```

- Each Go process runs a WebSocket hub using gorilla/websocket.
- Hub subscribes to Redis channels for the users connected to it.
- On disconnect, hub unsubscribes from unused channels.
- Feed boosts: when an interaction occurs, worker publishes to `feed:boosts`, the hub sends a "new high-scoring post" event to affected users' feed streams.

---

## 10. Lightweight Deployment

### Minimum Requirements
- **VPS**: 2 vCPU, 4GB RAM, 50GB SSD (~$15/mo on any provider)
- **PostgreSQL**: on same machine or $5/mo managed instance
- **Redis**: on same machine, minimal footprint
- **Static assets**: served via Go's embed or a CDN

### Startup Command
```bash
./polaris --db=postgres://... --redis=localhost:6379 --addr=:8080
```

Single binary, zero external dependencies at runtime (except DB and Redis).

### Performance Targets
- 10,000 concurrent WebSocket connections per instance
- 100,000 daily active users per instance
- Feed generation: < 50ms p99
- Post creation: < 30ms p95

---

## 11. Directory Structure

```
polaris/
├── cmd/
│   └── polaris/
│       └── main.go                  # Entry point
├── internal/
│   ├── api/
│   │   ├── router.go                # Gin router setup
│   │   ├── middleware/
│   │   │   ├── auth.go              # Session/JWT middleware
│   │   │   ├── trust.go             # Trust level gate
│   │   │   └── ratelimit.go         # Redis-based rate limiter
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── posts.go
│   │   │   ├── feeds.go
│   │   │   ├── users.go
│   │   │   ├── moderation.go
│   │   │   ├── circles.go
│   │   │   ├── bounties.go
│   │   │   ├── leaderboard.go
│   │   │   └── proximity.go
│   │   └── websocket/
│   │       ├── hub.go               # WebSocket connection manager
│   │       └── handlers.go          # WS message routing
│   ├── db/
│   │   ├── migrations/
│   │   │   └── 001_initial.sql
│   │   ├── db.go                    # Connection pool
│   │   └── queries/                 # Raw SQL or sqlc-generated
│   ├── model/
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── tag.go
│   │   └── ... (struct definitions)
│   ├── services/
│   │   ├── feed.go                  # Feed algorithm
│   │   ├── trust.go                 # Trust score computation
│   │   ├── moderation.go            # Jury selection, action execution
│   │   ├── pruning.go               # Content lifecycle
│   │   └── credits.go               # Ledger, dailies, bounties
│   └── worker/
│       ├── scheduler.go             # Cron-like scheduling
│       ├── feed_worker.go           # Rebuild feeds periodically
│       ├── trust_decay.go           # Decay trust for inactive users
│       └── prune_worker.go          # Archive old content
├── web/                             # Svelte frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/                 # API client (fetch wrapper)
│   │   │   ├── components/         # Reusable Svelte components
│   │   │   ├── stores/             # Svelte stores (auth, feed, etc.)
│   │   │   └── utils/              # Helpers
│   │   ├── routes/                  # SvelteKit pages
│   │   ├── app.html
│   │   └── app.css
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.js
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── docker-compose.yml              # Local dev: Go + Postgres + Redis
```

---

## 12. Svelte Frontend Component Tree

```
App
├── Layout
│   ├── Navbar (hamburger menu, search, notifications bell, profile avatar)
│   ├── MoodPicker (toggle: educational / entertaining / social / creative / all)
│   └── Footer (mod log link, health dashboard, credits balance)
├── Feed
│   ├── FeedItem (post card)
│   │   ├── AuthorInfo (avatar, trust level badge, name)
│   │   ├── PostBody (text, image carousel, video player, link preview)
│   │   ├── TagList (clickable tags; booru-style)
│   │   ├── InteractionBar (like, comment, share, gift buttons)
│   │   ├── CommentThread (expandable comments)
│   │   └── FlagButton (report content)
│   └── FeedSorter (hot / new / top / trusted)
├── Profile
│   ├── UserCard (bio, trust level, reputation, vouches)
│   ├── PostHistory (user's posts, filterable)
│   └── TrustGraph (visualization of connections)
├── Circles
│   ├── CircleCard (name, member count, tags)
│   ├── CircleFeed
│   └── CircleChat (WebSocket-based)
├── ModCenter
│   ├── ModLog (public, searchable, filterable)
│   ├── FlagQueue (for trust level 3+)
│   └── JuryVote (pending jury duties)
├── BountyBoard
│   ├── BountyCard (question, amount, time remaining)
│   ├── BountyDetail (answers, vote for best)
│   └── CreateBounty
├── ProximityPane
│   ├── LocationToggle (opt-out)
│   └── NearbyTopics ("5 people near you also like Stargazing")
├── HealthDashboard
│   ├── ActiveUsers (line chart)
│   ├── FlagResolutionRate
│   ├── AvgSentiment (over time)
│   └── TopTags (interest graph)
├── Leaderboard (posting, tagging, gifting, trust score)
└── AdminPanel (site config, pruning settings, content types)
```

---

## 13. Key Design Decisions & Rationale

### Why not ActivityPub/Fediverse?
Federation adds significant complexity (inbox/outbox, conflict resolution, state replication). For v1, keep it centralized. Design the data model so federation can be layered on later (every user has a public key, content is addressable by ID, trust graph is portable).

### Why not a real blockchain?
Real blockchains are slow, expensive, and confusing to end users. A cryptographic ledger in PostgreSQL provides the same auditability without any of the UX friction. Users don't care about consensus algorithms; they care about "can I spend my daily credits on a bounty."

### Why not microservices?
For an indie/team project, monolith + background workers is faster to develop, deploy, and debug. The API is separated into logical handlers but shares the same process. If the platform grows to massive scale, the feed worker can be extracted into its own service.

### Why SQL over an ORM?
Sqlc (or raw SQL with pgx) gives compile-time query checking, better performance, and full control. The schema is the source of truth. Go structs are generated from SQL. This is idiomatic Go and avoids ORM pain.

### Why Svelte over React?
Svelte compiles to vanilla JS, producing bundles ~60% smaller than equivalent React apps. Fewer client-side dependencies means faster load times on mobile. Reactive declarations `$:` map naturally to the mood-toggle and feed-filtering use cases. SvelteKit provides SSR for initial page loads (SEO) and then takes over as SPA for the logged-in experience.

---

## 14. Questions to Answer Before Coding

### Architecture & Scale
1. **What's the expected DAU at launch vs year 1?** — determines whether we need horizontal scaling from day one or can start with a monolith on a single $15 VPS.
2. **Do we need to support federation (ActivityPub) eventually, or is this always a walled garden?** — affects how we model identity and content IDs.
3. **What's the SLA?** — if this is a passion project we accept downtime; if users rely on it, we need health checks, backups, and failover.

### Moderation & Trust
4. **Who appoints the first elders (trust level 5)?** — bootstrap problem. The first 100 users need to be carefully onboarded. Do we start with trusted alpha testers?
5. **How do we handle edge cases where a trusted user goes rogue?** — e.g., a level 4 user starts mass-timeouting everyone. Do we need a "circuit breaker" (admin override)?
6. **How do we prevent Sybil farming of the daily reward?** — daily credits are non-withdrawable, but a bot farm could still accumulate them for on-platform influence (bounties, boosts). We need additional Sybil resistance (invite-only initially, CAPTCHA on registration, rate-limited daily claims).
7. **How do we prevent vote manipulation of the jury system?** — what if a coordinated group joins to vote on each other's favor? We need Sybil resistance at registration (invite-only initially?).

### Privacy & Legal
8. **Where will servers be located?** — GDPR (EU), CCPA (California), or both impose different data handling requirements. EU requires data deletion on request, right to explanation for algorithmic decisions.
9. **Is there any form of content that is always prohibited (CSAM, terrorist material)?** — if yes, we need automated detection + reporting. This is non-negotiable legally.
10. **How do we handle GDPR data export/deletion requests?** — need an automated mechanism before launch.
11. **What jurisdiction's laws govern the platform?** — affects everything from moderation liability to age requirements.

### Product & Growth
12. **What's the onboarding funnel?** — invite-only? Open registration with captcha? The vouch chain is great, but the first 1,000 users need to come from somewhere.
13. **How do we seed the initial content?** — empty platform = no retention. Do we have "seed groups" with curated content? Do we encourage early users to post by having them be "founding members" with special badges?
14. **What's the expected split between free and paying users?** — if 1% of DAU buys credits, that may be enough to sustain the platform. But we need to model unit economics: server costs per user vs. average credit spend per paying user.
15. **What languages/locales does the platform need to support at launch?** — affects full-text search config, UI strings, content filtering.

### Technical
16. **What's the backup strategy?** — PostgreSQL WAL streaming? Daily pg_dump? Point-in-time recovery? Without backups, a corrupted database kills the platform.
17. **How do we handle image/video uploads?** — direct to local disk? S3-compatible storage (MinIO, Backblaze)? File size limits? Thumbnail generation?
18. **Do we need a mobile app, or is PWA sufficient?** — PWA can do push notifications, offline support, and home screen install. Native app gives better camera access and background sync.
19. **What's the test strategy?** — Go backend: integration tests with test database? Frontend: Playwright for E2E? Without tests, refactoring becomes impossible.
20. **Who is the developer?** — is this a solo project? A team? An AI agent building the whole thing? The stack and tooling choices may shift based on who's writing the code. If it's an AI, we need extremely explicit specs and comprehensive tests, as AI tends to produce subtle bugs in edge cases.
21. **How do we handle spam/SEO abuse?** — if the platform is open, it will be hit by spam bots. Rate limiting, captcha on registration, email verification, automated spam detection (simple NLP on post body).
22. **What metrics define success for v1?** — 1000 DAU? 100 posts/day? < 10% flagged content? Having clear goals prevents scope creep.
