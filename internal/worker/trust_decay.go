package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunTrustDecay(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("trust_decay: applying trust decay")
}
