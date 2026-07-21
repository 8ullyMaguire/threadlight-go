-- Trust Penalty for Unfair Moderation Actions
-- When a moderation action receives enough "unfair" votes from the community,
-- the moderator who took the action loses trust. All parameters are
-- configurable via site_config so each community can set its own tolerance.

ALTER TABLE moderation_actions ADD COLUMN IF NOT EXISTS trust_penalty_applied BOOLEAN DEFAULT FALSE;

ALTER TABLE site_config ADD COLUMN IF NOT EXISTS unfair_threshold_pct      REAL    DEFAULT 0.70;
ALTER TABLE site_config ADD COLUMN IF NOT EXISTS unfair_penalty_amount     REAL    DEFAULT 15.0;
ALTER TABLE site_config ADD COLUMN IF NOT EXISTS unfair_min_reviews        INTEGER DEFAULT 5;
ALTER TABLE site_config ADD COLUMN IF NOT EXISTS unfair_penalty_cooldown_hrs INTEGER DEFAULT 24;
