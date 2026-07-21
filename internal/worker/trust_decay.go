package worker

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var trustLevelThresholds = []float64{0, 3, 6, 10, 15, 25}

func RunTrustDecay(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("trust_decay: running trust decay and auto-promotion...")

	rows, err := pg.Query(ctx, `
		SELECT id, trust_level, trust_score, last_active_at
		FROM users
		WHERE is_active = true AND deleted_at IS NULL
	`)
	if err != nil {
		log.Printf("trust_decay: query failed: %v", err)
		return
	}
	defer rows.Close()

	type userTrust struct {
		id         int64
		level      int16
		score      float64
		lastActive *time.Time
	}

	var users []userTrust
	for rows.Next() {
		var u userTrust
		if err := rows.Scan(&u.id, &u.level, &u.score, &u.lastActive); err != nil {
			log.Printf("trust_decay: scan failed: %v", err)
			continue
		}
		users = append(users, u)
	}

	for _, u := range users {
		decayed := u.score

		if u.lastActive != nil && time.Since(*u.lastActive) > 30*24*time.Hour {
			weeksInactive := int(time.Since(*u.lastActive).Hours() / (24 * 7))
			if weeksInactive > 0 {
				loss := float64(weeksInactive) * 0.5
				decayed -= loss
				if decayed < 0 {
					decayed = 0
				}
			}
		}

		newLevel := u.level
		for level := int(u.level) + 1; level < len(trustLevelThresholds); level++ {
			if decayed >= trustLevelThresholds[level] {
				newLevel = int16(level)
			} else {
				break
			}
		}

		if decayed != u.score || newLevel != u.level {
			_, err := pg.Exec(ctx,
				`UPDATE users SET trust_score = $1, trust_level = $2 WHERE id = $3`,
				decayed, newLevel, u.id,
			)
			if err != nil {
				log.Printf("trust_decay: update user %d failed: %v", u.id, err)
			} else if newLevel != u.level {
				log.Printf("trust_decay: user %d promoted to level %d (score: %.1f)", u.id, newLevel, decayed)
			}
		}
	}

	log.Printf("trust_decay: processed %d users", len(users))
}
