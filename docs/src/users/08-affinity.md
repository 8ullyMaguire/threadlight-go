# Affinity

This page explains what affinity scores are, how they are calculated, and how to use them to find people you might want to follow.

## Overview

Affinity is a measure of how connected you are to another user on Threadlight. It takes into account the things you share — tags, communities, reactions, and trust connections — and turns them into a single score.

A higher affinity score means you and that person have more in common. A lower score means you are less connected. Affinity is always between two people, not a global popularity score.

You can see your affinity with someone in two places:

- **The affinity badge** — a quick indicator on someone's profile page
- **The affinity page** — a full list of your highest-affinity users

## What Makes Up Your Affinity Score

The affinity score is composed of four components, each measuring a different kind of connection.

### Tag Overlap

The system counts how many tags (topics) you and the other person both use. If you both post about "gardening" and "cooking," that counts as overlap. The more shared tags, the higher this component.

Tag overlap is normalized as a ratio of shared tags to total unique tags across both users. If you share 5 tags out of 10 total, your tag overlap is 0.5.

### Co-Community Membership

If you are both members of the same communities, that counts toward affinity. The more communities you share, the stronger this signal.

This is normalized using a sigmoid function — one shared community gives a small bump, five shared communities gives a much bigger one, and the effect plateaus after about ten.

### Reaction Agreement

When you and another person react to the same posts, the system compares your reaction types. Did you both find the post insightful? Did you both flag it as off-topic?

If your reaction types match more often than not, this component goes up. If you tend to disagree on what is insightful versus misleading, it goes down.

Reaction agreement is the ratio of matching reactions to total shared reactions. If you both reacted the same way on 8 out of 10 shared posts, your agreement is 0.8.

### Trust Distance

This measures how close you are in the web of trust. A direct trust connection (someone vouched for you, or you for them) gives the highest score. The further away you are in the trust network, the lower this component.

The trust distance score is `1 / depth`, where depth is the shortest path through trust connections. A direct trust connection gives 1.0. A friend of a friend gives about 0.5. Someone five hops away gives 0.2.

### How They Combine

The four components are weighted and summed into a single score:

```
affinity_score = tag_overlap * 0.25 
                 + co_community * 0.20 
                 + reaction_agreement * 0.20 
                 + trust_distance * 0.20 
                 + interaction_weight * 0.15
```

The result is a number between 0 and 1, rounded to 3 decimal places. The **interaction weight** component counts direct interactions like replies and mentions, normalized on a log scale.

### Recency Factor

Affinity also takes into account how recently both people were active. The recency factor uses exponential decay — if both of you have been active recently, the factor is close to 1.0. If neither of you has logged in for months, it drops toward 0.1.

The half-life is about 90 days. So someone who was active yesterday has a much stronger recency factor than someone who was last active a year ago.

## The Affinity Badge

When you visit someone else's profile, you will see a small affinity badge. This is always visible and updates in real time because it reads from lightweight Redis counters instead of running a full database query.

The badge uses Option A — an event-driven accumulator that tracks follows, community joins, reactions, and trust connections as they happen. Every time you follow someone, join a community, or react to a post, the counters are bumped. The badge query is O(1) — it reads four hash fields and computes a score immediately.

The badge score range is 0 to 7.5, where:

- **0 — 2:** Low connection. You have very little in common.
- **2 — 4:** Some connection. You share a few tags or communities.
- **4 — 6:** Strong connection. You are in several shared communities and have similar reaction patterns.
- **6 — 7.5:** Very strong connection. You have direct trust connections, many shared interests, and consistent reaction agreement.

The badge is meant as a quick glance. For the full picture, use the affinity page.

## The Affinity Page

Your affinity page shows your top 20 similar users by default, ordered by affinity score. Each entry shows the score and a breakdown of the four components.

The affinity page uses Option C — on-demand computation that is cached for 24 hours. When you visit the page:

1. The server checks Redis for a cached result (`aff:cached:{yourUserID}`).
2. If found and less than 24 hours old, it returns the cached result.
3. If not found, it computes your affinity against all active users, stores the results in the database, writes them to Redis with a 24-hour TTL, and returns the result.

The cache means the first person to visit an affinity page after 24 hours triggers a computation, but subsequent visitors see the cached version instantly. If two requests trigger computation at the same time, one of them gets a 30-second lock and the other waits.

## Finding Similar Users

From the affinity page, you can also see a list of user IDs most similar to you. This is the same data but returned as a simple list of user IDs without the breakdown details. You can use this to:

- Find new people to follow who share your interests
- Discover communities you might like (check what communities your similar users are in)
- See who has similar taste in content

The similar users endpoint returns up to 20 users by default (max 100).

## API Routes

Affinity data is accessed through the following API endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/affinity` | Get your top affinities (with breakdowns) |
| GET | `/api/v1/affinity/similar` | Get a list of similar user IDs |

Both endpoints accept a `limit` query parameter (default 50 for affinities, 20 for similar).

## Privacy Notes

- Affinity scores are only visible to you. No one else can see who you have high affinity with.
- Affinity is computed from public data (tags, communities, reactions, trust connections). It is an analysis of things you have already made visible.
- You cannot opt out of affinity computation, but it only uses data that is already public.
- Blocking someone removes them from your affinity calculations.
