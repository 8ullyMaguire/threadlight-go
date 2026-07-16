package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunCommunityNotesProcessor(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("community_notes_worker: processing community notes")
}
