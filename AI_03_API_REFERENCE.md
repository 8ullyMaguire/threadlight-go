# Polaris — API Reference

## Base
`http://localhost:8080/api/v1`

## Auth
Header: `Authorization: Bearer <token>`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /auth/register | No | Register with username, email, password, invite_code |
| POST | /auth/login | No | Login with email, password → returns {token, user} |
| POST | /auth/logout | Yes | Invalidate session |
| GET | /auth/session | Yes | Current user session info |

## Posts
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /posts | Yes | Create post (title, body, mood, content_warning, scheduled_at, status) |
| GET | /posts/:id | Yes | Get single post |
| PATCH | /posts/:id | Yes | Update post (own only) |
| DELETE | /posts/:id | Yes | Soft delete post |
| GET | /posts | Yes | List posts with ?limit=20&offset=0 |
| GET | /posts/scheduled | Yes | List user's scheduled posts |
| POST | /posts/:id/interact | Yes | Like/comment/share/gift (interaction_type, metadata) |

## Tags
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /tags | Yes | Create tag (name, category) |
| GET | /tags/:id | Yes | Get tag |
| GET | /tags/search | Yes | Search tags (?q=) |
| POST | /posts/:id/tags | Yes | Add tag to post (tag_id) |
| DELETE | /posts/:id/tags/:tagId | Yes | Remove tag from post |
| GET | /posts/:id/tags | Yes | Get post's tags |

## Feed
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /feeds | Yes | Main algorithmic feed (?mood=&limit=20&offset=0) |
| GET | /feeds/hot | Yes | Trending globally |
| GET | /feeds/tag/:tagId | Yes | Tag-filtered feed |
| GET | /feeds/circle/:circleId | Yes | Circle feed |
| GET | /feeds/community/:slug | Yes | Community feed |
| GET | /feeds/proximity | Yes | "People near you also like..." |

## Custom Feeds
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /feed/custom | Yes | Create custom feed (name, slug, sort_order) |
| GET | /feed/custom | Yes | List my custom feeds |
| GET | /feed/custom/:slug | Yes | Get feed posts |
| PUT | /feed/custom/:slug | Yes | Update feed |
| DELETE | /feed/custom/:slug | Yes | Delete feed |
| POST | /feed/custom/:slug/sources | Yes | Add source (type, value, include) |
| DELETE | /feed/custom/:slug/sources/:id | Yes | Remove source |

## Search
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /search | Yes | Full-text search (?q=&limit=20) |
| POST | /search/advanced | Yes | Boolean search JSON body {q, limit} |

## Collections
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /collections | Yes | Create collection (name, visibility) |
| GET | /collections | Yes | List my collections |
| GET | /collections/:id | Yes | Get collection with posts |
| PUT | /collections/:id | Yes | Update collection |
| DELETE | /collections/:id | Yes | Delete collection |
| POST | /collections/:id/posts | Yes | Add post to collection (post_id) |
| DELETE | /collections/:id/posts/:postId | Yes | Remove post from collection |

## Communities
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /communities | Yes | Create community |
| GET | /communities | Yes | Browse (?sort=active_weekly) |
| GET | /communities/:slug | Yes | Community detail |
| POST | /communities/:slug/join | Yes | Join community |
| POST | /communities/:slug/leave | Yes | Leave community |
| POST | /communities/:slug/fork | Yes | Fork community |
| GET | /communities/:slug/members | Yes | List members |
| POST | /communities/:slug/curators | Yes | Add curator (community admin) |
| DELETE | /communities/:slug/curators/:userId | Yes | Remove curator |

## Circles
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /circles/suggestions | Yes | Pending circle suggestions |
| POST | /circles/:id/accept | Yes | Accept circle suggestion |
| POST | /circles/:id/dismiss | Yes | Dismiss circle suggestion |
| POST | /circles/:id/leave | Yes | Leave circle |
| GET | /circles/:id | Yes | Circle detail |

## Community Notes
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /notes | Yes | Create note (post_id, body) |
| POST | /notes/:id/vote | Yes | Vote (helpful: true/false) |
| GET | /notes/post/:postId | Yes | Notes on a post |
| GET | /notes/pending | Yes | Pending notes (threshold 3+) |

## Moderation
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /mod/flag | Yes | Flag content (post_id, reason) |
| GET | /mod/log | Yes | Public moderation log |
| POST | /mod/timeout | Yes | Issue timeout (threshold 3+) |
| POST | /mod/ban | Yes | Propose ban (threshold 4+) |
| POST | /mod/jury/:actionId/vote | Yes | Vote on jury action |
| GET | /mod/queue | Yes | Flag queue (threshold 3+) |

## Blocklist
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /mod/blocklist | Yes | List entries (?type=domain) |
| POST | /mod/blocklist | Yes | Add entry (entry_type, value, reason) |
| DELETE | /mod/blocklist/:id | Yes | Remove entry |
| GET | /mod/blocklist/export | Yes | JSON export for federation |

## Content Filters
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /filters | Yes | List my filters |
| POST | /filters | Yes | Create filter (filter_type, filter_value, filter_action) |
| PUT | /filters/:id | Yes | Update filter |
| DELETE | /filters/:id | Yes | Delete filter |

## Credits & Bounties
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /credits/claim-daily | Yes | Claim daily reward |
| GET | /credits/balance | Yes | Get credit balance |
| POST | /bounties | Yes | Create bounty (post_id, amount) |
| POST | /bounties/:id/award | Yes | Award bounty (answer_id) |
| GET | /bounties/open | Yes | Open bounties |

## Achievements & Leaderboards
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /achievements | Yes | All achievements + user progress |
| GET | /achievements/:id | Yes | Single achievement detail |
| GET | /leaderboard/:type | Yes | Leaderboard (?period=weekly) |

## Onboarding
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /onboarding/progress | Yes | Current onboarding stage |
| POST | /onboarding/stage | Yes | Advance stage |
| POST | /onboarding/tags | Yes | Pick initial tags (tag_ids) |
| GET | /onboarding/suggested-users | Yes | Suggested users to follow |

## Admin & Health
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | No | Postgres + Redis health check |
| GET | /api/v1/admin/status | Yes | Admin dashboard data |
| GET | /metrics | No | Prometheus metrics |

## WebSocket
`ws://localhost:8080/ws/v1/feed` — Real-time feed updates
`ws://localhost:8080/ws/v1/chat/:circleId` — Circle chat
`ws://localhost:8080/ws/v1/notifications` — Push notifications
`ws://localhost:8080/ws/v1/presence` — Online presence
