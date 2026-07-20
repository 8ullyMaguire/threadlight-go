package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunFeedRebuild(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("feed_worker: rebuilding feeds")
}
