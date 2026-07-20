# Trust, Communities, and the Web of Trust

This page explains the core systems that make Threadlight work: how trust is earned, how
communities function, how the Web of Trust prevents bots, and how circles help you discover
people.

## Discourse Trust Levels

Threadlight borrows the concept of **trust levels** from Discourse. These are stages of
earned access — gradually unlocked freedoms, not locked features. Everyone starts at Level 0
and progresses by participating constructively over time.

### Level 0: New

Read, basic interactions. Limits damage from drive-by spammers.

### Level 1: Basic

Create unrestricted posts after a few days of constructive participation.

### Level 2: Regular

Flag content, vote on tags, heavier trust weight in the Web of Trust. Reached after
sustained positive activity.

### Level 3: Leader

Vouch for new users, moderate content (M1), moderate tags. Leaders are users who have
consistently demonstrated good judgment.

### Level 4: Elder

Full trust, weighted moderation votes, ability to help resolve edge cases. Reached after
months of excellent standing.

Progression depends on time spent, positive interactions received, and absence of serious
moderation actions. It cannot be rushed.

## Web of Trust

The Web of Trust is a directed graph of trust connections between users. A→B means
"A vouches for B as a participant in good faith."

### Vouching

At **Leader** level or above, you can vouch for other users. Vouching is public — your name
appears on the trust connection. If a user you vouched for repeatedly breaks rules, your own
standing may be affected.

This makes bot networks structurally impossible: you cannot create a thousand bot accounts
because you cannot find a thousand real people willing to stake their reputation on them.

### Trust Weight

Each user's trust vote has weight determined by their own trust level:

| Level | Weight |
|-------|--------|
| 0-1 | 1 |
| 2 | 2 |
| 3 | 3 |
| 4 | 4 |

An Elder's trust matters more than a squad of brand-new accounts.

### Trust Score

Your overall trust score is a function of how many people trust you (weighted by their level),
how many distrust you, and whether the people trusting you are themselves trusted. Low scores
limit what you can do.

## Communities

Communities are groups of people who share a common interest. Each has its own rules,
moderators, and culture.

### Finding and Joining

Browse all public communities. Each shows name, description, member count, and whether it is
open or invite-only.

### Probation Period

When you join a new community, you begin in a short **probation period**. During this time,
your votes and flags within that community carry less weight. This prevents drive-by brigading
and lets you learn the community's norms before your actions have full effect.

### Tenure-Weighted Voting

Within a community, voting power on community-specific decisions grows with membership
duration. The relationship is logarithmic — early growth is fast, but long-time members
accumulate a meaningful voice. This rewards sustained contribution without giving newcomers
zero say.

### Community Curators

Curators are users who enforce a community's **scope tags** — the tags that define what the
community is about. Curators can add or remove scope tags and prune off-topic content. They
**cannot ban users**. Bans are handled by the moderation system (M1 + M2) or by administrators.

### Creating a Community

Any user at **Regular** level or above can create a community. New communities start in
**slow boot** mode (gradual member growth). Founders can set open or invite-only access.

### Forking a Community

If a community drifts away from what you wanted it to be, you can **fork** it. Forking
copies the member list and rules into a new community with a different direction. The
original is unchanged.

## Circles

Circles are auto-generated groups of people organized by shared interests and proximity.
Unlike communities, circles are discovery-focused and do not require joining.

- **Auto-Generated**: The system suggests circles of nearby people with overlapping tag
  subscriptions and interaction history.
- **Privacy-Preserving**: Location is handled through one-way hashing — you can find nearby
  people without ever revealing your exact address.
- **Lightweight**: No commitment, no notification spam. Browse a circle's trending topics
  and active users, follow interesting people, move on.

## Collaborative Tagging

Tags are not owned by post authors. Anyone can add a tag to any post; the community votes
on relevance. See [Posting and Tagging](02-posting.md) for details.
