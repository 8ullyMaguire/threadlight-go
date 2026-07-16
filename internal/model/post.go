package model

import "time"

type Post struct {
	ID                      int64      `json:"id"`
	AuthorID                int64      `json:"author_id"`
	Title                   string     `json:"title"`
	Body                    string     `json:"body"`
	ContentType             int16      `json:"content_type"`
	Mood                    int16      `json:"mood"`
	IsEducational           bool       `json:"is_educational"`
	IsEntertaining          bool       `json:"is_entertaining"`
	IsNsfw                  bool       `json:"is_nsfw"`
	ContentWarning          *string    `json:"content_warning,omitempty"`
	InteractionCount        int64      `json:"interaction_count"`
	CumulativeInteractions  int64      `json:"cumulative_interactions"`
	Status                  int16      `json:"status"`
	ScheduledAt             *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt               *time.Time `json:"created_at,omitempty"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty"`
	ArchivedAt              *time.Time `json:"archived_at,omitempty"`
	IsDeleted               bool       `json:"is_deleted"`
}

type CreatePostRequest struct {
	Title          string `json:"title" binding:"required"`
	Body           string `json:"body" binding:"required"`
	ContentType    int16  `json:"content_type"`
	Mood           int16  `json:"mood"`
	IsEducational  bool   `json:"is_educational"`
	IsEntertaining bool   `json:"is_entertaining"`
	IsNsfw         bool   `json:"is_nsfw"`
	ContentWarning string `json:"content_warning"`
}

type UpdatePostRequest struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	ContentType    int16  `json:"content_type"`
	Mood           int16  `json:"mood"`
	IsEducational  bool   `json:"is_educational"`
	IsEntertaining bool   `json:"is_entertaining"`
	IsNsfw         bool   `json:"is_nsfw"`
	ContentWarning string `json:"content_warning"`
}

type PostResponse struct {
	Post    Post   `json:"post"`
	Message string `json:"message"`
}
