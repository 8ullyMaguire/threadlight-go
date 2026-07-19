# Polaris Dream Features

This page lists every feature we want to build one day. They are not promised, not scheduled,
and some may never happen. They are dreams — ideas we think would be cool if someone had the
time and energy to build them.

---

## 1. Booru-Style Tag Curation

Tags should work like on imageboards (e.g. Danbooru):
categorized into **Artist**, **Character**, **Copyright**, **General**, **Meta**, and
**Rating**. Users apply tags to posts with a category. Tag suggestions pop up as you type.
Tag aliases redirect misspellings. Tag implications automatically add parent tags.

The tag page would show related tags, the most popular posts, and let you filter by tag
category. Tagging could be a game — users earn a bit of reputation for being the first to
tag a post correctly.

---

## 2. Slashdot-Style Moderation

Instead of the current jury system (which is complex and hard to explain), adopt a
Slashdot-like model:

- **Moderators** get a limited pool of moderation actions per day (say 5-10). They can apply
  labels like *Insightful*, *Funny*, *Interesting*, *Informative*, *Troll*, *Flamebait*,
  *Offtopic*.
- **Meta-moderation**: every few hours, moderators are asked to review other moderators'
  decisions. "Was this mod fair?" If a moderator consistently makes bad calls, they lose
  their moderation privileges.
- **No appeals** — once a moderation action is taken, it sticks. This discourages gaming the
  system and keeps the process fast.
- Users with high trust scores see more highly-rated content first. Users can filter by
  minimum rating.

---

## 3. Objective vs. Subjective Tags

Tags come in two flavours:

- **Objective**: factual, verifiable (e.g. "contains spoilers", "AI-generated", "NSFW").
  These are applied by the post author or by community vote with a high threshold.
- **Subjective**: opinion-based (e.g. "wholesome", "cringe", "underrated"). Anyone can add
  them, and the community votes them up or down. Highly-upvoted subjective tags appear
  prominently; downvoted ones fade.

This way objective metadata stays reliable while subjective opinions still have a place.

---

## 4. Proper Onboarding Flow

The current onboarding wizard (OnboardingWizard.svelte) calls backend endpoints that don't
exist. The dream version:

- Step 1: Pick tags you're interested in (uses real tag search with autocomplete).
- Step 2: Follow suggested users based on your tag picks (top users in those tags).
- Step 3: Create or join a community.
- Step 4: Make your first post (guided).

Each step calls a real backend endpoint that updates `onboarding_stage`. Users who skip
onboarding get full access anyway (stage 4 by default) — the wizard is just a helpful
suggestion.

---

## 5. Frontend That Doesn't Suck

A complete frontend rewrite (or major refactor) that fixes:

- **Routing**: Replace the 16-branch `{#if}` chain with a proper hash-based router using
  `<svelte:component>`.
- **Consistent design**: A shared CSS module for forms, buttons, cards, and modals. No
  duplicate styles across pages.
- **Error handling**: Every API error is shown to the user, not silently caught. Loading
  states and empty states everywhere.
- **Pagination**: Infinite scroll or a "Load more" button on the feed, profiles, and
  search results.
- **Real data**: No hardcoded sample entries. Every page calls real API endpoints.
- **Settings split**: The 672-line SettingsPage.svelte split into 4 independent components.
- **Svelte 5 migration**: Proper use of runes (`$state`, `$derived`, `$effect`) instead
  of legacy `$:` and `export let`.
- **Accessibility**: Proper aria labels, keyboard navigation, focus management, screen
  reader support.

---

## 6. Tags Search

A dedicated `GET /api/v1/tags/search?q=...` endpoint that searches tags by name. Returns
an autocomplete-friendly list sorted by post count. Frontend shows a tag search box on the
explore page and in the post composer.

---

## 7. Content Pruning Configuration

The prune worker already supports env var overrides (`PRUNE_AGE_DAYS`,
`PRUNE_MIN_INTERACTIONS`, `PRUNE_DISABLE`). The dream is a UI in the admin panel where
admins can tweak these values without restarting the server, and see how many posts would
be affected before the prune runs.

---

## 8. Federation (ActivityPub)

One day Polaris could speak ActivityPub, making it compatible with Mastodon, Lemmy, and
Piefed. Users on other instances could follow Polaris users, and Polaris communities could
federate with Lemmy communities. This is a massive undertaking — it would require mapping
Polaris's trust model to ActivityPub's simpler follow/like model.

---

## 9. Community Notes Evolution

The current community notes system (fact-check style) could be extended with:

- **Note templates**: Predefined note types like "Correction", "Context", "Source needed".
- **Expert verification**: Users with high trust in a topic area get a weighted vote on notes
  in that area.
- **Note history**: See how a note evolved over time as people edited it.

---

## 10. Economy & Achievements

The credits and achievements tables already exist in the schema. The dream:

- **Credits**: Earned by posting quality content, getting trusted, serving on juries (or
  meta-moderation). Spent on promoting posts or creating invite-only communities.
- **Achievements**: Badges for milestones — "First Post", "Trust Score 10", "Joined 5
  Communities", "100 Days Active". Displayed on user profiles.

Neither system is wired up in the UI yet.

---

## 11. Circle & Proximity Features

The schema has `circles`, `circle_members`, and `location_hash` in users. The dream:

- **Circles**: Private groups organised by topic or area. Unlike communities, circles are
  discovery-focused — they help you find people near you who share your interests, without
  the full commitment of joining a community.
- **Proximity**: Using `location_hash` (a one-way hash of a rough location), you could
  discover trending topics and users in your area without revealing your exact location.

---

## 12. Moderation Dashboard

A real moderation dashboard for admins and community moderators that shows:

- Current queue of flagged posts
- Moderator performance stats (fairness ratio, actions taken)
- Trust score distribution across the instance
- Most-reported users and posts
- Ability to browse trust connections between users (useful for detecting coordinated
  behavior)

---

## 13. Dark Mode

The CSS supports `prefers-color-scheme: dark` via CSS variables in `app.css`. But there's
no manual toggle. The dream is a theme switcher (Light / Dark / System) stored in
localStorage, with a smooth transition animation.

---

## 14. Search Improvements

The current search endpoint exists (`GET /api/v1/search`) but the frontend has only a basic
search box. The dream:

- Full-text search across posts, users, communities, and tags.
- Filters: by date range, by community, by content type, by trust score.
- Search results sorted by relevance, recency, or popularity.
- Autocomplete for tags and usernames.

---

## The Meta-Dream: Building Any of This

Despite the length of this list, most items are medium-sized features (2-5 files, ~200 lines
each). The architecture is simple enough that a competent developer can implement any single
item in a weekend. The hard part is not the coding — it's deciding what to build and in what
order.

If you want to help, pick the item that excites you most, open an issue on GitHub, and
start building.
