BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id               BIGSERIAL PRIMARY KEY,
    username         VARCHAR(30) UNIQUE NOT NULL,
    display_name     VARCHAR(100),
    bio              TEXT,
    email            VARCHAR(255) UNIQUE NOT NULL,
    password_hash    VARCHAR(255) NOT NULL,
    trust_level      SMALLINT DEFAULT 0,
    trust_score      REAL DEFAULT 0,
    reputation       BIGINT DEFAULT 0,
    invited_by       BIGINT REFERENCES users(id),
    invite_code      VARCHAR(32) UNIQUE,
    credits          BIGINT DEFAULT 0,
    is_active        BOOLEAN DEFAULT TRUE,
    last_active_at   TIMESTAMPTZ,
    public_key       TEXT,
    actor_id         VARCHAR(255) UNIQUE,
    is_local         BOOLEAN DEFAULT TRUE,
    onboarding_stage SMALLINT DEFAULT 0,
    proximity_opt_out BOOLEAN DEFAULT FALSE,
    location_hash    VARCHAR(64),
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS posts (
    id                      BIGSERIAL PRIMARY KEY,
    author_id               BIGINT NOT NULL REFERENCES users(id),
    title                   VARCHAR(200),
    body                    TEXT,
    content_type            SMALLINT DEFAULT 0,
    mood                    SMALLINT DEFAULT 0,
    is_educational          BOOLEAN DEFAULT FALSE,
    is_entertaining         BOOLEAN DEFAULT FALSE,
    is_nsfw                 BOOLEAN DEFAULT FALSE,
    content_warning         TEXT,
    interaction_count       BIGINT DEFAULT 0,
    cumulative_interactions BIGINT DEFAULT 0,
    status                  SMALLINT DEFAULT 0,
    scheduled_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ,
    archived_at             TIMESTAMPTZ,
    is_deleted              BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category    VARCHAR(50),
    is_wiki     BOOLEAN DEFAULT FALSE,
    created_by  BIGINT REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post_tags (
    post_id    BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id     INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    tagged_by  BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (post_id, tag_id, tagged_by)
);

CREATE TABLE IF NOT EXISTS interactions (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id),
    post_id          BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    interaction_type SMALLINT NOT NULL,
    metadata         JSONB,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, post_id, interaction_type)
);

CREATE TABLE IF NOT EXISTS trust_connections (
    id         BIGSERIAL PRIMARY KEY,
    truster_id BIGINT NOT NULL REFERENCES users(id),
    trustee_id BIGINT NOT NULL REFERENCES users(id),
    weight     REAL DEFAULT 1.0,
    signature  TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    UNIQUE (truster_id, trustee_id)
);

