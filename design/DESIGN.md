# Threadlight — Platform Design

**Tagline:** Bringing people together.
**Core Philosophy:** Replace outrage with connection. Replace algorithmic amplification with collaborative curation. Replace top-down moderation with trust-based democratic governance.

---

## 1. Why Threadlight Exists

Existing social platforms are designed for one thing: engagement. Engagement is maximized by outrage, division, and algorithmic manipulation. Threadlight exists to prove there is another way — a platform that brings people together around shared interests, trusts them to govern themselves, and refuses to treat human attention as a resource to be extracted.

The core insight: **most people want to connect, not fight.** The reason they fight on existing platforms is that the platform's algorithm rewards fighting. Threadlight's architecture is designed to reward connection instead.

---

## 2. Architecture Overview

```
Client (Svelte SPA + Vite)
    │ REST / WebSocket
    ▼
Go / Gin API Server (single binary)
    │
    ├── PostgreSQL (primary data, full-text search, JSONB)
    └── Redis (cache, pub/sub, rate limiting, presence)
```

**Why Go**: Single binary deploy, excellent concurrency for WebSocket, great performance on a $5 VPS or Orange Pi.

**Why Svelte**: Smallest bundle size of major frameworks, reactive by design, excellent for AI generation.

**Why PostgreSQL + Redis**: PostgreSQL handles everything relational — users, posts, communities, tags, interactions. Redis handles ephemeral state — rate limits, online presence, session cache, real-time pub/sub.

---

## 3. Trust System

Trust is the central mechanism of Threadlight. Instead of follower counts or like ratios, trust determines what you can do and how much your actions matter.

### 3.1 Discourse-Style Trust Levels

| Level | Name | How to Reach | Abilities |
|-------|------|-------------|-----------|
| 0 | New | Sign up | Read, basic interactions, limited posting |
| 1 | Basic | 2+ days, 5+ posts/replies, no flags | Unrestricted posting |
| 2 | Regular | 10+ days, 50+ interactions, consistent positive engagement | Flag content, vote on tags, heavier trust weight |
| 3 | Leader | 30+ days, sustained constructive participation | Vouch for new users, M1 moderation, moderate tags |
| 4 | Elder | 90+ days, excellent standing | Full trust, weighted moderation, oversight roles |

Progression is not a grind — it depends on time, positive interactions, and absence of moderation actions. It cannot be rushed.

### 3.2 Web of Trust (Vouching)

At Level 3+, users can vouch for others. Vouching is public — your name appears on the trust connection. If a user you vouched for breaks rules repeatedly, your own trust score is affected.

This makes bot networks structurally impossible: you cannot create 1,000 bot accounts because you cannot find 1,000 real people willing to stake their reputation on them.

### 3.3 Trust Weight

Each user's trust vote is weighted by their trust level:

| Level | Weight |
|-------|--------|
| 0-1 | 1 |
| 2 | 2 |
| 3 | 3 |
| 4 | 4 |

An Elder's trust matters more than a squad of brand-new accounts.

---

## 4. Moderation System (M1 + M2)

Inspired by Slashdot's two-tier moderation. Content is moderated by peers (M1). Moderators are held accountable by peers (M2).

### 4.1 The Rules

1. **You must not deny the humanity of another person.** — No dehumanization, no harassment, no advocating violence against a group.
2. **You must not lie about material conditions.** — No Holocaust denial, anti-vax pseudoscience, climate denial that delays action, or conspiracy theories that cause real-world harm.
3. **You must not astroturf.** — Every account must be a real person. Bots, bought engagement, coordinated sock-puppets are not allowed.
4. **You must not post illegal content.** — Whatever is illegal in the server's jurisdiction.

### 4.2 M1: Content Moderation

Users at trust level 2+ can apply labels to posts:

- **Positive**: Insightful, Interesting, Informative, Funny, Underrated
- **Corrective**: Misleading, Off-Topic, Troll, Flamebait

Labeling adds context — it does not remove content. Heavily corrective-labeled posts are visually marked and deprioritized in feeds. Heavily positive-labeled posts gain visibility.

Moderators receive **5-10 actions per day**. Unused budget does not roll over. This prevents any single user from dominating moderation.

### 4.3 M2: Meta-Moderation

Periodically, moderators are asked to review another moderator's recent decision: **"Was this moderation fair?"** Yes or No.

Moderators whose decisions are consistently rated unfair lose their privileges. Fair moderators may earn a slightly larger daily budget.

There is **no appeals process** for individual moderation decisions. The label sticks. This prevents the system from becoming a meta-game.

### 4.4 Admin Safety Enforcement

