package handlers
import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type FeedPluginHandler struct {
	feedPluginSvc *services.FeedPluginService
}

func NewFeedPluginHandler(feedPluginSvc *services.FeedPluginService) *FeedPluginHandler {
	return &FeedPluginHandler{feedPluginSvc: feedPluginSvc}
}

// Create handles WASM plugin upload via multipart form.
// POST /api/v1/feed-plugins
func (h *FeedPluginHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")

	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	description := c.PostForm("description")
	version := c.DefaultPostForm("version", "1.0.0")
	pluginTypeStr := c.DefaultPostForm("plugin_type", "0")
	pluginType, err := strconv.ParseInt(pluginTypeStr, 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin_type"})
		return
	}
	priceCreditsStr := c.DefaultPostForm("price_credits", "0")
	priceCredits, err := strconv.ParseInt(priceCreditsStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price_credits"})
		return
	}

	file, _, err := c.Request.FormFile("wasm")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wasm file is required"})
		return
	}
	defer file.Close()

	wasmBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read wasm file"})
		return
	}

	const maxWasmSize = 10 << 20
	if len(wasmBytes) > maxWasmSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wasm file exceeds maximum size of 10 MB"})
		return
	}

	// SHA256 hash for integrity verification
	wasmSHA256 := fmt.Sprintf("%x", sha256.Sum256(wasmBytes))

	req := model.CreateFeedPluginRequest{
		Name:         name,
		Description:  description,
		Version:      version,
		PluginType:   int16(pluginType),
		PriceCredits: priceCredits,
	}

	plugin, err := h.feedPluginSvc.CreatePlugin(c.Request.Context(), userID, req, wasmBytes, wasmSHA256)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plugin)
}

// Get returns a single plugin by ID.
// GET /api/v1/feed-plugins/:id
func (h *FeedPluginHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin id"})
		return
	}

	plugin, err := h.feedPluginSvc.GetPlugin(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plugin)
}

// List returns all reviewed plugins.
// GET /api/v1/feed-plugins
func (h *FeedPluginHandler) List(c *gin.Context) {
	reviewedStr := c.DefaultQuery("reviewed", "true")
	reviewed := reviewedStr == "true"
	sort := c.DefaultQuery("sort", "newest")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	plugins, err := h.feedPluginSvc.ListPlugins(c.Request.Context(), &reviewed, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if plugins == nil {
		plugins = []model.FeedPlugin{}
	}
	// Truncate to limit on handler side
	if len(plugins) > limit {
		plugins = plugins[:limit]
	}
	c.JSON(http.StatusOK, plugins)
}

// Install installs a plugin for the authenticated user.
// POST /api/v1/feed-plugins/:id/install
func (h *FeedPluginHandler) Install(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin id"})
		return
	}

	userID := c.GetInt64("user_id")
	install, err := h.feedPluginSvc.InstallPlugin(c.Request.Context(), userID, id, nil)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		if errors.Is(err, model.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "plugin already installed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, install)
}

// Uninstall removes a plugin installation.
// POST /api/v1/feed-plugins/:id/uninstall
func (h *FeedPluginHandler) Uninstall(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin id"})
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.feedPluginSvc.UninstallPlugin(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin uninstalled"})
}

// ListInstalls returns the authenticated user's installed plugins.
// GET /api/v1/feed-plugins/installs
func (h *FeedPluginHandler) ListInstalls(c *gin.Context) {
	userID := c.GetInt64("user_id")
	installs, err := h.feedPluginSvc.ListInstalls(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if installs == nil {
		installs = []model.FeedPluginInstall{}
	}
	c.JSON(http.StatusOK, installs)
}

// Review submits or updates a review for a plugin.
// POST /api/v1/feed-plugins/:id/review
func (h *FeedPluginHandler) Review(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin id"})
		return
	}

	var req struct {
		Rating int16  `json:"rating" binding:"required,min=1,max=5"`
		Review string `json:"review,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	review, err := h.feedPluginSvc.ReviewPlugin(c.Request.Context(), userID, id, req.Rating, req.Review)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, review)
}

// Execute runs a plugin against sample post IDs.
// POST /api/v1/feed-plugins/:id/execute
func (h *FeedPluginHandler) Execute(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plugin id"})
		return
	}

	var req struct {
		PostIDs []int64            `json:"post_ids" binding:"required"`
		Context map[string]any     `json:"context,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.feedPluginSvc.ExecutePlugin(c.Request.Context(), id, req.PostIDs, req.Context)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}
