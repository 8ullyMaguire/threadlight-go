# Polaris — Database Schema Reference

## Core Tables (31 original + 10 new = 41 tables)

### Users & Auth
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `users` | id, username, email, password_hash, trust_score, actor_id, onboarding_stage, avatar_url, banner_url, bio_html, email_verified, email_verify_token, password_reset_token, theme, hide_read_posts, is_deleted, deleted_at | Extended user accounts with verified email, password reset, theme, avatar/banner, soft delete |
| `trust_connections` | truster_id, trustee_id, weight, signature | Directed weighted edges for web of trust |
| `saved_searches` | user_id, name, query_json, notify_on_new | Saved search queries |

### Content
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `posts` | id, author_id, title, body, mood, content_warning, status, scheduled_at, edited_at, locked, sticky, sticky_at, language, is_ai_generated, license, cross_post_root_id, moved_from_community_id | Full post model with scheduling, editing, locking, stickying, cross-posting, AI flag, license |
| `tags` | id, name, category, is_wiki, created_by | Collaborative tags |
| `post_tags` | post_id, tag_id, tagged_by | M:N tagging with attribution |
| `interactions` | user_id, post_id, interaction_type (0=like,1=comment,2=share,3=gift,4=emoji), metadata | Supports emoji reactions |
| `feed_items` | user_id, post_id, score, seen, reason | Materialized feed per user |
| `custom_feeds` | owner_id, name, slug, is_public, sort_order | User-created weighted feeds |
| `feed_sources` | feed_id, source_type, include_mode | Sources within a custom feed |

### Collections
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `collections` | owner_id, name, visibility, is_default | Bookmark collections, 4 visibility layers |
| `collection_posts` | collection_id, post_id, added_by, notes | Posts within a collection |

### Communities & Circles
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `communities` | id, name, slug, tags[], curator_lock, slow_boot_days, forked_from | Public topic spaces with scope enforcement |
| `community_members` | community_id, user_id, role, status (0=probation,1=active) | Membership with slow-boot |
| `curators` | community_id, user_id, permission | Scope enforcers |
| `community_forks` | source_id, fork_id, reason, initiated_by | Fork history |
| `circles` | id, name, tag_id, grid_cell, member_count | Auto-generated interest+proximity groups |
| `circle_members` | circle_id, user_id, status (0=pending,1=accepted,2=dismissed) | Membership with suggestion state |

### Moderation
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `moderation_actions` | action_type, target_user_id, reason, is_jury_decision | Public append-only mod log |
| `jury_panels` | target_action_id, juror_id, vote | Jury votes |
| `blocklist_entries` | entry_type, shared | Shared blocklist for federation |
| `community_notes` | post_id, author_id, body, status, consensus_score | Misinformation warnings |
| `community_note_votes` | note_id, user_id, vote, trust_score_at_vote | Weighted consensus voting |

### Economy
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `credit_transactions` | from_user, to_user, amount, transaction_type, hash | Cryptographic ledger |
| `daily_rewards` | user_id, date, amount, claimed | Daily rewards with streak |
| `bounties` | post_id, creator_id, total_amount, best_answer_id | Question bounties |

### Discovery & Analytics
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `trending_topics` | topic, frequency, velocity, detection_window | Auto-detected trends |
| `active_stats` | scope_type, scope_id, active_1d/7d/30d | Fast community browsing sort |
| `content_filters` | user_id, filter_type, filter_action | Per-user content filters |
| `proximity_interests` | user_id, grid_cell, tag_id | Location+interest for circles |

### Gamification
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `achievements` | code, name, category, criteria_json | Achievement definitions |
| `user_achievements` | user_id, achievement_id, progress, visible | User unlock state |

### Post Moderation & User Controls (NEW)
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `blocked_users` | blocker_id, blocked_id | User blocking (suppress content + notifications) |
| `hidden_posts` | user_id, post_id | Per-user post hiding |
| `mod_notes` | user_id, noted_by, note | Private mod notes on users |
| `post_reports` | post_id, reporter_id, category, status, resolved_by | Structured content reports |
| `user_notifications` | user_id, notification_type, actor_id, post_id, body, is_read | In-app notification history |

### Infrastructure & Admin (NEW)
| Table | Key Columns | Purpose |
|-------|-------------|---------|
| `oauth_states` | provider, state, user_id, expires_at | OAuth flow state tracking |
| `registration_queue` | username, email, status, reviewed_by | Admin registration approval |
| `custom_pages` | slug, title, body, is_published | About, Terms, Privacy pages |
| `custom_emoji` | shortcode, image_url, alt_text | Custom emoji for instance |
| `email_queue` | to_email, subject, body, status | Outbound email queue |

## Key Indexes
- `idx_posts_fts` — GIN on `to_tsvector('english', body)`
- `idx_posts_scheduled` — partial on `(status, scheduled_at) WHERE status = 2`
- `idx_active_stats_7d` — for community sort by weekly active
- `idx_trending_window` — GiST on `detection_window`
- `idx_notifications_user` — on `(user_id, is_read)` for notification queries
- `idx_blocked_blocker/blocked` — for block checks
- `idx_reports_status` — for report queue queries

## Migration
`internal/db/migrations/001_initial.sql` — single migration, idempotent, 410+ lines.
