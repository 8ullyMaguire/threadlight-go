# ALL ROUTES (~95 in internal/api/router.go)

Prefix: `/api/v1`. Auth: `Authorization: Bearer <JWT>`. Middleware: `c.Set("user_id", int64)`.

## /auth
```
POST register          {username,email,password,invite_code?} → {token,user}
POST login             {email,password} → {token,user}
POST logout            [auth]
GET  session           [auth] → {user_id}
POST password-reset    {email}
POST password-reset/confirm {token,password}
POST verify-email      [auth]
POST verify-email/confirm {token}
```

## /posts [auth]
```
POST ""       {title,body,mood,content_warning?,scheduled_at?,status?} → Post
GET  ""       ?limit=20&offset=0 → []Post
GET  /:id     → Post
PATCH /:id    {body}              (15-min edit window)
DELETE /:id                        (soft delete)
GET  /scheduled                    → []Post
POST /:id/lock                     (mod, lock comments)
POST /:id/unlock                   (mod)
POST /:id/sticky                   (mod, pin)
POST /:id/hide                     (per-user)
DELETE /:id/hide
POST /:id/move    {community_id}   (mod)
POST /:id/interact {interaction_type,metadata?}  (like/comment/share/gift/emoji)
```

## /tags, /feeds, /feed/custom [auth]
```
POST /tags          {name,category} → Tag
GET  /tags/:id      → Tag
GET  /tags/search   ?q= → []Tag
POST /posts/:id/tags {tag_id}       (booru-style, any trusted user)
DELETE /posts/:id/tags/:tagId

GET  /feeds            ?mood=&limit=20&offset=0 → []Post
GET  /feeds/hot
GET  /feeds/tag/:tagId
GET  /feeds/community/:slug
GET  /feeds/proximity

POST /feed/custom          {name,slug,sort_order} → CustomFeed
GET  /feed/custom          → []CustomFeed
GET  /feed/custom/:slug    ?limit=&offset= → []Post
PUT  /feed/custom/:slug    {name,description}
DELETE /feed/custom/:slug
POST /feed/custom/:slug/sources  {source_type,source_value?,include_mode}
DELETE /feed/custom/:slug/sources/:id
```

## /search, /collections, /communities, /circles [auth]
```
GET  /search                ?q=&limit=20 → []Post
POST /collections            {name,visibility?} → Collection
GET  /collections            → []Collection
GET  /collections/:id        → Collection
PUT  /collections/:id        {name,visibility,description}
DELETE /collections/:id
POST /collections/:id/posts  {post_id}
DELETE /collections/:id/posts/:postId

POST /communities            {name,slug,tags,curator_lock?,slow_boot_days?} → Community
GET  /communities            ?sort=active_weekly → []Community
GET  /communities/:slug      → Community
POST /communities/:slug/join
POST /communities/:slug/leave
POST /communities/:slug/fork   {reason}
GET  /communities/:slug/members
POST /communities/:slug/curators {user_id}
DELETE /communities/:slug/curators/:userId

GET  /circles/suggestions    → []Circle
POST /circles/:id/accept
POST /circles/:id/dismiss
POST /circles/:id/leave
GET  /circles/:id
```

## /notes, /mod, /users, /blocks, /reports [auth]
```
POST /notes            {post_id,body} → CommunityNote
POST /notes/:id/vote   {helpful:bool}
GET  /notes/post/:postId  → []CommunityNote
GET  /notes/pending       → []CommunityNote (threshold 3+)

POST /mod/flag         {post_id,reason}
GET  /mod/log          ?limit=&offset= → []ModerationAction (public)
POST /mod/timeout      {target_user_id,duration,reason}  (threshold 3+)
POST /mod/ban          {target_user_id,reason}           (threshold 4+)
POST /mod/jury/:actionId/vote  {vote:bool,reason?}
GET  /mod/queue        → []pending flags
POST /mod/blocklist    {entry_type,entry_value,reason} → BlocklistEntry
GET  /mod/blocklist    ?type= → []BlocklistEntry
DELETE /mod/blocklist/:id      (threshold 5+)
GET  /mod/blocklist/export     → JSON (federation)
POST /mod/notes        {user_id,note} → ModNote
GET  /mod/notes/:userId → []ModNote
DELETE /mod/notes/:id
GET  /mod/reports      ?status= → []PostReport
POST /mod/reports/:id/resolve

PATCH /users/me        {bio?,display_name?,theme?,hide_read_posts?} → User
GET  /users/:id        → User (public)
GET  /users/:id/stats  → {post_count,comment_count,likes_received}
POST /users/:id/block
POST /users/:id/unblock

GET  /blocks           → []BlockedUser
POST /reports          {post_id,category,reason} → PostReport
```

## /notifications, /credits, /bounties, /achievements, /leaderboard, /onboarding [auth]
```
GET  /notifications           ?limit=20&offset=0 → []Notification
POST /notifications/:id/read
POST /notifications/read-all
GET  /notifications/unread-count → {count:N}

POST /credits/claim-daily     → {amount,streak}
GET  /credits/balance         → {credits:N}
POST /bounties                {post_id,amount} → Bounty
POST /bounties/:id/award      {answer_id}
GET  /bounties/open           → []Bounty

GET  /achievements            → []Achievement (with user progress)
GET  /leaderboard/:type       ?period=daily|weekly|monthly|yearly|all-time → []LeaderboardEntry

GET  /onboarding/progress           → {stage:N}
POST /onboarding/stage              {stage:N}
POST /onboarding/tags               {tag_ids:[]int64}
GET  /onboarding/suggested-users    → []User
```

## /admin (threshold 5+), Public (no auth)
```
GET  /admin/status       → {active_users,cache_hit_rate,queue_depths}
GET  /admin/emoji        → []CustomEmoji
POST /admin/emoji        {shortcode,image_url,alt_text}
DELETE /admin/emoji/:id
POST /admin/pages        {slug,title,body,is_published}

GET  /health             → {status,postgres:bool,redis:bool}
GET  /metrics            → Prometheus text
GET  /pages/:slug        → CustomPage (published only)
GET  /.well-known/webfinger  ?resource=acct:user@domain → ActivityPub WebFinger JSON
```

## WebSocket (query param token)
```
/ws/v1/feed              real-time feed + notifications
/ws/v1/chat/:circleId    chat (gorilla + Redis pub/sub)
/ws/v1/presence          online presence
```

## Category Codes
```
post mood: 0=general, 1=edu, 2=entertain, 3=social, 4=creative
interaction_type: 0=like, 1=comment, 2=share, 3=gift, 4=bounty_answer, 5=emoji
report category: 0=spam, 1=harassment, 2=misinformation, 3=illegal, 4=other
post status: 0=published, 1=draft, 2=scheduled
collection visibility: 0=private, 1=public_read, 2=public_edit, 3=anybody_edit
```
