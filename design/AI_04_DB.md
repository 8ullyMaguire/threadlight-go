# DATABASE SCHEMA (41 tables, idempotent migration in db/migrations/001_initial.sql)

## Core
```sql
CREATE TABLE users (id BIGSERIAL PRIMARY KEY, username VARCHAR(30) UNIQUE NOT NULL, email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL, invite_code VARCHAR(32) UNIQUE, trust_score REAL DEFAULT 0, reputation BIGINT DEFAULT 0,
  credits BIGINT DEFAULT 0, avatar_url VARCHAR(500), banner_url VARCHAR(500), bio_html VARCHAR(2000),
  email_verified BOOLEAN DEFAULT FALSE, email_verify_token VARCHAR(64), password_reset_token VARCHAR(64),
  theme VARCHAR(20) DEFAULT 'light', hide_read_posts BOOLEAN DEFAULT FALSE, is_deleted BOOLEAN DEFAULT FALSE,
  actor_id VARCHAR(255) UNIQUE, is_local BOOLEAN DEFAULT TRUE, onboarding_stage SMALLINT DEFAULT 0,
  invited_by BIGINT REFERENCES users(id), location_hash VARCHAR(64), created_at TIMESTAMPTZ DEFAULT NOW());

CREATE TABLE posts (id BIGSERIAL PRIMARY KEY, author_id BIGINT NOT NULL REFERENCES users(id),
  title VARCHAR(200), body TEXT, content_type SMALLINT DEFAULT 0, mood SMALLINT DEFAULT 0, content_warning TEXT,
  interaction_count BIGINT DEFAULT 0, status SMALLINT DEFAULT 0, scheduled_at TIMESTAMPTZ, locked BOOLEAN DEFAULT FALSE,
  sticky BOOLEAN DEFAULT FALSE, sticky_at TIMESTAMPTZ, language VARCHAR(10) DEFAULT 'en',
  is_ai_generated BOOLEAN DEFAULT FALSE, license VARCHAR(20) DEFAULT 'CC0',
  cross_post_root_id BIGINT REFERENCES posts(id), moved_from_community_id BIGINT REFERENCES communities(id),
  edited_at TIMESTAMPTZ, is_deleted BOOLEAN DEFAULT FALSE, created_at TIMESTAMPTZ DEFAULT NOW());

CREATE TABLE tags (id SERIAL PRIMARY KEY, name VARCHAR(100) UNIQUE NOT NULL, category VARCHAR(50));
CREATE TABLE post_tags (post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
  tagged_by BIGINT REFERENCES users(id), PRIMARY KEY(post_id,tag_id,tagged_by));
CREATE TABLE interactions (user_id BIGINT NOT NULL, post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  interaction_type SMALLINT NOT NULL, metadata JSONB, UNIQUE(user_id,post_id,interaction_type));
CREATE TABLE feed_items (user_id BIGINT NOT NULL, post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  score REAL NOT NULL, seen BOOLEAN DEFAULT FALSE, created_at TIMESTAMPTZ, PRIMARY KEY(user_id,post_id));
```

## Communities
```sql
CREATE TABLE communities (id BIGSERIAL PRIMARY KEY, name VARCHAR(100) NOT NULL, slug VARCHAR(100) UNIQUE NOT NULL,
  tags TEXT[] DEFAULT '{}', curator_lock BOOLEAN DEFAULT FALSE, slow_boot_days INTEGER DEFAULT 0,
  forked_from BIGINT REFERENCES communities(id), created_by BIGINT NOT NULL REFERENCES users(id), archived_at TIMESTAMPTZ);
CREATE TABLE community_members (community_id BIGINT REFERENCES communities(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, role SMALLINT DEFAULT 0, status SMALLINT DEFAULT 0, PRIMARY KEY(community_id,user_id));
CREATE TABLE curators (id BIGSERIAL, community_id BIGINT REFERENCES communities(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, permission SMALLINT DEFAULT 1, UNIQUE(community_id,user_id));
CREATE TABLE community_forks (source_id BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  fork_id BIGINT NOT NULL REFERENCES communities(id) ON DELETE CASCADE, reason TEXT NOT NULL);
```

