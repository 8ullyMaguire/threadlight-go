package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type CommunityHandler struct {
	communitySvc *services.CommunityService
}

func NewCommunityHandler(communitySvc *services.CommunityService) *CommunityHandler {
	return &CommunityHandler{communitySvc: communitySvc}
}

func (h *CommunityHandler) Create(c *gin.Context) {
	var req model.Community
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CreatedBy = c.GetInt64("user_id")

	community, err := h.communitySvc.CreateCommunity(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, community)
}

func (h *CommunityHandler) GetByID(c *gin.Context) {
	slug := c.Param("slug")

	community, err := h.communitySvc.GetCommunityBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, community)
}

func (h *CommunityHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	var req model.Community
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	community, err := h.communitySvc.UpdateCommunity(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, community)
}

func (h *CommunityHandler) Archive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	if err := h.communitySvc.ArchiveCommunity(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "community archived"})
}

func (h *CommunityHandler) List(c *gin.Context) {
	communities, err := h.communitySvc.ListCommunities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, communities)
}

func (h *CommunityHandler) Join(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	var req struct {
		Role   int16 `json:"role"`
		Status int16 `json:"status"`
	}
	c.ShouldBindJSON(&req)

	if err := h.communitySvc.JoinCommunity(c.Request.Context(), communityID, userID, req.Role, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "joined community"})
}

func (h *CommunityHandler) Leave(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}
	uid, _ := c.Get("user_id")

	if err := h.communitySvc.LeaveCommunity(c.Request.Context(), communityID, uid.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left community"})
}

func (h *CommunityHandler) Members(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	members, err := h.communitySvc.GetMembers(c.Request.Context(), communityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *CommunityHandler) UpdateMemberRole(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
		Role   int16 `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.communitySvc.UpdateMemberRole(c.Request.Context(), communityID, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *CommunityHandler) AddCurator(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	var req struct {
		UserID     int64 `json:"user_id" binding:"required"`
		Permission int16 `json:"permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	curator, err := h.communitySvc.AddCurator(c.Request.Context(), communityID, req.UserID, req.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, curator)
}

func (h *CommunityHandler) RemoveCurator(c *gin.Context) {
	communityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.communitySvc.RemoveCurator(c.Request.Context(), communityID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "curator removed"})
}

func (h *CommunityHandler) Fork(c *gin.Context) {
	sourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community id"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("user_id")

	fork, err := h.communitySvc.ForkCommunity(c.Request.Context(), sourceID, uid.(int64), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fork)
}
