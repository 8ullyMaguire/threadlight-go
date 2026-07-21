# Frontend Audit & Fix Plan

## The Problem

The frontend (`web/`) is a fork of Photon, a Lemmy/PieFed client. It works great if you're browsing a Lemmy instance. For Threadlight, there are three gaps:

1. **22 adapter methods are stubs** — they return `{} as Lemmy.X` empty objects instead of real data. The UI components that call them render nothing or crash silently.
2. **6 new Threadlight-specific pages exist but are unreachable** — the navbar, settings tabs, and routing don't know about them.
3. **The UI components expect Lemmy data shapes** — even when the adapter returns data, the components may not display Threadlight-specific fields (affinity breakdowns, list types, plugin metadata).

---

## Phase 1: Audit the Adapter (Who Calls What)

**Goal:** Know exactly which stubs are called by which UI components.

**Actions:**
1. Trace every stub method in `adapter.ts` back to its caller in `web/src/lib/feature/` and `web/src/routes/`
   ```
   grep -rn 'getUnreadCount\|listPostReports\|createPrivateMessage\|...' web/src/
   ```
2. Categorize each stub:
   - **Dead code** — method exists but no UI component calls it → delete or leave as-is
   - **Called by dead route** — method is called by a Lemmy-specific page that doesn't apply → add null check in component
   - **Called by active code** — method is called by a live component and needs a real implementation
3. For category 3, write a real `#api()` call that hits the Threadlight backend endpoint.

**Estimated time:** 1 hour.

**Example — `getUnreadCount()` is called by the inbox page. The stub returns `{ replies: 0, mentions: 0, private_messages: 0 }`. The inbox page renders a "0" badge. Since Threadlight has a different notification model, the badge is always wrong. Fix: call `GET /api/v1/notifications/unread-count` and map the response.**

---

## Phase 2: Wire the Adapter Methods (Threadlight Features)

**Goal:** Every new frontend page gets real data.

Already added (recent subagent work) but need verification:

| Method | Backend endpoint | Frontend caller | Status |
|--------|-----------------|----------------|--------|
| `getLists()` | `GET /api/v1/lists` | `/settings/lists/` | ⚠️ Verify |
| `getList(id)` | `GET /api/v1/lists/:id` | `/settings/lists/[id]` | ⚠️ Verify |
| `createList()` | `POST /api/v1/lists` | `/settings/lists/create` | ⚠️ Verify |
| `getAffinities()` | `GET /api/v1/affinity` | `/settings/affinity/` | ⚠️ Verify |
| `getSimilarUsers()` | `GET /api/v1/affinity/similar` | `/settings/affinity/` | ⚠️ Verify |
| `getPlugins()` | `GET /api/v1/feed-plugins` | `/plugins/` | ⚠️ Verify |
| `searchPosts()` | `GET /api/v1/search/posts` | `/search/` | ⚠️ Verify |
| `searchSuggest()` | `GET /api/v1/search/suggest` | `/search/` | ⚠️ Verify |

**Actions:**
1. For each method, hit the page in the browser and check the Network tab for errors.
2. Fix any 404/500 errors (wrong URL path, missing params, auth token not sent).
3. Fix any type mismatches between what the backend returns and what the TypeScript interface expects.

**Estimated time:** 2 hours.

---

## Phase 3: Fix the Navigation

**Goal:** Every Threadlight feature is reachable from the UI.

**Actions:**
1. **Settings layout** (`/settings/+layout.svelte`) — add tabs for Lists and Affinity.
   - The existing layout uses `TabbedLayoutShell` with tabs: App, Embeds, Lemmy, Moderation, Other.
   - Add two tabs: "Lists" → `/settings/lists`, "Affinity" → `/settings/affinity`

2. **Navbar** (`Navbar.svelte`) — add plugin marketplace link.
   - The sidebar has Explore, Search, Create, Admin, Moderation.
   - Add "Plugins" or put it under Explore as a sub-item.
   - Or add to the explore routes: `/explore/plugins`

