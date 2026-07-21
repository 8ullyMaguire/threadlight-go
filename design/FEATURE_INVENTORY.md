# Threadlight — A Living Map of What We Have and What Comes Next

This document tells the whole story of Threadlight — what works today, what is halfway to completion, what we dream of building someday, and how we think about what to do first. It is not a checklist or a spreadsheet. It is a living map of a community that is growing every day, made by many hands working together.

## What Already Works

The core of Threadlight is solid, working, and ready to welcome people. When someone arrives for the first time, they can sign up with an email and a password, verify their email address, and set up a profile that feels like them — with a display name, a short bio, an avatar, a banner image, a theme color, and privacy settings they control. They can write posts with Markdown formatting, add content warnings for sensitive topics, schedule them to appear at just the right time, and edit them after publishing. The feed shows everything in the order it was posted, with pagination that works smoothly on any screen size. A full rich Markdown renderer handles headings, lists, links, images, tables, code blocks, spoiler tags, subscripts, superscripts, strikethrough, blockquotes, and embedded media.

Communities are the beating heart of the platform. Anyone can create one, join one, or leave one. Inside each community, a trust progression worker runs every hour and automatically promotes members when they meet the criteria — account age, post count, reaction ratio, and clean moderation history. Communities support slow-boot mode for gradual growth, curators who enforce tag rules, forking when the community's direction changes, invite-only access with trust minimums, a trust graph, and a trust score API.

Search is surprisingly powerful. It understands queries with AND, OR, NOT, parentheses, and quoted phrases. It can filter by author, tag, community, mood, and date range. Behind the scenes it uses PostgreSQL tsvector indexes with weighted title and body fields, GIN-indexed full-text search on posts, and trigram-based autocomplete on usernames, display names, community names, and slugs — so typing a few letters brings up the right results. There are dedicated search endpoints for posts, users, and communities, plus a suggestion endpoint for autocomplete. The search parser lives in its own package and has its own test suite.

Tags are fully functional. Anyone can create, edit, and delete them. Any member can add tags to any post. The community votes on how useful each tag is through the TagVote component that handles both upvotes and downvotes right where posts are displayed. Trending topics are calculated over regular time windows and updated every minute by a dedicated worker.

Interactions support multiple types beyond a simple like: the model is ready for insightful, helpful, beautiful, funny, wholesome, and needs-sources reactions. Liking a post triggers the event-driven affinity accumulator and updates interaction counts in real time.

When someone needs space from another person, blocking works on an individual level and also through shared blocklists that can be imported and exported as JSON — communities that trust each other can coordinate to keep harmful people out. Hidden posts let each person hide individual posts they do not want to see. Content filters let each person hide posts containing words they would rather not see, using simple keywords they choose for themselves. If someone wants to leave the platform entirely, full account deletion removes everything cleanly.

Moderation is built on transparency and shared responsibility. Labels let moderators mark posts as insightful, misleading, off-topic, trolling, or flamebait, so everyone can see why something was flagged. Administrators can ban or suspend accounts when someone genuinely harms the community. Anyone can report content they are worried about, and reports flow into a queue for careful review. Every moderation action is open to public review — community members vote on whether each decision was fair or unfair, and a controversy score surfaces the most hotly debated calls so everyone can see where opinions differ. When a moderation action receives enough unfair votes from the community — by default, at least five reviews with at least seventy percent ruling unfair — the moderator who took the action automatically loses a configurable amount of trust. The penalty amount, threshold percentage, minimum votes required, and cooldown period are all adjustable through the admin settings, so each community can set its own tolerance. A cooldown prevents the same moderator from being penalized again within a configurable window, so one bad day does not erase years of earned trust. The penalty is applied transparently: the API response includes a human-readable explanation each time a penalty is triggered. The backend endpoints exist for the full moderation cycle: actions, jurors, jury voting, resolution, mod log, public review voting, controversial decisions, and trust penalty evaluation.

A whole economy of recognition is already running. Credits can be transferred between people, spent on community features, and tracked through a full transaction history. Daily rewards are claimed and tracked per-user per-date. Bounties can be created on posts, awarded, and expired by a background worker. Achievements can be checked and awarded when someone reaches a milestone, with progress tracking and visibility settings. A leaderboard API tracks scores, ready to be connected to a frontend that shows real data.

