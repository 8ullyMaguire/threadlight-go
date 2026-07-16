package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type ReportHandler struct {
	reportSvc *services.ReportService
}

func NewReportHandler(reportSvc *services.ReportService) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc}
}

type createReportRequest struct {
	PostID   int64  `json:"post_id" binding:"required"`
	Category int16  `json:"category" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

type resolveReportRequest struct {
	Status int16 `json:"status" binding:"required"`
}

func (h *ReportHandler) Create(c *gin.Context) {
	var req createReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("user_id")
	reporterID := uid.(int64)

	report := model.PostReport{
		PostID:     req.PostID,
		ReporterID: reporterID,
		Category:   req.Category,
		Reason:     req.Reason,
	}

	created, err := h.reportSvc.CreateReport(c.Request.Context(), report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ReportHandler) List(c *gin.Context) {
	reports, err := h.reportSvc.ListReports(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	report, err := h.reportSvc.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Resolve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	var req resolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, _ := c.Get("user_id")
	resolvedBy := uid.(int64)

	if err := h.reportSvc.ResolveReport(c.Request.Context(), id, req.Status, resolvedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "report resolved"})
}
