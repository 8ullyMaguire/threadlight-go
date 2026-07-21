package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type CommunityHandler struct {
	svc       *services.CommunityService
	configSvc *services.ConfigService
}

func NewCommunityHandler(svc *services.CommunityService, configSvc *services.ConfigService) *CommunityHandler {
	return &CommunityHandler{svc: svc, configSvc: configSvc}
}

func (h *CommunityHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Check trust-level privilege: user must meet the minimum trust level
	// for creating communities (configurable via site_config, default 0).
	cfg, cfgErr := h.configSvc.Get(c.Request.Context())
	if cfgErr == nil && cfg.MinTrustLevelForCommunityCreate > 0 {
		var userTrustLevel int16
		err := h.svc.GetUserTrustLevel(c.Request.Context(), userID.(int64), &userTrustLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify trust level"})
			return
		}
		if userTrustLevel < cfg.MinTrustLevelForCommunityCreate {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient trust level to create a community"})
			return
		}
	}

	var req model.Community
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CreatedBy = userID.(int64)

	// Default values
	if req.MinTrustScore == 0 {
		req.MinTrustScore = 0.5
	}

	community, err := h.svc.CreateCommunity(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, community)
}

func (h *CommunityHandler) List(c *gin.Context) {
	communities, err := h.svc.ListCommunities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if communities == nil {
		communities = []model.Community{}
	}
	c.JSON(http.StatusOK, communities)
}

func (h *CommunityHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}
	c.JSON(http.StatusOK, community)
}

func (h *CommunityHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	existing, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	var req model.Community
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.svc.UpdateCommunity(c.Request.Context(), existing.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *CommunityHandler) Archive(c *gin.Context) {
	slug := c.Param("slug")
	existing, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}
	if err := h.svc.ArchiveCommunity(c.Request.Context(), existing.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "archived"})
}

func (h *CommunityHandler) Join(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	// Check invite requirements
	if err := h.svc.CanJoin(c.Request.Context(), community.ID, userID.(int64)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	status := int16(1)
	if community.SlowBootDays > 0 {
		status = 0 // probation
	}
	if err := h.svc.JoinCommunity(c.Request.Context(), community.ID, userID.(int64), 0, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

func (h *CommunityHandler) Leave(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	if err := h.svc.LeaveCommunity(c.Request.Context(), community.ID, userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func (h *CommunityHandler) Members(c *gin.Context) {
	slug := c.Param("slug")
	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}
	members, err := h.svc.GetMembers(c.Request.Context(), community.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *CommunityHandler) AddCurator(c *gin.Context) {
	slug := c.Param("slug")
	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	userIDStr := c.PostForm("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Check trust-level privilege: curator candidate must meet the minimum
	// trust level (configurable via site_config, default level 1).
	cfg, cfgErr := h.configSvc.Get(c.Request.Context())
	if cfgErr == nil && cfg.MinTrustLevelForCurator > 0 {
		var targetTrustLevel int16
		if err := h.svc.GetUserTrustLevel(c.Request.Context(), userID, &targetTrustLevel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify trust level"})
			return
		}
		if targetTrustLevel < cfg.MinTrustLevelForCurator {
			c.JSON(http.StatusForbidden, gin.H{"error": "user does not meet the minimum trust level to become a curator"})
			return
		}
	}

	curator, err := h.svc.AddCurator(c.Request.Context(), community.ID, userID, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, curator)
}

func (h *CommunityHandler) RemoveCurator(c *gin.Context) {
	slug := c.Param("slug")
	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.svc.RemoveCurator(c.Request.Context(), community.ID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "curator removed"})
}

func (h *CommunityHandler) Fork(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fork, err := h.svc.ForkCommunity(c.Request.Context(), community.ID, userID.(int64), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fork)
}

// Trust graph endpoints
func (h *CommunityHandler) GetTrustGraph(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	tcs, err := h.svc.GetTrustGraph(c.Request.Context(), community.ID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tcs)
}

func (h *CommunityHandler) CreateTrustConnection(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	var req struct {
		TrusteeID int64 `json:"trustee_id" binding:"required"`
		Weight    int   `json:"weight" binding:"min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tc, err := h.svc.CreateTrustConnection(c.Request.Context(), community.ID, userID.(int64), req.TrusteeID, req.Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tc)
}

func (h *CommunityHandler) GetCommunityTrustScore(c *gin.Context) {
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		userID, _ := c.Get("user_id")
		userIDStr = strconv.FormatInt(userID.(int64), 10)
	}
	uid, _ := strconv.ParseInt(userIDStr, 10, 64)

	score, err := h.svc.GetCommunityTrustScore(c.Request.Context(), community.ID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trust_score": score})
}

func (h *CommunityHandler) CheckInviteStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	slug := c.Param("slug")

	community, err := h.svc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}

	err = h.svc.CanJoin(c.Request.Context(), community.ID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"can_join": false, "reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"can_join": true})
}
