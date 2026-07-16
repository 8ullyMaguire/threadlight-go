package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

type SearchService struct {
	pg *pgxpool.Pool
}

func NewSearchService(pg *pgxpool.Pool) *SearchService {
	return &SearchService{pg: pg}
}

func (s *SearchService) Search(ctx context.Context, query, searchType string, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tsquery := fmt.Sprintf("%s:*", sanitizeTSQuery(query))

	rows, err := s.pg.Query(ctx,
		`SELECT p.id, p.author_id, p.title, p.body, p.content_type, p.mood,
		        p.is_educational, p.is_entertaining, p.is_nsfw, p.content_warning,
		        p.interaction_count, p.cumulative_interactions, p.status,
		        p.scheduled_at, p.created_at, p.updated_at, p.archived_at, p.is_deleted
		 FROM posts p
		 WHERE to_tsvector('english', p.title || ' ' || p.body) @@ to_tsquery('english', $1)
		   AND p.is_deleted = FALSE
		   AND p.status = 0
		 ORDER BY ts_rank(to_tsvector('english', p.title || ' ' || p.body), to_tsquery('english', $1)) DESC
		 LIMIT $2 OFFSET $3`, tsquery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.Post
	for rows.Next() {
		p := &model.Post{}
		err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Body, &p.ContentType, &p.Mood,
			&p.IsEducational, &p.IsEntertaining, &p.IsNsfw, &p.ContentWarning,
			&p.InteractionCount, &p.CumulativeInteractions, &p.Status,
			&p.ScheduledAt, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt, &p.IsDeleted)
		if err != nil {
			return nil, err
		}
		results = append(results, p)
	}

	if results == nil {
		return []*model.Post{}, nil
	}
	return results, nil
}

func sanitizeTSQuery(s string) string {
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		} else {
			result = append(result, ' ')
		}
	}
	return string(result)
}
