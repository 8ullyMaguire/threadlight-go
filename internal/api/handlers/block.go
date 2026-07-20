package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/services"
)

type BlockHandler struct {
	blockSvc *services.BlockService
}

func NewBlockHandler(blockSvc *services.BlockService) *BlockHandler {
	return &BlockHandler{blockSvc: blockSvc}
}

type blockRequest struct {
	BlockedID int64 `json:"blocked_id" binding:"required"`
}

func (h *BlockHandler) Create(c *gin.Context) {
	var req blockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("user_id")
	blockerID := uid.(int64)

	block, err := h.blockSvc.BlockUser(c.Request.Context(), blockerID, req.BlockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, block)
}

func (h *BlockHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block id"})
		return
	}

	if err := h.blockSvc.UnblockUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user unblocked"})
}

func (h *BlockHandler) List(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	blocks, err := h.blockSvc.ListBlocks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, blocks)
}

func (h *BlockHandler) Check(c *gin.Context) {
	blockedID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	uid, _ := c.Get("user_id")
	blockerID := uid.(int64)

	blocked, err := h.blockSvc.IsBlocked(c.Request.Context(), blockerID, blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blocked": blocked})
}
