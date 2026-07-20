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
    -- ActivityPub
    actor_id        VARCHAR(255) UNIQUE,           -- https://example.com/users/username (populated for federation)
    is_local        BOOLEAN DEFAULT TRUE,          -- false for federated actors
    -- Onboarding
    onboarding_stage SMALLINT DEFAULT 0,           -- 0=pending, 1=tags_picked, 2=users_followed, 3=first_post, 4=complete
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
    content_warning TEXT,                           -- if set, post is blurred with this text as the warning label
    interaction_count BIGINT DEFAULT 0,             -- sum of likes, comments, shares in last 30 days
    cumulative_interactions BIGINT DEFAULT 0,       -- lifetime
    status          SMALLINT DEFAULT 0,             -- 0=published, 1=draft, 2=scheduled
    scheduled_at    TIMESTAMPTZ,                    -- post becomes visible at this time
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

-- Circles (small private groups for existing friend groups, max ~100 members)
CREATE TABLE circles (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    max_members     INTEGER DEFAULT 100,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE circle_members (
    circle_id       BIGINT NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            SMALLINT DEFAULT 0,             -- 0=member, 1=admin
    invited_by      BIGINT REFERENCES users(id),
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (circle_id, user_id)
);

-- Communities (topic-based public spaces, discoverable, with curators)
CREATE TABLE communities (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    tags             TEXT[] NOT NULL DEFAULT '{}',    -- scope-enforced tags (posts must use these)
    curator_lock    BOOLEAN DEFAULT FALSE,            -- if true, only curators can approve new members
    slow_boot_days  INTEGER DEFAULT 0,                -- mandatory waiting period before full participation
    forked_from     BIGINT REFERENCES communities(id),-- if this community was forked from another
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ                       -- auto-archived after 30 days of inactivity
);

CREATE TABLE community_members (
    community_id    BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            SMALLINT DEFAULT 0,             -- 0=member, 1=curator, 2=admin
    status          SMALLINT DEFAULT 0,             -- 0=probation (in slow_boot), 1=active, 2=left, 3=booted
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (community_id, user_id)
);

-- Community curators (users who can enforce scope and manage slow-boot)
CREATE TABLE curators (
    id              BIGSERIAL PRIMARY KEY,
    community_id    BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission      SMALLINT DEFAULT 1,             -- 1=curator, 2=admin
    appointed_by    BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (community_id, user_id)
);

-- Community forks (when a subset of members breaks off and creates a new community)
CREATE TABLE community_forks (
    id              BIGSERIAL PRIMARY KEY,
    source_id       BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    fork_id         BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    reason          TEXT NOT NULL,
    initiated_by    BIGINT NOT NULL REFERENCES users(id),
    member_count    INTEGER NOT NULL DEFAULT 0,      -- how many members moved
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Community Notes (misinformation warnings, Twitter/X-style)
CREATE TABLE community_notes (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id       BIGINT NOT NULL REFERENCES users(id),
    body            TEXT NOT NULL,                    -- the note content explaining the issue
    status          SMALLINT DEFAULT 0,              -- 0=pending, 1=visible, 2=rejected, 3=removed
    helpful_yes     INTEGER DEFAULT 0,
    helpful_no      INTEGER DEFAULT 0,
    consensus_score REAL DEFAULT 0,                  -- composite score for ranking
    requires_author  BOOLEAN DEFAULT FALSE,           -- note comes from the post author themselves
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

-- Votes on community notes (determines consensus)
CREATE TABLE community_note_votes (
    id              BIGSERIAL PRIMARY KEY,
    note_id         BIGINT NOT NULL REFERENCES community_notes(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    vote            BOOLEAN NOT NULL,                -- true=helpful, false=not helpful
    trust_score_at_vote REAL NOT NULL,               -- user's trust score when they voted (for weighting)
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (note_id, user_id)
);

-- Shared blocklist (domains, known spam, coordinated abuse networks)
CREATE TABLE blocklist_entries (
    id              BIGSERIAL PRIMARY KEY,
    entry_type      SMALLINT NOT NULL,               -- 0=domain, 1=ip_prefix, 2=public_key, 3=fingerprint
    entry_value     VARCHAR(500) NOT NULL,
    reason          TEXT,
    severity        SMALLINT DEFAULT 0,              -- 0=warning, 1=block, 2=shadow_ban
    added_by        BIGINT REFERENCES users(id),      -- threshold 4+ user who proposed it
    jury_approved   BOOLEAN DEFAULT FALSE,            -- whether jury confirmed it
    shared          BOOLEAN DEFAULT TRUE,             -- exportable to other instances
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (entry_type, entry_value)
);

-- Active user stats (materialized for fast circle/community browsing)
CREATE TABLE active_stats (
    id              BIGSERIAL PRIMARY KEY,
    scope_type      SMALLINT NOT NULL,                -- 0=circle, 1=community, 2=tag, 3=global
    scope_id        BIGINT NOT NULL,
    active_1d       INTEGER DEFAULT 0,
    active_7d       INTEGER DEFAULT 0,
    active_30d      INTEGER DEFAULT 0,
    computed_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (scope_type, scope_id, computed_at)
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

-- Custom weighted feeds (user-created feed definitions)
CREATE TABLE custom_feeds (
    id              BIGSERIAL PRIMARY KEY,
    owner_id        BIGINT NOT NULL REFERENCES users(id),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    slug            VARCHAR(100) NOT NULL,
    is_public       BOOLEAN DEFAULT FALSE,           -- shareable with others
    sort_order      SMALLINT DEFAULT 0,              -- 0=chrono, 1=hot, 2=top, 3=custom_score
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    UNIQUE (owner_id, slug)
);

-- Sources within a custom feed (tag, user, keyword, list)
CREATE TABLE feed_sources (
    id              BIGSERIAL PRIMARY KEY,
    feed_id         BIGINT NOT NULL REFERENCES custom_feeds(id) ON DELETE CASCADE,
    source_type     SMALLINT NOT NULL,               -- 0=tag, 1=user, 2=keyword, 3=user_list
    source_id       BIGINT,                          -- tag_id, user_id, or user_list_id
    source_value    VARCHAR(200),                    -- raw keyword or list name
    include_mode    BOOLEAN DEFAULT TRUE,            -- true=include, false=exclude
    sort_priority   SMALLINT DEFAULT 0,              -- ordering within the feed
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Collections (bookmarks with 4 visibility layers)
CREATE TABLE collections (
    id              BIGSERIAL PRIMARY KEY,
    owner_id        BIGINT NOT NULL REFERENCES users(id),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    visibility      SMALLINT NOT NULL DEFAULT 0,     -- 0=private, 1=public_read, 2=public_edit, 3=anybody_edit
    is_default      BOOLEAN DEFAULT FALSE,           -- auto-saved bookmarks go here
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

-- Posts within collections (with per-post visibility override)
CREATE TABLE collection_posts (
    id              BIGSERIAL PRIMARY KEY,
    collection_id   BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    added_by        BIGINT NOT NULL REFERENCES users(id),
    notes           TEXT,                             -- optional user note on this bookmark
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (collection_id, post_id)
);

-- Per-user content filters (regex, domain, keyword, user)
CREATE TABLE content_filters (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    filter_type     SMALLINT NOT NULL,               -- 0=keyword, 1=regex, 2=domain, 3=user
    filter_value    VARCHAR(500) NOT NULL,           -- the pattern or identifier
    filter_action   SMALLINT NOT NULL DEFAULT 0,     -- 0=hide, 1=blur, 2=low_priority
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, filter_type, filter_value)
);

-- Saved searches (user's saved search queries)
CREATE TABLE saved_searches (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    name            VARCHAR(100) NOT NULL,
    query_json      JSONB NOT NULL,                  -- structured query: { terms: [], tags: [], users: [], ... }
    notify_on_new   BOOLEAN DEFAULT FALSE,           -- push notification when new results found
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Trending topics (auto-detected from post content, not hashtags)
CREATE TABLE trending_topics (
    id              BIGSERIAL PRIMARY KEY,
    topic           VARCHAR(200) NOT NULL,
    frequency       INTEGER NOT NULL DEFAULT 0,      -- occurrences in detection window
    velocity        REAL NOT NULL DEFAULT 0,          -- rate of change
    detection_window TSTZRANGE,                      -- the time range this trend covers
    tag_id          INTEGER REFERENCES tags(id),     -- linked tag if one matches
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (topic, detection_window)
);

-- Achievement definitions
CREATE TABLE achievements (
    id              SERIAL PRIMARY KEY,
    code            VARCHAR(50) UNIQUE NOT NULL,     -- machine-readable key
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    icon            VARCHAR(200),                    -- icon URL or class
    category        SMALLINT NOT NULL DEFAULT 0,     -- 0=social, 1=curation, 2=quality, 3=longevity
    criteria_json   JSONB NOT NULL,                  -- threshold config: { "type": "post_count", "value": 100 }
    sort_order      SMALLINT DEFAULT 0
);

-- User unlock state for achievements
CREATE TABLE user_achievements (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    achievement_id  INTEGER NOT NULL REFERENCES achievements(id),
    unlocked_at     TIMESTAMPTZ DEFAULT NOW(),
    progress        REAL DEFAULT 0,                  -- 0.0–1.0 (partial progress for display)
    visible         BOOLEAN DEFAULT TRUE,            -- user can hide achievement
    UNIQUE (user_id, achievement_id)
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

-- Custom feeds
CREATE INDEX idx_custom_feeds_owner ON custom_feeds(owner_id);
CREATE INDEX idx_feed_sources_feed ON feed_sources(feed_id);

-- Collections
CREATE INDEX idx_collections_owner ON collections(owner_id);
CREATE INDEX idx_collection_posts_collection ON collection_posts(collection_id);
CREATE INDEX idx_collection_posts_post ON collection_posts(post_id);

-- Filters
CREATE INDEX idx_content_filters_user ON content_filters(user_id);

-- Saved searches
CREATE INDEX idx_saved_searches_user ON saved_searches(user_id);

-- Trending
CREATE INDEX idx_trending_window ON trending_topics USING GIST(detection_window);
CREATE INDEX idx_trending_frequency ON trending_topics(frequency DESC);

-- Achievements
CREATE INDEX idx_user_achievements_user ON user_achievements(user_id);

-- Communities
CREATE INDEX idx_communities_slug ON communities(slug);
CREATE INDEX idx_communities_tags ON communities USING GIN(tags);
CREATE INDEX idx_community_members_community ON community_members(community_id);
CREATE INDEX idx_community_members_user ON community_members(user_id);

-- Community notes
CREATE INDEX idx_community_notes_post ON community_notes(post_id);
CREATE INDEX idx_community_notes_status ON community_notes(status);
CREATE INDEX idx_note_votes_note ON community_note_votes(note_id);

-- Blocklist
CREATE INDEX idx_blocklist_type ON blocklist_entries(entry_type);

-- Active stats
CREATE INDEX idx_active_stats_scope ON active_stats(scope_type, scope_id);
CREATE INDEX idx_active_stats_7d ON active_stats(scope_type, scope_id, active_7d DESC);

-- Scheduled posts
CREATE INDEX idx_posts_scheduled ON posts(status, scheduled_at) WHERE status = 2;
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
| GET  | `/api/v1/feed/custom/:slug` | Feed from a custom feed definition |
| POST | `/api/v1/feed/custom` | Create a custom weighted feed |
| GET  | `/api/v1/feed/custom` | List user's custom feeds |
| PUT  | `/api/v1/feed/custom/:slug` | Update a custom feed |
| DELETE | `/api/v1/feed/custom/:slug` | Delete a custom feed |
| POST | `/api/v1/feed/custom/:slug/sources` | Add a source (tag/user/keyword) to feed |
| DELETE | `/api/v1/feed/custom/:slug/sources/:id` | Remove a source from feed |

### 3.5 Moderation

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/mod/flag` | Flag content |
| GET   | `/api/v1/mod/log` | Public moderation log (paginated, filterable) |
| POST  | `/api/v1/mod/timeout` | Issue timeout (trust level 3+) |
| POST  | `/api/v1/mod/ban` | Propose ban (requires jury) |
| POST  | `/api/v1/mod/jury/:actionId/vote` | Jury vote on action |
| GET   | `/api/v1/mod/queue` | Current flag queue (trust level 3+) |

### 3.6 Circles & Communities

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/circles` | Create a private circle |
| GET   | `/api/v1/circles` | List user's circles |
| GET   | `/api/v1/circles/:id` | Circle detail |
| POST  | `/api/v1/circles/:id/join` | Join circle (by invite) |
| POST  | `/api/v1/circles/:id/leave` | Leave circle |
| POST  | `/api/v1/communities` | Create a public community |
| GET   | `/api/v1/communities` | Browse communities (?sort=active_weekly\|active_daily\|active_monthly\|member_count\|newest, default=active_weekly) |
| GET   | `/api/v1/communities/:slug` | Community detail with active user counts |
| POST  | `/api/v1/communities/:slug/join` | Request to join community |
| POST  | `/api/v1/communities/:slug/leave` | Leave community |
| POST  | `/api/v1/communities/:slug/fork` | Fork community (creates new one with same tag scope) |
| GET   | `/api/v1/communities/:slug/members` | List members (with active status) |
| POST  | `/api/v1/communities/:slug/curators` | Add curator (admin only) |
| DELETE | `/api/v1/communities/:slug/curators/:userId` | Remove curator |

### 3.7 Bounties & Economy

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/bounties` | Create bounty on a post |
| POST  | `/api/v1/bounties/:id/award` | Award bounty to answer |
| GET   | `/api/v1/bounties/open` | Open bounties |
| POST  | `/api/v1/credits/claim-daily` | Claim daily reward |
| GET   | `/api/v1/leaderboard/:type` | Leaderboard (type=posting, tagging, gifts, trust) |

### 3.8 Community Notes

| Method | Path | Description |
|--------|------|-------------|
| POST  | `/api/v1/notes` | Create a community note on a post |
| POST  | `/api/v1/notes/:id/vote` | Vote helpful/not-helpful on a note |
| GET   | `/api/v1/notes/post/:postId` | Visible notes on a specific post |
| GET   | `/api/v1/notes/pending` | Pending notes awaiting consensus (threshold 3+) |

### 3.9 Collections

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/collections` | List user's collections |
| POST | `/api/v1/collections` | Create a collection |
| GET  | `/api/v1/collections/:id` | Collection detail with posts |
| PUT  | `/api/v1/collections/:id` | Update collection (name, visibility) |
| DELETE | `/api/v1/collections/:id` | Delete collection |
| POST | `/api/v1/collections/:id/posts` | Add post to collection |
| DELETE | `/api/v1/collections/:id/posts/:postId` | Remove post from collection |
| GET  | `/api/v1/collections/:id/share` | Get share link for public collection |

### 3.10 Search

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/search` | Full-text search (?q=terms&tags=1,2&author=3&after=2025-01-01&before=2025-06-01&sort=relevance) |
| POST | `/api/v1/search/advanced` | Boolean search with JSON body: `{ "and": [...], "or": [...], "not": [...], "phrase": "...", "fields": {...} }` |
| GET  | `/api/v1/search/saved` | List saved searches |
| POST | `/api/v1/search/saved` | Save a search query |
| DELETE | `/api/v1/search/saved/:id` | Delete saved search |

### 3.11 Content Filters

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/filters` | List user's content filters |
| POST | `/api/v1/filters` | Create a filter (type=keyword/regex/domain/user, value, action=hide/blur) |
| PUT  | `/api/v1/filters/:id` | Update filter |
| DELETE | `/api/v1/filters/:id` | Delete filter |

### 3.12 Trending

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/trending` | Current trending topics (?window=24h&limit=20) |
| GET  | `/api/v1/trending/:topicId` | Posts in a trending topic |

### 3.13 Achievements & Leaderboards

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/achievements` | All achievement definitions + current user's progress |
| GET  | `/api/v1/achievements/:id` | Single achievement detail with recent unlockers |
| GET  | `/api/v1/leaderboard/:type` | Leaderboard (type=posting, tagging, gifts, trust_network, achievement_points; query: ?period=weekly&limit=50) |
| GET  | `/api/v1/leaderboard/:type/:userId` | User's rank on a leaderboard |

### 3.14 Onboarding

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/onboarding/progress` | Current onboarding stage |
| POST | `/api/v1/onboarding/stage` | Advance to next stage |
| POST | `/api/v1/onboarding/tags` | Pick initial tags (minimum 5) |
| GET  | `/api/v1/onboarding/suggested-users` | Suggested users based on selected tags |

### 3.15 RSS/Atom Feeds

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/feed/rss` | Main feed as RSS/Atom (?format=rss\|atom) |
| GET  | `/api/v1/feed/tag/:tagId/rss` | Tag feed as RSS |
| GET  | `/api/v1/feed/custom/:slug/rss` | Custom feed as RSS |
| GET  | `/api/v1/feed/community/:slug/rss` | Community feed as RSS |
| GET  | `/api/v1/users/:id/posts/rss` | User's posts as RSS |
| GET  | `/api/v1/collections/:id/rss` | Collection as RSS |
| GET  | `/api/v1/search/rss` | Search results as RSS (saved searches) |

### 3.16 Shared Blocklist

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/api/v1/mod/blocklist` | List blocklist entries (?type=domain) |
| POST | `/api/v1/mod/blocklist` | Add blocklist entry (threshold 4+) |
| DELETE | `/api/v1/mod/blocklist/:id` | Remove blocklist entry (threshold 5+) |
| GET  | `/api/v1/mod/blocklist/export` | Export blocklist for federation (JSON) |

### 3.17 Observability & Admin

| Method | Path | Description |
|--------|------|-------------|
| GET  | `/metrics` | Prometheus metrics (request duration, DB latency, WS connections) |
| GET  | `/health` | Health check (DB + Redis connectivity) |
| GET  | `/api/v1/admin/status` | Admin panel: current state, active users, queue depths |

### 3.18 WebSocket

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

## 5. Custom Feeds & Weighted Sources

Custom feeds let users build their own algorithmic mix. They are shareable, composable, and blend social + interest graphs.

### 5.1 Source Types

| Source | Include/Exclude | Behavior |
|--------|----------------|----------|
| Tag | Both | All posts with this tag |
| User | Include | All posts from this user |
| User list | Include | Posts from all users in the list |
| Keyword | Both | Posts containing (or avoiding) this keyword in title/body |

A feed can have any number of sources. Exclude sources override include sources. Sources are ANDed within a category, ORed across categories.

### 5.2 Feed Generation

Custom feeds bypass the main algorithm entirely. They query posts matching the source rules, sorted by the feed's `sort_order` (chronological, hot, top, or custom score). No diversity guarantees — the user explicitly chose these sources.

### 5.3 Sharing

Public feeds can be shared via URL. Any user can subscribe to another user's feed. Subscribed feeds appear in the sidebar. Feed authors can see subscriber count.

### 5.4 API Integration

The feed endpoint accepts a `feed=` query parameter. If a user has a custom feed with slug `photography`, they can link to `/feed/custom/photography`. Feeds are also available as RSS/Atom for external readers.

---

## 6. Collections & Bookmarks

Every user gets a default collection (Twitter-style bookmark) for saving posts. They can create additional collections with 4 visibility layers.

### 6.1 Visibility Layers

| Level | Behavior |
|-------|----------|
| Private | Only the owner sees the collection and its contents |
| Public read | Anyone can view, share, and subscribe to changes |
| Public edit | Trusted users (threshold 2+) can add/remove posts |
| Anybody edit | Any logged-in user can add/remove posts (wiki-style) |

### 6.2 Default Collection

Created on registration. The bookmark button saves to this collection by default. Users can change which collection is their default.

### 6.3 Collections Feed

Each collection automatically generates a feed (sorted by `sort_order`). Public collections appear on user profiles. Collections can also be followed — new additions appear in the subscriber's feed.

### 6.4 Use Cases

- **Reading list**: save posts to read later
- **Community bibliography**: public collection tagged `#rust` with best Rust tutorials
- **Collaborative gallery**: public-edit collection where event attendees add their photos
- **Watchlist**: private collection of users/content to monitor

---

## 7. Advanced Search

Full boolean search over posts, tags, users, and metadata. Powered by PostgreSQL full-text search with a structured query parser.

### 7.1 Query Syntax

| Feature | Syntax | Example |
|---------|--------|---------|
| AND | space or `+` | `rust async` |
| OR | `\|` | `rust \| go` |
| NOT | `-` | `rust -borrow` |
| Exact phrase | `"..."` | `"borrow checker"` |
| Field: title | `title:...` | `title:generics` |
| Field: author | `author:...` | `author:alice` |
| Field: tag | `tag:...` | `tag:photography` |
| Field: user list | `in:list_name` | `in:favorites` |
| Field: collection | `collection:...` | `collection:reading` |
| Date range | `after:2025-01-01 before:2025-06-01` | |
| Sort | `sort:relevance\|new\|top` | `sort:new` |

### 7.2 Saved Searches

Users can save any query with a name. Saved searches appear in the sidebar with a badge showing new results count. Optionally, users can subscribe to push notifications when new results match a saved search.

### 7.3 Implementation

The search endpoint parses the query string into a structured query JSON, then builds a parameterized SQL query. For performance, the `posts` table has a GIN index on `to_tsvector('english', body)`. For field filters, the query planner uses b-tree indexes on `author_id`, `created_at`, and the `post_tags` join.

---

## 8. Per-User Content Filters

Users control what they see. Filters are processed client-side and server-side for feed generation.

### 8.1 Filter Types

| Type | Action | Example |
|------|--------|---------|
| Keyword (exact/substring) | Hide, blur, low-priority | Hide all posts mentioning "spoiler" |
| Regex pattern | Hide, blur, low-priority | Blur posts matching `\bnsfw\b` |
| Domain | Hide, blur | Hide links from `example.com` |
| User | Hide, blur | Mute a specific user |

### 8.2 Filter Processing

- **Server-side (feed generation)**: Hidden posts are excluded from the candidate pool. Low-priority posts get a score penalty (-0.3).
- **Client-side (render)**: Blurred posts are rendered with a blurred overlay that the user can click to reveal. This gives users full control without the platform removing content.
- **Applied before the feed algorithm**: Diversity guarantees and scoring operate on the remaining pool.

### 8.3 Filter Management

Users manage filters from their settings page. Filters have an `is_active` toggle for temporary disable. Users can also temporarily disable all filters (e.g., "show everything mode") from the mood picker dropdown.

---

## 9. Trending Topics

Trending topics are auto-detected from post content — not hashtags. This prevents hashtag spam and captures organic conversation.

### 9.1 Detection Algorithm

A background worker runs every 15 minutes:

1. **Extract**: Collect posts from the last hour. Tokenize titles + body, remove stop words, extract noun phrases.
2. **Count**: Frequency of each term in the current window vs. the baseline (last 7 days, averaged).
3. **Score**: `trending_score = frequency_in_window / baseline_frequency * log(window_frequency)`. This favors terms that are unusually hot right now.
4. **Cluster**: Group similar terms (stemming + edit distance < 2) into topics.
5. **Store**: Write to `trending_topics` with the detection window.

### 9.2 Display

- Trending topics appear in a sidebar widget: "🔥 Rising: Rust, Generics, Photography"
- Each topic links to a feed filtered by that term
- If a topic matches an existing tag (via fuzzy match), it links to the tag page
- Topics with velocity > 2.0 get a "🔥" indicator; > 5.0 get "🚀"

### 9.3 Anti-abuse

- Terms must appear in at least 3 different users' posts
- Terms must appear in at least 2 different tag categories
- Accounts under 30 days old do not contribute to trending detection
- Manual override: mods can pin, unpin, or hide topics from trending

---

## 10. Achievements & Leaderboards

Gamification focused on positive behaviors: curation, quality contributions, and community health.

### 10.1 Achievement Catalog

| Category | Example Achievements |
|----------|---------------------|
| Social | First like received, first comment, first follower, first gift |
| Curation | 10 tags added, 50 tags added, 100 tags added, tag wiki created |
| Quality | 10 likes received, 100 likes received, top 10% scorers by month |
| Longevity | 30-day streak, 90-day streak, 1 year on platform |
| Community | First circle created, circle reaches 50 members, served on 10 juries |
| Special | Bug report accepted, first-time poster vouched for |

### 10.2 Progression Display

- Achievements can be hidden from profile (privacy) — `visible` flag in `user_achievements`
- Progress is shown as a progress bar (e.g., "8/10 tags added") before unlock
- Unlocked achievements show the date and a brief animation
- Achievement points sum is shown on the profile: "🏆 245 pts — Level 4 Curator"

### 10.3 Leaderboards

Leaderboards are computed for multiple periods to keep them fresh:

| Period | Reset | Meant For |
|--------|-------|-----------|
| Daily | Every 24h | Quick dopamine, catch rising stars |
| Weekly | Every Monday | Consistent engagement |
| Monthly | 1st of month | Serious contributors |
| Yearly | Jan 1 | Hall of fame |
| All-time | Never | Platform legends |

Leaderboard types:
- **Posting**: posts created (quality-weighted: upvotes × post length factor)
- **Tagging**: tags added, wiki pages created, tag conflicts resolved
- **Gifts**: credits gifted (not received — incentivizes generosity)
- **Trust network**: incoming trust connections × trust score
- **Achievement points**: sum of all achievement point values

### 10.4 Anti-gaming

- Posts in the first 7 days of an account don't count toward leaderboards (prevents bot farms)
- Leaderboard position is updated once per period, not in real-time
- Only the top 100 positions are shown publicly; the rest show percentile

---

## 11. Communities (Public Topic Spaces)

Communities are the public face of Polaris — discoverable topic-based spaces where people gather around shared interests. Unlike Circles (private, invite-only, max ~100 members), Communities are open, browsable, and designed for growth.

### 11.1 Circles vs. Communities

| Aspect | Circles | Communities |
|--------|---------|-------------|
| Visibility | Private, invite-only | Public, discoverable |
| Max size | ~100 members | Unlimited |
| Discovery | None (not listed) | Browseable, tag-indexed |
| Joining | By invitation only | Open join or curator-approved |
| Moderation | Circle admin | Platform trust system + curators |
| Forking | N/A | Members can fork if they disagree |
| Scope enforcement | None | Tags are enforced (posts must match) |
| Slow boot | No | Optional probation period |
| Active user stats | Not shown | Shown on community page (default sort) |

### 11.2 Curators

Each community has curators (appointed by existing curators) who:
- Approve/deny join requests when `curator_lock` is enabled
- Remove posts that violate the community's tag scope
- Set the community's slow boot duration (0-14 days)
- Cannot ban users (only platform trust system can do that)

Curators are not moderators. They are scope enforcers. All moderation (timeouts, suspensions, bans) still goes through the platform-wide trust system.

### 11.3 Slow Boot

When a community has `slow_boot_days > 0`, new members are on probation:
- They can read and browse, but their posts are held for curator review
- They cannot vote in community-specific polls
- After the probation period, they become active members automatically
- If a curator rejects 3+ posts during probation, the membership is revoked

This prevents drive-by spam and gives communities time to socialize new members.

### 11.4 Scope Enforcement

Communities can declare a set of scope tags. Any post in the community's feed must use at least one of these tags. Curators can remove posts that don't fit the scope. Posts with irrelevant tags are flagged by the system.

Example: a `#photography` community would have scope tags `photography, landscape, portrait, street-photography, macro, film`. A post about cooking would be removed from that community.

### 11.5 Forking

If a group of members disagrees with the direction of a community, they can fork it. A fork:
1. Copies the community's name (with "-fork" suffix) and tag scope
2. Copies the member list (each user must confirm within 7 days)
3. Sets `forked_from` on the new community
4. Does NOT copy curators — the fork starts with the initiators as curators

Forks are public. The `community_forks` table tracks the relationship. Users can browse a community's fork history.

### 11.6 Active User Analytics

Every community page displays:
- **Daily active users**: users who interacted (posted, commented, voted, tagged) in the last 24h
- **Weekly active users**: last 7 days (default sort metric for community browse)
- **Monthly active users**: last 30 days

These stats are computed by a background worker every 30 minutes and stored in the `active_stats` table. The API defaults to sorting communities by weekly active users, with options for daily, monthly, member count, and newest.

This same analytics model applies to tags and circles for consistency.

---

## 12. Community Notes (Misinformation Warnings)

Inspired by Twitter/X Community Notes, this feature allows trusted users to add context to posts. Notes appear inline below the post with a colored banner when they reach consensus.

### 12.1 Creating a Note

Any user at threshold 2+ can add a note to any post. The note should explain:
- What is misleading or inaccurate
- What the correct information is
- Sources or references

Notes that are purely opinion-based ("I disagree with this") are discouraged and will be rejected by the consensus system.

### 12.2 Consensus Algorithm

A note becomes visible when it reaches a consensus threshold:

1. **Voting period**: 48 hours after creation. Users at threshold 1+ can vote "Helpful" or "Not Helpful".
2. **Weighted votes**: Each vote is weighted by the voter's trust_score. A threshold 4 user's vote counts more than a threshold 1 user's vote.
3. **Diversity requirement**: For a note to become visible, it must have helpful votes from users who don't share overlapping tag preferences with the unhelpful voters. This prevents coordinated mobbing and ensures cross-ideological consensus.
4. **Threshold**: A note becomes visible when `weighted_helpful_ratio > 0.65` AND `unique_author_groups > 2` (voters from at least 3 different tag clusters).

### 12.3 Display

- Visible notes appear below the post with a pale yellow background and "Community Note" header
- If the post author themselves adds a note (retraction/correction), it is always shown with a green "Author's note" banner
- Notes are sorted by helpfulness score (best first)
- Users can collapse notes with a toggle

### 12.4 Abuse Prevention

- Each user can create at most 3 notes per day
- A user whose note is rejected 5+ times in 30 days is blocked from creating notes for 90 days
- Notes on notes are not allowed (no nesting)
- A post with an active visible note gets a small indicator in feeds: "ⓘ" icon

---

## 13. Post Scheduling & Content Warnings

### 13.1 Scheduled Posts

Users can compose posts and schedule them for future publication:

- **Draft status** (`status = 1`): Saved but not visible. Only the author can see drafts. Drafts count toward storage but not toward posting leaderboards.
- **Scheduled status** (`status = 2`): Post will publish at `scheduled_at`. A background worker checks every minute for posts where `status = 2 AND scheduled_at <= NOW()` and sets them to `status = 0`.
- **Published status** (`status = 0`): Normal visible post.

API:
- `POST /api/v1/posts` — accepts optional `scheduled_at` and `status` fields
- `GET /api/v1/posts/scheduled` — list user's scheduled posts
- `PATCH /api/v1/posts/:id` — can change scheduled_at, status, or body before publish

### 13.2 Content Warnings

When `content_warning` is set on a post:
- In feeds, the post body is replaced with: **[Content Warning: {reason}] — Click to reveal**
- The post title and author are still visible (so users can decide)
- The interaction bar is hidden until the post is revealed
- Content warnings are different from NSFW flags — they're for any potentially sensitive content (spoilers, medical images, strong language, etc.)

The author sets the content warning on create. Curators and threshold 3+ users can retroactively add content warnings to others' posts (logged in mod log).

---

## 14. RSS/Atom Everywhere

Every feed on Polaris is available as RSS or Atom. This ensures content is accessible to users who prefer traditional feed readers and provides a migration path if the platform ever shuts down.

### 14.1 Available Feeds

| Resource | RSS URL | Atom URL |
|----------|---------|----------|
| Main feed | `/api/v1/feed/rss` | `/api/v1/feed/atom` |
| Tag feed | `/api/v1/feed/tag/:tagId/rss` | `/api/v1/feed/tag/:tagId/atom` |
| Custom feed | `/api/v1/feed/custom/:slug/rss` | `/api/v1/feed/custom/:slug/atom` |
| Community feed | `/api/v1/feed/community/:slug/rss` | `/api/v1/feed/community/:slug/atom` |
| User posts | `/api/v1/users/:id/posts/rss` | `/api/v1/users/:id/posts/atom` |
| Collection | `/api/v1/collections/:id/rss` | `/api/v1/collections/:id/atom` |
| Search results | `/api/v1/search/rss?saved=:searchId` | `/api/v1/search/atom?saved=:searchId` |
| Circle feed | `/api/v1/circles/:id/rss` | `/api/v1/circles/:id/atom` |

### 14.2 Implementation

- RSS/Atom feeds are generated from the same query logic as the JSON API, but rendered as XML templates
- Feeds are cached in Redis for 15 minutes (TTL) to reduce DB load
- ETag headers are set based on last content change date
- Each feed entry includes: title, link, author, published date, summary (first 500 chars of body), tags as categories
- Full content can be included via `?full=true` query parameter

### 14.3 Autodiscovery

HTML `<link>` tags for RSS/Atom are injected into every page's `<head>`:
```html
<link rel="alternate" type="application/rss+xml" title="Feed" href="/api/v1/feed/rss">
<link rel="alternate" type="application/atom+xml" title="Feed" href="/api/v1/feed/atom">
```

Community, tag, and user pages each point to their respective feeds.

---

## 15. Onboarding Wizard

The first-time user experience is critical for retention. Polaris uses a guided onboarding flow with 4 stages.

### 15.1 Stages

| Stage | What Happens | API |
|-------|-------------|-----|
| 0: Pending | User registered but hasn't started | Show welcome screen |
| 1: Tags | User picks 5+ tags of interest | `POST /api/v1/onboarding/tags` |
| 2: Users | Suggested users based on tags, user follows 3+ | `POST /api/v1/onboarding/follows` |
| 3: First post | User is prompted to create their first post (optional, skippable) | `POST /api/v1/onboarding/first-post` |
| 4: Complete | Onboarding done, show main feed | `POST /api/v1/onboarding/complete` |

### 15.2 Stage Enforcement

If `onboarding_stage < 4`:
- The API returns the onboarding page instead of the main feed
- Feed endpoints return a "Please complete onboarding" error
- Push notifications are suppressed
- The user can't create content (except during stage 3)

### 15.3 Suggested Users Algorithm

For stage 2, the server queries:
1. Users who have posted in the user's selected tags (filtered by trust score > 0.5)
2. Sorted by interaction count in those tags
3. Excludes users below threshold 1 (new accounts are not good suggestions)
4. Returns top 20 with a "Why suggested" reason: "Alice posts a lot in #photography"

### 15.4 Skipping

Users can skip the first post stage. They cannot skip tag selection (minimum 5 tags) because the algorithm needs data to build their feed.

---

## 16. Observability & Monitoring

A social platform must be observable to operators. The design prioritizes lightweight, zero-overhead instrumentation.

### 16.1 Structured Logging

- Use `zerolog` for structured JSON logging
- Every request logs: method, path, status, duration, user_id (anonymized)
- Every DB query logs: query hash, duration, rows returned
- Every background worker run logs: task name, duration, items processed, errors
- Logs are written to stdout (captured by the container runtime)
- No sensitive data (passwords, email bodies, tokens) ever enters logs

### 16.2 Prometheus Metrics (`/metrics`)

| Metric | Type | Labels |
|--------|------|--------|
| `polaris_requests_total` | Counter | method, path, status |
| `polaris_request_duration_ms` | Histogram | method, path |
| `polaris_db_query_duration_ms` | Histogram | query_name |
| `polaris_ws_connections` | Gauge | — |
| `polaris_ws_messages_total` | Counter | event_type |
| `polaris_users_online` | Gauge | — |
| `polaris_queue_depth` | Gauge | queue_name |
| `polaris_cache_hit_ratio` | Gauge | cache_name |

### 16.3 Health Checks

- `GET /health`: Returns 200 if PostgreSQL and Redis are reachable
- `GET /health/ready`: Returns 200 when the server is ready to accept traffic (migrations complete, caches warmed)

### 16.4 Dashboard (Admin)

The `/api/v1/admin/status` endpoint returns a JSON snapshot:
- Active users (1h, 24h, 7d, 30d)
- Queue depths (feed generation, trending, scheduled posts)
- Cache hit rates
- Error rates by endpoint
- Current rate limit state
- Database connection pool stats

No separate admin UI — the admin panel is a simple HTML page rendered by the Go server.

---

## 17. Trust System

### 17.1 Trust Levels

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

### 17.2 Trust Score

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

### 17.3 Vouch Chain

When a new user joins, they must be vouched for by someone at threshold 3+. That vouch is a **directed trust edge** from the inviter to the invitee.

**Responsibility**: For 90 days after the invite, if the invitee's content is flagged 3+ times, the inviter's trust_score takes a hit proportional to the severity. This makes invite chains self-policing.

### 17.4 Decay

trust_score decays by 5% per week of inactivity after 30 days. This is additive — 6 months of inactivity reduces trust score to near zero. Trust score never goes below 0.

### 17.5 Web of Trust Graph

Every user has an ed25519 keypair. Trust connections are signed statements: "I trust user X at weight Y."

The trust graph is **personalized** — when showing content to user A, we favor posts that scored highly with users who are close in A's trust graph (weighted by path distance).

**Sybil resistance**: One user can vouch for at most 5 invites per 90 days. A new account must reach threshold 3 before it can vouch for others.

### 17.6 Revocation

Users can revoke a trust connection at any time. When they do, the edge is removed, and any content scored using that edge is re-scored on the next feed rebuild. Revocations are public (logged in a dedicated table) to prevent abuse.

---

## 18. Moderation System

### 18.1 Soft Actions (Timeouts)

- Any user at threshold 3+ can issue a warning or timeout to a user they do not share a trust connection with
- **Timeout duration**: max 24h for users at threshold 3+, max 72h for threshold 4+
- Users can timeout a user they share a trust connection with, but this costs the issuer 10% of their own trust_score — a steep price that makes it reserved for serious cases
- All timeouts are logged to the public mod log automatically

### 18.2 Hard Actions (Suspension & Ban)

| Action | Who Can Propose | Who Decides |
|--------|-----------------|-------------|
| Warning | threshold 3+ | Single moderator, logged |
| Timeout | threshold 3+ | Single moderator, logged |
| Suspension (7-30 days) | threshold 4+ | 60% jury (threshold 2+, n=12) |
| Permanent ban | threshold 5 | 90% jury (threshold 2+, n=20) |

### 18.3 Jury Selection

1. A random pool is selected from active (last 7 days) users at the required threshold
2. Pool is balanced by:
   - Timezone (±3 hours of the accused's active hours)
   - Trust score distribution (mix of high and mid-score users)
   - No direct trust connections to the accused or accuser
3. Jury votes blind (cannot see other votes) for 48 hours
4. Votes are public after the decision
5. Appeals go to a new pool of threshold 4+ users (n=24 for bans, n=12 for suspensions)

### 18.4 Collaborative Curation (Boorus)

Any user at threshold 1+ can add tags to any post. Tags are **not author-controlled** — they're community-controlled. This prevents authors from hiding problematic content behind misleading tags.

Tags can have wiki pages (created by any threshold 3+ user). Wiki pages are collaboratively edited (Wikipedia-style) with full revision history.

**Tag conflict**: If multiple users disagree on a tag for a post, a vote is triggered among users at threshold 2+ who have tagged > 10 posts anywhere. Winner is tag with most votes after 24 hours.

### 18.5 Shared Blocklist

The platform maintains a shared blocklist of known spam domains, IP prefixes, and coordinated abuse networks. Entries are added by threshold 4+ users and approved by a jury of threshold 3+ users (60% majority, n=8).

| Entry Type | Example | Action |
|------------|---------|--------|
| Domain | `spamlink.example.com` | Posts with this domain are held for mod review |
| IP prefix | `192.168.1.0/24` | Registration from these IPs requires manual approval |
| Public key | `ed25519:abc...` | Associated accounts are flagged for review |
| Fingerprint | Browser fingerprint hash | Rate-limited heavily |

Blocklist entries can be exported as JSON (`/api/v1/mod/blocklist/export`) for sharing with other instances. When another instance pulls the export, they can:
- Apply the blocklist as-is
- Mirror it with their own severity overrides
- Use it as a training signal for their own detection

Entries are shared with the `shared = TRUE` flag. Some entries (e.g., IPs of targeted harassment victims) are kept local.

### 18.6 Admin Safety Circuit Breaker

A single admin account exists outside the trust system. It is:
- **Disabled by default** — enabled only via an environment variable + hardware 2FA key
- **Fully logged** — every action is recorded in the public mod log with an "Admin Override" banner
- **Limited scope** — can reverse jury decisions, restore banned users, and revert timeouts. Cannot create content, vote, or interact as a regular user.
- **Emergency-only** — meant for system-level emergencies (juror coordinated manipulation, database corruption, legal takedown requests)

The admin account's existence and every action it takes is visible on the health dashboard. If the admin account is used more than once per quarter, an automatic community review is triggered.

---

## 19. Web of Trust & Proof of Humanity

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

## 20. Blockchain Credits System

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

## 21. Real-Time Architecture

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

## 22. Lightweight Deployment

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

### Rate Limiting

Concrete tiers per endpoint group, enforced by Redis-based sliding window:

| Tier | Endpoints | Limit | Window |
|------|-----------|-------|--------|
| Auth | login, register, refresh | 5 req | 1 min |
| Write | create post, comment, interact | 30 req | 1 min |
| Read | feed, search, profile | 100 req | 1 min |
| Mod | flag, timeout, ban proposal | 10 req | 1 min |
| Onboarding | tag pick, first post | 20 req | 1 min |
| Blocklist | add/remove entries | 5 req | 1 min |
| RSS | feed exports | 60 req | 1 min |
| WebSocket | connect | 10 conn | 1 min per IP |

Limits are per-user (authenticated) or per-IP (unauthenticated). When exceeded, the server returns `429 Too Many Requests` with a `Retry-After` header.

### Backup Strategy

| Type | Frequency | Retention | Tool |
|------|-----------|-----------|------|
| WAL archive | Continuous (every 60s) | 7 days | `pg_receivewal` → S3 |
| Full DB dump | Daily | 30 days | `pg_dump --format=custom` |
| Point-in-time recovery | On-demand | Within WAL window | `pg_restore + WAL replay` |
| Redis dump | Every 6h | 3 days | `SAVE` + S3 copy |
| Uploaded files | Continuous | Until post deleted | S3 versioning |

Backups are encrypted (AES-256) before upload. A restore drill is run quarterly. The restore script is checked into the repo at `scripts/restore.sh`.

---

## 23. Directory Structure

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
│   │   │   ├── custom_feeds.go        # Custom weighted feed CRUD
│   │   │   ├── users.go
│   │   │   ├── collections.go         # Bookmark collections
│   │   │   ├── search.go              # Advanced search + saved searches
│   │   │   ├── filters.go             # Per-user content filters
│   │   │   ├── trending.go            # Trending topics
│   │   │   ├── moderation.go
│   │   │   ├── circles.go
│   │   │   ├── bounties.go
│   │   │   ├── leaderboard.go
│   │   │   ├── achievements.go        # Achievements + progress
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
│   │   ├── collection.go
│   │   ├── feed.go                  # Custom feed + feed_source structs
│   │   ├── filter.go
│   │   ├── trending.go
│   │   ├── achievement.go
│   │   └── ... (struct definitions)
│   ├── services/
│   │   ├── feed.go                  # Feed algorithm
│   │   ├── custom_feed.go           # Weighted feed generation
│   │   ├── communities.go           # Community management + curators
│   │   ├── community_notes.go       # Community Notes CRUD + consensus
│   │   ├── scheduling.go            # Scheduled post publishing
│   │   ├── blocklist.go             # Shared blocklist
│   │   ├── rss.go                   # RSS/Atom feed generation
│   │   ├── onboarding.go            # Wizard stages + suggested users
│   │   ├── trust.go                 # Trust score computation
│   │   ├── moderation.go            # Jury selection, action execution
│   │   ├── pruning.go               # Content lifecycle
│   │   └── credits.go               # Ledger, dailies, bounties
│   └── worker/
│       ├── scheduler.go             # Cron-like scheduling
│       ├── feed_worker.go           # Rebuild feeds periodically
│       ├── trust_decay.go           # Decay trust for inactive users
│       ├── trending_worker.go       # Detect trending topics every 15m
│       ├── achievement_worker.go    # Evaluate achievement progress
│       ├── scheduled_post_worker.go # Publish scheduled posts every 60s
│       ├── community_notes_worker.go# Evaluate note consensus
│       ├── active_stats_worker.go   # Compute active user stats every 30m
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

## 24. Svelte Frontend Component Tree

```
App
├── Layout
│   ├── Navbar (hamburger menu, search, notifications bell, profile avatar)
│   ├── Sidebar
│   │   ├── FeedList (main feed, custom feeds, circle feeds)
│   │   ├── TrendingWidget (auto-detected trending topics)
│   │   └── SavedSearches (saved searches with new-result badges)
│   ├── MoodPicker (toggle: educational / entertaining / social / creative / all)
│   └── Footer (mod log link, health dashboard, credits balance)
├── Feed
│   ├── CustomFeedBar (selector for user's custom feeds, create new)
│   ├── FeedItem (post card)
│   │   ├── AuthorInfo (avatar, trust level badge, name)
│   │   ├── PostBody (text, image carousel, video player, link preview)
│   │   ├── TagList (clickable tags; booru-style)
│   │   ├── InteractionBar (like, comment, share, gift, bookmark buttons)
│   │   ├── CommentThread (expandable comments)
│   │   └── FlagButton (report content)
│   └── FeedSorter (hot / new / top / trusted / custom)
├── CustomFeedBuilder
│   ├── SourceList (tags, users, keywords with include/exclude toggles)
│   ├── SortOrderPicker (chrono, hot, top, custom)
│   └── ShareToggle (public / unlisted / private)
├── Search
│   ├── SearchBar (query input with syntax help dropdown)
│   ├── SearchFilters (date range, tag filter, author filter, sort)
│   ├── AdvancedToggle (boolean search builder)
│   └── SavedSearchList (save, load, delete, subscribe to searches)
├── Collections
│   ├── CollectionList (user's collections, default bookmark collection)
│   ├── CollectionDetail (posts in collection, sort, filter)
│   ├── CollectionShare (visibility toggle, share link)
│   └── BookmarkButton (inline on FeedItem, adds to default collection)
├── Profile
│   ├── UserCard (bio, trust level, reputation, vouches, achievement badge showcase)
│   ├── AchievementWall (grid of unlocked achievements with progress bars)
│   ├── PostHistory (user's posts, filterable)
│   └── TrustGraph (visualization of connections)
├── Settings
│   ├── ContentFilters (add/edit/delete keyword, regex, domain, user filters)
│   ├── FilterPreview (test a filter against recent feed)
│   └── FilterToggle (temporarily disable all filters)
├── Circles (private groups)
│   ├── CircleCard (name, member count)
│   ├── CircleFeed
│   └── CircleChat (WebSocket-based)
├── Communities (public topic spaces)
│   ├── BrowseCommunities (sortable by active weekly/daily/monthly, member count)
│   ├── CommunityPage (header with active user stats, curator list, scope tags)
│   ├── CommunityFeed (scope-enforced, with community note indicators)
│   ├── JoinCommunity (with slow-boot countdown when applicable)
│   └── ForkCommunity (initiate a community fork)
├── ModCenter
│   ├── ModLog (public, searchable, filterable)
│   ├── FlagQueue (for trust level 3+)
│   ├── JuryVote (pending jury duties)
│   ├── SharedBlocklist (view blocklist entries, propose new ones)
│   └── AdminOverride (circuit breaker panel, disabled by default)
├── CommunityNotes
│   ├── NoteCard (inline on post, colored banner, helpfulness score)
│   ├── CreateNote (threshold 2+, for any post)
│   ├── NoteVoteButton (helpful / not helpful)
│   └── PendingNotesQueue (threshold 3+, review pending notes)
├── BountyBoard
│   ├── BountyCard (question, amount, time remaining)
│   ├── BountyDetail (answers, vote for best)
│   └── CreateBounty
├── OnboardingWizard (stages 0-4, shown before main app)
│   ├── StageTags (pick 5+ tags of interest)
│   ├── StageUsers (follow suggested users)
│   ├── StageFirstPost (optional, skippable)
│   └── StageComplete (onboarding complete, enter app)
├── PostComposer (rich markdown editor, image upload with client-side compression)
│   ├── SchedulePicker (optional: post now or schedule for later)
│   ├── ContentWarningToggle (optional: add content warning label)
│   └── PollBuilder (optional: create poll with up to 4 options)
├── ProximityPane
│   ├── LocationToggle (opt-out)
│   └── NearbyTopics ("5 people near you also like Stargazing")
├── HealthDashboard
│   ├── ActiveUsers (line chart)
│   ├── FlagResolutionRate
│   ├── AvgSentiment (over time)
│   └── TopTags (interest graph)
├── Leaderboard (period selector: daily/weekly/monthly/yearly/all-time)
└── AdminPanel (site config, trending overrides, pruning settings, content types)
```

### Keyboard Shortcuts

All shortcuts are discoverable via `?` (help overlay). Implemented as a Svelte action that binds to `document.onkeydown`.

| Key | Action |
|-----|--------|
| `?` | Show keyboard shortcuts help overlay |
| `j` / `k` | Navigate up/down through feed items |
| `l` | Like/unlike focused post |
| `s` | Save/unsave (bookmark) focused post |
| `b` | Toggle sidebar visibility |
| `/` | Focus search bar |
| `Enter` | Open focused post in detail view |
| `Esc` | Close overlay / return to feed |
| `n` | Compose new post |
| `p` | Go to profile |
| `.` | Reload feed (pull-to-refresh equivalent) |
| `=` | Open mood picker |

All shortcuts are disabled when typing in an input field (`document.activeElement?.tagName === 'INPUT'`).

### Mental Health Features

Polaris includes passive mental health guardrails:

- **Scroll timer**: After 30 continuous minutes of scrolling (no user interaction gap > 2 min), a gentle modal appears: "You've been scrolling for a while. Would you like to take a break?" with options "Take a 5 min break" (shows a nature photo) and "I'm fine, continue".
- **Notification bundling**: Push notifications are bundled — at most one per 15 minutes per user. No "like bomb" notifications.
- **No infinite scroll**: Pagination uses "Load more" buttons with visible page numbers instead of infinite scroll. Users see "Page 3 of 12" rather than an endless abyss.
- **Reading time estimates**: Every post shows estimated reading time ("3 min read"). Users can filter to "Quick reads" (< 5 min) when they're short on time.
- **Sleep mode**: Users can set quiet hours (e.g., 11pm-7am) during which push notifications are suppressed and the UI uses warm color tones. Configurable per user in settings.

These features are purely client-side (no additional server cost) except notification bundling.

---

## 25. Key Design Decisions & Rationale

### Why not ActivityPub/Fediverse?

Federation adds significant complexity (inbox/outbox, conflict resolution, state replication). For v1, keep it centralized. However, the data model is designed with federation in mind:

**User → Actor mapping**:
```
User.actor_id        => "https://polari.social/users/alice"
User.public_key      => ed25519 public key (used for HTTP Signatures)
User.is_local        => true (false for federated actors)
```

**Post → Note / Article mapping**:
```
Post.id              => ActivityPub object ID: "https://polari.social/posts/123"
Post.body            => content (in `content` field as HTML)
Post.content_type    => activity type hint (Note, Article, Image, Video)
Tags                 => Hashtag objects in `tag` array
```

**Collection → OrderedCollection mapping**:
```
Collections are exposed as ActivityPub OrderedCollection
Paginated via ActivityPub's `first`, `next`, `last` properties
```

In a future federated version:
- `GET /api/v1/users/:id/outbox` — returns posts as ActivityPub `OrderedCollection`
- `POST /api/v1/users/:id/inbox` — receives ActivityPub activities from other instances
- `GET /.well-known/webfinger?resource=acct:alice@polari.social` — user discovery
- HTTP Signatures (using ed25519 keys in `users.public_key`) for request authentication
- Follows and likes federate via standard ActivityPub activities (Follow, Like, Announce)
- Blocklist entries can be shared via a simple JSON endpoint, pulled by other instances periodically

This design ensures that if federation is ever added, no data migration is needed — the existing tables have the right fields.

### Why not a real blockchain?
Real blockchains are slow, expensive, and confusing to end users. A cryptographic ledger in PostgreSQL provides the same auditability without any of the UX friction. Users don't care about consensus algorithms; they care about "can I spend my daily credits on a bounty."

### Why not microservices?
For an indie/team project, monolith + background workers is faster to develop, deploy, and debug. The API is separated into logical handlers but shares the same process. If the platform grows to massive scale, the feed worker can be extracted into its own service.

### Why SQL over an ORM?
Sqlc (or raw SQL with pgx) gives compile-time query checking, better performance, and full control. The schema is the source of truth. Go structs are generated from SQL. This is idiomatic Go and avoids ORM pain.

### Why Svelte over React?
Svelte compiles to vanilla JS, producing bundles ~60% smaller than equivalent React apps. Fewer client-side dependencies means faster load times on mobile. Reactive declarations `$:` map naturally to the mood-toggle and feed-filtering use cases. SvelteKit provides SSR for initial page loads (SEO) and then takes over as SPA for the logged-in experience.

---

## 26. Questions to Answer Before Coding

### Architecture & Scale
1. **What's the expected DAU at launch vs year 1?** — determines whether we need horizontal scaling from day one or can start with a monolith on a single $15 VPS.
2. **Do we need to support federation (ActivityPub) eventually, or is this always a walled garden?** — the data model is ready for federation, but the actual outbox/inbox/webfinger endpoints would need to be built. Is federation a v2 goal or never?
3. **What's the SLA?** — if this is a passion project we accept downtime; if users rely on it, we need health checks, backups, and failover.

### Moderation & Trust
4. **Who appoints the first elders (threshold 5)?** — bootstrap problem. The first 100 users need to be carefully onboarded. Do we start with trusted alpha testers, or do we seed with threshold 3+ accounts and let them vote in the first elders?
5. **How do we handle edge cases where a trusted user goes rogue?** — the admin circuit breaker exists but is meant for emergencies only. What's the process for a threshold 4 user who abuses their timeout privilege? The jury system can suspend them, but who proposes it?
6. **How do we prevent vote manipulation of the jury system and community notes?** — the diversity requirement in community notes (cross-tag-cluster consensus) helps. For juries, the balancing by timezone and trust score helps. But a coordinated group could still infiltrate over months.

### Privacy & Legal
7. **Where will servers be located?** — GDPR (EU), CCPA (California), or both impose different data handling requirements. EU requires data deletion on request, right to explanation for algorithmic decisions.
8. **Is there any form of content that is always prohibited (CSAM, terrorist material)?** — if yes, we need automated detection + reporting. This is non-negotiable legally.
9. **How do we handle GDPR data export/deletion requests?** — need an automated mechanism before launch.
10. **What jurisdiction's laws govern the platform?** — affects everything from moderation liability to age requirements.

### Product & Growth
11. **What's the onboarding funnel?** — invite-only? Open registration with captcha? The vouch chain is great, but the first 1,000 users need to come from somewhere.
12. **How do we seed the initial content and communities?** — empty platform = no retention. Do we have "seed groups" with curated content? Do the first communities need curators appointed by the founding team?
13. **What's the expected split between free and paying users?** — if 1% of DAU buys credits, that may be enough to sustain the platform. But we need to model unit economics: server costs per user vs. average credit spend per paying user.
14. **What languages/locales does the platform need to support at launch?** — affects full-text search config, UI strings, content filtering.
15. **Will community forking be used constructively or destructively?** — forks could split communities over legitimate disagreements, or be used to harass. Do we need a cooldown period between forks?

### Technical
16. **What's the backup strategy?** — documented above (WAL archiving, daily dumps, quarterly restore drills). But who pays for S3 storage? Who runs the restore drills?
17. **How do we handle image/video uploads?** — direct to local disk? S3-compatible storage (MinIO, Backblaze)? File size limits? Thumbnail generation?
18. **Do we need a mobile app, or is PWA sufficient?** — PWA can do push notifications, offline support, and home screen install. Native app gives better camera access and background sync.
19. **What's the test strategy?** — Go backend: integration tests with test database? Frontend: Playwright for E2E? Without tests, refactoring becomes impossible.
20. **Who is the developer?** — is this a solo project? A team? An AI agent building the whole thing? The stack and tooling choices may shift based on who's writing the code.
21. **How do we handle spam/SEO abuse?** — rate limiting, shared blocklist, captcha on registration, email verification, slow-boot on communities. But we need automated spam detection for the main feed.
22. **What metrics define success for v1?** — 1000 DAU? 100 posts/day? < 10% flagged content? Having clear goals prevents scope creep.
23. **Who maintains the shared blocklist?** — the community votes entries in via jury. But someone needs to be the initial curator of the blocklist. The admin account? That centralizes power.
24. **How do we handle RSS feed abuse?** — RSS endpoints are rate-limited, but someone could still scrape the entire platform via RSS. Do we need to limit full-content RSS to authenticated users?