CREATE TABLE IF NOT EXISTS moderation_actions (
    id              BIGSERIAL PRIMARY KEY,
    action_type     SMALLINT NOT NULL,
    target_user_id  BIGINT REFERENCES users(id),
    target_post_id  BIGINT REFERENCES posts(id),
    moderator_id    BIGINT REFERENCES users(id),
    reason          TEXT NOT NULL,
    duration        INTERVAL,
    is_jury_decision BOOLEAN DEFAULT FALSE,
    jury_yes        INTEGER DEFAULT 0,
    jury_no         INTEGER DEFAULT 0,
    jury_total      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS jury_panels (
    id               BIGSERIAL PRIMARY KEY,
    target_action_id BIGINT REFERENCES moderation_actions(id),
    juror_id         BIGINT REFERENCES users(id),
    vote             BOOLEAN,
    reason           TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (target_action_id, juror_id)
);

CREATE TABLE IF NOT EXISTS circles (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    description   TEXT,
    tag_id        INTEGER REFERENCES tags(id),
    grid_cell     VARCHAR(20) NOT NULL,
    member_count  INTEGER DEFAULT 0,
    is_active     BOOLEAN DEFAULT TRUE,
    last_activity TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS circle_members (
    circle_id    BIGINT NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       SMALLINT DEFAULT 0,
    suggested_at TIMESTAMPTZ DEFAULT NOW(),
    joined_at    TIMESTAMPTZ,
    PRIMARY KEY (circle_id, user_id)
);

CREATE TABLE IF NOT EXISTS communities (
    id             BIGSERIAL PRIMARY KEY,
    name           VARCHAR(100) NOT NULL,
    description    TEXT,
    slug           VARCHAR(100) UNIQUE NOT NULL,
    tags           TEXT[] DEFAULT '{}',
    curator_lock   BOOLEAN DEFAULT FALSE,
    slow_boot_days INTEGER DEFAULT 0,
    forked_from    BIGINT REFERENCES communities(id),
    created_by     BIGINT NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ,
    archived_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS community_members (
    community_id BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         SMALLINT DEFAULT 0,
    status       SMALLINT DEFAULT 0,
    joined_at    TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (community_id, user_id)
);

CREATE TABLE IF NOT EXISTS curators (
    id           BIGSERIAL PRIMARY KEY,
    community_id BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission   SMALLINT DEFAULT 1,
    appointed_by BIGINT NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (community_id, user_id)
);

CREATE TABLE IF NOT EXISTS community_forks (
    id            BIGSERIAL PRIMARY KEY,
    source_id     BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    fork_id       BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    reason        TEXT NOT NULL,
    initiated_by  BIGINT NOT NULL REFERENCES users(id),
    member_count  INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_notes (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id       BIGINT NOT NULL REFERENCES users(id),
    body            TEXT NOT NULL,
    status          SMALLINT DEFAULT 0,
    helpful_yes     INTEGER DEFAULT 0,
    helpful_no      INTEGER DEFAULT 0,
    consensus_score REAL DEFAULT 0,
    requires_author BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS community_note_votes (
    id                BIGSERIAL PRIMARY KEY,
    note_id           BIGINT NOT NULL REFERENCES community_notes(id) ON DELETE CASCADE,
    user_id           BIGINT NOT NULL REFERENCES users(id),
    vote              BOOLEAN NOT NULL,
    trust_score_at_vote REAL NOT NULL,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (note_id, user_id)
);

CREATE TABLE IF NOT EXISTS blocklist_entries (
    id            BIGSERIAL PRIMARY KEY,
    entry_type    SMALLINT NOT NULL,
    entry_value   VARCHAR(500) NOT NULL,
    reason        TEXT,
    severity      SMALLINT DEFAULT 0,
    added_by      BIGINT REFERENCES users(id),
    jury_approved BOOLEAN DEFAULT FALSE,
    shared        BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (entry_type, entry_value)
);

CREATE TABLE IF NOT EXISTS active_stats (
    id         BIGSERIAL PRIMARY KEY,
    scope_type SMALLINT NOT NULL,
    scope_id   BIGINT NOT NULL,
    active_1d  INTEGER DEFAULT 0,
    active_7d  INTEGER DEFAULT 0,
    active_30d INTEGER DEFAULT 0,
    computed_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (scope_type, scope_id, computed_at)
);

CREATE TABLE IF NOT EXISTS custom_feeds (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT NOT NULL REFERENCES users(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    slug        VARCHAR(100) NOT NULL,
    is_public   BOOLEAN DEFAULT FALSE,
    sort_order  SMALLINT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    UNIQUE (owner_id, slug)
);

CREATE TABLE IF NOT EXISTS feed_sources (
    id            BIGSERIAL PRIMARY KEY,
    feed_id       BIGINT NOT NULL REFERENCES custom_feeds(id) ON DELETE CASCADE,
    source_type   SMALLINT NOT NULL,
    source_id     BIGINT,
    source_value  VARCHAR(200),
    include_mode  BOOLEAN DEFAULT TRUE,
    sort_priority SMALLINT DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS collections (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT NOT NULL REFERENCES users(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    visibility  SMALLINT DEFAULT 0,
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS collection_posts (
    id            BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    post_id       BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    added_by      BIGINT NOT NULL REFERENCES users(id),
    notes         TEXT,
    sort_order    INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (collection_id, post_id)
);

CREATE TABLE IF NOT EXISTS content_filters (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id),
    filter_type   SMALLINT NOT NULL,
    filter_value  VARCHAR(500) NOT NULL,
    filter_action SMALLINT DEFAULT 0,
    is_active     BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, filter_type, filter_value)
);

CREATE TABLE IF NOT EXISTS saved_searches (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id),
    name          VARCHAR(100) NOT NULL,
    query_json    JSONB NOT NULL,
    notify_on_new BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trending_topics (
    id               BIGSERIAL PRIMARY KEY,
    topic            VARCHAR(200) NOT NULL,
    frequency        INTEGER DEFAULT 0,
    velocity         REAL DEFAULT 0,
    detection_window TSTZRANGE,
    tag_id           INTEGER REFERENCES tags(id),
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS achievements (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    icon        VARCHAR(200),
    category    SMALLINT DEFAULT 0,
    sort_order  SMALLINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_achievements (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id),
    achievement_id INTEGER NOT NULL REFERENCES achievements(id),
    unlocked_at    TIMESTAMPTZ DEFAULT NOW(),
    progress       REAL DEFAULT 0,
    visible        BOOLEAN DEFAULT TRUE,
    UNIQUE (user_id, achievement_id)
);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id               BIGSERIAL PRIMARY KEY,
    from_user        BIGINT REFERENCES users(id),
    to_user          BIGINT REFERENCES users(id),
    amount           BIGINT NOT NULL,
    transaction_type SMALLINT NOT NULL,
    reference_id     BIGINT,
    hash             VARCHAR(64),
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS daily_rewards (
    id       BIGSERIAL PRIMARY KEY,
    user_id  BIGINT NOT NULL REFERENCES users(id),
    date     DATE NOT NULL,
    amount   BIGINT NOT NULL,
    claimed  BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE TABLE IF NOT EXISTS bounties (
    id             BIGSERIAL PRIMARY KEY,
    post_id        BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    creator_id     BIGINT NOT NULL REFERENCES users(id),
    total_amount   BIGINT NOT NULL,
    status         SMALLINT DEFAULT 0,
    best_answer_id BIGINT REFERENCES interactions(id),
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feed_items (
    user_id    BIGINT NOT NULL REFERENCES users(id),
    post_id    BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    score      REAL NOT NULL,
    reason     TEXT,
    seen       BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE IF NOT EXISTS proximity_interests (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    grid_cell  VARCHAR(20) NOT NULL,
    tag_id     INTEGER NOT NULL REFERENCES tags(id),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, grid_cell, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_invite_code ON users(invite_code);
CREATE INDEX IF NOT EXISTS idx_posts_content_type_created ON posts(content_type, created_at);
CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts(author_id, created_at);
CREATE INDEX IF NOT EXISTS idx_posts_interaction_count ON posts(interaction_count);
CREATE INDEX IF NOT EXISTS idx_posts_scheduled ON posts(status, scheduled_at) WHERE status = 2;
CREATE INDEX IF NOT EXISTS idx_posts_fts ON posts USING GIN (to_tsvector('english', body));
CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_post_tags_post ON post_tags(post_id);
CREATE INDEX IF NOT EXISTS idx_feed_items_user_score ON feed_items(user_id, score);
CREATE INDEX IF NOT EXISTS idx_feed_items_created ON feed_items(created_at);
CREATE INDEX IF NOT EXISTS idx_mod_actions_target_user ON moderation_actions(target_user_id);
CREATE INDEX IF NOT EXISTS idx_mod_actions_created ON moderation_actions(created_at);
CREATE INDEX IF NOT EXISTS idx_trust_trustee ON trust_connections(trustee_id);
CREATE INDEX IF NOT EXISTS idx_trust_truster ON trust_connections(truster_id);
CREATE INDEX IF NOT EXISTS idx_proximity_grid_tag ON proximity_interests(grid_cell, tag_id);
CREATE INDEX IF NOT EXISTS idx_custom_feeds_owner ON custom_feeds(owner_id);
CREATE INDEX IF NOT EXISTS idx_feed_sources_feed ON feed_sources(feed_id);
CREATE INDEX IF NOT EXISTS idx_collections_owner ON collections(owner_id);
CREATE INDEX IF NOT EXISTS idx_collection_posts_collection ON collection_posts(collection_id);
CREATE INDEX IF NOT EXISTS idx_collection_posts_post ON collection_posts(post_id);
CREATE INDEX IF NOT EXISTS idx_content_filters_user ON content_filters(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_searches_user ON saved_searches(user_id);
CREATE INDEX IF NOT EXISTS idx_trending_window ON trending_topics USING GIST (detection_window);
CREATE INDEX IF NOT EXISTS idx_trending_frequency ON trending_topics(frequency);
CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_communities_slug ON communities(slug);
CREATE INDEX IF NOT EXISTS idx_communities_tags ON communities USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_community_members_community ON community_members(community_id);
CREATE INDEX IF NOT EXISTS idx_community_members_user ON community_members(user_id);
CREATE INDEX IF NOT EXISTS idx_community_notes_post ON community_notes(post_id);
CREATE INDEX IF NOT EXISTS idx_community_notes_status ON community_notes(status);
CREATE INDEX IF NOT EXISTS idx_note_votes_note ON community_note_votes(note_id);
CREATE INDEX IF NOT EXISTS idx_blocklist_type ON blocklist_entries(entry_type);
CREATE INDEX IF NOT EXISTS idx_active_stats_scope ON active_stats(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_active_stats_7d ON active_stats(active_7d);

COMMIT;
