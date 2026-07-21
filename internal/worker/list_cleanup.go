package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunListCleanup deletes user lists that have no active subscribers.
// A list with zero active subscriptions serves no purpose and wastes
// resources on refresh cycles and storage.
func RunListCleanup(ctx context.Context, pg *pgxpool.Pool) {
	log.Println("list-cleanup: checking for lists with no active subscribers")

	result, err := pg.Exec(ctx, `
		DELETE FROM user_lists ul
		WHERE NOT EXISTS (
		    SELECT 1 FROM list_subscriptions ls
		    WHERE ls.list_id = ul.id AND ls.active = TRUE
		)
	`)
	if err != nil {
		log.Printf("list-cleanup: query failed: %v", err)
		return
	}

	removed := result.RowsAffected()
	if removed > 0 {
		log.Printf("list-cleanup: removed %d lists with no active subscribers", removed)
	} else {
		log.Println("list-cleanup: no lists to remove")
	}
}
