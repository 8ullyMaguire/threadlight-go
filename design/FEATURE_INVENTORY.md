# Threadlight — Complete Feature Inventory & Next Steps

Based on README, AGENTS.md, ARCHITECTURE.md, DESIGN.md, TODO.md, and all source code.

---

## 1. Implemented (core works, may need polish)

| Feature | Description |
|---------|------------|
| Auth (register/login/logout) | Email/password with JWT session management |
| User profiles | Display name, bio, avatar, banner, theme, privacy settings |
| Text posts | Create/update/delete with content warnings, scheduling, Markdown |
| Chronological feed | Basic feed ordered by time with pagination |
| Trust levels 0-2 | Model + DB for Discourse-style trust progression |
| Reactions | Like/unlike (basic), multi-type reactions model exists |
| Communities | Create/join/leave with slow-boot, probation, tenure-voting, fork |
| Search | Full-text with Boolean query parser (AND/OR/NOT/parentheses/phrases), field filters (`author:`, `tag:`, `community:`, `mood:`, `is:`, `before:`, `after:`), tsvector GIN index, trigram autocomplete, facet counts |
| Email verification | Verification email on signup |
| Password reset | Token-based (needs SMTP to actually send) |
| User blocking | Per-user block + shared blocklist JSON import/export |
| Account deletion | Full account removal |
| Content filters | Per-user keyword filters to hide posts |
| Tags CRUD | Create/read/update/delete tags |
| Collaborative tagging | Anyone can tag any post (post_tags table) |
| Tag voting API + frontend | Upvote/downvote tags via TagVote.svelte component |
| Trending | Period-based aggregation of popular topics |
| Collections | Create collections, add/remove posts |
| Manual circles | User-created groups of people |
| Moderation labels (M1) | Insightful/Misleading/Off-Topic/Troll/Flamebright labels |
| Admin ban/suspend | Direct admin enforcement tools |
| Reports + review queue | Report content, queue for moderator review |
| Public mod decision reviews | Fair/Unfair voting on any moderation action, controversy scoring (`(fair+unfair) * min(fair,unfair) / max(fair,unfair)`), `/modlog/controversial` endpoint |
| Credits economy | Award, spend, getBalance, getHistory |
| Achievements | Check + award milestones |
| Leaderboard API | Top users endpoint |
| Workers (13) | Feed rebuild, trending, trust decay, prune, achievements, notes promote, credits, active stats, scheduled posts, affinity computation, algorithmic list refresh |
| WebSocket hub | Connection manager exists but not wired to feed yet |
| Community notes | Fact-check notes on posts with voting |
| Dockerfile | Multi-stage build at repo root |
| Migration runner | SQL migrations run on startup (4 migrations) |
| User lists | Create/manage follow/block lists, add/remove members, subscribe/unsubscribe |
| Collaborative lists | Editor/viewer roles, invite flow, collaborator management |
| Algorithmic lists | Criteria-based auto-population (`min_trust_level`, `min_posts_30d`, `max_negative_ratio`, etc.), refresh worker |
| User affinity | Event-driven Redis accumulator (O(1) badge lookup), on-demand SQL compute + 24h cache (affinity page), 4-component scoring (tag overlap, co-community, reaction agreement, trust distance) |
| Feed plugins | WASM-sandboxed custom feed algorithms (wazero runtime, 10ms timeout, 1MB memory), plugin marketplace with install/uninstall/review, execution worker |
| Handler integration tests | 4 tests covering user lists, algorithmic lists, affinity endpoints |

## 2. Partially Built / Needs Finish Work