For cases the community system cannot handle — real-world threats, organized harassment, legal takedowns — administrators have direct tools: warnings, suspensions, bans, and overrides. All admin actions are logged in a **public moderation log** with justification. This ensures administrators are as accountable as moderators.

---

## 5. Collaborative Tagging (Booru-Style)

Tags are not owned by post authors. Anyone can add a tag to any post; the community votes on relevance.

### 5.1 Tag Categories

- **Content**: what the post is about (technology, music, gardening, philosophy)
- **Descriptors**: tone or quality (wholesome, thoughtful, controversial, funny)
- **Moderation flags**: spoiler, AI-generated, NSFW

### 5.2 Tag Voting

Tags can be upvoted or downvoted. Widely-upvoted tags are prominently displayed. Downvoted tags fade. Tags are not deleted by a single moderator — they fade through consensus.

### 5.3 Subjective Tags (Phase 2)

Subjective tags (e.g., "underrated," "wholesome," "cringe") can be voted on separately from objective tags. Both coexist: objective tags provide reliable metadata; subjective tags surface community sentiment.

---

## 6. Communities

Communities are topic-based groups with independent rule sets. They do **not** degenerate over time because of several design protections:

- **Slow boot**: New communities grow gradually (configurable 0-14 days). This gives founders time to set culture.
- **Probation**: New members' votes and flags have reduced weight for a period.
- **Tenure-weighted voting**: Voting power within a community grows logarithmically with membership duration.
- **Curators**: Users who enforce scope tags (what the community is about). Curators can prune off-topic content but cannot ban users.
- **Forking**: If a community drifts, any member can fork it — copy the member list and rules into a new community with a different direction.

---

## 7. The Positive Algorithm

The feed algorithm ranks content by substance and connection, not outrage:

1. Content from trusted users is ranked higher
2. Content rated "insightful" or "interesting" by M1 is ranked higher
3. Content rated "misleading" or "troll" is ranked lower
4. Repeatedly flagged content reduces the author's visibility
5. "Why this post?" button exposes the exact ranking factors

Content that generates high interaction but low trust-weighted quality scores (heated arguments, trolling) is automatically downranked.

### 7.1 Custom Feeds

Users can create custom feeds from any combination of:
- Specific communities
- Specific tags
- Specific people
- Any combination of the above

Each custom feed has an RSS/Atom export.

### 7.2 Advanced Search

Boolean operators (AND, OR, NOT), fielded search (author:, community:, tag:), date range filtering (after:, before:), and saved searches.

---

## 8. Circles (Proximity + Interest)

Circles are auto-generated groups of nearby people with overlapping interests. They use one-way location hashing — you can find people near you without revealing your exact location.

Unlike communities, circles are discovery-focused and require no commitment. Browse a circle's trending topics and active members, follow interesting people, move on.

---

## 9. Reactions

Beyond a single "like," users can react with:
- **Insightful**, **Helpful**, **Beautiful**, **Funny**, **Wholesome**, **Needs Sources**

Each reaction gives different signals to the feed algorithm. A post with many "Wholesome" reactions is surfaced differently from one with many "Insightful" reactions.

---

## 10. Economy & Achievements (Phase 2)

- **Credits**: Earned by posting, getting trusted, moderating fairly. Spent on promoting posts or creating invite-only communities. Credit transactions use a hash-chain ledger for tamper evidence.
- **Streak bonuses**: Activity rewards for consecutive days.
- **Giftable credits**: Users can send credits to each other.
- **Achievements**: Badges for milestones — "First Post," "Level 4," "100 Days Active."

---

## 11. Federation (Phase 3)

Architecture is designed for ActivityPub compatibility:
- Every user has an ed25519 keypair and actor_id
- Public timeline can be mapped to outbox/inbox
- Shared blocklist format for inter-instance defense

---

## 12. Anti-Fascism Architecture

Threadlight is explicitly anti-fascist. This is not about left vs. right — it is about recognizing that authoritarian patterns can emerge from any direction.

**Horseshoe theory**: the far left and far right bend toward each other, developing the same authoritarian habits with different branding. On the right: nativist scapegoating and strongman worship. On the left: cancel culture and moral enforcement. Both abandon free speech, reduce people to identity labels, and impose their will through intimidation.

Threadlight rejects both patterns. The platform's defenses:
- **Web of Trust** prevents bot armies and coordinated harassment
- **M1 + M2 moderation** distributes power and holds moderators accountable
- **Tag voting** prevents any single group from controlling the taxonomy
- **Community forking** prevents any single community from becoming a monopoly
- **Public mod logs** ensure admin accountability

There are 138 million children in child labor worldwide. Many producing goods for the countries that bomb their neighborhoods. If this does not bother you, we would prefer you do not use this platform.
