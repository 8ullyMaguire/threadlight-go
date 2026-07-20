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
- [x] Search (PostgreSQL full-text)
- [x] Email verification
- [x] Password reset
- [x] User blocking + shared blocklist
- [x] Account deletion
- [x] Content filters (per-user keyword filters)
- [x] Feed pagination
- [x] Responsive layout (CSS variables, media queries)
- [ ] Trust Level auto-progression (cron job that promotes users based on time/activity/flags)
- [ ] Email-sent password reset (currently returns token directly — needs SMTP config + email sending)
- [ ] Mod post deletion (MODREMOVE endpoint exists, needs frontend button)
- [ ] Collaborative tag voting (tag_votes table + API exist, needs frontend UI)

## Tier 2 — Community Health

- [x] Moderation labels (M1)
- [x] Admin ban/suspend
- [x] Reports + review queue
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
- [x] Tag voting API (vote up/down)
- [x] Advanced search (boolean, fielded, date range)
- [x] Trending (period-based aggregation)
- [x] Collections (create, add/remove posts)
- [x] Circles (manual groups)
- [ ] The Positive Algorithm (non-chronological feed ranker)
- [ ] "Why this post?" button
- [ ] Custom feeds with RSS export
- [ ] Full reaction set (Insightful, Helpful, Beautiful, etc.)
- [ ] Tag search with autocomplete
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
- [x] Migration runner (runs on startup)
- [ ] ActivityPub (ed25519 keys, actor_id, inbox/outbox, HTTP Signatures)
- [ ] WebSocket (live feed, chat, presence)
- [ ] Image upload (model exists, service exists, needs endpoint + CDN/storage config)
- [ ] Seed data (sample users, posts, communities, tags for fresh instances)
- [ ] Test coverage (currently ~50% on backend, <10% on frontend)
- [ ] CI/CD pipeline (Forgejo Actions)

## Currently Incomplete (Needs Finish Work)

| Item | Status | What's Missing |
|------|--------|----------------|
| Onboarding wizard | BACKEND ONLY | All 3 onboarding API endpoints exist? Suggest active posters from tag picks + recommend circles |
| Password reset | DONE | Features implemented. Needs SMTP to actually send the email. |
| Admin seeding | DONE | `seedAdmin` creates admin user on first start if env vars set. |
| Worker system | DONE | 11 workers running on scheduler. |
| CORS | DONE | Middleware exists in `internal/api/middleware/cors.go`. |
| Vite proxy | DONE | Configured in `vite.config.js` for dev. |
| Dockerfile | DONE | Multi-stage build at repo root. |
| Advanced search | DONE | Boolean + fielded operators, date range filtering. |
| Tag voting | DONE | `tag_votes` table + API. Needs frontend UI. |
| Mod post removal | DONE | `POST /posts/:id/remove` endpoint + handler. Needs frontend UI. |
| Migration runner | DONE | `internal/db/migrations.go` runs `.sql` files on startup. |
