# Writing Your Own Feed Rules

Most social networks decide what you see. An algorithm — built by a company, for its own goals — chooses which posts appear in your feed, in what order, and how often. You have no say in it. You cannot say "show me more of this kind of thing" or "never show me that kind of thing." The algorithm knows best, and the algorithm works for its shareholders, not for you.

Threadlight gives you that power back. You can write a small program — a plugin — that decides how your feed works. You can install plugins written by other people. You can share your own. The feed becomes yours.

This page explains how plugins work and how to write one.

## What a Plugin Does

A plugin is a small program that receives a list of posts and returns a decision about each one. There are two kinds of decisions:

- **Ranking** — assign a score to each post. Higher-scoring posts appear higher in your feed.
- **Filtering** — decide whether to keep or hide each post. Hidden posts never reach your feed.

You can do either one or both. A rank plugin might boost posts from people you trust. A filter plugin might hide posts about topics you are tired of. A hybrid plugin does both in one pass.

The plugin runs on your server, not in your browser. You upload it once, and it runs every time your feed refreshes. It does not send your data anywhere — it only sees the posts that are already in your feed queue.

## The Sandbox

Plugins run inside a sandbox. This is a protected environment that gives the plugin exactly what it needs and nothing more.

The sandbox has no access to the network. A plugin cannot call out to the internet, send data to another server, or download anything. It has no access to the filesystem. It cannot read your files or write anything to disk. It cannot access any other user's data. It receives only the posts you give it and returns only scores or decisions.

Each plugin has a strict time limit: 10 milliseconds. If a plugin takes longer, it is stopped automatically. This prevents a slow or broken plugin from delaying your feed. The memory limit is 1 megabyte. If a plugin tries to use more, it is stopped.

These limits mean plugins cannot do harm — not intentionally, not accidentally. They can only process the list of posts you gave them and return results.

## Three Kinds of Plugins

**Rank plugins** are the simplest. They receive a list of posts and return a score for each one. The scores can be any number. A score of 10 means "show this prominently." A score of -5 means "bury this." The feed sorts posts by score, highest first.

A rank plugin might give a bonus to posts from people you have high affinity with. Or it might penalize posts that use certain words. Or it might boost posts that are getting a lot of positive reactions. The logic is entirely yours.

**Filter plugins** do not score posts — they decide which to keep. They return a list of the posts that should stay. Everything else is removed. A filter plugin might hide posts about a topic you are taking a break from, or hide posts from accounts that match a pattern you find harmful.

Filtering happens before ranking. If you use both a filter plugin and a rank plugin, the filter plugin runs first to remove unwanted posts, and then the rank plugin scores the survivors.

**Hybrid plugins** do both. They receive all posts, filter some out, and assign scores to the rest. This is useful when your filtering and ranking logic are connected — for example, if you want to penalize certain posts instead of hiding them entirely, or if your scoring depends on something you learned during filtering.

## How to Write One

Plugins are compiled to WebAssembly, a format that runs safely on any server. You can write them in Go, Rust, or any language that compiles to WebAssembly.

A plugin must export at least one of three functions: `rank_posts`, `filter_post`, or `plugin_metadata`. The metadata function tells the server your plugin's name, version, and description. The other two do the actual work.

The input to your plugin is always a list of posts in JSON format. Each post includes its title, body text, author, tags, community, mood, and reaction counts. Your plugin reads this list, applies its logic, and returns either a list of scores (for ranking) or a list of kept post IDs (for filtering).

Here is the simplest possible rank plugin, in pseudocode:

```
for each post in the input list:
    score = 0
    if the post is from someone I trust, add 5 to score
    if the post has helpful reactions, add 3 to score
    if the post has a content warning, add 2 to score
    return score for each post
```

That is the entire idea. The real code is about twenty lines in any language.

## Submitting a Plugin

When you are ready to share your plugin, you upload it through the plugins page. You give it a name, a description, a version number, and the compiled WebAssembly file. The file must be under 1 megabyte.

After uploading, your plugin is not immediately public. An admin or trusted reviewer checks it first. The review is not about whether your plugin is good or useful — it is about whether it is safe. Does it stay within the sandbox limits? Does it do what it claims? Does it have any obvious bugs that could crash the server?

The review process is transparent. Once approved, your plugin appears in the marketplace. Anyone can install it, use it, and leave a review. If people find it useful, it gets more installs and higher ratings. If it has problems, reviewers and users can flag them.

You can also update your plugin. Upload a new version, and anyone who installed the old version gets notified. They can upgrade or stay on the old version. You cannot push updates silently — the user always chooses.

## Installing Someone Else's Plugin

Browse the marketplace. Read the description. Check the rating and reviews. When you find a plugin that sounds useful, click install. It starts working immediately on your feed.

You can install as many plugins as you want. They run in order — first all filter plugins, then all rank plugins. The order matters, so you can rearrange them.

You can uninstall a plugin at any time. Your feed goes back to how it was before. No data is lost. No harm done.

The marketplace is not a store. There is no payment required, no subscription, no tracking. Plugins on Threadlight are free, because the purpose is not to sell you something — it is to give you control over your own experience.
