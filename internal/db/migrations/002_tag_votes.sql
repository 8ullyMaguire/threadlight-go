-- Tag voting for collaborative curation
CREATE TABLE IF NOT EXISTS tag_votes (
    id BIGSERIAL PRIMARY KEY,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote SMALLINT NOT NULL DEFAULT 1 CHECK (vote IN (-1, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tag_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tag_votes_tag_id ON tag_votes(tag_id);
CREATE INDEX IF NOT EXISTS idx_tag_votes_user_id ON tag_votes(user_id);
