# Polaris Codebase — Ultra-Compressed Reference (~2,000 words)

## Stack
**Backend**: Go 1.23, Gin (HTTP), pgx v5 (PostgreSQL), go-redis, zerolog (logging), golang-jwt.
**Frontend**: Svelte 3/4 (with `runes: true` in config but no runes used), svelte-spa-router (hash-based), Vite, plain CSS, Vitest.
**Deployment**: Single Go binary + separate Vite-built SPA. PostgreSQL 16 + Redis 7. Cloudflare Tunnel for public access.

---

## Database (PostgreSQL) — Key Tables

**users** (25 cols): id, username, email, password_hash, display_name, bio, trust_score (0-10), trust_level, reputation, invited_by (FK self), invite_code, credits, onboarding_stage (0-4), is_admin, is_local, location_hash, created_at.

**posts** (23 cols): id, author_id (FK users), title, body, content_type (0=text,1=link,2=image,3=video), mood (0=none,1=curious,2=excited,...), is_educational, is_entertaining, is_nsfw, content_warning, interaction_count, cumulative_interactions, status, scheduled_at, archived_at, is_deleted, created_at, updated_at.

**communities** (11 cols): id, name, slug (UNIQUE), description, rules, owner_id (FK users), is_invite_only, slow_boot_days, forked_from (FK self), created_at.

**trust_connections** (7 cols): id, truster_id, trustee_id (both FK users), weight, signature, created_at, expires_at.

**tags** (5 cols): id, name, slug (UNIQUE), post_count, created_at.
**post_tags**: post_id FK + tag_id FK (M:N join).

**community_notes** (9 cols): id, post_id FK, author_id FK users, body, is_verified, notes_count, upvotes, downvotes, created_at.
**community_note_votes**: id, note_id FK, user_id FK, vote (SMALLINT), created_at.

**reports** (7 cols): id, reporter_id FK, reported_id FK, reason, status, created_at, resolved_at.
**jury_votes**: id, case_id FK reports, juror_id FK users, vote, created_at.
**moderation_actions**: id, case_id FK reports, action (warn/suspend/ban), duration, created_by, created_at.

**subscriptions** (3 cols): user_id FK + tag_id FK + created_at (for tag following).

**site_config** (single row): name, description, contact_email, registration_mode, prune_age_days, prune_min_interactions.

Full schema: `internal/db/migrations/001_initial.sql` (586 lines, includes indexes + FKs with cascades).

---

## Backend Architecture

### Request Flow
Browser → Gin Router (`internal/api/router.go`) → Auth Middleware (JWT validation, sets `c.Set("user_id", uid)`) → Handler (parse input, call service, return JSON) → Service (business logic + raw SQL via pgx) → PostgreSQL.

### Router (`internal/api/router.go` — 327 lines)
All routes wired in `NewRouter(pg, rdb, jwtSecret)`. 23 services instantiated, routed to handlers. Groups:
- `POST /api/v1/auth/register, /login` (no auth)
- `apiV1 := r.Group("/api/v1")` with `RequireAuth()` middleware
  - `GET /posts` (feed), `POST /posts`, `GET/PUT/DELETE /posts/:id`
  - `GET/PUT /users/me`, `GET /users/:id`, `POST /users/me/trust/:id`
  - `GET /communities`, `POST /communities`, `GET /communities/:id`, `POST /communities/:id/join`
  - `GET /tags`, `GET /tags/:id`, `POST/DELETE /tags/posts/:post_id/tags/:tag_id`
  - `GET /search` (posts, users, tags, communities)
  - `GET /feeds`, `POST /feeds`, `GET /feeds/:id`
  - `GET /trending`, `GET /trending/:type`
  - `GET /notifications`
  - `GET /mod/jury`, `POST /mod/jury/:id/vote`, `GET/POST /mod/community`
  - `POST /posts/:id/notes` (community notes)
  - `GET /about` (instance info)
  - `GET /admin/stats`, `PUT /admin/site-config`
  - Admin seeding: `POST /api/v1/admin/seed`, `POST /api/v1/admin/seed/tags`

**IMPORTANT**: No onboarding routes exist (POST /onboarding/tags, GET /onboarding/suggested-users, POST /onboarding/stage all return 404). Tags search route missing. These are dead features in the frontend.

### Handler Pattern (`internal/api/handlers/*.go` — 21 files)
Each handler is a struct holding its service dependency. Methods do: read input (c.Param, c.ShouldBindJSON, c.Get("user_id")), call service, return c.JSON.

Common errors: model.ErrNotFound (404), ErrForbidden (403), ErrUnauthorized (401). Test behind the `handlers_test.go` pattern.

### Service Pattern (`internal/services/*.go` — 21 files)
Each service is a struct holding `pg *pgxpool.Pool`. Every method takes `ctx context.Context` first. Raw SQL (no ORM). Errors are model.Err* sentinels.

**IDOR FIXED**: UpdatePost, ArchivePost now check `author_id == userID` before mutating. NULL bio scan bug fixed (Bio *string).