Behind the scenes, background workers keep everything humming — they rebuild feeds, calculate trending topics, decay trust for inactive members, prune old data, check achievements, promote trust levels, process community notes, award credits, update active statistics, publish scheduled posts, compute affinity between users, refresh algorithmic lists, and clean up user lists with no active subscribers. A WebSocket connection manager is already built, ready to light up live features when we wire it in. Community notes let people fact-check posts with voting, with consensus scoring and trust-weighted vote influence.

User lists are fully functional down to every detail. You can create and manage lists, add and remove members, subscribe and unsubscribe with batch follow or block action for every member on the list. Lists support public, private, and shared visibility. Collaborators can be invited with editor or viewer roles, an invite flow with accept/decline, and the owner can remove any collaborator at any time. Algorithmic lists populate themselves based on rules like minimum trust level, minimum posts in the last thirty days, maximum negative ratio, and tag affinity. They refresh automatically respecting each list's configured interval — on creation they are populated immediately, then daily afterward. Lists that have no active subscribers are automatically cleaned up daily. Four handler integration tests cover all of this.

The user affinity engine is event-driven — it uses Redis counters for instant badge lookups and a twenty-four-hour cache for the full affinity page. It scores connections across four dimensions: tag overlap, shared communities, reaction agreement, and trust distance. Similar users can be discovered through a dedicated endpoint.

Feed plugins run inside a WASM sandbox using the wazero runtime with a ten-millisecond timeout and one megabyte memory limit. They can be created, installed, reviewed with ratings and reviews, uninstalled, and executed through a marketplace. A worker handles execution so the main server stays responsive.

The platform has a full invite system with per-user invite limits keyed to trust level, a registration queue for applications, OAuth state management, custom emoji, custom pages (legal, about, code of conduct, etc.), saved searches with notification-on-new, proximity-based interests for circles, and site-wide config that controls every aspect of the platform's behavior.

The technical foundation is solid. Six database migrations run automatically on every startup. A multi-stage Dockerfile makes deployment straightforward. Integration tests cover user lists, algorithmic lists, and affinity endpoints. The frontend has over 130 Svelte components, shared UI primitives (modals, toasts, popovers, menus, tabs, buttons, forms, badges, progress bars, skeletons, search bars, sidebars, navbars, pagination), and dedicated pages for the feed, communities, profiles, search, settings, moderation, the plugin marketplace, themes, the mod log, inbox, saved posts, reports, and admin.

## What Is Halfway There

Several important features have their backend built but are waiting for a frontend to match, or need one missing piece before they become usable.

The password reset flow works but returns the token on screen instead of sending it by email — the email queue table exists, and the model supports it, but SMTP configuration and an email-sending step need to be wired in. Trust progression is actually running hourly as a worker, but the frontend has no way to show a user their current trust level, what they need to do to reach the next one, or that they have been promoted at all. The MODREMOVE endpoint for moderators to delete bad posts exists on the backend but needs a frontend button and a confirmation dialog. The moderator dashboard has backend endpoints for actions, jurors, and votes but needs a frontend that shows the review queue, lets mods act on reports, and tracks their activity. The public moderation log has its backend endpoints (controversial decisions, review voting, my vote lookup) and a frontend page shell exists, but the page needs to be populated with real data from those endpoints.

The M2 meta-moderation system has a jury foundation (panel creation, juror addition, voting, resolution) but needs the "Was this moderation fair?" prompt logic to be fully wired in the frontend. Community curators have backend endpoints to add and remove curators but no frontend to manage them. The appeal system needs both backend and frontend work so users can contest decisions. Image upload has a complete model (avatar_url, banner_url on users) and service methods written but the frontend upload endpoints are not fully connected — the ImageAttachForm, ImageInputModal, ImageInputUpload, and ImagePreviewInput components exist but need storage configuration.

The feed ranking Positive Algorithm needs to be built and tuned. The "Why this post?" button needs a frontend tooltip that shows the ranking factors that brought something into your feed. Custom feeds with RSS export need the export layer added. The full set of reactions — beyond the current like — needs to be activated in the frontend. Tag autocomplete for post creation needs a frontend component that suggests tags as someone types.

WebSocket is fully built — the hub, client manager, and broadcast logic all exist — but nothing is wired to it yet. No live feed, no presence indicators, no real-time notifications. ActivityPub federation has ed25519 keys stored on users, actor_id and public_key columns in the database, and an OAuth state table ready for the handshake, but there is no inbox, no outbox, no HTTP Signatures, no webfinger, and no activity delivery.

