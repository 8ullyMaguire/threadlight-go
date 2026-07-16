# Polaris — API Reference (~95 endpoints)

## Base
`http://localhost:8080/api/v1` • Header: `Authorization: Bearer <token>`

## Auth & Accounts
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /auth/register | No | Register (username, email, password, invite_code) |
| POST | /auth/login | No | Login → {token, user} |
| POST | /auth/logout | Yes | Invalidate session |
| GET | /auth/session | Yes | Current session info |
| POST | /auth/password-reset | No | Request reset token (email) |
| POST | /auth/password-reset/confirm | No | Confirm reset (token, password) |
| POST | /auth/verify-email | Yes | Send verification email |
| POST | /auth/verify-email/confirm | No | Confirm email (token) |
| DELETE | /api/v1/account | Yes | Soft-delete account |

## Posts
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /posts | Yes | Create (title, body, mood, content_warning, scheduled_at, status) |
| GET | /posts/:id | Yes | Get single post |
| PATCH | /posts/:id | Yes | Edit body (15-min window) |
| DELETE | /posts/:id | Yes | Soft delete |
| GET | /posts | Yes | List (?limit=&offset=) |
| GET | /posts/scheduled | Yes | User's scheduled posts |
| POST | /posts/:id/interact | Yes | Like/comment/share/gift/emoji (type, metadata) |
| POST | /posts/:id/lock | Yes | Lock comments (mod) |
| POST | /posts/:id/unlock | Yes | Unlock comments |
| POST | /posts/:id/sticky | Yes | Pin to community |
| POST | /posts/:id/hide | Yes | Per-user hide |
| DELETE | /posts/:id/hide | Yes | Unhide |
| POST | /posts/:id/move | Yes | Move to another community (mod) |

## Tags
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /tags | Yes | Create tag (name, category) |
| GET | /tags/:id | Yes | Get tag |
| GET | /tags/search | Yes | Search (?q=) |
| POST | /posts/:id/tags | Yes | Add tag (tag_id) |
| DELETE | /posts/:id/tags/:tagId | Yes | Remove tag |
| GET | /posts/:id/tags | Yes | Get post tags |

## Feed
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /feeds | Yes | Algorithmic feed (?mood=&limit=&offset=) |
| GET | /feeds/hot | Yes | Trending globally |
| GET | /feeds/tag/:tagId | Yes | Tag-filtered |
| GET | /feeds/circle/:circleId | Yes | Circle feed |
| GET | /feeds/community/:slug | Yes | Community feed |
| GET | /feeds/proximity | Yes | "People near you also like..." |

## Custom Feeds
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /feed/custom | Yes | Create (name, slug, sort_order) |
| GET | /feed/custom | Yes | List my feeds |
| GET | /feed/custom/:slug | Yes | Get feed posts |
| PUT | /feed/custom/:slug | Yes | Update |
| DELETE | /feed/custom/:slug | Yes | Delete |
| POST | /feed/custom/:slug/sources | Yes | Add source (type, value, include) |
| DELETE | /feed/custom/:slug/sources/:id | Yes | Remove source |

## Search
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /search | Yes | Full-text (?q=&limit=) |
| POST | /search/advanced | Yes | Boolean search {q, limit} |

## Collections
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /collections | Yes | Create (name, visibility) |
| GET | /collections | Yes | List |
| GET | /collections/:id | Yes | Detail with posts |
| PUT | /collections/:id | Yes | Update |
| DELETE | /collections/:id | Yes | Delete |
| POST | /collections/:id/posts | Yes | Add post (post_id) |
| DELETE | /collections/:id/posts/:postId | Yes | Remove post |

## Communities
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /communities | Yes | Create |
| GET | /communities | Yes | Browse (?sort=active_weekly) |
| GET | /communities/:slug | Yes | Detail |
| POST | /communities/:slug/join | Yes | Join |
| POST | /communities/:slug/leave | Yes | Leave |
| POST | /communities/:slug/fork | Yes | Fork |
| GET | /communities/:slug/members | Yes | Members |
| POST | /communities/:slug/curators | Yes | Add curator |
| DELETE | /communities/:slug/curators/:userId | Yes | Remove curator |

