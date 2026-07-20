# What Comes Next

Threadlight is not finished. There is a lot more we want to build. This page lists the
things we hope to add one day, if we have the time and energy.

None of these are promises. They are just ideas we think would be cool.

---

## Things We Could Build Soon

### Full Booru-Style Tagging

Tags are powerful already, but we want to make them even better. Tags could be organized
into categories like Artist, Character, General, and Rating. Misspelled tags could
redirect to the correct spelling. Tagging could be a fun game — people earn reputation
for being the first to tag a post correctly.

### Economy and Achievements

The database already has tables for credits and achievements, but they are not connected
to the user interface yet. The plan:

- **Credits:** Earn them by posting, getting trusted, and helping with moderation. Spend
  them on promoting posts or creating invite-only communities. Credits could eventually
  use a tamper-evident transaction log.
- **Streaks:** Get rewarded for visiting multiple days in a row.
- **Gifting:** Send credits to other users.
- **Badges:** Get a badge for your first post, for reaching Level 4, or for being active
  for 100 days.

### Dark Mode Toggle

The website already supports dark mode if your computer or phone is set to dark mode. But
there is no manual switch yet. We want to add a theme picker (Light / Dark / System) and
remember your choice.

### Frontend Cleanup

The user interface works, but it could be much better. We want to:

- Replace the 16-if-statement page system with a proper router
- Split the 672-line settings page into smaller pieces
- Stop swallowing errors silently — show them to the user
- Add loading animations and "load more" buttons everywhere

---

## Things For Later

### Federation (ActivityPub)

One day Threadlight might speak ActivityPub, the language that Mastodon and other
platforms use to talk to each other. This would let Threadlight users follow people on
other servers, and vice versa. The architecture already supports it — every user already
has the right kind of keys — but the actual messaging code is not written yet.

### Shared Blocklist Exchange

The shared blocklist exists as a JSON file. A future version could let trusted servers
automatically share blocklists with each other.

---

## The Big Picture

Most of these features are medium-sized — a couple of files and a couple hundred lines
of code. Any competent programmer could build any single item in a weekend. The hard part
is not the coding. It is deciding what to build and in what order.

If you want to help, pick the item that excites you most and start building.
