# KEY MODELS (26 files in internal/model/)

## errors.go
```go
var ( ErrNotFound = errors.New("not found"); ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden = errors.New("forbidden"); ErrConflict = errors.New("conflict"); ErrValidation = errors.New("validation error") )
```

## User (user.go)
```go
type User struct {
    ID int64 `json:"id"`; Username string `json:"username"`; DisplayName string `json:"display_name,omitempty"`
    Bio string `json:"bio,omitempty"`; Email string `json:"email,omitempty"`; PasswordHash string `json:"-"`  // never serialized
    TrustLevel int16 `json:"trust_level"`; TrustScore float64 `json:"trust_score"`; Reputation int64 `json:"reputation"`
    Credits int64 `json:"credits"`; InvitedBy *int64 `json:"invited_by,omitempty"`; InviteCode string `json:"invite_code,omitempty"`
    AvatarURL string `json:"avatar_url"`; BannerURL string `json:"banner_url"`; BioHTML string `json:"bio_html"`
    EmailVerified bool `json:"email_verified"`; Theme string `json:"theme"`; HideReadPosts bool `json:"hide_read_posts"`
    IsLocal bool `json:"is_local"`; OnboardingStage int16 `json:"onboarding_stage"`; ProximityOptOut bool `json:"proximity_opt_out"`
    ActorID string `json:"actor_id,omitempty"`; LocationHash string `json:"location_hash,omitempty"`
    CreatedAt time.Time `json:"created_at"`; IsDeleted bool `json:"is_deleted"`; DeletedAt *time.Time `json:"deleted_at"`
}
type RegisterRequest struct { Username string `json:"username" binding:"required,min=3,max=30"`
    Email string `json:"email" binding:"required,email"`; Password string `json:"password" binding:"required,min=8"`
    InviteCode string `json:"invite_code"` }
type LoginRequest struct { Email string `json:"email" binding:"required,email"`; Password string `json:"password" binding:"required"` }
type AuthResponse struct { Token string `json:"token"`; User User `json:"user"` }
```
Passwords: bcrypt hashed. JWT: HS256 with claims `{user_id, exp(7d), iat}`.

## Post (post.go)
```go
type Post struct {
    ID int64; AuthorID int64; Title string; Body string
    ContentType int16; Mood int16; ContentWarning *string
    InteractionCount int64; CumulativeInteractions int64
    Status int16; ScheduledAt *time.Time `json:"scheduled_at"`; EditedAt *time.Time `json:"edited_at"`
    Locked bool; Sticky bool; StickyAt *time.Time; Language string; IsAiGenerated bool; License string
    CrossPostRootID *int64; MovedFromCommunityID *int64
    IsEducational bool; IsDeleted bool
    CreatedAt *time.Time; UpdatedAt *time.Time; ArchivedAt *time.Time
}
// content_type: 0=text,1=image,2=video,3=link,4=poll
// mood: 0=general,1=edu,2=entertain,3=social,4=creative
// status: 0=published,1=draft,2=scheduled
```

## Tags + Interactions (tag.go, interaction.go)
```go
type Tag struct { ID int; Name, Description, Category string; IsWiki bool; CreatedBy int64; CreatedAt time.Time }
type Interaction struct { ID int64; UserID int64; PostID int64; InteractionType int16; Metadata map[string]interface{}; CreatedAt time.Time }
// interaction_type: 0=like,1=comment,2=share,3=gift,4=bounty_answer,5=emoji
```

## Community + Collection (community.go, collection.go)
```go
type Community struct { ID int64; Name, Description, Slug string; Tags []string; CuratorLock bool
    SlowBootDays int; ForkedFrom *int64; CreatedBy int64; CreatedAt time.Time; ArchivedAt *time.Time }
type Collection struct { ID int64; OwnerID int64; Name, Description string; Visibility int16; IsDefault bool; CreatedAt,UpdatedAt *time.Time }
// visibility: 0=private,1=public_read,2=public_edit,3=anybody_edit
```

