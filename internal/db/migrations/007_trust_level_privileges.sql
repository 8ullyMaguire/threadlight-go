-- Trust-Level Privilege Thresholds
-- Each privilege in the platform can require a minimum trust level.
-- All values are configurable via site_config so admins can tune
-- participation requirements for their community.

ALTER TABLE site_config ADD COLUMN IF NOT EXISTS min_trust_level_for_review_voting   SMALLINT DEFAULT 1;
ALTER TABLE site_config ADD COLUMN IF NOT EXISTS min_trust_level_for_community_create SMALLINT DEFAULT 0;
ALTER TABLE site_config ADD COLUMN IF NOT EXISTS min_trust_level_for_curator           SMALLINT DEFAULT 1;