### Auth Flow
Register → bcrypt hash → INSERT users → JWT created (24h expiry). Login → bcrypt compare → return JWT. Middleware validates JWT signature + expiry, extracts user_id into context. Auth handler also has `GetUserFromToken`.

### Middleware
- `CORS("*")` — permissive for dev
- `Logging(logger)` — zerolog request logging
- `RequireAuth()` — JWT validation, sets user_id in context
- `OptionalAuth()` — sets user_id if JWT present, continues if absent (for public endpoints)

---

## Frontend Architecture

### Tech Stack
Svelte 3/4 (with `runes: true` in svelte.config.js but NO component uses runes — mismatch), svelte-spa-router (hash-based), Vite dev server (proxies /api to Go backend on :8080), plain CSS (296 lines global + per-component style blocks), Vitest + jsdom for testing.

### Routing (`web/src/App.svelte`)
No router component. Uses 16 `{#if}` branches on a `page` string variable. Navigation via `goTo(page, params)` which mutates top-level state. No URL hash sync — browser back/forward doesn't work. Pages mounted with `<svelte:component>` would be cleaner.

### Component Tree
```
App.svelte
├── Navbar.svelte (top bar, auth state, nav links)
├── {#if page === 'feed'} → FeedPage.svelte (534 lines — too big)
├── {#if page === 'profile'} → ProfilePage.svelte (497 lines)
├── {#if page === 'settings'} → SettingsPage.svelte (672 lines — monolith)
├── ... 13 more #if branches
├── TrendingSidebar.svelte (right column, hardcoded data)
└── Footer.svelte
```

### Stores (`lib/stores/`)
- **auth.js**: currentUser (writable), token, isAuthenticated (derived), login/logout/register functions
- **feed.js**: posts, feedType, page, pagination (feedOffset, loadMore — but pagination NEVER called from UI)
- **notifications.js**: notifications array, unreadCount, fetchNotifs, markRead

### API Client (`lib/api/client.js` — 700+ lines)
Every method is `request('METHOD', '/api/v1/...', body)`. 20+ wrapper methods, many unused. Generic pattern: `const token = get(authToken); return fetch(url, { headers: { Authorization: token ? `Bearer ${token}` : '' }, body: JSON.stringify(body) }).then(r => r.json())`. No timeout, no retry, no request cancellation.

### Known Frontend Problems
1. **Svelte 5 runes mismatch**: Config has `runes: true` but all components use legacy syntax (`export let`, `$:`, `on:click`, `createEventDispatcher`). Only `Counter.svelte` uses `$state()`. Fix: remove `runes: true`.
2. **Silent error catching**: FeedPage, SearchPage, CommunitiesPage, OnboardingWizard all have `catch(e) {/* ignore */}` — users never see errors.
3. **Missing pagination**: feed.js has pagination code but FeedPage has no "load more" button.
4. **Hardcoded data**: LeaderboardPage (359 lines entirely fake), ModCenter (half fake), TrendingSidebar (fake users).
5. **Settings monolith**: 672 lines handling 4 independent features in one file.
6. **XSS risk**: `{@html notif.body}` in NotificationsPage (line 88).
7. **Broken onboarding**: Calls backend endpoints that don't exist.
8. **Duplicate CSS**: LoginPage + RegisterPage duplicate 150+ lines of auth form styles.
9. **Weak tests**: Only 4 test files (12 tests) for 5,300+ lines of frontend code.
10. **Unused imports**: Many .svelte files import stores/components they don't use.

### Page List (16 routes)
FeedPage (534), ProfilePage (497), SettingsPage (672), LoginPage (216), RegisterPage (138), SearchPage (234), CommunitiesPage (170), CollectionsPage (197), NotificationsPage (187), LeaderboardPage (359 — fake data), AdminPage (327), AboutPage (123), OnboardingWizard (331 — broken), ModCenter (529 — half fake), BlockedUsersPage (56), AdminPage/seed.

---

## Background Workers (`internal/worker/`)

Scheduler runs jobs in goroutines: feed_refresh (30s), trending (60s), trust_decay (6h — stub), prune (1h — env-configurable), achievements (10m), notes_promote (5m), credits (1h), active_stats (15m), scheduled_posts (30s). Workers accessible at `/admin/stats`.

---

## Critical Missing Features (for LLM refactoring context)

| Missing | Impact |
|---------|--------|
| Onboarding backend (3 routes) | OnboardingWizard silently fails, users stuck on step 1 |
| Tags search route | Tag search returns 400 (routes to /tags/:id) |
| Real leaderboard data | LeaderboardPage shows fake data |
| Pagination UI | Feed never loads more than initial batch |
| Error surfacing | 5+ pages silently catch API errors |
| Settings modularity | 672-line monolith |
| Invite code UX | RegisterPage has no guidance |
| Backend search endpoint | SearchPage may return empty |
| Moderation API parity | ModCenter may call non-existent endpoints |

---

## Key File Reference

