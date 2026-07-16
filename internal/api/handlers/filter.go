package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type FilterHandler struct {
	filterSvc *services.FilterService
}

func NewFilterHandler(filterSvc *services.FilterService) *FilterHandler {
	return &FilterHandler{filterSvc: filterSvc}
}

func (h *FilterHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req model.ContentFilter
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter, err := h.filterSvc.CreateFilter(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, filter)
}

func (h *FilterHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter id"})
		return
	}
	var req model.ContentFilter
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter, err := h.filterSvc.UpdateFilter(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, filter)
}

func (h *FilterHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter id"})
		return
	}
	if err := h.filterSvc.DeleteFilter(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "filter deleted"})
}

func (h *FilterHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	filters, err := h.filterSvc.ListUserFilters(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, filters)
}

func (h *FilterHandler) Check(c *gin.Context) {
	userID := c.GetInt64("user_id")
	filterType, err := strconv.ParseInt(c.DefaultQuery("type", "0"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}
	filterValue := c.Query("value")
	if filterValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value required"})
		return
	}
	filter, err := h.filterSvc.CheckFilter(c.Request.Context(), userID, int16(filterType), filterValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if filter == nil {
		c.JSON(http.StatusOK, gin.H{"blocked": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blocked": true, "filter": filter})
}
