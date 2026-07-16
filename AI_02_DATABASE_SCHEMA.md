# Polaris — Database Schema Reference

## Connection
PostgreSQL 17, accessed via pgx v5 pool. Redis 7 for caching, pub/sub, rate limiting.

## Core Tables (31 total)

### Users & Auth
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `users` | id, username, email, password_hash, trust_score, actor_id, onboarding_stage | User accounts with trust system and ActivityPub-ready actor_id |
| `trust_connections` | truster_id, trustee_id, weight, signature | Directed weighted edges for web of trust |
| `saved_searches` | user_id, name, query_json, notify_on_new | Saved search queries with optional notifications |

### Content
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `posts` | id, author_id, title, body, mood, content_warning, status, scheduled_at | Content items with scheduling + content warnings |
| `tags` | id, name, category, is_wiki, created_by | Collaborative tags (booru-style) |
| `post_tags` | post_id, tag_id, tagged_by | M:N tagging with attribution (multiple users can tag) |
| `interactions` | user_id, post_id, interaction_type (0=like,1=comment,2=share,3=gift), metadata | Likes, comments, shares, gifts |
| `feed_items` | user_id, post_id, score, seen, reason | Materialized feed per user (partitioned by hash) |
| `custom_feeds` | owner_id, name, slug, is_public, sort_order | User-created weighted feeds |
| `feed_sources` | feed_id, source_type (0=tag,1=user,2=keyword,3=list), include_mode | Sources within a custom feed |

### Collections
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `collections` | owner_id, name, visibility (0=private,1=public_read,2=public_edit,3=anybody_edit), is_default | Bookmark collections with 4 visibility layers |
| `collection_posts` | collection_id, post_id, added_by, notes | Posts within a collection |

### Communities & Circles
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `communities` | id, name, slug, tags[], curator_lock, slow_boot_days, forked_from | Public topic spaces with scope enforcement |
| `community_members` | community_id, user_id, role, status (0=probation,1=active) | Membership with slow-boot tracking |
| `curators` | community_id, user_id, permission (1=curator,2=admin) | Community scope enforcers |
| `community_forks` | source_id, fork_id, reason, initiated_by | Community fork history |
| `circles` | id, name, tag_id, grid_cell, member_count | Auto-generated interest+proximity groups |
| `circle_members` | circle_id, user_id, status (0=pending,1=accepted,2=dismissed) | Membership with suggestion state |

### Moderation
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `moderation_actions` | action_type, target_user_id, reason, is_jury_decision, jury_yes/no/total | Public append-only moderation log |
| `jury_panels` | target_action_id, juror_id, vote | Jury votes (random selection, balanced) |
| `blocklist_entries` | entry_type (0=domain,1=ip,2=key,3=fingerprint), shared | Shared blocklist for federation |
| `community_notes` | post_id, author_id, body, status, consensus_score | Misinformation warnings (Twitter/X-style) |
| `community_note_votes` | note_id, user_id, vote, trust_score_at_vote | Weighted consensus voting |

### Economy
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `credit_transactions` | from_user, to_user, amount, transaction_type, hash | Cryptographic ledger (SHA-256 chain) |
| `daily_rewards` | user_id, date, amount, claimed | Daily credit rewards with streak tracking |
| `bounties` | post_id, creator_id, total_amount, best_answer_id | Question bounties |

### Discovery & Analytics
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `trending_topics` | topic, frequency, velocity, detection_window (TSTZRANGE) | Auto-detected trends (not hashtags) |
| `active_stats` | scope_type (0=circle,1=community,2=tag,3=global), scope_id, active_1d/7d/30d | Fast community browsing sort |
| `content_filters` | user_id, filter_type (0=keyword,1=regex,2=domain,3=user), filter_action (0=hide,1=blur) | Per-user content filters |
| `proximity_interests` | user_id, grid_cell (geohash), tag_id | Privacy-safe location+interest for circle generation |

### Gamification
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `achievements` | code, name, category, criteria_json | Achievement definitions |
| `user_achievements` | user_id, achievement_id, progress, visible | User unlock state |

## Key Indexes
- `idx_posts_fts` — GIN on `to_tsvector('english', body)` for full-text search
- `idx_posts_scheduled` — partial index on `(status, scheduled_at) WHERE status = 2`
- `idx_active_stats_7d` — for community sort by weekly active users
- `idx_trending_window` — GiST on `detection_window` for time-range queries
- `idx_post_tags_tag/post` — fast tag joins for feed generation

## Migration File
`internal/db/migrations/001_initial.sql` — single migration with all tables + indexes, idempotent (IF NOT EXISTS), wrapped in BEGIN/COMMIT.
