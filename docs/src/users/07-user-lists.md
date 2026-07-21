# User Lists: Organizing Together

On most social networks, you are alone. You choose who to follow, who to block, and what to see. If a wave of harassment hits your community, each person has to defend themselves one by one. If someone has spent years curating a list of excellent scientists to follow, you cannot benefit from their work — you start from zero.

User lists change that. They turn individual actions into shared tools. They are one of the simplest and most powerful ideas on Threadlight: a group of accounts, named and described, that anyone can use, share, and build together.

## What a List Is

A user list is just a collection of accounts with a name and a description. You might call one "climate scientists" or "known spammers on our instance" or "thoughtful writers about parenting." The name is for you and anyone you share it with.

Every list has a **type** that says what happens when someone uses it:

- A **follow list** contains accounts worth following. When someone subscribes to your follow list, the system follows every person on it for them.
- A **block list** contains accounts that should be blocked. When someone subscribes, everyone on the list gets blocked for them.

This is the difference between a recommendation and a protection. Follow lists help people find good voices. Block lists help people avoid harm.

Every list also has a **visibility** that says who can see it:

- **Private** — only you can see it. Good for personal organization.
- **Public** — anyone can see it. Good for sharing your curation with the world.
- **Shared** — people you invite can help edit it. Good for building a list together.

## Creating a List

Go to your lists page and click Create List. Give it a name that tells people what it is for. "bad actors" is clear. "my list" is not. Choose the type — follow or block — and the visibility.

Once created, you can add accounts to it. You add them one at a time by username. There is no limit to how many accounts a list can hold, though a list with thousands of members becomes hard to manage.

You can also delete a list. Deleting a list removes all memberships and subscriptions. Anyone who subscribed to it will no longer be following or blocking its members through that list.

## Subscribing to a List

When you find a public list that interests you, you can subscribe to it. Subscribing applies the list's action — follow or block — to every account on the list.

If the list grows later (someone adds new accounts), those new accounts are automatically followed or blocked for you. You do not need to check back. You also stay subscribed even if the list changes hands or becomes collaborative.

You can unsubscribe at any time. Unsubscribing reverses the action for accounts that were added to the list after you subscribed — it does not unfollow or unblock accounts that were already on your personal list before.

This makes lists useful for ongoing protection. Subscribe once to a well-maintained block list, and you are shielded from new bad actors without having to watch for them yourself.

## Building Lists Together

A collaborative list is one where multiple people share the work of maintaining it. The person who created it is the owner. The owner can invite others as editors or viewers.

**Editors** can add and remove accounts from the list. They cannot change the list's name, type, or visibility, and they cannot delete it. This makes them trusted curators who share the workload.

**Viewers** can see the list and its members but cannot change anything. This is useful for transparency — a community might want everyone to see who is on a block list and why, even if only a few people maintain it.

Collaboration turns a single person's effort into a community resource. A group of moderators from different communities can maintain a shared block list. A collective of science communicators can curate a follow list of accurate voices. The work is distributed, and everyone benefits.

## Lists That Build Themselves

An algorithmic list is one that fills itself based on rules. Instead of manually adding accounts, you tell the system what kinds of accounts you want, and it finds them.

The rules are simple:

- **Minimum trust level** — only include accounts above this trust level
- **Minimum posts in the last 30 days** — skip inactive accounts
- **Minimum reaction score** — only accounts whose posts get enough positive reactions
- **Maximum negative reaction ratio** — skip accounts whose posts get too many negative reactions
- **Tag score threshold** — boost accounts that contribute to a specific tag
- **Maximum freshness in days** — only accounts that have posted recently

You might create an algorithmic follow list called "top rust contributors this month" that looks for accounts with at least ten posts tagged "rust" in the last 30 days, a high positive reaction rate, and a trust level above 1. The system evaluates this rule every 30 minutes and updates the list automatically.

Algorithmic lists can be public and collaborative too. You could share a "rising scientists" list that automatically includes accounts meeting certain criteria, and let a panel of editors override the algorithm by manually adding or removing specific accounts.
