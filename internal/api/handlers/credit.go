package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type CreditHandler struct {
	creditSvc *services.CreditService
	configSvc *services.ConfigService
}

func NewCreditHandler(creditSvc *services.CreditService, configSvc *services.ConfigService) *CreditHandler {
	return &CreditHandler{creditSvc: creditSvc, configSvc: configSvc}
}

func (h *CreditHandler) TransferCredits(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		ToUser int64 `json:"to_user" binding:"required"`
		Amount int64 `json:"amount" binding:"required"`
		TxType int16 `json:"tx_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := h.creditSvc.TransferCredits(c.Request.Context(), &userID, &req.ToUser, req.Amount, req.TxType, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *CreditHandler) GetTransactions(c *gin.Context) {
	userID := c.GetInt64("user_id")
	txs, err := h.creditSvc.GetTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}

func (h *CreditHandler) ClaimDailyReward(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		req.Amount = 50
	}
	reward, err := h.creditSvc.ClaimDailyReward(c.Request.Context(), userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reward)
}

func (h *CreditHandler) GetDailyRewardStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")
	status, err := h.creditSvc.GetDailyRewardStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (h *CreditHandler) CreateBounty(c *gin.Context) {
	var req model.Bounty
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CreatorID = c.GetInt64("user_id")
	bounty, err := h.creditSvc.CreateBounty(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bounty)
}

func (h *CreditHandler) GetBounty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bounty id"})
		return
	}
	bounty, err := h.creditSvc.GetBounty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bounty)
}

func (h *CreditHandler) AwardBounty(c *gin.Context) {
	bountyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bounty id"})
		return
	}
	var req struct {
		AnswerID int64 `json:"answer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.creditSvc.AwardBounty(c.Request.Context(), bountyID, req.AnswerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bounty awarded"})
}

// ── New Credit Economy Endpoints ──────────────────────────────────────────

// GetBalance returns the user's credit balance, streak, and today's quests.
func (h *CreditHandler) GetBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx := c.Request.Context()

	balance, err := h.creditSvc.GetBalance(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	streak, err := h.creditSvc.GetStreak(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	quests, err := h.creditSvc.GetDailyQuests(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
		"streak":  streak,
		"quests":  quests,
	})
}

// CompleteQuest marks a daily quest as completed.
func (h *CreditHandler) CompleteQuest(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		QuestType int16 `json:"quest_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.creditSvc.CompleteQuest(c.Request.Context(), userID, req.QuestType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "quest completed"})
}

// GetCosts returns the credit action cost configuration.
func (h *CreditHandler) GetCosts(c *gin.Context) {
	cfg, err := h.configSvc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"credit_action_costs": cfg.CreditActionCosts})
}
