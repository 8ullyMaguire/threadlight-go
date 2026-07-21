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
	c.JSON(http.StatusOK, results)
}

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
	c.JSON(http.StatusOK, results)
}

// SearchPosts handles GET /api/v1/search/posts — full-text post search with filters
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

	results, total, facets, err := h.searchSvc.SearchPosts(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"page":    q.Page,
		"limit":   q.Limit,
		"facets":  facets,
	})
}

// SearchUsers handles GET /api/v1/search/users
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

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// SearchCommunities handles GET /api/v1/search/communities
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

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// SearchSuggest handles GET /api/v1/search/suggest — prefix-based suggestions for autocomplete
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
