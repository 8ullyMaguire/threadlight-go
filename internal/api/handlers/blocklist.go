package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type BlocklistHandler struct {
	blocklistSvc *services.BlocklistService
}

func NewBlocklistHandler(blocklistSvc *services.BlocklistService) *BlocklistHandler {
	return &BlocklistHandler{blocklistSvc: blocklistSvc}
}

func (h *BlocklistHandler) Create(c *gin.Context) {
	var req model.BlocklistEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}
	req.AddedBy = &userID

	entry, err := h.blocklistSvc.AddEntry(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *BlocklistHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}

	if err := h.blocklistSvc.RemoveEntry(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "entry removed"})
}

func (h *BlocklistHandler) List(c *gin.Context) {
	entries, err := h.blocklistSvc.ListEntries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *BlocklistHandler) Check(c *gin.Context) {
	entryType, err := strconv.ParseInt(c.Query("type"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}
	entryValue := c.Query("value")
	if entryValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value required"})
		return
	}

	entry, err := h.blocklistSvc.CheckEntry(c.Request.Context(), int16(entryType), entryValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entry == nil {
		c.JSON(http.StatusOK, gin.H{"blocked": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blocked": true, "entry": entry})
}