Collection visibility layers have the database model with a visibility column but no frontend UI to choose who can read and edit. Auto-generated circles based on interest and proximity need a ranker to sit on top of the manual circle system. The leaderboard frontend shows hardcoded sample data instead of real scores. Achievements have a full backend (definition, unlock, progress tracking) but badges need a home on profiles in the frontend. Credit streak bonuses for consecutive active days need backend logic to track streaks. Giftable credits need both an API and a user interface. The credit hash-chain ledger has the hash column on the transactions table but needs the append-only chain logic.

The search frontend needs a redesign with tabs, filters, snippets, and URL-synced state to match the powerful backend. Error handling across the frontend has many places where errors are silently swallowed instead of shown to users. The monolithic SettingsPage needs to be split into smaller, manageable components with proper navigation.

Seed data — sample users, posts, communities, and tags — would make fresh instances feel welcoming from day one. Test coverage is strong on the backend (with integration tests for lists, affinity, and plugins) but thin on the frontend. A CI/CD pipeline would catch regressions before they reach users.

## Features That Need Work

Some features have been started in the codebase but are incomplete in ways that matter.

Own post archiving exists but the frontend has no action menu item for it. Post locking (preventing new interactions), stickying (pinning to top of community), and language tagging are all columns on the posts table with no frontend controls. The cross-post root tracking and moved-from-community tracking columns exist on posts but the linking UI does not. Mod notes on users have a full table and model but no frontend to view or write them.

The community slow-boot feature has a column but no worker or frontend enforcement. Community member roles beyond basic membership and curator are not utilized in the frontend. The privacy policy, terms of service, and code of conduct URLs in site_config have no dedicated admin UI — they are set in the monolithic config endpoint. Custom emoji have a full table but no frontend for uploading or using them in posts.

The registration queue processes applications but the admin frontend for reviewing them is basic. The email queue table exists but nothing reads from it — no worker sends queued emails. The OAuth state table is ready for federated login but no provider integration is built. Proximity-based interests (grid_cell + tag for circles) are stored in the database but the auto-circle ranker does not exist yet.

The frontend build has known layout issues: the community settings team page, the inbox messages page, and several admin pages have limited functionality compared to their backend APIs. The theme customizer has presets and color swatches but does not persist theme selections to the backend.

## Possibilities for the Future

Looking further ahead, there are wonderful things Threadlight could become. Here are ten ideas that would fit naturally with the platform's values and existing architecture.

### 1. Localized Communities with Geographic Proximity
The proximity_interest table already stores grid cells and tags per user. Building on this, Threadlight could show local communities organized by geographic region. A person could discover neighbors who share their interests, find local events, and build place-based connections. The location_hash column is already on the users table, ready for this. This would make Threadlight feel grounded and real, connecting the online community to the places people actually live.

### 2. Collaborative Post Editing (Community Wiki Posts)
Posts already support editing, tracking, and locking. A natural extension is collaborative editing where multiple trusted members can contribute to a single post, turning it into a living document — like a community FAQ, a resource guide, or a collaborative announcement. The post's edit history and the trust system together provide the accountability needed for this to work well.

### 3. Multi-Curator Workflows with Approval Chains
The curators table already has a permission field. Expanding this to support approval chains — where certain actions (pinning a post, banning a user, changing community settings) need approval from two curators or a curator plus a moderator — would make community governance more resilient. No single curator could act alone on the most sensitive decisions, and the community would know that important actions were reviewed by multiple trusted people.

### 4. Saved Search Alerts
The saved_searches table already stores query JSON and has a notify_on_new boolean. Completing this feature would let people save a search and receive notifications when new posts match their criteria. A person could save a search for "hiking trails in the Pacific Northwest" and get a friendly ping when someone posts about it, without having to check manually. The notification system and the search engine are both ready for this.

### 5. Private Message Inbox with Read Receipts
The inbox pages exist in the frontend skeleton, and the user_notifications table tracks every kind of alert. Adding private person-to-person messaging would give people a way to have conversations that do not belong on the public feed. Read receipts, message threading, and the ability to block someone from messaging you would all build naturally on top of the existing block and notification systems.

### 6. Community Polls and Decision-Making Tools
The interaction model already supports multiple types, and the jury system shows that Threadlight knows how to run votes. Adding native polls to communities — with configurable duration, anonymous voting, multiple choice or ranked choice, and result visibility options — would give communities a built-in way to make decisions together. Should the community add a new rule? Should we change the description? A poll makes that decision visible and democratic.

