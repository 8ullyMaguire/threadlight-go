package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type ConfigHandler struct {
	configSvc *services.ConfigService
}

func NewConfigHandler(configSvc *services.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: configSvc}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configSvc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var req model.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.configSvc.Update(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *ConfigHandler) GenerateInvite(c *gin.Context) {
	userID, _ := c.Get("user_id")

	invite, err := h.configSvc.GenerateInvite(c.Request.Context(), userID.(int64))
	if err != nil {
		if err == model.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "no invites remaining this month"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, invite)
}

func (h *ConfigHandler) ListInvites(c *gin.Context) {
	userID, _ := c.Get("user_id")

	invites, err := h.configSvc.ListMyInvites(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invites)
}

func (h *ConfigHandler) GetInviteLimit(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, err := h.configSvc.GetInviteLimit(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, err := h.configSvc.InviteCountThisMonth(c.Request.Context(), userID.(int64))
	if err != nil {
		count = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"limit": limit,
		"used":  count,
	})
}
