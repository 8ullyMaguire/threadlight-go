package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunScheduledPostPublisher(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("scheduled_post_worker: publishing scheduled posts")
}
