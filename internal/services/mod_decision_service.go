package services

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

// ModDecisionReviewsService handles public moderation decision reviews.
// Any logged-in user can vote "Fair" (1) or "Unfair" (-1) on any
// moderation action. A controversy score surfaces the most debated
// decisions for the /modlog/controversial view.
type ModDecisionReviewsService struct {
	pg *pgxpool.Pool
}

func NewModDecisionReviewsService(pg *pgxpool.Pool) *ModDecisionReviewsService {
	return &ModDecisionReviewsService{pg: pg}
}

// CastVote inserts or updates a user's review of a moderation action.
func (s *ModDecisionReviewsService) CastVote(ctx context.Context, userID, actionID int64, vote int16) (*model.ModDecisionReview, error) {
	if vote != 1 && vote != -1 {
		return nil, model.ErrValidation
	}
	review := &model.ModDecisionReview{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO mod_decision_reviews (user_id, moderation_action_id, vote)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, moderation_action_id)
		 DO UPDATE SET vote = EXCLUDED.vote, created_at = NOW()
		 RETURNING id, user_id, moderation_action_id, vote, created_at`,
		userID, actionID, vote,
	).Scan(&review.ID, &review.UserID, &review.ModerationActionID, &review.Vote, &review.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("cast vote: %w", err)
	}
	return review, nil
}

// GetUserVote returns the user's vote for a specific action, or nil if not voted.
func (s *ModDecisionReviewsService) GetUserVote(ctx context.Context, userID, actionID int64) (*int16, error) {
	var vote int16
	err := s.pg.QueryRow(ctx,
		`SELECT vote FROM mod_decision_reviews WHERE user_id = $1 AND moderation_action_id = $2`,
		userID, actionID,
	).Scan(&vote)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

// GetActionReviews returns all reviews for a single moderation action.
func (s *ModDecisionReviewsService) GetActionReviews(ctx context.Context, actionID int64) ([]model.ModDecisionReview, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, user_id, moderation_action_id, vote, created_at
		 FROM mod_decision_reviews WHERE moderation_action_id = $1
		 ORDER BY created_at DESC`, actionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.ModDecisionReview
	for rows.Next() {
		var r model.ModDecisionReview
		if err := rows.Scan(&r.ID, &r.UserID, &r.ModerationActionID, &r.Vote, &r.CreatedAt); err != nil {
			return reviews, err
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}

// GetReviewCounts returns aggregated fair/unfair counts for an action.
func (s *ModDecisionReviewsService) GetReviewCounts(ctx context.Context, actionID int64) (*model.ModDecisionReviewCounts, error) {
	var counts model.ModDecisionReviewCounts
	counts.ActionID = actionID
	err := s.pg.QueryRow(ctx,
		`SELECT
		   COUNT(*) FILTER (WHERE vote = 1) AS fair,
		   COUNT(*) FILTER (WHERE vote = -1) AS unfair,
		   COUNT(*) AS total
		 FROM mod_decision_reviews WHERE moderation_action_id = $1`, actionID,
	).Scan(&counts.Fair, &counts.Unfair, &counts.Total)
	if err != nil {
		return nil, err
	}
	return &counts, nil
}

// controversyScore computes (fair+unfair) * min(fair,unfair) / max(fair,unfair).
// Ranges from 0 (unanimous) upward. 0 if either side has 0 votes.
func controversyScore(fair, unfair int) float64 {
	total := fair + unfair
	if total == 0 || fair == 0 || unfair == 0 {
		return 0
	}
	f, u := float64(fair), float64(unfair)
	return float64(total) * math.Min(f, u) / math.Max(f, u)
}

// GetControversial returns moderation actions ordered by controversy score,
// including aggregated review counts.
func (s *ModDecisionReviewsService) GetControversial(ctx context.Context, limit, offset int) ([]model.ControversialDecision, error) {
	rows, err := s.pg.Query(ctx, `
		SELECT
			a.id, a.action_type, a.target_user_id, a.target_post_id,
			a.moderator_id, a.reason, a.duration, a.is_jury_decision,
			a.jury_yes, a.jury_no, a.jury_total, a.created_at,
			COALESCE(r.fair, 0)   AS fair,
			COALESCE(r.unfair, 0) AS unfair,
			COALESCE(r.total, 0)  AS total
		FROM moderation_actions a
		LEFT JOIN (
			SELECT
				moderation_action_id,
				COUNT(*) FILTER (WHERE vote = 1) AS fair,
				COUNT(*) FILTER (WHERE vote = -1) AS unfair,
				COUNT(*) AS total
			FROM mod_decision_reviews
			GROUP BY moderation_action_id
		) r ON r.moderation_action_id = a.id
		WHERE r.total > 0
		ORDER BY (r.total * LEAST(r.fair, r.unfair)::float8 / GREATEST(r.fair, r.unfair)::float8) DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []model.ControversialDecision
	for rows.Next() {
		var d model.ControversialDecision
		var dur *string
		if err := rows.Scan(
			&d.Action.ID, &d.Action.ActionType, &d.Action.TargetUserID, &d.Action.TargetPostID,
			&d.Action.ModeratorID, &d.Action.Reason, &dur, &d.Action.IsJuryDecision,
			&d.Action.JuryYes, &d.Action.JuryNo, &d.Action.JuryTotal, &d.Action.CreatedAt,
			&d.ReviewCounts.Fair, &d.ReviewCounts.Unfair, &d.ReviewCounts.Total,
		); err != nil {
			return decisions, err
		}
		d.ReviewCounts.ActionID = d.Action.ID
		d.Controversy = controversyScore(d.ReviewCounts.Fair, d.ReviewCounts.Unfair)
		decisions = append(decisions, d)
	}
	return decisions, nil
}
