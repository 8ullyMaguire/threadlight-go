package worker

import (
	"context"
	"log"
	"time"
)

func (w *Worker) calculateTrending(ctx context.Context) {
	log.Println("worker: calculating trending topics")
	start := time.Now()

	_, err := w.pg.Exec(ctx, `
		INSERT INTO trending_topics (topic, frequency, velocity)
		SELECT
			LOWER(TRIM(unnest(regexp_matches(p.body, '#([A-Za-z0-9_]+)', 'g')))),
			COUNT(*)::int,
			COUNT(*)::float8 / 15.0
		FROM posts p
		WHERE p.created_at > NOW() - INTERVAL '6 hours'
		AND p.status = 1
		GROUP BY 1
		ORDER BY frequency DESC
		LIMIT 50
		ON CONFLICT (topic) DO UPDATE SET
			frequency = EXCLUDED.frequency,
			velocity = EXCLUDED.velocity,
			created_at = NOW()
	`)
	if err != nil {
		log.Printf("worker: trending calculation error: %v", err)
		return
	}

	_, err = w.pg.Exec(ctx, `
		DELETE FROM trending_topics WHERE created_at < NOW() - INTERVAL '24 hours'
	`)
	if err != nil {
		log.Printf("worker: trending cleanup error: %v", err)
		return
	}

	log.Printf("worker: trending calculated in %v", time.Since(start))
}
