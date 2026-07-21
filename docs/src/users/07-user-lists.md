# User Lists

User lists let you create curated groups of accounts. You can use them to batch-follow or batch-block people, share lists with others, or let the system build a list for you based on rules.

## What Are User Lists?

A user list is a named group of accounts. Think of it like a playlist, but for people instead of songs. You can use lists to:

- **Organize people** you want to keep track of
- **Batch follow** a whole group at once
- **Batch block** spammers or bad actors
- **Share** a list so others can use it too
- **Collaborate** with friends on a shared list

Lists come in three types: manual (you choose every member), algorithmic (the system fills based on rules), and collaborative (multiple people help curate).

## Creating a List

To create a list, go to your lists page and click Create List. You will be asked for:

- **Name** — a short label for your list (required, up to 100 characters)
- **Description** — what this list is for (optional)
- **Type** — Follow list or Block list
- **Visibility** — who can see it:
  - **Private** — only you can see the list
  - **Public** — anyone can see the list
  - **Shared** — people you invite can help edit the list

Private lists are invisible to everyone except you. Public lists are visible to anyone but only you can edit them (unless you add collaborators).

## Follow Lists vs Block Lists

Every list has a type that changes what happens when someone subscribes:

### Follow Lists

A follow list contains accounts you recommend. When someone subscribes to your follow list, the system automatically follows every person on that list for them. This is useful for:

- New users who want a good starting set of people to follow
- Topic-specific lists (e.g., "scientists to follow")
- Community onboarding

### Block Lists

A block list contains accounts you have blocked or identified as problematic. When someone subscribes to your block list, everyone on that list gets blocked for them. This is useful for:

- Sharing known spammers with your community
- Coordinating blocks against coordinated harassment
- Importing block lists from other instances

You can tell whether a list is a follow list or a block list by its icon in the UI.

## Adding and Removing Members

To add someone to a list:

1. Go to the list you want to edit.
2. Click Add Members.
3. Search for the person you want to add.
4. Select them and confirm.

You can remove someone the same way — click Remove next to their name in the member list.

Only the list owner and editors can add or remove members. Viewers can see the list but cannot change it.

## Subscribing to Shared Lists

When someone shares a public list, you can subscribe to it:

1. Find the list (browse public lists or get a direct link).
2. Click Subscribe.
3. The system will immediately follow (or block) every person on that list.

Subscribing is a one-time action — it follows or blocks everyone currently on the list. If the list owner adds more people later, you will not automatically follow them. You can manually re-apply the list or use the Apply button to sync again.

To stop using a list:

1. Go to your subscriptions.
2. Click Unsubscribe next to the list.

Unsubscribing does **not** unfollow or unblock those people — it just stops the connection. You can unfollow or unblock them individually if you want to.

## Collaborative Lists

If you want to manage a list with other people, you can create a shared list and invite collaborators:

1. Create a list with visibility set to Shared.
2. Go to the Collaborators section.
3. Click Invite and enter the person's username.
4. Choose their role:
   - **Editor** — can add and remove members, update the list description
   - **Viewer** — can see the list and member list but cannot make changes

The person you invited will receive a notification. They must accept the invitation before they can access the list.

You can remove a collaborator at any time from the same page.

## Algorithmic Lists

Algorithmic lists automatically fill themselves based on rules you define. Instead of picking members one by one, you set criteria and the system finds matching accounts.

### Creating an Algorithmic List

1. Go to your lists page.
2. Click Create Algorithmic List.
3. Fill in the same name, type, and visibility settings as a regular list.
4. Set criteria rules:

| Criteria | What it does |
|----------|-------------|
| Minimum trust level | Only include users at this trust level or above |
| Min posts in last 30 days | Exclude users who haven't posted enough recently |
| Min reaction score | Only include users whose posts have a minimum total reaction count |
| Min positive ratio | Users must have at least this fraction of positive interactions |
| Max negative ratio | Users must have at most this fraction of negative interactions |
| Max freshness days | Only include users active within this many days |
| Tag score threshold | For per-tag lists, minimum vote score on the relevant tag |

5. Choose the scope:
   - **Global** — matches against all users
   - **Per-tag** — only matches users who have posted about a specific tag

6. Set a refresh interval (default: 24h).

### How Algorithmic Lists Refresh

The system runs a background worker every 30 minutes that checks every algorithmic list. If a list's `last_refreshed_at + refresh_interval` is in the past, the worker:

1. Runs the criteria query against the user table
2. Deletes all existing algorithmic members from the list
3. Inserts the newly matching users
4. Updates `last_refreshed_at` to now

So if you set a refresh interval of 24 hours, the list updates once per day. If you need more frequent updates, you can set a shorter interval, but 30 minutes is the minimum (the worker's own interval).

### Evaluating an Algorithmic List

You can manually evaluate an algorithmic list to see who would match without actually changing the list. This is useful for tweaking your criteria before the next automatic refresh.

Go to the list and click Evaluate. The system will show you the matching users without modifying the list members.

## API Routes

User lists are accessed through the following API endpoints:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/lists` | Create a new list |
| GET | `/api/v1/lists` | List your lists |
| GET | `/api/v1/lists/:id` | Get list details |
| PUT | `/api/v1/lists/:id` | Update list |
| DELETE | `/api/v1/lists/:id` | Delete list |
| POST | `/api/v1/lists/:id/members` | Add a member |
| DELETE | `/api/v1/lists/:id/members/:user_id` | Remove a member |
| GET | `/api/v1/lists/:id/members` | List members |
| POST | `/api/v1/lists/:id/subscribe` | Subscribe to a list |
| POST | `/api/v1/lists/:id/unsubscribe` | Unsubscribe |
| GET | `/api/v1/lists/:id/subscribers` | See who else subscribed |
| GET | `/api/v1/lists/subscriptions` | Your subscriptions |
| POST | `/api/v1/lists/:id/collaborators` | Invite a collaborator |
| POST | `/api/v1/lists/:id/collaborators/accept` | Accept an invitation |
| DELETE | `/api/v1/lists/:id/collaborators/:user_id` | Remove collaborator |
| GET | `/api/v1/lists/:id/collaborators` | List collaborators |
| POST | `/api/v1/lists/algorithmic` | Create an algorithmic list |
| POST | `/api/v1/lists/:id/evaluate` | Preview algorithmic matches |

## Privacy Notes

- Your private lists are never visible to anyone else.
- Public lists show your name as the owner. Anyone can see the list name, description, and member list.
- When you subscribe to a list, the list owner can see that you subscribed (but not who you follow individually).
- Block lists are always processed silently — the blocked user is never notified.