## Circles (auto-generated)
```sql
CREATE TABLE circles (id BIGSERIAL PRIMARY KEY, name VARCHAR(200), tag_id INTEGER REFERENCES tags(id),
  grid_cell VARCHAR(20) NOT NULL, member_count INTEGER DEFAULT 0, is_active BOOLEAN DEFAULT TRUE);
CREATE TABLE circle_members (circle_id BIGINT REFERENCES circles(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, status SMALLINT DEFAULT 0, PRIMARY KEY(circle_id,user_id));
```

## Moderation
```sql
CREATE TABLE moderation_actions (id BIGSERIAL PRIMARY KEY, action_type SMALLINT NOT NULL,
  target_user_id BIGINT REFERENCES users(id), moderator_id BIGINT REFERENCES users(id),
  reason TEXT NOT NULL, duration INTERVAL, is_jury_decision BOOLEAN DEFAULT FALSE,
  jury_yes INTEGER DEFAULT 0, jury_no INTEGER DEFAULT 0);
CREATE TABLE jury_panels (target_action_id BIGINT REFERENCES moderation_actions(id),
  juror_id BIGINT REFERENCES users(id), vote BOOLEAN, UNIQUE(target_action_id, juror_id));
CREATE TABLE blocklist_entries (entry_type SMALLINT NOT NULL, entry_value VARCHAR(500) NOT NULL,
  severity SMALLINT DEFAULT 0, shared BOOLEAN DEFAULT TRUE, UNIQUE(entry_type,entry_value));
CREATE TABLE community_notes (post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, author_id BIGINT REFERENCES users(id),
  body TEXT NOT NULL, status SMALLINT DEFAULT 0, helpful_yes INTEGER, helpful_no INTEGER, consensus_score REAL);
CREATE TABLE community_note_votes (note_id BIGINT REFERENCES community_notes(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id), vote BOOLEAN NOT NULL, trust_score_at_vote REAL NOT NULL);
```

## Economy
```sql
CREATE TABLE credit_transactions (id BIGSERIAL PRIMARY KEY, from_user BIGINT REFERENCES users(id),
  to_user BIGINT REFERENCES users(id), amount BIGINT NOT NULL, transaction_type SMALLINT NOT NULL, hash VARCHAR(64));
CREATE TABLE daily_rewards (user_id BIGINT REFERENCES users(id), date DATE NOT NULL, amount BIGINT NOT NULL, UNIQUE(user_id,date));
CREATE TABLE bounties (post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, creator_id BIGINT REFERENCES users(id),
  total_amount BIGINT NOT NULL, status SMALLINT DEFAULT 0, best_answer_id BIGINT REFERENCES interactions(id));
```

## Discovery
```sql
CREATE TABLE trending_topics (topic VARCHAR(200), frequency INTEGER, velocity REAL, detection_window TSTZRANGE);
CREATE TABLE active_stats (scope_type SMALLINT, scope_id BIGINT, active_1d INTEGER, active_7d INTEGER, active_30d INTEGER, UNIQUE(scope_type,scope_id,computed_at));
CREATE TABLE content_filters (user_id BIGINT REFERENCES users(id), filter_type SMALLINT, filter_value VARCHAR(500), UNIQUE(user_id,filter_type,filter_value));
CREATE TABLE proximity_interests (user_id BIGINT REFERENCES users(id), grid_cell VARCHAR(20), tag_id INTEGER REFERENCES tags(id), UNIQUE(user_id,grid_cell,tag_id));
```

