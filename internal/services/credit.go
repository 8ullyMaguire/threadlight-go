package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

type CreditService struct {
	pg *pgxpool.Pool
}

func NewCreditService(pg *pgxpool.Pool) *CreditService {
	return &CreditService{pg: pg}
}

func (s *CreditService) TransferCredits(ctx context.Context, fromUser, toUser *int64, amount int64, txType int16, refID *int64) (*model.CreditTransaction, error) {
	hashInput := fmt.Sprintf("%d-%d-%d-%d-%d", time.Now().UnixNano(), fromUser, toUser, amount, txType)
	hash := sha256.Sum256([]byte(hashInput))
	hashStr := hex.EncodeToString(hash[:])

	tx := &model.CreditTransaction{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO credit_transactions (from_user, to_user, amount, transaction_type, reference_id, hash)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, from_user, to_user, amount, transaction_type, reference_id, hash, created_at`,
		fromUser, toUser, amount, txType, refID, hashStr,
	).Scan(&tx.ID, &tx.FromUser, &tx.ToUser, &tx.Amount, &tx.TransactionType, &tx.ReferenceID, &tx.Hash, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *CreditService) GetTransactions(ctx context.Context, userID int64) ([]model.CreditTransaction, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, from_user, to_user, amount, transaction_type, reference_id, hash, created_at
		 FROM credit_transactions WHERE from_user = $1 OR to_user = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []model.CreditTransaction
	for rows.Next() {
		var t model.CreditTransaction
		if err := rows.Scan(&t.ID, &t.FromUser, &t.ToUser, &t.Amount, &t.TransactionType, &t.ReferenceID, &t.Hash, &t.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, nil
}

func (s *CreditService) ClaimDailyReward(ctx context.Context, userID int64, amount int64) (*model.DailyReward, error) {
	today := time.Now().Format("2006-01-02")
	reward := &model.DailyReward{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO daily_rewards (user_id, date, amount, claimed)
		 VALUES ($1, $2, $3, TRUE)
		 ON CONFLICT (user_id, date) DO UPDATE SET claimed = TRUE, amount = $3
		 RETURNING id, user_id, date, amount, claimed, created_at`,
		userID, today, amount,
	).Scan(&reward.ID, &reward.UserID, &reward.Date, &reward.Amount, &reward.Claimed, &reward.CreatedAt)
	if err != nil {
		return nil, err
	}
	return reward, nil
}

func (s *CreditService) GetDailyRewardStatus(ctx context.Context, userID int64) (*model.DailyReward, error) {
	today := time.Now().Format("2006-01-02")
	reward := &model.DailyReward{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, user_id, date, amount, claimed, created_at
		 FROM daily_rewards WHERE user_id = $1 AND date = $2`,
		userID, today,
	).Scan(&reward.ID, &reward.UserID, &reward.Date, &reward.Amount, &reward.Claimed, &reward.CreatedAt)
	if err != nil {
		return nil, nil
	}
	return reward, nil
}

func (s *CreditService) CreateBounty(ctx context.Context, req model.Bounty) (*model.Bounty, error) {
	bounty := &model.Bounty{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO bounties (post_id, creator_id, total_amount, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, post_id, creator_id, total_amount, status, best_answer_id, expires_at, created_at`,
		req.PostID, req.CreatorID, req.TotalAmount, req.Status, req.ExpiresAt,
	).Scan(&bounty.ID, &bounty.PostID, &bounty.CreatorID, &bounty.TotalAmount, &bounty.Status,
		&bounty.BestAnswerID, &bounty.ExpiresAt, &bounty.CreatedAt)
	if err != nil {
		return nil, err
	}
	return bounty, nil
}

func (s *CreditService) GetBounty(ctx context.Context, bountyID int64) (*model.Bounty, error) {
	bounty := &model.Bounty{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, post_id, creator_id, total_amount, status, best_answer_id, expires_at, created_at
		 FROM bounties WHERE id = $1`, bountyID,
	).Scan(&bounty.ID, &bounty.PostID, &bounty.CreatorID, &bounty.TotalAmount, &bounty.Status,
		&bounty.BestAnswerID, &bounty.ExpiresAt, &bounty.CreatedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return bounty, nil
}

func (s *CreditService) AwardBounty(ctx context.Context, bountyID, answerID int64) error {
	_, err := s.pg.Exec(ctx,
		`UPDATE bounties SET status = $2, best_answer_id = $3 WHERE id = $1`,
		bountyID, 1, answerID)
	return err
}
