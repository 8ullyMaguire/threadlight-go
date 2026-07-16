# Polaris — Feature TODO

Features from PieFed that suit Polaris, organized by priority.

---

## Tier 1 — Foundational (next 5 iterations)

### Auth & Accounts
- [ ] **Email verification** — token-link verification on registration; block unverified accounts from posting
- [ ] **Password reset** — email token flow with expiry
- [ ] **User blocking** — dedicated `POST /api/v1/users/:id/block` + `GET /api/v1/blocks` that fully hides blocked user's content from feed, search, notifications; separate from content filters (keyword/domain)
- [ ] **Account deletion** — soft delete with data retention window; `DELETE /api/v1/account` with confirmation flow

### Posts
- [ ] **Post editing** — editable within 15-minute window, tracked via `edited_at` column in posts table; PATCH endpoint already exists but needs time-window enforcement
- [ ] **Post locking (lock comments)** — community mods can disable new comments on a post; `locked` boolean on posts
- [ ] **Post hiding** — per-user hide without blocking (different from filters); `hidden_posts` M:N table
- [ ] **Post stickying** — community-level pin (mods) and instance-wide pin (admins); `sticky` boolean + `sticky_at` timestamp on posts

### Profile
- [ ] **Avatar upload** — auto-thumbnails at 32px, 64px, 128px, 256px; store to disk/S3
- [ ] **Banner image upload** — 1500x500 cover image
- [ ] **Bio with Markdown** — render markdown to HTML safely (use a Go markdown lib)
- [ ] **User stats on profile** — post count, reply count, likes received, days active

---

## Tier 2 — Growth (iterations 6–10)

### Content
- [ ] **Cross-posting** — same post body to multiple communities with aggregated reply count; `cross_post_root_id` on posts
- [ ] **Post language** — single language per post (ISO 639-1); filter feeds by language
- [ ] **Post moving** — mods can move a post from one community to another
- [ ] **OAuth providers** — Google OAuth, Discord OAuth as alternative login; requires provider config in admin panel
- [ ] **"Mea culpa"** — author can mark own post as mistaken/retracted; shows banner on post

### Moderation
- [ ] **Mod notes** — private per-user notes only visible to moderators; `mod_notes` table (user_id, note, created_by)
- [ ] **Report categories** — structured flag reasons (spam, harassment, misinformation, illegal, other) instead of free-text
- [ ] **Report queue** — dedicated page in ModCenter showing flagged posts grouped by severity
- [ ] **Same-IP detection** — flag accounts sharing an IP with a banned user for mod review
- [ ] **Ban evasion protection** — rate-limit registration from banned IP ranges

### UI/UX
- [ ] **Read post tracking** — optional setting to hide already-seen posts from feed; uses `feed_items.seen` flag
- [ ] **Post embed (oEmbed)** — `GET /api/v1/oembed?url=` returns embed JSON; add `<link>` tags for oEmbed discovery
- [ ] **Notifications page** — history of all notifications (likes, replies, flags, invites) — currently only WebSocket push
- [ ] **Light/dark theme** — CSS variables already support it; add theme toggle in settings that sets `data-theme` on `<html>`
- [ ] **Community application process** — written applications for joining curator-locked communities (beyond simple approval)

---

## Tier 3 — Ecosystem (iterations 11–15)

### Economy & Gamification
- [ ] **Reputation leaderboard** — extend existing leaderboard with reputation decay and rank changes over time
- [ ] **Emoji reactions** — extend interactions table to support emoji type (👍 ❤️ 😄 😢 😡) instead of just "like"
- [ ] **User titles** — earned titles based on achievements/trust level displayed below username on profile

### Content Curation
- [ ] **Post archiving (freeze)** — replace post body with archive.org snapshot link
- [ ] **Creative Commons licensing** — `license` field on posts (CC0, CC-BY, CC-BY-SA, etc.)
- [ ] **AI-generated content flag** — `is_ai_generated` boolean on posts; filter feed to show/hide

### Community
- [ ] **Mod mailing list** — private mod-only communication within a community
- [ ] **Mod wiki** — shared documentation page editable by community curators
- [ ] **Community stats page** — growth chart, top contributors, active hours heatmap

### Federation
- [ ] **WebFinger endpoint** — `GET /.well-known/webfinger?resource=acct:user@domain` for ActivityPub discovery
- [ ] **ActivityPub actor page** — `GET /api/v1/users/:id` serves `application/activity+json` with actor JSON
- [ ] **HTTP Signatures** — verify incoming ActivityPub requests using ed25519 keys
- [ ] **Outbox/Inbox endpoints** — minimal ActivityPub outbox (public posts) and inbox (receive follows/likes)

---

## Tier 4 — Platform Maturity (iterations 16–20)

### Hosting & Admin
- [ ] **Custom pages** — editable About, Terms of Service, Privacy Policy pages served at `/about`, `/terms`, `/privacy`
- [ ] **Admin dashboard enhancements** — active users chart, content growth, moderation stats, storage usage
- [ ] **Custom emoji** — upload and manage custom emoji per instance; replace `:emoji:` in post bodies
- [ ] **Security headers** — CSP, HSTS, X-Content-Type-Options, Referrer-Policy via middleware
- [ ] **Registration modes** — Open, Invite-only, Approval Required, Applications (admin review)
- [ ] **Bot challenge** — challenge-response (simple quiz or CAPTCHA) for suspicious signups
- [ ] **Account waiting queue** — admin dashboard for reviewing pending registrations

### Notification & Communication
- [ ] **Email notifications** — configurable: new reply, like, community invite, moderation action; queue via Redis + worker
- [ ] **Email digests** — weekly summary of top posts from subscribed communities and trust network
- [ ] **Web push notifications** — service worker push via VAPID (browser push API), complementing WebSocket

### Search & Discovery
- [ ] **Related posts** — tag-based "Similar Posts" widget at bottom of post detail page
- [ ] **Trending communities** — surface growing communities on the browse page
- [ ] **Interest graph recommendations** — "Users who liked this also like..." based on tag overlap

---

## Currently Incomplete (needs finish work)

These are already partially built in the codebase but need completion:

- [ ] **Worker system** — `internal/worker/scheduler.go` missing; workers for feed rebuild, trust decay, scheduled posts, community notes consensus, active stats, pruning all need writing
- [ ] **CORS middleware** — Gin router needs `gin-contrib/cors` for frontend dev on `:5173`
- [ ] **Vite proxy** — `web/vite.config.js` needs `server.proxy` to forward `/api` to `:8080`
- [ ] **Production Dockerfile** — multi-stage build for Go binary + static frontend assets
- [ ] **Image upload** — `POST /api/v1/uploads` with thumbnail generation; storage to disk or S3
- [ ] **ActivityPub stubs** — WebFinger, outbox, inbox endpoints exist in schema but not in router
- [ ] **Password reset implementation** — endpoint exists in design but no handler/service yet
- [ ] **Seed data** — no seed script for development/demo (sample posts, communities, users)
- [ ] **Test coverage** — only 14 integration tests; needs unit tests for services, E2E tests
- [ ] **Polls** — schema has `content_type = 4` for polls but no poll-specific endpoints or logic
