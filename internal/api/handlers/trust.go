package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/services"
)

type TrustHandler struct {
	trustSvc *services.TrustService
}

func NewTrustHandler(trustSvc *services.TrustService) *TrustHandler {
	return &TrustHandler{trustSvc: trustSvc}
}

type createConnectionRequest struct {
	TrusteeID int64   `json:"trustee_id" binding:"required"`
	Weight    float64 `json:"weight" binding:"required"`
}

func (h *TrustHandler) CreateConnection(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req createConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conn, err := h.trustSvc.CreateConnection(c.Request.Context(), userID, req.TrusteeID, req.Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, conn)
}

func (h *TrustHandler) GetConnection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}
	conn, err := h.trustSvc.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conn)
}

func (h *TrustHandler) GetOutgoingConnections(c *gin.Context) {
	userID := c.GetInt64("user_id")
	conns, err := h.trustSvc.GetOutgoingConnections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conns)
}

func (h *TrustHandler) GetIncomingConnections(c *gin.Context) {
	userID := c.GetInt64("user_id")
	conns, err := h.trustSvc.GetIncomingConnections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conns)
}

type updateConnectionRequest struct {
	Weight float64 `json:"weight" binding:"required"`
}

func (h *TrustHandler) UpdateConnection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}
	var req updateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conn, err := h.trustSvc.UpdateConnection(c.Request.Context(), id, req.Weight)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conn)
}

func (h *TrustHandler) DeleteConnection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}
	if err := h.trustSvc.DeleteConnection(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "connection deleted"})
}
