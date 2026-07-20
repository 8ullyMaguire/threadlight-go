# Feeds, Circles, Search, and The Positive Algorithm

This page explains how Threadlight decides what to show you and how to find what you're
looking for.

## The Positive Algorithm

Most platforms optimize for engagement — time spent, clicks, reactions. These algorithms
learn that anger and outrage produce the most engagement, so they show you things that make
you angry.

Threadlight's feed algorithm is different. It is called **The Positive Algorithm**, and it
optimizes for substance and connection:

- Content from people you trust is ranked higher
- Content rated "insightful" or "interesting" by moderators is ranked higher
- Content rated "misleading" or "troll" is ranked lower
- Repeatedly flagged content from a user reduces that user's visibility
- Recency is balanced against quality

### Why This Post?

Every post in your feed has a **"Why this post?"** button. Click it to see exactly why the
algorithm surfaced that specific piece of content — "because you follow @user," "because
this community matches your tags," "because this is trending in your circles." This makes
the algorithm fully transparent.

### Outrage Downranking

Content that generates high interaction but low trust-weighted quality scores (heated
arguments, trolling, bad-faith engagement) is automatically downranked. The algorithm
prefers posts that generate thoughtful reactions over posts that generate angry ones.

## Custom Feeds

You can create custom feeds that combine different sources:

- Posts from specific communities
- Posts with specific tags
- Posts from specific people
- Any combination of the above

Each custom feed can be exported as **RSS/Atom** — subscribe in any RSS reader. Feeds are
also available for any tag, community, or collection.

## Trending Topics

Threadlight detects trending topics from **post content**, not hashtags. The system analyzes
the text of posts that are receiving disproportionate attention in a time window and surfaces
the emerging themes. This prevents hashtag spam and catches trends the community hasn't named
yet.

Trending is scoped to your communities and circles — you see what matters to the people you
care about, not a global feed.

## Advanced Search

Search across posts, users, communities, and tags with:

- **Boolean operators**: `cats AND dogs`, `cats OR dogs`, `cats NOT dogs`
- **Fielded search**: `author:username`, `community:name`, `tag:tagname`
- **Date range**: `after:2025-01-01 before:2025-06-01`
- **Saved searches**: Save frequently used searches for quick access

## Circles

Circles are auto-generated groups of people who share your interests and are near you.
Unlike communities, circles are discovery-focused. The system proposes circles based on
overlapping tag subscriptions and interaction patterns.

Circles use one-way location hashing — you can find people nearby without revealing your
exact location. Browse a circle's trending topics and active members, follow interesting
people, and move on.

## Shared Blocklist

For federation-ready instances, a **shared blocklist** in JSON format lists domains, IPs,
and server fingerprints that the instance blocks. The blocklist is publicly exportable so
other instance admins can import it. This is defensive infrastructure — it exists for
federation readiness even if no federation code is active yet.
