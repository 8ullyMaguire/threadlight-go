# Affinity: Finding Your People

Most social platforms have one goal: keep you watching. They learn what makes you angry, what makes you afraid, and what makes you argue, because those things keep you scrolling. They do not care if you find people who share your values or your interests. They care that you stay.

Threadlight is different. Instead of an algorithm that exploits your attention, it offers you a tool that helps you find connection. We call it affinity.

## What Affinity Means

Affinity is not popularity. It is not a count of followers or likes. It is a measure of how much common ground you share with another specific person. It looks at what you both post about, the communities you both belong to, how you both react to things, and how close you are in the network of people who trust each other.

Think of it like this: if two people attend the same community meetings, talk about the same topics, tend to agree on what is valuable, and are connected through people they both trust, they probably have something in common. The affinity score captures that.

The score is a number between zero and about seven. A score above five means you share a lot. A score below one means you have not crossed paths much. Zero means the system has never seen you interact with the same things.

## How the System Learns Your Connections

Every action you take on Threadlight feeds into affinity quietly, in the background. There is nothing you need to enable or configure.

When you follow someone, the system makes a note. When you join a community, it notes that too. When you react to a post — calling it insightful, flagging it as off-topic, or any other reaction — the system checks who else reacted to that same post and compares your reactions. When someone trusts you, or you trust someone, the system maps that connection.

None of this data leaves your server. It is stored in Redis, a fast in-memory store, and kept only as long as it is useful — three days for the live counters, 24 hours for the full page cache. Old data disappears on its own.

## The Badge on Every Profile

When you visit someone's profile, you might see a small badge that shows your affinity with them. This badge is fast — it reads four numbers from Redis and combines them in a fraction of a millisecond. There is no database query, no waiting, no batch job.

The badge updates in real time. If you follow someone and then visit their profile, the badge reflects that immediately. If you both react to the same post, the badge adjusts on your next visit.

This instant feedback makes the badge feel alive. It is not a static number computed once a day. It is a living signal that grows and shifts as your relationship with someone changes.

## The Full Affinity Page

Your affinity page shows you a list of everyone you have a meaningful connection with, ordered by how strong that connection is. For each person, you can see the four components that make up the score:

- **Tag overlap** — how many topics you both post about. If you both write about gardening and astronomy, that counts as two shared tags. The more you share, the higher this number.
- **Co-community** — how many communities you both belong to. Shared spaces mean shared context.
- **Reaction agreement** — when you both react to the same post, do you tend to agree on what kind of reaction it deserves? High agreement means you share values about what is good and what is not.
- **Trust distance** — how many hops apart you are in the trust network. A direct trust connection is a score of 1.0. A friend of a friend is about 0.5. Someone five hops away is 0.2.

The page also has a "similar users" tab. You set a minimum score, and the system shows you everyone above that threshold. This is how you discover people you might want to follow — not through an algorithmic recommendation engine that has been trained on billions of data points, but through the actual things you and others have done together on this platform.

## Private and Secure

Affinity is computed from data that is already public — your posts, your communities, your reactions, your trust connections — but the scores themselves are only visible to you. No one else can see your affinity page. No one else can see their score with you unless you choose to show it.

The system never sells or shares affinity data. There are no advertisers to sell it to. The only purpose of affinity is to help you find meaningful connection on a platform that is designed for that, not for profit.
