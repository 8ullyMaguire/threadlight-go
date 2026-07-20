# Selected Community Design Ideas for Polaris

## From the r/programming Post

### 1. Tenure-Weighted Influence (Community Trust Scores)

**The problem:** In any growing community, the people who built it lose control to the flood of new users. The "first 100" effect disappears at 10,000.

**The Polaris approach:** Each community has its own trust graph — a web of trust connections weighted 1–10. When someone vouches for a new member inside a community, that's a trust edge. The community's algorithm weights member input (votes, flags, suggestions) by their tenure-weighted trust score within that *specific* community, not the global platform.

**Edge cases handled:**
- Community creator gets unlimited invites (cannot be locked out)
- Forking a community does NOT copy trust edges — the fork starts fresh
- Trust connections can be revoked at any time; revocations are public
- Minimum global trust score floor of 0.5 prevents brand-new accounts from joining (default, configurable per community)
- Trust score within a community = sum of incoming trust connection weights

### 2. Invite-Only Communities (Optional, Per-Community)

**When it applies:** A community can toggle `invite_only = true`. When enabled:
- Only existing members can vouch for new members (create trust connections)
- The community creator always has unlimited invites
- Anyone below the min trust score is blocked regardless of invite
- Open communities coexist — `invite_only` is per-community, not global

**Edge case — the bootstrap problem:** When a community first enables invite-only mode, existing members at join time get automatic trust weight. No one is grandfathered into being locked out.

### 3. Slow-Boot Probation (Reused from original design)

Even with invite-only, new members can do damage. Slow-boot gives communities a probation period (0–14 days) where a new member's posts are held for review by existing members before going public.

---

## Slashdot-Style Moderation

### The original idea
Slashdot gave users a finite pool of "moderation points" that replenished. You earned the right to moderate by being a good commenter. A few key design choices made it work:
- **Scarce resource:** Each user got 5 points, not infinite votes
- **Score capping:** Comments capped at +5/-1 — no dogpiling
- **Meta-moderation:** Users reviewed whether votes were fair; abusers lost moderation privileges

### How it fits Polaris

Polaris already has most of the infrastructure. The additions would be:

**Finite moderation actions:**
```
moderation_actions_per_month = floor(trust_score * 3) + 2
```
A user with trust score 1.5 gets 6 moderation actions per month. They can flag, timeout, or vote in juries. Once they're out, they're out until next month.

**Why this matters with community trust graphs:**
In an invite-only community, trust connections ARE the moderation point allocation. Vouching for someone is a limited resource — you only get so many invites per month based on your trust score. This makes every invite meaningful. You don't invite randoms because you might need to use that invite on someone better later.

**Combined flow:**
1. Community is invite-only
2. Existing members have a limited monthly invite budget (based on trust level)
3. To invite someone, you create a trust connection inside that community
4. The invitee can now join
5. If the invitee behaves badly, the inviter's trust takes a hit (vouch chain responsibility)
6. Meta-moderation: if someone abuses their invite power (invites spammers), the community can vote to reduce their invite limit

**Score capping in communities:**
Within a community's feed, posts can reach at most `+N / -1` where `N = community_size / 100 + 3`. This prevents massive dogpiles and ensures one controversial post doesn't dominate the entire community's discourse.

### Why not the full Slashdot system?

The Slashdot system was designed for the *entire* site. Polaris's approach is:
- Global moderation is done by the trust system + juries (for serious actions)
- Community moderation is about curation and gatekeeping (vouching, flagging within the community)
- Finite moderation per month prevents burnout (a problem Slashdot also solved by cycling moderators)

### Implementation status in Polaris

| Component | Status |
|-----------|--------|
| Community trust connections | ✅ Newly implemented |
| Invite-only toggle per community | ✅ Newly implemented |
| Soft timeouts (existing) | ✅ Already in trust system |
| Meta-moderation (planned) | ⬜ Jury system already exists, just needs extension |
| Finite moderation budget | ⬜ Simple calculation, not yet implemented |
| Score capping in community feeds | ⬜ Feed algorithm needs community-specific max score |
