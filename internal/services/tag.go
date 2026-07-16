package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

type TagService struct {
	pg *pgxpool.Pool
}

func NewTagService(pg *pgxpool.Pool) *TagService {
	return &TagService{pg: pg}
}

func (s *TagService) CreateTag(ctx context.Context, req model.Tag) (*model.Tag, error) {
	tag := &model.Tag{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO tags (name, description, category, is_wiki, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, description, category, is_wiki, created_by, created_at`,
		req.Name, req.Description, req.Category, req.IsWiki, req.CreatedBy,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.Category, &tag.IsWiki, &tag.CreatedBy, &tag.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) GetTag(ctx context.Context, tagID int64) (*model.Tag, error) {
	tag := &model.Tag{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, name, description, category, is_wiki, created_by, created_at
		 FROM tags WHERE id = $1`, tagID,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.Category, &tag.IsWiki, &tag.CreatedBy, &tag.CreatedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return tag, nil
}

func (s *TagService) ListTags(ctx context.Context) ([]model.Tag, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, name, description, category, is_wiki, created_by, created_at
		 FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Category, &t.IsWiki, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *TagService) UpdateTag(ctx context.Context, tagID int64, req model.Tag) (*model.Tag, error) {
	tag := &model.Tag{}
	err := s.pg.QueryRow(ctx,
		`UPDATE tags SET name = COALESCE(NULLIF($2, ''), name),
		                 description = COALESCE(NULLIF($3, ''), description),
		                 category = COALESCE(NULLIF($4, ''), category),
		                 is_wiki = $5
		 WHERE id = $1
		 RETURNING id, name, description, category, is_wiki, created_by, created_at`,
		tagID, req.Name, req.Description, req.Category, req.IsWiki,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.Category, &tag.IsWiki, &tag.CreatedBy, &tag.CreatedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return tag, nil
}

func (s *TagService) DeleteTag(ctx context.Context, tagID int64) error {
	_, err := s.pg.Exec(ctx, `DELETE FROM tags WHERE id = $1`, tagID)
	if err != nil {
		return model.ErrNotFound
	}
	return nil
}

func (s *TagService) TagPost(ctx context.Context, postID, tagID, taggedBy int64) error {
	_, err := s.pg.Exec(ctx,
		`INSERT INTO post_tags (post_id, tag_id, tagged_by, created_at)
		 VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING`,
		postID, tagID, taggedBy)
	return err
}

func (s *TagService) UntagPost(ctx context.Context, postID, tagID int64) error {
	_, err := s.pg.Exec(ctx,
		`DELETE FROM post_tags WHERE post_id = $1 AND tag_id = $2`,
		postID, tagID)
	if err != nil {
		return errors.New("tag not found on post")
	}
	return nil
}

func (s *TagService) GetPostTags(ctx context.Context, postID int64) ([]model.PostTag, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT post_id, tag_id, tagged_by, created_at
		 FROM post_tags WHERE post_id = $1`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.PostTag
	for rows.Next() {
		var pt model.PostTag
		if err := rows.Scan(&pt.PostID, &pt.TagID, &pt.TaggedBy, &pt.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, pt)
	}
	return tags, nil
}
