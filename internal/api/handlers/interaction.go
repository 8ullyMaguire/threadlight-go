package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type InteractionHandler struct {
	interactionSvc *services.InteractionService
}

func NewInteractionHandler(interactionSvc *services.InteractionService) *InteractionHandler {
	return &InteractionHandler{interactionSvc: interactionSvc}
}

func (h *InteractionHandler) CreateInteraction(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req model.Interaction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = userID
	interaction, err := h.interactionSvc.CreateInteraction(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, interaction)
}

func (h *InteractionHandler) GetPostInteractions(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}
	interactionType, _ := strconv.ParseInt(c.DefaultQuery("type", "0"), 10, 64)
	interactions, err := h.interactionSvc.GetPostInteractions(c.Request.Context(), postID, int16(interactionType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, interactions)
}

func (h *InteractionHandler) RemoveInteraction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interaction id"})
		return
	}
	if err := h.interactionSvc.RemoveInteraction(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "interaction removed"})
}

func (h *InteractionHandler) HasUserInteracted(c *gin.Context) {
	userID := c.GetInt64("user_id")
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}
	interactionType, _ := strconv.ParseInt(c.DefaultQuery("type", "0"), 10, 64)
	has, err := h.interactionSvc.HasUserInteracted(c.Request.Context(), userID, postID, int16(interactionType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"has_interacted": has})
}
