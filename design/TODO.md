# Threadlight — TODO

Priority tiers. Tier 1 = ship-blocking. Tier 5 = nice-to-have.

---

## Tier 1 — Core Experience (Ship Without These and People Complain)

- [x] Email/password registration + login
- [x] User profiles (display name, bio, avatar)
- [x] Text posts with content warnings + scheduling
- [x] Basic chronological feed
- [x] Trust levels 0-2 (model + database)
- [x] Basic reactions (like/unlike)
- [x] Communities (create/join/leave)
- [x] Search (PostgreSQL full-text with Boolean operators, field filters, trigram autocomplete)
- [x] Email verification
- [x] Password reset
- [x] User blocking + shared blocklist
- [x] Account deletion
- [x] Content filters (per-user keyword filters)
- [x] Feed pagination
- [x] Responsive layout (CSS variables, media queries)
- [x] Tag voting UI (frontend TagVote component)
- [ ] Trust Level auto-progression (cron job that promotes users based on time/activity/flags)
- [ ] Email-sent password reset (currently returns token directly — needs SMTP config + email sending)
- [ ] Mod post deletion (MODREMOVE endpoint exists, needs frontend button)

## Tier 2 — Community Health

- [x] Moderation labels (M1)
- [x] Admin ban/suspend
- [x] Reports + review queue
- [x] Public moderation decision reviews (Fair/Unfair voting + controversy scoring)
- [ ] Meta-moderation (M2) — jury-on-moderation system
- [ ] Community curators (scope tag enforcement without ban power)
- [ ] Community tenure-weighted voting
- [ ] Community probation for new members
- [ ] Public moderation log endpoint
- [ ] Mod dashboard (frontend queue)
- [ ] Appeal system

## Tier 3 — Discovery & Curation

- [x] Tags (CRUD, associate with posts)
- [x] Collaborative tagging (anyone can tag, post_tags table)
- [x] Tag voting API + frontend (vote up/down)
- [x] Search (boolean, fielded, date range, facets, autocomplete)
- [x] Trending (period-based aggregation)
- [x] Collections (create, add/remove posts)
- [x] Circles (manual groups)
- [x] User lists (CRUD, collaborative, algorithmic)
- [x] User affinity engine (event-driven Redis + on-demand cache)
- [x] Feed plugins (WASM marketplace sandbox)
- [ ] The Positive Algorithm (non-chronological feed ranker)
- [ ] "Why this post?" button
- [ ] Custom feeds with RSS export
- [ ] Full reaction set (Insightful, Helpful, Beautiful, etc.)
- [ ] Tag autocomplete (frontend)
- [ ] Collection visibility layers (public read, public edit)
- [ ] Auto-generated circles (by interest + proximity)

## Tier 4 — Gamification & Economy

- [x] Credits (award, spend, getBalance, getHistory)
- [x] Achievements (check + award)
- [x] Leaderboard API
- [ ] Credit streak bonuses
- [ ] Giftable credits (transfer between users)
- [ ] Credit hash-chain ledger (tamper-evident transaction log)
- [ ] Leaderboard frontend (currently uses sample data)
- [ ] Achievements frontend (show badges on profile)

## Tier 5 — Federation & Scale

- [x] Shared blocklist (JSON import/export)
- [x] Dockerfile (multi-stage build)
- [x] Migration runner (runs on startup, 4 migrations)
- [ ] ActivityPub (ed25519 keys, actor_id, inbox/outbox, HTTP Signatures)
- [ ] WebSocket (live feed, chat, presence)
- [ ] Image upload (model exists, service exists, needs endpoint + CDN/storage config)
- [ ] Seed data (sample users, posts, communities, tags for fresh instances)
- [ ] Test coverage (currently ~50% on backend, <10% on frontend)
- [ ] CI/CD pipeline (Forgejo Actions)

## Recently Completed

| Item | Status | Notes |
|------|--------|-------|
| Tag voting frontend | ✅ | TagVote.svelte with upvote/downvote, integrated into PostMeta |
| User lists | ✅ | CRUD, members, subscriptions, collaborators, batch follow/block |
| Algorithmic lists | ✅ | Criteria-based auto-population, refresh worker |
| User affinity | ✅ | Event-driven Redis counters (badge) + on-demand cache (page) |
| Feed plugins | ✅ | WASM marketplace, install, sandbox execution, reviews |
| Mod decision reviews | ✅ | Fair/Unfair voting, controversy scoring, /modlog/controversial |
| Search redesign | ✅ | Boolean parser, tsvector, trigram autocomplete, facets |
| Handler integration tests | ✅ | 4 new handler tests for lists, affinity |