| Feature | Description | What's Needed |
|---------|------------|--------------|
| Trust Level auto-progression | Promote users based on time/activity/flags | Cron job to evaluate + promote |
| SMTP password reset | Email the reset token instead of returning it | SMTP config + email sending |
| Mod post deletion (frontend) | MODREMOVE endpoint exists | Frontend button + confirmation |
| M2 Meta-moderation | Jury system exists but not M2-style review | "Was this fair?" prompt logic |
| Community curators | Scope enforcement without ban power | Endpoint exists, needs frontend |
| Public moderation log | Transparency for admin actions | Endpoint + frontend |
| Mod dashboard | Queue for moderators | Frontend pages |
| Appeal system | Let users contest decisions | Backend + frontend |
| Positive Algorithm | Non-chronological feed ranker | Weighted scoring by trust/reactions/recency/affinity |
| "Why this post?" button | Transparency for feed ranking | Frontend tooltip showing rank factors |
| Custom feeds with RSS | Combine communities/tags/people | Feeds exist, RSS export needed |
| Full reaction set | Insightful/Helpful/Beautiful/Funny/Wholesome/NeedsSources | Only generic like/unlike works |
| Tag autocomplete | Suggest tags as user types | Frontend component |
| Collection visibility layers | Public read / public edit / invite-only | Model exists, UI missing |
| Auto-generated circles | By interest + proximity | Manual circles exist, auto needs ranker |
| Credit streak bonuses | Reward consecutive active days | Backend logic |
| Giftable credits | Transfer credits between users | API + UI |
| Credit hash-chain ledger | Tamper-evident transaction log | Append-only chain |
| Leaderboard frontend | Shows real data | Currently uses hardcoded sample data |
| Achievements frontend | Show badges on profile | No UI for achievements yet |
| ActivityPub federation | ed25519 keys exist, no inbox/outbox | HTTP Signatures, webfinger, activity delivery |
| WebSocket live feed | Real-time post updates | Hub exists, wire to feed |
| Image upload | Model + service exist | Missing endpoint + storage config |
| Seed data | Sample users/posts/communities for fresh instances | Data + migration |
| Search frontend redesign | Backend has full Boolean + facets | Replace search page with tabs/filters/snippets/URL-sync UI |
| Frontend error handling | Many `catch(e) { /* ignore */ }` | Show errors to users |
| SettingsPage refactor | Monolithic settings page | Split into smaller components |

---

## 3. Future Features (From Docs / Design)

| Feature | Description | Next Step |
|---------|------------|-----------|
| Federation (ActivityPub) | Interoperate with Mastodon/Pleroma | Wire ed25519 keys to webfinger, implement inbox/outbox |
| Slashdot-style limited moderation | 5-10 actions/day, no rollover | Already planned in M1 design |
| Tag categories | Artist/Character/General/Rating booru-style | New DB columns + UI |
| Tag redirects | Misspelled tags → correct spelling | Tag alias table |
| Gamified tagging | Earn rep for first to tag a post correctly | Achievement check + bonus |
| Economy: credit bounties | Offer credits for good posts | Bounty endpoints exist |
| Slow-boot + curator review | New member posts held for review | Backend done, frontend missing |
| Invite-only communities | Gated communities with invite codes | Backend partially done |
| WebSocket presence | See who's online in real time | Hub exists, wire to UI |
| "No /All feed" option | Community-only content, no global firehose | Feed filter toggle |
| Mod bot downvote filter | Auto-mod based on bot comment reactions | Custom worker |
| Community forking with spirit-change | Better fork UX than full copy | Improve fork UX |
| Reputation vesting (fixed period) | Cap at full member after vesting, not infinite log | Configurable probation |
| Credit-bounties for community work | Reward mods/curators with credits | Extend bounty system |
| Multi-instance shared blocklists | Auto-exchange blocklists between trusted servers | ActivityPub extension |

---

## 4. Suggested Priority Order

**Immediate (finish existing frontends):**
1. User list management pages (CRUD, members, subscribe UI) — backend done
2. Affinity display page (table + similar users) — backend done
3. Plugin marketplace frontend (browse, install, review) — backend done
4. Search frontend redesign (tabs, filters, URL sync, snippets) — backend done
5. Moderation appeal system

**Short-term (high user-facing impact):**
6. Image upload — highest user-facing gap
7. Trust level auto-progression cron — closes a tier-1 gap
8. Mod post deletion button (frontend) — quick fix
9. SMTP config + password reset email — security gap

**Medium-term:**
10. Positive Algorithm + "Why this post?" — replaces chronological feed
11. M2 meta-moderation UI — closes the accountability loop
12. Full reaction set — richer interaction
13. Tag autocomplete — UX polish

**Long-term:**
14. ActivityPub federation
15. Full economy (streaks, gifting, hash-chain ledger)
16. WebSocket live feed
17. Auto-generated circles
