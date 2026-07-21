# Features That Give Power Back to You

This page is for the person running a Threadlight server. It explains the new pieces that make the platform more democratic — user lists, affinity, and plugins — and how they fit together without needing you to be a technical expert.

## Shared Power, Not Central Control

Most social networks are built like a castle: a few people at the top decide what you see, who you hear from, and what is allowed. Threadlight is built differently. It is more like a town square where everyone has a voice and everyone helps keep things running.

The features in this update continue that idea. They let people organize together, discover each other through genuine connection, and even write their own feed rules — all without needing permission from an admin or a company.

## User Lists: Organizing Together

A user list is simply a group of accounts. You might make one called "science writers I trust" or "accounts that spread harm." What makes lists powerful is that they are **shared tools**, not private notes.

When you create a list, you decide who can see it and who can help maintain it. A list can be just for you, visible to everyone, or something you and a few friends build together. When someone subscribes to your follow list, the system follows every person on that list for them. When they subscribe to a block list, everyone on that list gets blocked.

This turns a single person's careful curation into something the whole community benefits from. One person spots a pattern of harm, adds the accounts to a block list, and everyone who subscribes is protected. A group of people who know a topic well build a follow list of good voices, and newcomers have a healthy starting point instead of staring at an empty feed.

The server runs this automatically. You do not need to configure anything. Lists, their members, subscriptions, and collaborators are stored in the database and updated whenever someone subscribes or a list changes.

## Affinity: Finding Your People

Most platforms try to keep you engaged by showing you what makes you angry or afraid. Threadlight does the opposite. It tries to connect you with people who share your interests and values.

Affinity is a measure of how much you have in common with another person. It looks at four things:

- **Tag overlap** — do you both post about the same topics?
- **Shared communities** — are you both members of the same spaces?
- **Reaction agreement** — when you both react to a post, do you tend to agree or disagree?
- **Trust distance** — how close are you in the web of people who trust each other?

These four signals are combined into a single score. The score is not stored in a central database and batch-processed once a day like in older platforms. Instead, every time you follow someone, join a community, or react to a post, a lightweight counter updates in Redis. When you look at someone's profile, the badge shows your affinity instantly — there is no waiting for a nightly job to run.

For the full affinity page, where you can see your top connections and find similar users, the system computes scores on demand and keeps them cached for a day. You get fresh results when you need them without wasting resources computing scores nobody looks at.

The server does not need special tuning for this. The Redis counters have a three-day expiry so stale data cleans itself up. The on-demand cache refreshes automatically when you visit the page.

## Feed Plugins: Your Feed, Your Rules

The most radical idea here is that you should be able to decide how your feed works. Not through a settings panel with five presets chosen by a corporation, but by writing — or installing — a small program that ranks and filters posts however you want.

These programs run in a sandbox. They cannot access your data, the network, or anything else on the server. They receive a list of posts and return either a score for each one (to rank them) or a decision to keep or hide each one (to filter them). You can install plugins made by other people or write your own.

Plugins are reviewed before they become public, so bad actors cannot sneak something harmful into the marketplace. The review process is transparent — anyone can see who reviewed a plugin and what they said.

The server runs plugins using WebAssembly, a safe and fast runtime that is built into the Go code. Each plugin gets a strict time limit and memory limit. If a plugin tries to do too much, it is stopped automatically. No harm done.

## How the Server Handles All This

The database tables for these features are created automatically when the server starts. There is nothing to run manually. Two background workers keep things fresh:

- **Affinity computation** runs once an hour and updates affinity scores between users who have been active recently. This is a safety net — the real-time Redis counters handle the instant lookups, and the hourly worker fills in any gaps.
- **Algorithmic list refresh** runs every 30 minutes. If you have an algorithmic list (a list that fills itself based on rules), this worker evaluates the rules and updates the membership automatically.

These workers are gentle on resources. They run inside the same Go process, log what they do, and catch any panics so they never crash the server.

If you need to see what is happening, the server logs show each worker run — how long it took, how many rows it processed, and any errors it hit. The default settings work for a small community of a few hundred people without any tuning.
