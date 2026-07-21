package model

// SearchPostsQuery holds all search parameters for post search.
type SearchPostsQuery struct {
	Query         string `json:"query"`
	Author        string `json:"author"`
	Tag           string `json:"tag"`
	Community     string `json:"community"`
	Mood          int16  `json:"mood"`
	ContentType   int16  `json:"content_type"`
	IsEducational *bool  `json:"is_educational,omitempty"`
	IsNSFW        *bool  `json:"is_nsfw,omitempty"`
	DateFrom      string `json:"date_from"`
	DateTo        string `json:"date_to"`
	Sort          string `json:"sort"` // "relevance", "newest", "oldest"
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
}

// PostWithHeadline embeds a Post and adds a headline snippet from ts_headline.
type PostWithHeadline struct {
	Post
	Headline string `json:"headline"`
}

// SearchFacets holds aggregate counts that can be returned alongside results.
type SearchFacets struct {
	Tags        map[string]int `json:"tags"`
	Mood        map[int]int    `json:"mood"`
	ContentType map[int]int    `json:"content_type"`
}

// Suggestions holds auto-complete style results.
type Suggestions struct {
	Users       []string `json:"users"`
	Communities []string `json:"communities"`
	Tags        []string `json:"tags"`
}