| Path | Role |
|------|------|
| `cmd/polaris/main.go` | Entry point: env vars, DB connect, seed admin, serve frontend |
| `internal/api/router.go` | All 327 route definitions, service/handler wiring |
| `internal/api/handlers/*.go` | 21 handler files (1 per resource) |
| `internal/services/*.go` | 21 service files (1 per domain) |
| `internal/model/*.go` | Go structs (User, Post, Community, etc.) + error sentinels |
| `internal/db/migrations/001_initial.sql` | Full schema (586 lines, 30+ tables) |
| `internal/worker/scheduler.go` | Worker registry + goroutine manager |
| `web/src/App.svelte` | App shell + 16-branch router |
| `web/src/app.css` | Global CSS (variables, dark mode, layout grid) |
| `web/src/lib/api/client.js` | All 20+ API call functions |
| `web/src/lib/stores/` | 3 Svelte stores (auth, feed, notifications) |
| `web/src/routes/*.svelte` | 16 page components |
| `web/src/lib/components/*.svelte` | Navbar, Footer, TrendingSidebar |



Reddthat

    Communities
    Create Post
    Create Community

I've been working on a social platform and it's finally at a point where it actually runs on my laptop.
PumpkinDramaM to PumpkinDrama ·

I’ve been working on a social platform called Polaris and it’s finally at a point where it actually runs on my laptop.

The core idea is: what if a social network was designed to be popular but positive? No outrage amplification, no influencer hierarchy, no shadowy moderation. Instead:

    Trust-based moderation — you earn the ability to moderate through positive contributions. Permanent bans require a 90% jury consensus. A random, balanced panel reviews every ban proposal. Every moderation action is public in an append-only log.

    Collaborative tagging (booru-style) — any trusted user can tag any post. Tags aren’t controlled by the author. If the community disagrees on a tag, a vote is triggered among experienced taggers. The community curates the taxonomy together, wiki-style.

    Auto-generated circles — the system groups nearby people who share your interests. “5 people near you also like astrophotography” — you accept or dismiss.

    Community Notes (Twitter/X-style consensus warnings) — weighted by trust score, with a diversity requirement to prevent coordinated voting. Authors can also add correction notes that always appear with a green banner.

    Algorithmic feed you can fully override — the default feed blends recency, engagement velocity, tag affinity, and trust network signals. But you can build your own custom feeds from any mix of tags, users, and keywords with include/exclude rules. RSS export for every feed, tag, collection, and community.

    Communities with slow-boot, curators, and forking — topic-based public spaces. New members can be required to go through a probation period. Curators enforce scope tags but can’t ban anyone. If the community disagrees with the direction, members can fork it — creating a new community with the same tag scope and member list.

    Credit economy — daily rewards with streak bonuses, giftable credits between users, bounty system for questions. Fully auditable ledger using a hash chain in Postgres. Not a real blockchain — no mining, no gas, no wallet.

    Achievements and multi-period leaderboards — daily, weekly, monthly, yearly, and all-time rankings for posting, tagging, gifts, trust network, and achievement points. Achievements focus on positive behaviors like curation, quality contributions, and community service.

    Advanced search — full boolean operators (AND, OR, NOT), field search (title, author, tag, date range), exact phrase matching. Saved searches with new-result notifications.

    Collections with four visibility layers — private, public read, public edit, and anybody edit. Perfect for collaborative link collections, reading lists, or community wikis.

    Per-user content filters — hide, blur, or deprioritize content matching keywords, regex patterns, domains, or specific users. Processed both server-side (feed ranking) and client-side (blur overlay).

    Trending topic detection from content, not hashtags — the system automatically detects rising terms by analyzing post text and comparing frequency against baseline. No hashtag spam gaming the algorithm.

    Onboarding wizard — new users pick 5+ tags of interest, follow suggested users, and optionally make their first post before they see the main feed. The algorithm needs data to work with.

    Content warnings and post scheduling — authors can add content warnings (blurred until clicked). Posts can be drafted, scheduled for future publication, or published immediately.

    Shared blocklist for federation — domains, IP prefixes, and fingerprint-based blocks can be exported as JSON for other instances to consume.

    ActivityPub-ready — every user has an ed25519 keypair and an actor_id. The data model supports federation without migration. Public keys are already stored for HTTP Signatures.

Tech stack is Go + Gin + Svelte + PostgreSQL + Redis. Single lightweight binary.

Would love feedback on it.
Body

    0ndead
    @infosec.pub
    English

    How much did AI code?

PumpkinDrama
PumpkinDrama
Community locked: only moderators can create posts. You can still comment on posts.
Visibility: Public

This community can be federated to other instances and be posted/commented in by their users.

    3 users / day12 users / week12 users / month58 users / 6 months1 local subscriber55 subscribers44 Posts11 CommentsModlog

    mods:
    PumpkinDrama

    UI: 0.19.19
    BE: 0.19.19-jemalloc
    Modlog
    Legal
    Instances
    Docs
    Code
    join-lemmy.org


programming.dev

    Communities
    Create Post
    Create Community

    1

