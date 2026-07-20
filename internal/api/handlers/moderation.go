package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type ModerationHandler struct {
	moderationSvc *services.ModerationService
}

func NewModerationHandler(moderationSvc *services.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationSvc: moderationSvc}
}

func (h *ModerationHandler) CreateAction(c *gin.Context) {
	var req model.ModerationAction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ModeratorID = c.GetInt64("user_id")
	action, err := h.moderationSvc.CreateAction(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, action)
}

func (h *ModerationHandler) GetAction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}
	action, err := h.moderationSvc.GetAction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, action)
}

func (h *ModerationHandler) ListActions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	actions, err := h.moderationSvc.ListActions(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actions)
}

func (h *ModerationHandler) AddJuror(c *gin.Context) {
	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}
	userID := c.GetInt64("user_id")
	panel, err := h.moderationSvc.AddJuror(c.Request.Context(), actionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, panel)
}

func (h *ModerationHandler) VoteJury(c *gin.Context) {
	panelID, err := strconv.ParseInt(c.Param("panel_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid panel id"})
		return
	}
	var req struct {
		Vote   bool   `json:"vote" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.moderationSvc.VoteJury(c.Request.Context(), panelID, req.Vote, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vote submitted"})
}

func (h *ModerationHandler) ResolveJury(c *gin.Context) {
	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}
	if err := h.moderationSvc.ResolveJury(c.Request.Context(), actionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "jury resolved"})
}
