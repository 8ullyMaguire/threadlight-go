package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunAchievementCheck(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("achievement_worker: checking achievements")
}
