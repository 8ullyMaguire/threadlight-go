package services

import (
	"context"
	"fmt"
	"strings"

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

	baseQuery := `SELECT p.id, p.author_id, p.title, p.body, p.content_type, p.mood,
	        p.is_educational, p.is_entertaining, p.is_nsfw, p.content_warning,
	        p.interaction_count, p.cumulative_interactions, p.status,
	        p.scheduled_at, p.created_at, p.updated_at, p.archived_at, p.is_deleted
	 FROM posts p`

	var conditions []string
	var args []interface{}
	argIdx := 1

	tsquery, fields := parseAdvancedQuery(query)

	if tsquery != "" {
		conditions = append(conditions,
			fmt.Sprintf(`to_tsvector('english', p.title || ' ' || p.body) @@ to_tsquery('english', $%d)`, argIdx))
		args = append(args, tsquery)
		argIdx++
	}

	if author, ok := fields["author"]; ok {
		conditions = append(conditions,
			fmt.Sprintf(`p.author_id IN (SELECT id FROM users WHERE username ILIKE $%d OR display_name ILIKE $%d)`, argIdx, argIdx))
		args = append(args, "%"+author+"%")
		argIdx++
	}

	if community, ok := fields["community"]; ok {
		conditions = append(conditions,
			fmt.Sprintf(`p.id IN (SELECT post_id FROM community_posts cp JOIN communities c ON cp.community_id = c.id WHERE c.name ILIKE $%d OR c.slug ILIKE $%d)`, argIdx, argIdx))
		args = append(args, "%"+community+"%")
		argIdx++
	}

	if tag, ok := fields["tag"]; ok {
		conditions = append(conditions,
			fmt.Sprintf(`p.id IN (SELECT post_id FROM post_tags pt JOIN tags t ON pt.tag_id = t.id WHERE t.name ILIKE $%d)`, argIdx))
		args = append(args, "%"+tag+"%")
		argIdx++
	}

	if after, ok := fields["after"]; ok {
		conditions = append(conditions, fmt.Sprintf(`p.created_at >= $%d::timestamp`, argIdx))
		args = append(args, after)
		argIdx++
	}

	if before, ok := fields["before"]; ok {
		conditions = append(conditions, fmt.Sprintf(`p.created_at <= $%d::timestamp`, argIdx))
		args = append(args, before)
		argIdx++
	}

	conditions = append(conditions, "p.is_deleted = FALSE", "p.status = 0")

	querySQL := baseQuery + " WHERE " + strings.Join(conditions, " AND ")

	if tsquery != "" {
		querySQL += fmt.Sprintf(` ORDER BY ts_rank(to_tsvector('english', p.title || ' ' || p.body), to_tsquery('english', $%d)) DESC`, argIdx)
		args = append(args, tsquery)
		argIdx++
	} else {
		querySQL += " ORDER BY p.created_at DESC"
	}

	querySQL += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pg.Query(ctx, querySQL, args...)
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

func parseAdvancedQuery(raw string) (tsquery string, fields map[string]string) {
	fields = make(map[string]string)
	var tokens []string

	for _, part := range strings.Fields(raw) {
		if colonIdx := strings.Index(part, ":"); colonIdx > 0 {
			key := strings.ToLower(part[:colonIdx])
			val := part[colonIdx+1:]
			switch key {
			case "author", "community", "tag", "after", "before":
				fields[key] = val
				continue
			}
		}

		upper := strings.ToUpper(part)
		switch upper {
		case "AND":
			tokens = append(tokens, "&")
		case "OR":
			tokens = append(tokens, "|")
		case "NOT":
			tokens = append(tokens, "!")
		default:
			sanitized := sanitizeTSWord(part)
			if sanitized != "" {
				tokens = append(tokens, sanitized+":*")
			}
		}
	}

	if len(tokens) > 0 {
		tsquery = strings.Join(tokens, " ")
	}
	return
}

func sanitizeTSWord(s string) string {
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		}
	}
	return string(result)
}
