package worker

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunPrune(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("prune: starting daily content pruning")

	// Get prune config from site_config
	var ageDays, minInteractions int
	err := pg.QueryRow(ctx,
		`SELECT COALESCE(prune_age_days, 180), COALESCE(prune_min_interactions, 3)
		 FROM site_config WHERE id = 1`).Scan(&ageDays, &minInteractions)
	if err != nil {
		log.Printf("prune: failed to read config: %v, using defaults", err)
		ageDays = 180
		minInteractions = 3
	}

	cutoff := time.Now().AddDate(0, 0, -ageDays)

	// Archive posts that:
	// 1. Are older than the cutoff
	// 2. Have fewer than minInteractions interactions
	// 3. Are not already archived, deleted, or scheduled
	result, err := pg.Exec(ctx,
		`UPDATE posts SET archived_at = NOW()
		 WHERE archived_at IS NULL
		 AND is_deleted = FALSE
		 AND status = 0
		 AND created_at < $1
		 AND interaction_count < $2`,
		cutoff, minInteractions,
	)
	if err != nil {
		log.Printf("prune: error archiving posts: %v", err)
		return
	}

	rowsAffected := result.RowsAffected()
	log.Printf("prune: archived %d posts (age > %dd, interactions < %d)", rowsAffected, ageDays, minInteractions)

	// Also prune the feed_items table for archived posts to save space
	pruneResult, err := pg.Exec(ctx,
		`DELETE FROM feed_items
		 WHERE post_id IN (SELECT id FROM posts WHERE archived_at IS NOT NULL AND archived_at < NOW() - INTERVAL '7 days')`)
	if err == nil {
		log.Printf("prune: cleaned %d feed_items", pruneResult.RowsAffected())
	}
}
