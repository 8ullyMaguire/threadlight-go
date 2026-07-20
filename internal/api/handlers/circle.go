package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type CircleHandler struct {
	circleSvc *services.CircleService
}

func NewCircleHandler(circleSvc *services.CircleService) *CircleHandler {
	return &CircleHandler{circleSvc: circleSvc}
}

func (h *CircleHandler) CreateCircle(c *gin.Context) {
	var req model.Circle
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	circle, err := h.circleSvc.CreateCircle(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, circle)
}

func (h *CircleHandler) GetCircle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	circle, err := h.circleSvc.GetCircle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, circle)
}

func (h *CircleHandler) UpdateCircle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	var req model.Circle
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	circle, err := h.circleSvc.UpdateCircle(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, circle)
}

func (h *CircleHandler) ListCircles(c *gin.Context) {
	circles, err := h.circleSvc.ListCircles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, circles)
}

type joinCircleRequest struct {
	Status int16 `json:"status"`
}

func (h *CircleHandler) JoinCircle(c *gin.Context) {
	circleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	userID := c.GetInt64("user_id")
	var req joinCircleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Status = 1
	}
	if err := h.circleSvc.JoinCircle(c.Request.Context(), circleID, userID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "joined circle"})
}

func (h *CircleHandler) LeaveCircle(c *gin.Context) {
	circleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.circleSvc.LeaveCircle(c.Request.Context(), circleID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left circle"})
}

func (h *CircleHandler) SuggestMember(c *gin.Context) {
	circleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.circleSvc.SuggestMember(c.Request.Context(), circleID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member suggested"})
}

func (h *CircleHandler) GetMembers(c *gin.Context) {
	circleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid circle id"})
		return
	}
	members, err := h.circleSvc.GetMembers(c.Request.Context(), circleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}