How would you design online communities so they don't degenerate over time?
CoderSupreme to ProgrammingEnglish ·

The pattern is always the same:

    Someone passionate creates a community
    Early members are deeply invested — high-quality discussions, strong norms
    Community grows, hits the front page / “all”
    New users flood in, post memes, low-effort content, drive-by hot takes
    Old-timers get drowned out
    Mods either burn out and quit, or become dictators ruling their own fiefdom

Reddit’s model accelerates this. Any subreddit that gets popular enough hits the All feed, and at that point the incentives shift from “quality contribution” to “what gets upvotes from the broadest audience.” The people who built the community lose control of it.

I’m working on a social platform and I’m trying to design communities that don’t go through this lifecycle. This is what I have in mind right now. A user’s voting power within a community is proportional to their tenure in that specific community. So if someone’s been there for 2 years, their vote counts more than someone who joined yesterday. No matter how many new users show up, they can’t swing the community away from the people who built it.

The weighting wouldn’t be dramatic, maybe a logarithmic curve where the first month gives you baseline power and it grows slowly over years. The goal isn’t to create an aristocracy, just to make sure the signal from committed members isn’t buried by a flash mob.

Implementation-wise I’m thinking:

    Each community membership has a joined_at timestamp
    When scoring a post within that community’s feed, each boost/vote is multiplied by log2(days_since_join + 1) / log2(30) — so 1 day = 0.05x, 30 days = 1x, 1 year = ~1.7x, 5 years = ~2.1x
    The algorithm still surfaces new content to everyone, but the ranking is weighted toward longer-tenured members’ tastes
    Also considering: a slow-boot period where new members’ posts are held for review by existing members (like a probation phase)

But I’m second-guessing myself. Some concerns:

    Does this just create a gerontocracy where old members gatekeep forever?
    How do you handle the first 30 days when no one has much weight?
    Would this actually prevent the decay, or just slow it down?
    Is there a simpler/better mechanism I’m missing?

Curious what other approaches people have seen work (or fail). Has anyone implemented tenure-weighted voting before?

Edit: I appreciate the suggestions. A few of them actually align with features I’ve already built:

    Trust graph: The platform already uses directed trust edges, with vouching and inviter accountability (if an invitee misbehaves, the inviter is penalized). I can extend this model from the platform level down to individual communities.

    Invite-only mode: I already have trust-based monthly invite limits for the whole platform, so adding the same option for communities seems like a natural extension.

    Slow-boot / probation: This is already in place, new members’ posts are held for curator review before being published.

    Different interaction types: The system already supports multiple interaction types (like, comment, share, gift, emoji). I’m not sure how to expand these into Slashdot-style categories (e.g. agree/disagree, insightful/funny, quality/shitpost, high/low effort, etc.), but it’s something I’d like to explore.

    Moderation limits: I like Slashdot’s approach of limiting moderation to prevent burnout. I’m less certain whether meta-moderation would be a good fit here, but I’m open to considering it.