## Circle + CommunityNote (circle.go, note.go)
```go
type Circle struct { ID int64; Name, Description string; TagID *int64; GridCell string; MemberCount int; IsActive bool; LastActivity *time.Time; CreatedAt time.Time }
type CircleMember struct { CircleID, UserID int64; Status int16; SuggestedAt,JoinedAt *time.Time }
// status: 0=pending,1=accepted,2=dismissed,3=left
type CommunityNote struct { ID int64; PostID int64; AuthorID int64; Body string; Status int16; HelpfulYes,HelpfulNo int; ConsensusScore float64; CreatedAt,UpdatedAt *time.Time }
type CommunityNoteVote struct { ID int64; NoteID int64; UserID int64; Vote bool; TrustScoreAtVote float64; CreatedAt time.Time }
```

## Moderation (moderation.go, blocklist.go)
```go
type ModerationAction struct { ID int64; ActionType int16; TargetUserID *int64; TargetPostID *int64; ModeratorID int64; Reason string; Duration *time.Duration; IsJuryDecision bool; JuryYes,JuryNo,JuryTotal int; CreatedAt time.Time }
type JuryPanel struct { ID int64; TargetActionID int64; JurorID int64; Vote *bool; Reason string; CreatedAt time.Time }
type BlocklistEntry struct { ID int64; EntryType int16; EntryValue string; Reason string; Severity int16; AddedBy *int64; JuryApproved bool; Shared bool; CreatedAt time.Time }
type PostReport struct { ID int64; PostID int64; ReporterID int64; Category int16; Reason string; Status int16; ResolvedBy *int64; CreatedAt time.Time; ResolvedAt *time.Time }
```

## Economy (credit.go)
```go
type CreditTransaction struct { ID int64; FromUser *int64; ToUser *int64; Amount int64; TransactionType int16; ReferenceID *int64; Hash string; CreatedAt time.Time }
type DailyReward struct { ID int64; UserID int64; Date string; Amount int64; Claimed bool }
type Bounty struct { ID int64; PostID int64; CreatorID int64; TotalAmount int64; Status int16; BestAnswerID *int64; ExpiresAt *time.Time; CreatedAt time.Time }
```

## Social (trust.go, block.go, notification.go)
```go
type TrustConnection struct { ID int64; TrusterID,TrusteeID int64; Weight float64; Signature string; CreatedAt time.Time; ExpiresAt *time.Time }
type BlockedUser struct { ID int64; BlockerID,BlockedID int64; CreatedAt time.Time }
type Notification struct { ID int64; UserID int64; NotificationType int16; ActorID *int64; PostID *int64; Body string; IsRead bool; CreatedAt time.Time }
type ModNote struct { ID int64; UserID,NotedBy int64; Note string; CreatedAt time.Time }
```

## Achievement + Feed + Filter (achievement.go, feed.go, filter.go)
```go
type Achievement struct { ID int; Code,Name,Description,Icon string; Category int16; SortOrder int16 }
type UserAchievement struct { ID int64; UserID int64; AchievementID int; UnlockedAt time.Time; Progress float64; Visible bool }
type CustomFeed struct { ID int64; OwnerID int64; Name,Description,Slug string; IsPublic bool; SortOrder int16; CreatedAt,UpdatedAt time.Time }
type FeedSource struct { ID int64; FeedID int64; SourceType int16; SourceID *int64; SourceValue string; IncludeMode bool; SortPriority int16 }
type ContentFilter struct { ID int64; UserID int64; FilterType int16; FilterValue string; FilterAction int16; IsActive bool; CreatedAt time.Time }
```

## Misc (trending.go, oauth.go, page.go, emoji.go, reg_queue.go)
```go
type TrendingTopic struct { ID int64; Topic string; Frequency int; Velocity float64; TagID *int64; DetectionWindow interface{}; CreatedAt time.Time }
type OAuthState struct { ID int64; Provider,State string; UserID *int64; CreatedAt,ExpiresAt time.Time }
type CustomPage struct { ID int; Slug,Title,Body string; IsPublished bool; UpdatedAt time.Time }
type CustomEmoji struct { ID int; Shortcode,ImageURL,AltText string; CreatedAt time.Time }
```

## DB Nullable Scan Rule
All TIMESTAMPTZ columns are nullable → scan into `*time.Time`, not `time.Time`.  
All VARCHAR columns without NOT NULL → scan into `*string`.  
All BIGINT references without NOT NULL → scan into `*int64`.