3. **Search page** — the search page was redesigned by a subagent. Verify it uses the new `searchPosts()` adapter method and that the results render correctly in the existing `PostItem` / `PostFeed` components.

**Estimated time:** 1 hour.

---

## Phase 4: Fix the Stubs That Matter

**Goal:** The most-visible Lemmy stubs that real users hit every day return real data.

Priority ordered by user impact:

| # | Method | Used by | Current stub | Fix |
|---|--------|---------|-------------|-----|
| 1 | `getUnreadCount()` | Inbox badge | `{ replies:0, mentions:0, pms:0 }` | Call `GET /api/v1/notifications/unread-count` |
| 2 | `getReplies()` | Inbox page | `{ replies: [] }` | Call `GET /api/v1/notifications` |
| 3 | `getUnreadRegistrationApplicationCount()` | Admin badge | `{ applications: 0 }` | Call `GET /api/v1/admin/invites?status=pending` |
| 4 | `listPostReports()` | Mod queue | `{ post_reports: [] }` | Call `GET /api/v1/reports?type=post` |
| 5 | `listCommentReports()` | Mod queue | `{ comment_reports: [] }` | Call `GET /api/v1/reports?type=comment` |
| 6 | `search()` | Global search | calls `searchPosts()` | Already fixed — verifies |
| 7 | `getPerson()` | Profile page | real call exists | Verify |
| 8 | `getModlog()` | Mod log | `{ actions: [] }` | Call `GET /api/v1/modlog/actions` |

**Estimated time:** 2 hours.

---

## Phase 5: Type Alignment

**Goal:** The TypeScript interfaces match what the backend actually returns.

**Actions:**
1. Compare every response from `GET /api/v1/...` with its TypeScript interface in `types.ts`.
2. Fix mismatches — either change the interface or add a mapping function in the adapter.
3. Pay special attention to nested objects: Lemmy's `PostView` wraps a post + creator + community, while Threadlight's API returns them flat or with different nesting.

**Common mismatches:**
- Threadlight returns `{ id: 1, title: "x", author_id: 1 }` — Lemmy expects `{ post: { id, name, body }, creator: { id, name }, community: { id, name } }`
- Threadlight pagination uses `page` + `limit` — Lemmy uses `page` + `limit` too but returns `posts` array differently
- Threadlight `GET /api/v1/posts` returns `[{ ...post, tags: [...], author: { ... } }]` — Lemmy expects `{ posts: [PostView] }`

**Estimated time:** 3 hours.

---

## Phase 6: Test Every Route

**Goal:** Every route loads without JS errors and shows meaningful data or an empty state.

**Route checklist:**

- [ ] `/` — feed
- [ ] `/search` — with Boolean syntax, tabs, filters
- [ ] `/settings/lists` — create, detail, members
- [ ] `/settings/affinity` — table, similar users
- [ ] `/plugins` — browse, detail, upload
- [ ] `/post/[instance]/[id]` — post detail with tags, reactions
- [ ] `/u/[name]` — user profile with affinity badge
- [ ] `/c/[name]` — community
- [ ] `/inbox` — notifications
- [ ] `/moderation` — reports queue
- [ ] `/modlog` — controversial decisions
- [ ] `/create/post` — post creation
- [ ] `/explore/communities` — community browser
- [ ] `/admin` — admin panel (if admin)

**Estimated time:** 1.5 hours.

---

## Summary

| Phase | What | Time |
|-------|------|------|
| 1 | Audit adapter stubs & their callers | 1h |
| 2 | Wire new feature adapter methods | 2h |
| 3 | Fix navigation (settings tabs, navbar) | 1h |
| 4 | Fix priority stubs (notifications, reports) | 2h |
| 5 | Align TypeScript types with backend responses | 3h |
| 6 | Test every route | 1.5h |
| **Total** | | **~10.5h** |

The quickest path to "feels functional" is **Phase 3 + Phase 4 items 1-3** (~3 hours): wire the settings tabs so user lists and affinity are reachable, then fix the inbox badge and notifications so the nav bar stops showing "0" everywhere.
