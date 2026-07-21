-- User Lists (private, collaborative, algorithmic)
CREATE TABLE IF NOT EXISTS user_lists (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    list_type   SMALLINT NOT NULL DEFAULT 0, -- 0=follow, 1=block
    visibility  SMALLINT NOT NULL DEFAULT 0, -- 0=private, 1=public, 2=shared (editable with invite)
    is_algorithmic BOOLEAN DEFAULT FALSE,
    criteria_json JSONB,                     -- rules for algorithmic lists
    scope       SMALLINT DEFAULT 0,          -- 0=global, 1=per-tag
    tag_id      BIGINT REFERENCES tags(id) ON DELETE SET NULL,
    refresh_interval TEXT DEFAULT '24h',
    last_refreshed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_lists_owner ON user_lists(owner_id);
CREATE INDEX IF NOT EXISTS idx_user_lists_type ON user_lists(list_type);
CREATE INDEX IF NOT EXISTS idx_user_lists_visibility ON user_lists(visibility);

CREATE TABLE IF NOT EXISTS list_members (
    id            BIGSERIAL PRIMARY KEY,
    list_id       BIGINT NOT NULL REFERENCES user_lists(id) ON DELETE CASCADE,
    target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by      BIGINT NOT NULL REFERENCES users(id),
    added_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(list_id, target_user_id)
);

CREATE INDEX IF NOT EXISTS idx_list_members_list ON list_members(list_id);
CREATE INDEX IF NOT EXISTS idx_list_members_target ON list_members(target_user_id);

CREATE TABLE IF NOT EXISTS list_subscriptions (
    id         BIGSERIAL PRIMARY KEY,
    list_id    BIGINT NOT NULL REFERENCES user_lists(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action     SMALLINT NOT NULL DEFAULT 0, -- 0=follow, 1=block
    active     BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(list_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_list_subs_user ON list_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_list_subs_list ON list_subscriptions(list_id);

CREATE TABLE IF NOT EXISTS list_collaborators (
    id          BIGSERIAL PRIMARY KEY,
    list_id     BIGINT NOT NULL REFERENCES user_lists(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        SMALLINT NOT NULL DEFAULT 0, -- 0=viewer, 1=editor
    invited_by  BIGINT NOT NULL REFERENCES users(id),
    accepted_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(list_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_list_collab_list ON list_collaborators(list_id);
CREATE INDEX IF NOT EXISTS idx_list_collab_user ON list_collaborators(user_id);

-- User Affinity
CREATE TABLE IF NOT EXISTS user_affinities (
    user_a_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    affinity_score REAL DEFAULT 0,
    recency_factor REAL DEFAULT 1.0,
    breakdown      JSONB,                    -- component scores (tag_overlap, co_community, reaction_agreement, trust_distance)
    computed_at    TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_a_id, user_b_id)
);

CREATE INDEX IF NOT EXISTS idx_affinity_a ON user_affinities(user_a_id);
CREATE INDEX IF NOT EXISTS idx_affinity_b ON user_affinities(user_b_id);
CREATE INDEX IF NOT EXISTS idx_affinity_score ON user_affinities(affinity_score DESC);

-- Feed Plugins (Custom Feeds Marketplace)
CREATE TABLE IF NOT EXISTS feed_plugins (
    id             BIGSERIAL PRIMARY KEY,
    name           VARCHAR(100) NOT NULL,
    description    TEXT DEFAULT '',
    author_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wasm_bytes     BYTEA,                     -- compiled WASM plugin binary
    wasm_sha256    VARCHAR(64),               -- integrity hash
    version        VARCHAR(20) DEFAULT '1.0.0',
    plugin_type    SMALLINT DEFAULT 0,        -- 0=rank, 1=filter, 2=hybrid
    price_credits  BIGINT DEFAULT 0,          -- 0 = free
    rating         REAL DEFAULT 0,
    install_count  BIGINT DEFAULT 0,
    enabled        BOOLEAN DEFAULT TRUE,
    reviewed       BOOLEAN DEFAULT FALSE,     -- reviewed by admin before publishing
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_plugins_author ON feed_plugins(author_id);
CREATE INDEX IF NOT EXISTS idx_feed_plugins_rating ON feed_plugins(rating DESC);
CREATE INDEX IF NOT EXISTS idx_feed_plugins_reviewed ON feed_plugins(reviewed);

CREATE TABLE IF NOT EXISTS feed_plugin_installs (
    id         BIGSERIAL PRIMARY KEY,
    plugin_id  BIGINT NOT NULL REFERENCES feed_plugins(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    config_json JSONB,
    enabled    BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plugin_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_plugin_installs_user ON feed_plugin_installs(user_id);

CREATE TABLE IF NOT EXISTS feed_plugin_reviews (
    id         BIGSERIAL PRIMARY KEY,
    plugin_id  BIGINT NOT NULL REFERENCES feed_plugins(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating     SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review     TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plugin_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_plugin_reviews_plugin ON feed_plugin_reviews(plugin_id);

-- User follows (for batch follow via ApplyListAction)
CREATE TABLE IF NOT EXISTS user_follows (
    id          BIGSERIAL PRIMARY KEY,
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (follower_id, followee_id)
);
CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_followee ON user_follows(followee_id);
