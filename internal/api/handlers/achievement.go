package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/services"
)

type AchievementHandler struct {
	achievementSvc *services.AchievementService
}

func NewAchievementHandler(achievementSvc *services.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievementSvc: achievementSvc}
}

func (h *AchievementHandler) List(c *gin.Context) {
	achievements, err := h.achievementSvc.ListAchievements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, achievements)
}

func (h *AchievementHandler) GetUserAchievements(c *gin.Context) {
	uid, _ := c.Get("user_id")

	achievements, err := h.achievementSvc.GetUserAchievements(c.Request.Context(), uid.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, achievements)
}

func (h *AchievementHandler) Unlock(c *gin.Context) {
	uid, _ := c.Get("user_id")

	var req struct {
		AchievementID int     `json:"achievement_id" binding:"required"`
		Progress      float64 `json:"progress"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ua, err := h.achievementSvc.UnlockAchievement(c.Request.Context(), uid.(int64), req.AchievementID, req.Progress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ua)
}
