-- Mod Decision Reviews
-- Every logged-in user can vote "Fair" or "Unfair" on moderation actions.
-- A controversy score surfaces the most debated decisions.
CREATE TABLE IF NOT EXISTS mod_decision_reviews (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    moderation_action_id BIGINT NOT NULL REFERENCES moderation_actions(id) ON DELETE CASCADE,
    vote                SMALLINT NOT NULL CHECK (vote IN (1, -1)), -- 1=Fair, -1=Unfair
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, moderation_action_id)
);

CREATE INDEX IF NOT EXISTS idx_mod_reviews_action ON mod_decision_reviews(moderation_action_id);
CREATE INDEX IF NOT EXISTS idx_mod_reviews_user ON mod_decision_reviews(user_id);
