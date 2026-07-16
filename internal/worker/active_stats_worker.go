package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func RunActiveStatsUpdater(ctx context.Context, pg *pgxpool.Pool, rdb *redis.Client) {
	log.Println("active_stats_worker: updating active user stats")
}
