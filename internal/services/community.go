package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

type CommunityService struct {
	pg *pgxpool.Pool
}

func NewCommunityService(pg *pgxpool.Pool) *CommunityService {
	return &CommunityService{pg: pg}
}

func (s *CommunityService) CreateCommunity(ctx context.Context, req model.Community) (*model.Community, error) {
	comm := &model.Community{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO communities (name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by, created_at, updated_at`,
		req.Name, req.Description, req.Slug, req.Tags, req.CuratorLock, req.SlowBootDays, req.ForkedFrom, req.CreatedBy,
	).Scan(&comm.ID, &comm.Name, &comm.Description, &comm.Slug, &comm.Tags, &comm.CuratorLock, &comm.SlowBootDays,
		&comm.ForkedFrom, &comm.CreatedBy, &comm.CreatedAt, &comm.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return comm, nil
}

func (s *CommunityService) GetCommunity(ctx context.Context, communityID int64) (*model.Community, error) {
	comm := &model.Community{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by, created_at, updated_at, archived_at
		 FROM communities WHERE id = $1 AND archived_at IS NULL`, communityID,
	).Scan(&comm.ID, &comm.Name, &comm.Description, &comm.Slug, &comm.Tags, &comm.CuratorLock, &comm.SlowBootDays,
		&comm.ForkedFrom, &comm.CreatedBy, &comm.CreatedAt, &comm.UpdatedAt, &comm.ArchivedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return comm, nil
}

func (s *CommunityService) GetCommunityBySlug(ctx context.Context, slug string) (*model.Community, error) {
	comm := &model.Community{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by, created_at, updated_at, archived_at
		 FROM communities WHERE slug = $1 AND archived_at IS NULL`, slug,
	).Scan(&comm.ID, &comm.Name, &comm.Description, &comm.Slug, &comm.Tags, &comm.CuratorLock, &comm.SlowBootDays,
		&comm.ForkedFrom, &comm.CreatedBy, &comm.CreatedAt, &comm.UpdatedAt, &comm.ArchivedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return comm, nil
}

func (s *CommunityService) UpdateCommunity(ctx context.Context, communityID int64, req model.Community) (*model.Community, error) {
	comm := &model.Community{}
	err := s.pg.QueryRow(ctx,
		`UPDATE communities SET name = COALESCE(NULLIF($2, ''), name),
		                        description = COALESCE(NULLIF($3, ''), description),
		                        slug = COALESCE(NULLIF($4, ''), slug),
		                        tags = $5,
		                        curator_lock = $6,
		                        slow_boot_days = COALESCE(NULLIF($7, 0), slow_boot_days),
		                        updated_at = NOW()
		 WHERE id = $1 AND archived_at IS NULL
		 RETURNING id, name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by, created_at, updated_at, archived_at`,
		communityID, req.Name, req.Description, req.Slug, req.Tags, req.CuratorLock, req.SlowBootDays,
	).Scan(&comm.ID, &comm.Name, &comm.Description, &comm.Slug, &comm.Tags, &comm.CuratorLock, &comm.SlowBootDays,
		&comm.ForkedFrom, &comm.CreatedBy, &comm.CreatedAt, &comm.UpdatedAt, &comm.ArchivedAt)
	if err != nil {
		return nil, model.ErrNotFound
	}
	return comm, nil
}

func (s *CommunityService) ArchiveCommunity(ctx context.Context, communityID int64) error {
	_, err := s.pg.Exec(ctx,
		`UPDATE communities SET archived_at = NOW() WHERE id = $1 AND archived_at IS NULL`,
		communityID)
	return err
}

func (s *CommunityService) ListCommunities(ctx context.Context) ([]model.Community, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, name, description, slug, tags, curator_lock, slow_boot_days, forked_from, created_by, created_at, updated_at
		 FROM communities WHERE archived_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []model.Community
	for rows.Next() {
		var c model.Community
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Slug, &c.Tags, &c.CuratorLock, &c.SlowBootDays,
			&c.ForkedFrom, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		communities = append(communities, c)
	}
	return communities, nil
}

func (s *CommunityService) JoinCommunity(ctx context.Context, communityID, userID int64, role, status int16) error {
	_, err := s.pg.Exec(ctx,
		`INSERT INTO community_members (community_id, user_id, role, status, joined_at)
		 VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (community_id, user_id) DO UPDATE SET status = $4`,
		communityID, userID, role, status)
	return err
}

func (s *CommunityService) LeaveCommunity(ctx context.Context, communityID, userID int64) error {
	_, err := s.pg.Exec(ctx,
		`DELETE FROM community_members WHERE community_id = $1 AND user_id = $2`,
		communityID, userID)
	return err
}

func (s *CommunityService) UpdateMemberRole(ctx context.Context, communityID, userID int64, role int16) error {
	_, err := s.pg.Exec(ctx,
		`UPDATE community_members SET role = $3 WHERE community_id = $1 AND user_id = $2`,
		communityID, userID, role)
	return err
}

func (s *CommunityService) GetMembers(ctx context.Context, communityID int64) ([]model.CommunityMember, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT community_id, user_id, role, status, joined_at
		 FROM community_members WHERE community_id = $1`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.CommunityMember
	for rows.Next() {
		var m model.CommunityMember
		if err := rows.Scan(&m.CommunityID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (s *CommunityService) AddCurator(ctx context.Context, communityID, userID int64, permission int16) (*model.Curator, error) {
	curator := &model.Curator{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO curators (community_id, user_id, permission)
		 VALUES ($1, $2, $3)
		 RETURNING id, community_id, user_id, permission, created_at`,
		communityID, userID, permission,
	).Scan(&curator.ID, &curator.CommunityID, &curator.UserID, &curator.Permission, &curator.CreatedAt)
	if err != nil {
		return nil, err
	}
	return curator, nil
}

func (s *CommunityService) RemoveCurator(ctx context.Context, communityID, userID int64) error {
	_, err := s.pg.Exec(ctx,
		`DELETE FROM curators WHERE community_id = $1 AND user_id = $2`,
		communityID, userID)
	return err
}

func (s *CommunityService) ForkCommunity(ctx context.Context, sourceID, initiatedBy int64, reason string) (*model.CommunityFork, error) {
	fork := &model.CommunityFork{}
	err := s.pg.QueryRow(ctx,
		`INSERT INTO community_forks (source_id, fork_id, initiated_by, reason, member_count, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 RETURNING id, source_id, fork_id, initiated_by, reason, member_count, created_at`,
		sourceID, nil, initiatedBy, reason, 0,
	).Scan(&fork.ID, &fork.SourceID, &fork.ForkID, &fork.InitiatedBy, &fork.Reason, &fork.MemberCount, &fork.CreatedAt)
	if err != nil {
		return nil, err
	}
	return fork, nil
}
