# Threadlight Roadmap

This page lists features we want to build. They are not promised or scheduled — they are
things we think would make Threadlight better if someone had the time and energy.

---

## Near-Term (Could Build This Weekend)

### 1. Full Booru-Style Tag Curation

Collaborative tagging exists. The next phase adds categorization (Artist, Character,
Copyright, General, Meta, Rating), tag aliases for misspellings, and tag implications
that automatically add parent tags. Tagging becomes a game — users earn reputation for
being the first to tag a post correctly.

### 2. Economy & Achievements

Credits and achievements already exist in the database schema but are not wired into the
UI. The plan:

- **Credits**: Earned by posting, getting trusted, moderating fairly. Spent on promoting
  posts or creating invite-only communities. Credit ledger uses a hash-chain for
  tamper-evident audit (each transaction references the previous one's hash).
- **Streak bonuses**: Consecutive-day activity rewards.
- **Giftable**: Credits can be sent to other users.
- **Achievements**: Badges for milestones — "First Post," "Level 4," "100 Days Active."

### 3. Dark Mode Toggle

The CSS already supports `prefers-color-scheme: dark` but there is no manual toggle. A
theme switcher (Light / Dark / System) stored in localStorage.

### 4. Frontend Refactor

- Replace the 16-branch `{#if}` router with `<svelte:component>`
- Split the 672-line SettingsPage into sub-components
- Fix all silent error catching
- Add pagination and loading states everywhere

---

## Longer-Term

### 5. Federation (ActivityPub)

Threadlight could speak ActivityPub, becoming compatible with Mastodon, Lemmy, and Piefed.
The architecture is designed for this — every user has an ed25519 keypair and an `actor_id`
— but federation code is not yet implemented.

### 6. Shared Blocklist Exchange

The shared blocklist format exists. A future version could add automated exchange between
trusted instances.

---

## Meta

Most items here are medium-sized features (2-5 files, ~200 lines each). Pick the one that
excites you and start building.