## Circles
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /circles/suggestions | Yes | Pending suggestions |
| POST | /circles/:id/accept | Yes | Accept suggestion |
| POST | /circles/:id/dismiss | Yes | Dismiss |
| POST | /circles/:id/leave | Yes | Leave |
| GET | /circles/:id | Yes | Detail |

## Community Notes
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /notes | Yes | Create (post_id, body) |
| POST | /notes/:id/vote | Yes | Vote (helpful: bool) |
| GET | /notes/post/:postId | Yes | Notes on post |
| GET | /notes/pending | Yes | Pending (threshold 3+) |

## Moderation
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /mod/flag | Yes | Flag content (post_id, reason) |
| GET | /mod/log | Yes | Public mod log |
| POST | /mod/timeout | Yes | Issue timeout (threshold 3+) |
| POST | /mod/ban | Yes | Propose ban (threshold 4+) |
| POST | /mod/jury/:actionId/vote | Yes | Jury vote |
| GET | /mod/queue | Yes | Flag queue (threshold 3+) |
| POST | /mod/notes | Yes | Add mod note (user_id, note) |
| GET | /mod/notes/:userId | Yes | Get mod notes for user |
| DELETE | /mod/notes/:id | Yes | Delete own mod note |

## User Controls
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /users/:id/block | Yes | Block user |
| POST | /users/:id/unblock | Yes | Unblock user |
| GET | /blocks | Yes | List blocked users |
| PATCH | /users/me | Yes | Update profile (bio, display_name, theme, hide_read_posts) |
| GET | /users/:id/stats | Yes | User stats (post count, likes, etc.) |

## Reports
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /reports | Yes | Create report (post_id, category, reason) |
| GET | /mod/reports | Yes | List reports (threshold 3+) |
| POST | /mod/reports/:id/resolve | Yes | Resolve report |

## Notifications
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /notifications | Yes | List (?limit=&offset=) |
| POST | /notifications/:id/read | Yes | Mark read |
| POST | /notifications/read-all | Yes | Mark all read |
| GET | /notifications/unread-count | Yes | Unread count |

## Credits & Bounties
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /credits/claim-daily | Yes | Claim daily reward |
| GET | /credits/balance | Yes | Balance |
| POST | /bounties | Yes | Create (post_id, amount) |
| POST | /bounties/:id/award | Yes | Award (answer_id) |
| GET | /bounties/open | Yes | Open bounties |

## Achievements & Leaderboards
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /achievements | Yes | All + user progress |
| GET | /achievements/:id | Yes | Single detail |
| GET | /leaderboard/:type | Yes | (?period=weekly) |

## Onboarding
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /onboarding/progress | Yes | Current stage |
| POST | /onboarding/stage | Yes | Advance stage |
| POST | /onboarding/tags | Yes | Pick initial tags (tag_ids) |
| GET | /onboarding/suggested-users | Yes | Suggested users |

## Admin
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /admin/status | Yes | Dashboard data |
| GET | /admin/emoji | Yes | List custom emoji |
| POST | /admin/emoji | Yes | Create emoji |
| DELETE | /admin/emoji/:id | Yes | Delete emoji |
| POST | /admin/pages | Yes | Upsert custom page |

## Public
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | No | PG + Redis health |
| GET | /metrics | No | Prometheus metrics |
| GET | /pages/:slug | No | Public custom page |
| GET | /.well-known/webfinger | No | ActivityPub WebFinger |

## WebSocket
| Path | Description |
|------|-------------|
| /ws/v1/feed | Real-time feed updates |
| /ws/v1/chat/:circleId | Circle chat |
| /ws/v1/notifications | Push notifications |
| /ws/v1/presence | Online presence |
