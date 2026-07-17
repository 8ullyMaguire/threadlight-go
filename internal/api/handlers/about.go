package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/services"
)

type AboutHandler struct {
	configSvc *services.ConfigService
}

func NewAboutHandler(configSvc *services.ConfigService) *AboutHandler {
	return &AboutHandler{configSvc: configSvc}
}

func (h *AboutHandler) GetAbout(c *gin.Context) {
	about, err := h.configSvc.GetAbout(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, about)
}
