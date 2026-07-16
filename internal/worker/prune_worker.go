package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunPruneWorker(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("prune_worker: pruning stale data")
}
