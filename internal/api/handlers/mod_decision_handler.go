package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type ModDecisionHandler struct {
	svc      *services.ModDecisionReviewsService
	configSvc *services.ConfigService
}

func NewModDecisionHandler(svc *services.ModDecisionReviewsService, configSvc *services.ConfigService) *ModDecisionHandler {
	return &ModDecisionHandler{svc: svc, configSvc: configSvc}
}

// CastVote   POST /api/v1/modlog/:id/review
func (h *ModDecisionHandler) CastVote(c *gin.Context) {
	userID := c.GetInt64("user_id")
	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}

	// Check trust-level privilege: user must meet the minimum trust level
	// for review voting (configurable via site_config, default level 1).
	cfg, cfgErr := h.configSvc.Get(c.Request.Context())
	if cfgErr == nil && cfg.MinTrustLevelForReviewVoting > 0 {
		var userTrustLevel int16
		if err := h.svc.GetUserTrustLevel(c.Request.Context(), userID, &userTrustLevel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify trust level"})
			return
		}
		if userTrustLevel < cfg.MinTrustLevelForReviewVoting {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient trust level to vote on moderation reviews"})
			return
		}
	}

	var req struct {
		Vote int16 `json:"vote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	review, err := h.svc.CastVote(c.Request.Context(), userID, actionID, req.Vote)
	if err == model.ErrValidation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vote must be 1 (fair) or -1 (unfair)"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// After each vote, check if the action now meets the unfair penalty threshold.
	// This is best-effort — a failure here does not block the vote itself.
	penaltyReason, penErr := h.svc.EvaluateUnfairPenalty(c.Request.Context(), actionID)
	if penErr == nil && penaltyReason != "" {
		// Include the penalty info in the response so the UI can show it
		c.JSON(http.StatusOK, gin.H{
			"review":         review,
			"trust_penalty":  penaltyReason,
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetActionReviews   GET /api/v1/modlog/:id/reviews
func (h *ModDecisionHandler) GetActionReviews(c *gin.Context) {
	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}
	reviews, err := h.svc.GetActionReviews(c.Request.Context(), actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	counts, err := h.svc.GetReviewCounts(c.Request.Context(), actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"counts":  counts,
	})
}

// GetMyVote   GET /api/v1/modlog/:id/my-review
func (h *ModDecisionHandler) GetMyVote(c *gin.Context) {
	userID := c.GetInt64("user_id")
	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}
	vote, err := h.svc.GetUserVote(c.Request.Context(), userID, actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vote": vote})
}

// GetControversial   GET /api/v1/modlog/controversial
func (h *ModDecisionHandler) GetControversial(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}

	decisions, err := h.svc.GetControversial(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if decisions == nil {
		decisions = []model.ControversialDecision{}
	}
	c.JSON(http.StatusOK, gin.H{
		"decisions": decisions,
		"limit":     limit,
		"offset":    offset,
	})
}
