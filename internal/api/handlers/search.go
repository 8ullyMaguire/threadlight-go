package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type SearchHandler struct {
	searchSvc *services.SearchService
}

func NewSearchHandler(searchSvc *services.SearchService) *SearchHandler {
	return &SearchHandler{searchSvc: searchSvc}
}

// @Summary      Search
// @Description  Perform a global search across posts, users, and communities
// @Tags         search
// @Param        q query string true "Search query"
// @Param        limit query int false "Max results (default 20)"
// @Param        offset query int false "Result offset (default 0)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}
	limit := parseIntParam(c.DefaultQuery("limit", "20"), 20)
	offset := parseIntParam(c.DefaultQuery("offset", "0"), 0)

	results, err := h.searchSvc.Search(c.Request.Context(), query, "", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if results == nil {
		results = []*model.Post{}
	}
	page := (offset / limit) + 1
	total := offset + len(results)
	if len(results) >= limit {
		total = offset + limit + 1
	}
	c.JSON(http.StatusOK, model.NewPaginated(results, page, limit, total))
}

// @Summary      Advanced search
// @Description  Perform an advanced search with a JSON body for more options
// @Tags         search
// @Param        body body object true "Advanced search query"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /search/advanced [post]
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var req struct {
		Query string `json:"q" binding:"required"`
		Limit int    `json:"limit"`
		Offset int   `json:"offset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.searchSvc.Search(c.Request.Context(), req.Query, "advanced", req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if results == nil {
		results = []*model.Post{}
	}
	page := (req.Offset / req.Limit) + 1
	total := req.Offset + len(results)
	if len(results) >= req.Limit {
		total = req.Offset + req.Limit + 1
	}
	c.JSON(http.StatusOK, model.NewPaginated(results, page, req.Limit, total))
}

// @Summary      Search posts
// @Description  Full-text post search with filters (author, tag, community, date range, mood, etc.)
// @Tags         search
// @Param        q query string true "Search query"
// @Param        author query string false "Filter by author username"
// @Param        tag query string false "Filter by tag"
// @Param        community query string false "Filter by community"
// @Param        date_from query string false "Filter posts from date (RFC3339)"
// @Param        date_to query string false "Filter posts to date (RFC3339)"
// @Param        sort query string false "Sort order (relevance, newest, oldest) default relevance"
// @Param        page query int false "Page number (default 1)"
// @Param        limit query int false "Results per page (default 20)"
// @Param        mood query int false "Filter by mood"
// @Param        content_type query int false "Filter by content type"
// @Param        is_educational query bool false "Filter educational posts"
// @Param        is_nsfw query bool false "Filter NSFW posts"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /search/posts [get]
func (h *SearchHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}

	q := model.SearchPostsQuery{
		Query:         query,
		Author:        c.Query("author"),
		Tag:           c.Query("tag"),
		Community:     c.Query("community"),
		DateFrom:      c.Query("date_from"),
		DateTo:        c.Query("date_to"),
		Sort:          c.DefaultQuery("sort", "relevance"),
		Page:          parseIntParam(c.DefaultQuery("page", "1"), 1),
		Limit:         parseIntParam(c.DefaultQuery("limit", "20"), 20),
	}

	if mood := c.Query("mood"); mood != "" {
		if m, err := strconv.Atoi(mood); err == nil {
			q.Mood = int16(m)
		}
	}
	if ct := c.Query("content_type"); ct != "" {
		if ctv, err := strconv.Atoi(ct); err == nil {
			q.ContentType = int16(ctv)
		}
	}
	if ie := c.Query("is_educational"); ie != "" {
		v := ie == "true"
		q.IsEducational = &v
	}
	if ins := c.Query("is_nsfw"); ins != "" {
		v := ins == "true"
		q.IsNSFW = &v
	}

	results, total, _, err := h.searchSvc.SearchPosts(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []model.PostWithHeadline{}
	}
	c.JSON(http.StatusOK, model.NewPaginated(results, q.Page, q.Limit, total))
}

// @Summary      Search users
// @Description  Search for users by username or display name
// @Tags         search
// @Param        q query string true "Search query"
// @Param        page query int false "Page number (default 1)"
// @Param        limit query int false "Results per page (default 20)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /search/users [get]
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}

	page := parseIntParam(c.DefaultQuery("page", "1"), 1)
	limit := parseIntParam(c.DefaultQuery("limit", "20"), 20)

	results, total, err := h.searchSvc.SearchUsers(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []model.User{}
	}
	c.JSON(http.StatusOK, model.NewPaginated(results, page, limit, total))
}

// @Summary      Search communities
// @Description  Search for communities by name or description
// @Tags         search
// @Param        q query string true "Search query"
// @Param        page query int false "Page number (default 1)"
// @Param        limit query int false "Results per page (default 20)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /search/communities [get]
func (h *SearchHandler) SearchCommunities(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}

	page := parseIntParam(c.DefaultQuery("page", "1"), 1)
	limit := parseIntParam(c.DefaultQuery("limit", "20"), 20)

	results, total, err := h.searchSvc.SearchCommunities(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []model.Community{}
	}
	c.JSON(http.StatusOK, model.NewPaginated(results, page, limit, total))
}

// @Summary      Search suggestions
// @Description  Get prefix-based autocomplete suggestions for users, communities, and tags
// @Tags         search
// @Param        q query string true "Search prefix"
// @Success      200 {object} map[string]interface{}
// @Router       /search/suggest [get]
func (h *SearchHandler) SearchSuggest(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}
	suggestions, err := h.searchSvc.GetSuggestions(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Flatten all suggestions into a single list
	var all []string
	all = append(all, suggestions.Users...)
	all = append(all, suggestions.Communities...)
	all = append(all, suggestions.Tags...)
	c.JSON(http.StatusOK, gin.H{"suggestions": all})
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var v int
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal
		}
		v = v*10 + int(c-'0')
	}
	return v
}