## Gamification
```sql
CREATE TABLE achievements (code VARCHAR(50) UNIQUE NOT NULL, name VARCHAR(100), category SMALLINT DEFAULT 0);
CREATE TABLE user_achievements (user_id BIGINT REFERENCES users(id), achievement_id INTEGER REFERENCES achievements(id),
  progress REAL DEFAULT 0, visible BOOLEAN DEFAULT TRUE, UNIQUE(user_id,achievement_id));
```

## Collections + Custom Feeds
```sql
CREATE TABLE collections (owner_id BIGINT REFERENCES users(id), name VARCHAR(100) NOT NULL, visibility SMALLINT DEFAULT 0, is_default BOOLEAN DEFAULT FALSE);
CREATE TABLE collection_posts (collection_id BIGINT REFERENCES collections(id) ON DELETE CASCADE,
  post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, added_by BIGINT REFERENCES users(id), UNIQUE(collection_id,post_id));
CREATE TABLE custom_feeds (owner_id BIGINT REFERENCES users(id), name VARCHAR(100), slug VARCHAR(100), is_public BOOLEAN DEFAULT FALSE, UNIQUE(owner_id,slug));
CREATE TABLE feed_sources (feed_id BIGINT REFERENCES custom_feeds(id) ON DELETE CASCADE, source_type SMALLINT NOT NULL, include_mode BOOLEAN DEFAULT TRUE);
```

## User Control (NEW)
```sql
CREATE TABLE blocked_users (blocker_id BIGINT REFERENCES users(id), blocked_id BIGINT REFERENCES users(id), UNIQUE(blocker_id,blocked_id));
CREATE TABLE hidden_posts (user_id BIGINT REFERENCES users(id), post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, UNIQUE(user_id,post_id));
CREATE TABLE mod_notes (user_id BIGINT REFERENCES users(id), noted_by BIGINT REFERENCES users(id), note TEXT NOT NULL);
CREATE TABLE post_reports (post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, reporter_id BIGINT REFERENCES users(id),
  category SMALLINT DEFAULT 0, status SMALLINT DEFAULT 0, resolved_by BIGINT REFERENCES users(id));
CREATE TABLE user_notifications (user_id BIGINT REFERENCES users(id), notification_type SMALLINT NOT NULL,
  actor_id BIGINT REFERENCES users(id), post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE, body TEXT NOT NULL, is_read BOOLEAN DEFAULT FALSE);
```

## Infrastructure
```sql
CREATE TABLE oauth_states (provider VARCHAR(50), state VARCHAR(64) UNIQUE NOT NULL, user_id BIGINT REFERENCES users(id), expires_at TIMESTAMPTZ);
CREATE TABLE custom_pages (slug VARCHAR(100) UNIQUE NOT NULL, title VARCHAR(200), body TEXT, is_published BOOLEAN DEFAULT FALSE);
CREATE TABLE custom_emoji (shortcode VARCHAR(100) UNIQUE NOT NULL, image_url VARCHAR(500));
CREATE TABLE email_queue (to_email VARCHAR(255), subject VARCHAR(200), body TEXT, status SMALLINT DEFAULT 0);
CREATE TABLE saved_searches (user_id BIGINT REFERENCES users(id), name VARCHAR(100), query_json JSONB, notify_on_new BOOLEAN DEFAULT FALSE);
```

## Key Indexes
```sql
CREATE INDEX idx_posts_fts ON posts USING GIN(to_tsvector('english',body));
CREATE INDEX idx_posts_scheduled ON posts(status,scheduled_at) WHERE status=2;
CREATE INDEX idx_active_stats_7d ON active_stats(scope_type,scope_id,active_7d DESC);
CREATE INDEX idx_trending_window ON trending_topics USING GIST(detection_window);
CREATE INDEX idx_notifications_user ON user_notifications(user_id,is_read);
CREATE INDEX idx_blocked_blocker ON blocked_users(blocker_id);
CREATE INDEX idx_blocked_blocked ON blocked_users(blocked_id);
CREATE INDEX idx_reports_status ON post_reports(status);
```
