# Feeds, Circles, and Search

This page explains how Threadlight decides what to show you and how to find things you
are interested in.

## The Positive Algorithm

Most social networks use an algorithm that decides what you see. The problem is that these
algorithms learn that anger and outrage get the most attention. So they show you things
that make you angry, because an angry person scrolls more, and a person who scrolls more
sees more ads.

Threadlight does not have ads, so it does not need to make you angry.

Our feed system is called **The Positive Algorithm.** It ranks content based on substance
and connection, not outrage:

- Posts from people you trust appear higher
- Posts that other people found insightful or interesting appear higher
- Posts that were rated misleading or trolling appear lower
- People who get flagged a lot have their posts shown less
- Recent posts are balanced against quality

### Why This Post?

Every post in your feed has a **"Why this post?"** button. Click it to see exactly why
that post appeared. You will see something like "because you follow @username" or "because
this is trending in your circles." There is no secret sauce. The algorithm is completely
transparent.

### What Gets Deprioritized

Posts that get a lot of attention from angry or bad-faith reactions are automatically
shown to fewer people. The system prefers conversations that make people think over
conversations that make people fight.

## Custom Feeds

You can create your own custom feeds that mix together different sources:

- Posts from specific communities
- Posts with specific tags
- Posts from specific people
- Any combination you want

Each custom feed can be exported as RSS. You can subscribe to it in any RSS reader
(some email programs and news readers support RSS).

## Trending

Threadlight detects trending topics by reading what people are posting about, not by
counting hashtags. The system analyzes the text of posts that are getting a lot of
attention and surfaces the themes that are emerging.

Trending is scoped to your communities and circles. You see what is popular among the
people you care about, not what is popular globally.

## Search

You can search for posts, people, communities, and tags. Threadlight's search supports:

- **Boolean operators:** You can search for "cats AND dogs" to find posts about both, or
  "cats NOT dogs" to find posts about cats that are not about dogs.
- **Fielded search:** You can search within specific fields like `author:username`,
  `community:name`, or `tag:tagname`.
- **Date range:** You can filter by date using `after:2025-01-01 before:2025-06-01`.
- **Saved searches:** You can save a search you use a lot for quick access later.

## Circles

Circles are groups of people who share your interests and are near you. The system
automatically suggests circles based on the topics you follow and where you are located.

Circles are different from communities. You do not have to join a circle. It is just
a way to discover new people. Browse the circle, find someone interesting, follow them,
move on.

Circles are designed to protect your privacy. You can find people nearby without ever
revealing your exact address.

## Shared Blocklist

The server maintains a shared blocklist of domains and IP addresses that are known to be
problematic. This list can be exported as a JSON file so that other server administrators
can import it. It is defensive infrastructure — it does not do anything visible to regular
users, but it helps keep the platform safe.
