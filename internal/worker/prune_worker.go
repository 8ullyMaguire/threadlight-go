package worker

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunPrune(ctx context.Context, pg *pgxpool.Pool) {
	// Allow full disable via env var
	if os.Getenv("PRUNE_DISABLE") == "true" {
		return
	}

	log.Println("prune: starting daily content pruning")

	// Get prune config from site_config, overridable by env vars
	var ageDays, minInteractions int
	err := pg.QueryRow(ctx,
		`SELECT COALESCE(prune_age_days, 180), COALESCE(prune_min_interactions, 3)
		 FROM site_config WHERE id = 1`).Scan(&ageDays, &minInteractions)
	if err != nil {
		log.Printf("prune: failed to read config: %v, using defaults", err)
		ageDays = 180
		minInteractions = 3
	}

	// Env overrides
	if v := os.Getenv("PRUNE_AGE_DAYS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 {
			ageDays = d
		}
	}
	if v := os.Getenv("PRUNE_MIN_INTERACTIONS"); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m >= 0 {
			minInteractions = m
		}
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
