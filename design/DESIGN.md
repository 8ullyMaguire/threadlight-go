# Polaris — Design

A social platform for positive connection: trust-based moderation, collaborative curation, and democratic self-governance. No outrage amplification.

---

## Philosophy

- **Positive addiction** — replace outrage with connection, algorithmic amplification with collaborative curation
- **Self-governing** — moderation power flows from trust, not central authority; bans require 90% jury consensus
- **Privacy-respecting** — proximity uses geohashes, not exact locations; no KYC or phone number required
- **Community-owned** — transparent mod log, public health dashboard, portable identities
- **Lightweight** — single Go binary serving thousands of users; no microservices, no Kubernetes

## Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Backend | Go 1.26 + Gin | Single binary, excellent concurrency (WebSocket), low cost VPS |
| Frontend | Svelte 4 + Vite | Smallest bundle size, minimal boilerplate, reactive |
| Database | PostgreSQL 17 | JSONB, full-text search (GIN), recursive CTEs, pgvector |
| Cache | Redis 7 | Pub/sub for real-time, sorted sets for feeds, rate limiting |
| Workers | Go goroutines | 11 background workers, ticker-based scheduling |

## Architecture

```
Client (Svelte SPA) → REST / WebSocket → Go / Gin API + WebSocket Hub
                                           ├── PostgreSQL (primary)
                                           ├── Redis (cache, pub/sub, queues)
                                           └── Workers (11 goroutines)
```

Single binary monolith + background goroutines. No microservices. Static files served separately via nginx in production.

## Core Systems

### Trust (score-driven, not level-gated)
- Single `trust_score` float per user computed from: reputation (0.5), incoming trust connections (0.3), time active (0.1), diversity of interactions (0.1)
- Thresholds gate privileges (timeout at 0.5, mod queue at 2.0, ban vote at 5.0) — but no fixed levels
- Trust decays 5%/week after 30 days of inactivity
- Vouch chain: inviter is penalized if invitee's content is flagged 3+ times in 90 days
- Web of trust: ed25519 signed edges, personalized graph ranking

### Moderation
- Soft actions (timeouts): threshold 3+ can issue. Duration: 24h (3+), 72h (4+). Costs 10% trust_score if same trust network.
- Hard actions (bans): threshold 4+ proposes. Jury of 16 (balanced by timezone + trust score). 60% for suspension, 90% for permanent ban. Votes are public after decision.
- Admin circuit breaker: single account (disabled by default), limited scope (can only reverse mod actions, not create content), all actions logged with "Admin Override" banner.
- Shared blocklist: domain/IP/key entries voted in by jury; exportable JSON for federation.

### Communities (public topic spaces)
- Discoverable, topic-based, with curators (scope enforcers, not moderators)
- Slow-boot probation (0-14 days): new members' posts held for review
- Scope enforcement: posts must use at least one of the community's declared tags
- Forking: if members disagree, they can fork — copies member list (each confirms within 7 days), resets curators

### Circles (auto-generated small groups)
- System-generated from `(tag_id, geohash)` overlap in `proximity_interests`
- Min 3 users, max 20 members. Named: `"{Tag} in {Neighborhood}"`
- Suggested to users; they accept or dismiss. Dedup: don't re-suggest declined circles for 30 days.

### Content Curation
- Booru-style collaborative tagging: any trusted user can add tags to any post
- Tags are NOT author-controlled — prevents hiding problematic content behind deceptive tags
- Tag conflict resolution: vote among users who've tagged > 10 posts
- Community Notes: Twitter/X-style consensus-driven misinformation warnings. Weighted voting with cross-tag diversity requirement to prevent mobbing.

### Feed Algorithm
```
score(p,u) = recency(0.3) + engagement_velocity(0.2) + tag_affinity(0.25) + trust_network(0.15) + proximity(0.05) + quality(0.05)
```
Diversity guarantee: max 3 posts from same tag cluster, max 2 from same author in first 20 results. Content pruned after 6 months with zero interactions.

### Economy
- Credits: daily rewards (10/day + streak bonus), purchased via Stripe (100 credits = $1), gifted between users
- 1% platform tax on all transfers (burned from circulation)
- Cash-out via Stripe Connect (min $100, KYC required, weekly limit)
- No real-money exchange for daily reward credits (non-withdrawable)

### ActivityPub Readiness
- `users.actor_id`, `users.public_key` (ed25519), `users.is_local` already in schema
- WebFinger, outbox/inbox endpoints stubbed
- Blocklist exportable as JSON for instance-to-instance sharing

## Database (41 tables)

Key tables: `users`, `posts`, `tags`, `post_tags`, `interactions`, `feed_items`, `custom_feeds`, `communities`, `community_members`, `curators`, `circles`, `circle_members`, `community_notes`, `moderation_actions`, `jury_panels`, `blocklist_entries`, `credit_transactions`, `achievements`, `blocked_users`, `post_reports`, `user_notifications`, `custom_pages`, `custom_emoji`.

Full schema in `AI_04_DB.md`. All migrations in `db/migrations/001_initial.sql` (idempotent, 410 lines).

## API (~95 endpoints)

All under `/api/v1`. Auth via `Authorization: Bearer <JWT>`. Full reference in `AI_03_API.md`.

Key groups: auth, posts, tags, feeds (algorithmic + custom), search, collections, communities, circles, notes (community notes), moderation (flag, timeout, jury, blocklist, reports), notifications, blocks, credits, bounties, achievements, leaderboards, onboarding, admin, emoji, pages.

## Frontend

Svelte 4 SPA with custom event routing (`navigate` events on `window`). No SvelteKit. Plain fetch API client. Vite proxies `/api` to backend `:8080`. Full reference in `AI_05_FRONTEND.md`.

## Key Design Decisions

- **No ORM** — raw SQL with pgx. Compile-time checked, full control.
- **No microservices** — monolith + goroutine workers. Simpler to develop, deploy, debug.
- **No real blockchain** — cryptographic ledger in PostgreSQL. Same auditability, no UX friction.
- **No SvelteKit** — pure SPA with custom routing. Smaller bundle, simpler build chain.
- **ActivityPub-ready but not federated yet** — data model supports it; endpoints are stubs.
- **No KYC for platform use** — only for credit cash-out. Proof of humanity via web of trust, not ID documents.

## Build Progress (~85% complete)

See `AI_05_BUILD_PLAN.md` for full iteration status.

Still needed: Dockerfile, image upload, ActivityPub endpoints, email delivery, WebSocket frontend client, seed data, expanded test coverage.