### 7. Digest Emails (Daily or Weekly Summaries)
The email queue table exists but is not used. A daily or weekly digest email — summarizing top posts from your communities, new replies to your threads, trending topics you might have missed, and your weekly achievement summary — would bring people back to the platform and make sure nobody misses what matters to them. The feed worker, trending system, and achievement system together produce exactly the data a digest needs.

### 8. Reading List with Progress Tracking
Collections already let you gather posts around a theme. Adding a reading-list mode with read/unread status, progress tracking (X of Y posts read), and estimated reading time would turn collections into a powerful way to work through content at your own pace. This is especially useful for educational series, community onboarding guides, or long-form discussions that unfold over many posts.

### 9. Trusted Referral Paths (Who Brought You Here)
The invited_by column on users and the invite system already track who brought someone into the community. Building a visible referral path — showing new members "You were invited by Alice, who was invited by Bob, who was one of the first members" — would make the community's growth visible and personal. It would give people a sense of where they fit in the community's story and recognize the people who have helped it grow over time.

### 10. Community Onboarding Quests
The achievement system and trust progression already define milestones. A structured onboarding quest — complete your profile, make your first post, join a community, react to someone's post, use your first tag, earn your first achievement — would guide new members through the things that make Threadlight feel like home. Each step could unlock a small credit reward or a badge, making the learning process feel like an accomplishment rather than a chore.

### 11. Collaborative Moderation Sandbox
The moderation system is powerful but irreversible actions can feel stressful for new moderators. A sandbox mode where moderators can preview the effect of an action — "If I mark this post as off-topic, here is how it will look to the author and to everyone else" — before committing would reduce mistakes and build confidence. The moderation endpoints already return the data needed to build this preview.

### 12. Community Member Map
The proximity_interest grid cells could power a community member map — not showing exact locations, but showing a gentle heatmap of where members are concentrated. This would help people feel the geographic diversity of the community and might spark conversations between members who live near each other. The privacy controls (proximity_opt_out on users) already exist so that no one appears on the map without choosing to.

### 13. Plugin SDK and Example Repository
The WASM plugin system is powerful but writing a plugin requires understanding the internal API. An SDK with documentation, example plugins, type definitions, and a local testing harness would make plugin development accessible to anyone who knows Rust, Go, or any language that compiles to WASM. A public plugin repository (beyond the in-app marketplace) where people can browse source code, leave feedback, and fork plugins would turn the plugin system into a community of its own.

### 14. Cross-Community Reputation
Today, trust and reputation are per-user but not per-community. A cross-community reputation system would let people see that someone who is highly trusted in the gardening community also brings thoughtful contributions to the philosophy community. This would help moderators in one community know that a newcomer from another part of the platform is already a respected member, reducing the friction of moving between communities.

### 15. Scheduled Content Curation (Editorial Calendar)
Scheduled posts already exist for individual authors. An editorial calendar for communities — where curators can schedule posts, coordinate topic days, and see what is coming up — would help communities plan their content together. The scheduled post worker and the community curator system together provide the foundation for this.

## How We Decide What to Build Next

Deciding what to build next is not about picking the shiniest idea. It is about understanding what will make the biggest difference for the most people, and then doing that first — carefully, together.

Right now the most important work is finishing the frontends that already have working backends. User list management pages, the affinity display page, the plugin marketplace frontend, and the search frontend redesign all have solid backend foundations just waiting for a user interface. The moderation appeal system is the one big piece in this group that needs both sides built at once.

Next come the things that people will notice and benefit from every day. Image upload closes the biggest gap in how people share. Trust level display on profiles closes a gap in the community health system. The mod post deletion button is a quick frontend fix that makes a difference. SMTP configuration for password reset emails closes a real security gap.

In the medium term, the Positive Algorithm and its "Why this post?" companion will transform the feed from a simple timeline into a thoughtful curator. The M2 meta-moderation user interface will close the accountability loop that makes moderation truly democratic. The full reaction set will give people richer ways to express themselves. Tag autocomplete will make the simple act of tagging feel effortless. And the digest email system will bring people back to the platform regularly, making sure nobody misses what matters to them.

And in the long run, ActivityPub federation will connect Threadlight to the wider world. The full economy with streaks, gifting, and the hash-chain ledger will make contribution and recognition feel meaningful. WebSocket live feed will make the community feel instant and alive. Auto-generated circles will help people find the others they most belong with. Geographic proximity features will ground the community in the places people actually live.

Every step of this journey is guided by the same idea: that a good community is built by everyone who lives in it, that the tools should serve the people and not the other way around, and that the work of building something together is itself one of the most rewarding things a group of people can do.