Body

    Marija_

    Strong community norms matter more than clever ranking algorithms, in my experience.
    bitfucker

    I would like to add something too because I only remember it just now. But I recall someone mentioned that bluesky moderation is unique and I quite like the idea but I’ve never used bluesky myself so idk how true it is. Basically, an account chooses who to trusts to be a moderator for said account. So that content that is curated by said moderator is the one that the account sees. If the account opt to not be moderated then it can see everything
        CoderSupreme
        OP

        Sorry, I couldn’t understand that moderation system. I’ve made a post about moderations systems here.
    schnurrito
    @discuss.tchncs.de

    We had a solution to this problem in the 2000s and early 2010s when we were on topic-specific forums where each (broad) topic was a separate website where you needed a separate account.

    But that had its disadvantages too…
    jtrek
    @startrek.website

    I wouldn’t do it based on join time. Age doesn’t always map to quality, and it might encourage people to make sleeper accounts to age up.

    I would do it based on a graph of trust.

    Let’s say you’re user 0. You trust yourself. You can then invite or vouch for a user. If they misbehave, it should reflect badly on you. If they post good content, you should share some of the glory. If one of those users invites someone, the good and bad flow upstream.

    People no one has vouched for have very little power. You can filter them out with a toggle.

    You’d be able to see the web of behavior. If someone keeps inviting or endorsing assholes, then they’re a problem.

    Endorsing is different than upvote. You can upvote whatever, but endorsing is saying you trust this person and you’re willing to put your own reputation and access on the line.

    Something like that.
        CoderSupreme
        OP

        Good thinking. The tenure wasn’t gained when idle and was lost at 5% per week after a month of inactivity, but I like using the same trust system I had designed for the platform for each community as well.
    Modern_medicine_isnt
    @lemmy.world

    Anything that involves more than a few hundred people goes bad. So create a new community every time you hit a certain criteria.
        CoderSupreme
        OP

        I already added community forking but it copies all the users, I’ll have to see how I can make the forker change the spirit of the new community.
    talkingpumpkin
    @lemmy.world

        Does this just create a gerontocracy where old members gatekeep forever?

    I don’t think it determines a gerontocracy (do people really care about upvotes?) and I don’t think it solves the problem you’re worried about either.

        How do you handle the first 30 days when no one has much weight?

    You show upvotes as a %, so if a post has +.01/-0 it’s still 100% upvoted.

        Would this actually prevent the decay, or just slow it down? Is there a simpler/better mechanism I’m missing?

    In your scenario, “hardcore” members leave because the signal/noise ratio is so low that the community is not worth following anymore.

    That has nothing to do with upvotes and all to do with mods failing to take down meme posts (assuming members don’t like memes).

    A community is its mods: if they allow low-effort posts (without meaning to) they are not doing their job properly, and that’s probably due to the job being too difficult and/or requiring too much effort.

    You could look into allowing older members to help moderate, or maybe into allowing mods to flag “trusted” users whose downvotes (possibly unbeknown to the users themself) will result in posts being taken down “for review”, or otherwise figure out some other way to help mods out, both when making up the rules (that too takes a lot of effort!) and when moderating posts/comments.

    A lemmy/reddit-style community is its mods. Think of how to help them be effective.
    atzanteol
    @sh.itjust.works
    English

    Invite only or a fee to join. Keep it small. No anonymous accounts.

    Just like in-person clubs.

    Humans aren’t adapted to interacting with thousands of anonymous people at a time. It brings out the worst in people (Gabe’s greater Internet fuckwad theory).

    You can deal with “that jerk” but it’s exponentially harder to deal with dozens or hundreds of jerks in a pile on.
        Skullgrid
        @lemmy.world

        Then you have no content, no reason to use, no network effects. People get bored and leave
            atzanteol
            @sh.itjust.works
            English

            Right. On-line communities that work well are effectively impossible.
                staircase

                I’m starting to wonder if “online community” is a contradiction in terms.
        queerlilhayseed
        @piefed.blahaj.zone
        English

        What makes an account non-anonymous?

        I wouldn’t say there’s no room for non-anonymous services, I run a few of them for my family. But one of the cool things about the internet is that you can talk to people (a lot of people) you would have never interacted with without the internet to facilitate.

        I think the Internet Fuckwad Theory is bunk. There are lots of people on the internet who are fully anonymous and still capable of being chill, empathetic, decent human beings. I would estimate that to be the majority of internet users, the majority of the time. Fuckwaddery just has a tendency to draw attention to itself and strain everyone else’s time and patience. And I think OP is right, I think Fuckwaddery is partially design dependent and that most platforms design for because enragement is engagement. I think it’s highly likely that there’s a forum design out there that mitigates it, even if it doesn’t eliminate it entirely.
            atzanteol
            @sh.itjust.works
            English

                Fuckwaddery just has a tendency to draw attention to itself and strain everyone else’s time and patience.

            Bingo. Like I said - one jerk is something you can deal with. But lets say only a small percentage of any user population are fuckwads. As you add users the number of fuckwads scales up linearly.

            But an individual person does not scale up. It’s still just you. And you’re now dealing with a lot more fuckwads as the community scales up. That’s not tenable. One person even politely pointing out a mistake you’ve made is “a bit annoying but okay”. dozens of people doing it in an asshole way is infuriating.

                I think it’s highly likely that there’s a forum design out there that mitigates it, even if it doesn’t eliminate it entirely.

            Nobody has discovered it yet. And I honestly doubt it’s possible. We’re not made to communicate at such scale.
                queerlilhayseed
                @piefed.blahaj.zone
                English

                Well, we weren’t made to fly over thousands of miles but we figured it out. I think we can figure out how to scale a free, durable, peaceful community to the globe, and I think we have to if we’re ever going to figure out world peace.
                    atzanteol
                    @sh.itjust.works
                    English

                    Despite all evidence to the contrary eh?

                    Flight is a physics problem. World peace is a human nature problem. And keep in mind that we’ve used flight to wage war more effectively.
                        queerlilhayseed
                        @piefed.blahaj.zone
                        English

                        Despite all evidence to the contrary, I still think it’s possible. Or maybe just hope, because I rather think we (the human species) have to if we’re going to survive.

                        And, the internet isn’t that old, humanity is still getting acclimated to it. I think there’s still a lot of room for improvement on what the internet is, how it functions, and how to make it work well.
                            atzanteol
                            @sh.itjust.works
                            English

                                And, the internet isn’t that old, humanity is still getting acclimated to it. I think there’s still a lot of room for improvement on what the internet is, how it functions, and how to make it work well.

                            This is not an “internet problem” it’s a “humanity problem.”

                            For 10,000+ years humanity has not been able to solve this problem without the internet. The more people you add to a community the more likely you are to have problems. The internet just makes this happen more easily and across national borders.
                                queerlilhayseed
                                @piefed.blahaj.zone
                                English

                                Right, it’s been a problem forever, but it’s only just now that we have the technology to let everyone communicate with everyone else in near real time. that’s huge. There are still a lot of hurdles to overcome, like how to get everyone who wants to be connected connected, and respecting those who choose not to engage with the internet, how to ensure that powerful individuals or governments can’t deny communications between one group and another (for example), but I think it’s a powerful tool for fostering human connection that we have, so far, been using pretty poorly. I think the question of “how can we use the internet better” is one with a lot of excellent answers that we haven’t found yet.
    Skullgrid
    @lemmy.world

    Your ideas, and some in this thread are quite neat.

    One thing to bear in mind when working with systems is that you need to use the right tool for the right problem, and to use people tools for people problems and technical tools for technical problems.

    This is a people problem.

    One of the things that have been successful is branching communities; a good example on lemmy is linux.

    There are communities specifically for linux shitposts and memes and easy communication. There are ones for news. There are ones for noobs. And there are ones for discussion of linux in general.

    This way, you have easy shit that doesn’t drown out the main idea, while still feeding back into and out of the main idea.

    However, this can lead to mod burnout if too many people leak into the wrong area with the wrong posts.

    Something that might work nicely is to allow the user to set up ratios of how often they want to see stuff from whichever theme/variation on that theme. 30% news, 20% memes, 50% general discussion.

    Or they can set it to 90% memes and 10% general discussion, and see stuff that’s different.
    Ftumch
    @lemmy.today

    I think it might help to have different vote types. Instead of up and downvotes there could be things like agree/disagree, insightful, funny, quality shitpost, high/low effort post or things like that. That way people who want to see quality discussion could sort by insightful, while people looking for low effort shitposts or memes could sort by something else.
        CoderSupreme
        OP

            I think it might help to have different vote types. Instead of up and downvotes there could be things like agree/disagree, insightful, funny, quality shitpost, high/low effort post or things like that. That way people who want to see quality discussion could sort by insightful, while people looking for low effort shitposts or memes could sort by something else.

        Slashdot moderation system, I like it, I’ll see how it fits.
            Cricket@lemmy.zip
            @lemmy.zip
            English

            My first thought when I read (skimmed) your post was Slashdot’s moderation system. I feel that there were many really good ideas there that should be attempted again. Limited, random ability to vote, categories of votes (insightful, funny, etc.), meta-moderation (i.e., ask others to review someone else’s votes). I think that those were all good ideas.

            One other thing that Slashdot had that minimized the memes we see here: no images on posts.
            bitfucker

            Implementation wise, I think it would be like tagging. See how many people tag the post as what
            Ftumch
            @lemmy.today

            The comparison to tags got me thinking it might be cool if mods or posters could set custom tags/vote types per community or thread.
                CoderSupreme
                OP

                I don’t get it, could you expand on that? The idea I’m getting is something like “flairs” on Reddit that people vote on. Right now I’m thinking since all posts will probably be tagged thanks to booru-style collaborative curation, people could vote on the tags that are subjective to see which tags apply best to the post. Then we could sort based on some of the tags like the ones you mentioned.
                    Ftumch
                    @lemmy.today

                    For example, an advice community could have a “good answer” vote. Each post in an “Am I the Asshole” community could have a “You’re the Asshole”/“You’re NOT the Asshole” vote. That sort of thing.
                    bitfucker

                    That’s more in line with what I was thinking, yeah. The rest is just hoping that people will not abuse the tag (which, with enough mass, the noise would probably be drowned out by genuine people anyway)
        staircase

        The single vote approach is absolutely awful. Adding nuance is an improvement, but to be honest the whole idea of voting is insane. Imagine doing that to someone talking to you in person. It would be seriously weird at best.

        Voting is a bad system and needs to be dug out from the roots and burned.
            Ftumch
            @lemmy.today

            I’d say it’s similar to applauding or booing someone who’s addressing a crowd.
    MonkderVierte
    @lemmy.zip

    No /all feed?
    astronaut_sloth
    @mander.xyz
    English

    Honestly, this feels like trying to program ourselves out of a fundamentally social/human issue. I think that the best way to design a strong-lasting platform is to not over-complicate it. I would look to successful forums, those that are still running or ran for many years. How did they maintain their community?

    Ultimately, I suspect that the answer ultimately comes down to the growth curve. Slow and steady growth where new users get assimilated into the platform’s culture is likely how this works.

    I think your logarithmic curve is an interesting idea, and if you do implement it, I’d be interested to see the results. However, I don’t know if it will be enough to “maintain” a community. In fact, you may end up with the opposite problem if the algorithm is successful. You’ll have a community that doesn’t change where the oldest accounts rule.
    InvalidName2
    @lemmy.zip

    I’m not saying this to be discouraging, but the more complicated you make it, the more likely it will fail.

    If you don’t want an online community to degenerate over time, then pick a very small time scale to measure (ex: 10 days) or choose easy to measure but ultimately worthless metrics to measure it (ex: # of words in posts that aren’t slang). The fact is, long term communities change over time, and the overwhelming majority of the time, those changes can and are regarded as degenerating.

    A lot of the things are you suggest have the common pitfall of disincentivizing new users. The more hoops and the less reward people get, particularly when some of the hoops feel punitive, the more likely you are to deter new users and less likely you are to cultivate the future foundations of your community.

    Lemmy, and it’s not even all that difficult sign-up hoops, are a perfect example. The place has grown a bit. It would be hard to convince me that there’s been no degeneration. We pick up the ones that Reddit rejected for basically being the worst of the worst along side the small number of decent users who just happen to have a burning dispassion for Reddit. And even so, this place has really just moved closer and closer to being what Reddit used to be – so, hard to say that’s not degenerating a bit compared to the happy fun nice pleasant place it was in the early days when I first joined.

    Okay, that rant aside, the other thing I want to say is that a lot of these suggestions you’ve made and a lot of the suggestions I’ve seen in the comments: They sound good in theory, but they pretty much only work in a vacuum. If there were no real “competition”, no other communities of merit, they’d probably work out alright. But in reality, people have lots of options. You have to work extra hard to incentivize them to come and stay AND you have to avoid all the traps, nuances, etc that are inherent to this kind of thing. It’s a very fine, tricky line that pretty much nobody succeeds with long term. Not because they haven’t tried new and inventive things.

    And so that I’m not just the naysayer doomsday poopoohead: If you really want to make this work, you’d have to keep a small, well-curated community expertly moderated by one or two extremely prescient moderators with super well defined guidelines and who are in total agreement. Basically unicorns, but if at least one unicorn existed. So, it’s a lot of work, doesn’t rely on complicated procedural / algorithmic gimmicks and it’s not really scalable – but technically doable. I’ve seen it done. Not truly long term, because moderators burn out and then things degenerate.

    OOPS: One more addition here – The other pitfall is that communities have to change or be open to change over time. If they don’t or can’t, they stagnate, and rot. Basically, they degenerate anyway.
    HubertManne
    @piefed.social
    English

    that pattern does not apply here. give users as much control as possible to control their feed. that is the best thing to do.
    Eggymatrix
    @sh.itjust.works

    Imo the issue is the scope. The best communities online are independent (no reddit or other) forums dedicated to specific things, often moderated by someone payed to do that.

    If your incentive is growth and income and the community’s incentive is pertinence either you or the community will have a problem, unless you understand how to manage that equilibrium.

    The issue imo is not self regulation mechanisms like voting and weights, but law enforcement and responsability towards the main subject and the community
    Noja
    @sopuli.xyz

    Have a mod bot comment, if that gets downvoted too much, delete the post. This filters all the “front page” voters who only upvote based on content and don’t participate in the community.

Programming
programming

Welcome to the main community in programming.dev! Feel free to post anything relating to programming here!

Cross posting is strongly encouraged in the instance. If you feel your post or another person’s post makes sense in another community cross post into it.

Hope you enjoy the instance!
Rules

Wormhole

Follow the wormhole through a path of communities !webdev@programming.dev

Visibility: Public

This community can be federated to other instances and be posted/commented in by their users.

    319 users / day1.6K users / week2.73K users / month8.55K users / 6 months3.05K local subscribers27.8K subscribers3.43K Posts50.4K CommentsModlog

    mods:
    snowe
    Ategon
    UlrikHD
    bugsmith
    Spyro

    BE: 0.19.18
    Modlog
    Legal
    Instances
    Docs
    Code
    join-lemmy.org


programming.dev

    Communities
    Create Post
    Create Community

    1

What is the best moderation system for a social platform?
CoderSupreme to ProgrammingEnglish ·

One of the main features of my site will be Discourse-style trust levels, the higher your trust level, the more moderation privileges you gain. I’m also considering either a jury system, where multiple users collaborate on moderation decisions, or a meta-moderation system, where users can vote “fair” or “unfair” on moderation actions taken.

Someone mentioned to me in another post that Bluesky also has a very good moderation system, but the best one I’ve found so far is Discourse. What I really like about it is that you earn trust levels by reading posts, not by taking moderation actions. I definitely don’t want to gamify moderation the way Stack Overflow did, that was the worst I’ve ever seen.
Body

    litchralee
    @sh.itjust.works
    English

    “best” is relative, and while moderation can achieve a great deal of success on a smaller scale, it’s proven to be extremely tricky at large scale. Mike Masnick of TechDirt (and on the BlueSky board of directors) has documented this rather well, using some very notable case studies.

    For example, moderation is both necessary and complex (Twitter post buyout): https://www.techdirt.com/2022/11/02/hey-elon-let-me-help-you-speed-run-the-content-moderation-learning-curve/

    Decentralized moderation like Mastodon is good (by blocking Gab from the get-go) but requires active coordination from federated operators: https://www.techdirt.com/2019/07/16/gab-mastodon-challenges-content-moderation-more-distributed-social-network/

    At a slightly larger scale, volunteer moderators can be overburdened and lead to adverse infrastructure shutdowns (Parler getting bumped for violating AWS policies): https://www.techdirt.com/2021/01/13/parlers-laughably-bad-antitrust-lawsuit-against-amazon/

    My point with these examples is that a moderation system solely based on volunteer effort can be: 1) overwhelmed, 2) manipulated, or 3) antithetical to the purposes of the community and/or platform. There is no magic moderation system that will be 100% effective for your environment at all times.
        CoderSupreme
        OP

        I think we can most agree that while “best” may be relative Disourse is definitely better than StackOverflow in regards to moderation system.
            litchralee
            @sh.itjust.works
            English

            By what criteria do you say that? You might be right, but without any qualification for how one might be better than another, it’s as useful as saying “purple is better than yellow”.

            But ultimately, you’re looking for a moderation paradigm and don’t owe me or anyone else a criteria or explanation for what you ultimately choose. It merely means that the question as posed is unanswerable, no different than “why is green the best color?”.
    CameronDev

    Had a quick look at your doc, and the juries idea does not sound good.

        Moderation is not something loads of people want to do. Take a look at the list of unmoderated communities: https://programming.dev/u/Vacant - Forcing random users to be involved would likely annoy your users.

        For certain types of reports, CSAM, gore, etc, you absolutely do not want to share that to random users. If your users are underage, it could be illegal to do that anyway, but sensitive users may not want to be exposed to triggering content.

    Maybe you make being in the jury opt-in, but if most don’t opt in your back to regular old moderation by the few who have the stomach for it.
        CoderSupreme
        OP

        deleted by creator
            CameronDev

                Criticizing that system without offering any better alternative isn’t useful.

            Disagree. Pointing out flaws in a proposal is important, it raises awareness so that they can be resolved.

                In the jury system, content could be hidden as soon as it is reported

            True of all moderation systems, but is open to abuse. Removing content, even temporarily is good way to censor opinions, by hiding them until they fall off the “new” page. Moderation usually happens quite slowly if the mods arent terminally online. A jury system may exacerbate that if the jurors aren’t available to respond in time.

            Penalising false reports also risks discouraging reports, and may actually just encourage “group think”. And if accounts are easy to spin up, there is no real cost to abusing reports to hide content.

            I actually think the current systems are actually adequate. We have a moderators who can be replaced at any time by admins, which gives oversight, and we have the ability to simply move communities to other instances if that fails.
                CoderSupreme
                OP

                deleted by creator
                    CameronDev

                    5 posts in the last week vs 100s of moderation actions taken in the last week. Thats actually a really good hit rate, which suggests that the system actually works really well.
    Sir_Kevin
    @lemmy.dbzer0.com
    English

    Slashdot did it best.
    bitfucker

    This is how bluesky does it. So as I mentioned earlier, it is the user that opt to be moderated and they can choose their moderator.
    0xtero
    @beehaw.org

    I tend to like systems that give the user tools to moderate their own feed and content they see. Then it’s up to me to decide what I want in my timeline and what I don’t want. This can be a bit heavy on the system though, especially if you don’t have lot of metadata and have to rely on full-text search blocking. Seems to work on Mastodon though.

    This then leaves only certain moderation tasks to the server admins (automating CSAM scanning/reporting) - and filtering out content that breaks the server ToS.

    Back in the days, Slashdot had a system where you could “opt-in” as a moderator, and this would randomly give you “mod-points” every now and then. You could use your points to up- or downvote posts and comments. I quite enjoyed it and it didn’t feel like I had to apply for a 2nd job (which is the case with lot of other moderation systems)

