package worker

import (
	"context"
	"log"
	"time"
)

func (w *Worker) distributeDailyRewards(ctx context.Context) {
	log.Println("worker: distributing daily rewards")
	start := time.Now()

	today := time.Now().UTC().Format("2006-01-02")

	_, err := w.pg.Exec(ctx, `
		WITH eligible AS (
			SELECT u.id AS user_id, u.credits
			FROM users u
			WHERE u.is_active = true
			AND NOT EXISTS (
				SELECT 1 FROM daily_rewards dr
				WHERE dr.user_id = u.id AND dr.date = $1
			)
		)
		INSERT INTO daily_rewards (user_id, date, amount, claimed)
		SELECT user_id, $1,
			CASE
				WHEN credits < 1000 THEN 50
				WHEN credits < 5000 THEN 30
				ELSE 15
			END,
			false
		FROM eligible
	`, today)
	if err != nil {
		log.Printf("worker: daily reward distribution error: %v", err)
		return
	}

	log.Printf("worker: daily rewards distributed in %v", time.Since(start))
}

func (w *Worker) expireBounties(ctx context.Context) {
	log.Println("worker: expiring stale bounties")
	start := time.Now()

	_, err := w.pg.Exec(ctx, `
		UPDATE bounties SET status = 2
		WHERE status = 1 AND expires_at IS NOT NULL AND expires_at < NOW()
	`)
	if err != nil {
		log.Printf("worker: bounty expiry error: %v", err)
		return
	}

	log.Printf("worker: bounties expired in %v", time.Since(start))
}

func (w *Worker) archiveOldContent(ctx context.Context) {
	log.Println("worker: archiving old content")
	start := time.Now()

	_, err := w.pg.Exec(ctx, `
		UPDATE posts SET status = 2, archived_at = NOW()
		WHERE status = 1
		AND interaction_count = 0
		AND created_at < NOW() - INTERVAL '90 days'
	`)
	if err != nil {
		log.Printf("worker: archival error: %v", err)
		return
	}

	log.Printf("worker: old content archived in %v", time.Since(start))
}
