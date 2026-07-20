package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/services"
)

type TrendingHandler struct {
	trendingSvc *services.TrendingService
}

func NewTrendingHandler(trendingSvc *services.TrendingService) *TrendingHandler {
	return &TrendingHandler{trendingSvc: trendingSvc}
}

func (h *TrendingHandler) ListTrending(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	topics, err := h.trendingSvc.GetTrendingTopics(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topics)
}

type incrementTopicRequest struct {
	Topic string `json:"topic" binding:"required"`
}

func (h *TrendingHandler) IncrementTopic(c *gin.Context) {
	var req incrementTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.trendingSvc.IncrementTopic(c.Request.Context(), req.Topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "topic incremented"})
}

type addTopicRequest struct {
	Topic string `json:"topic" binding:"required"`
	TagID *int64 `json:"tag_id"`
}

func (h *TrendingHandler) AddTopic(c *gin.Context) {
	var req addTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topic, err := h.trendingSvc.AddTopic(c.Request.Context(), req.Topic, req.TagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, topic)
}