Programming
programming

Welcome to the main community in programming.dev! Feel free to post anything relating to programming here!

Cross posting is strongly encouraged in the instance. If you feel your post or another person’s post makes sense in another community cross post into it.

Hope you enjoy the instance!
Rules

Wormhole

Follow the wormhole through a path of communities !webdev@programming.dev

Visibility: Public

This community can be federated to other instances and be posted/commented in by their users.

    319 users / day1.6K users / week2.73K users / month8.55K users / 6 months3.05K local subscribers27.8K subscribers3.43K Posts50.4K CommentsModlog

    mods:
    snowe
    Ategon
    UlrikHD
    bugsmith
    Spyro

    BE: 0.19.18
    Modlog
    Legal
    Instances
    Docs
    Code
    join-lemmy.org



I asked the llm to implement discourse style trust level moderation and web of trust as two separate things but it merged them into a single thing. also in the docs the llm said things about the platform being open source and having no algorithm where users take all decisions and that's just not what I asked for. What I want is a platform that tries to bring people together and as popular as possible to compete directly with twitter but without pushing outrage or allowing fascism or bots pretending to be human, twitter is full of those and it makes me think everyone in the platform is just a bot so I end up not interacting with anyone. I also want booru-style image board tag curation, so that every tag is tagged, subjective tags allow votes, images get alt descriptions. leaderboards to promote all actions like tagging, posting, commenting, etc. I want threaded collapsable comments like reddit. communities like reddit too but that don't degenerate over time to appeal to the lowest common denominator. I wrote a bunch of features I like, make a new design for this platform, remember that the most important thing is bringing people together, not pushing outrage, not allowing fascism or bots pretending to be humans.
