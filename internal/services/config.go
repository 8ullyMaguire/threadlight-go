package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
)

type ConfigService struct {
	pg *pgxpool.Pool
}

func NewConfigService(pg *pgxpool.Pool) *ConfigService {
	return &ConfigService{pg: pg}
}

func (s *ConfigService) Get(ctx context.Context) (*model.SiteConfig, error) {
	cfg := &model.SiteConfig{}
	err := s.pg.QueryRow(ctx, `SELECT id, registration_mode, invite_limit_threshold_0,
		invite_limit_threshold_1, invite_limit_threshold_2, invite_limit_threshold_3,
		invite_limit_threshold_4, invite_limit_threshold_5, updated_at
		FROM site_config WHERE id = 1`).Scan(&cfg.ID, &cfg.RegistrationMode,
		&cfg.InviteLimitThreshold0, &cfg.InviteLimitThreshold1, &cfg.InviteLimitThreshold2,
		&cfg.InviteLimitThreshold3, &cfg.InviteLimitThreshold4, &cfg.InviteLimitThreshold5, &cfg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *ConfigService) Update(ctx context.Context, req model.UpdateConfigRequest) (*model.SiteConfig, error) {
	query := `UPDATE site_config SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.RegistrationMode != "" {
		query += `, registration_mode = $` + strconv.Itoa(argIdx)
		args = append(args, req.RegistrationMode)
		argIdx++
	}
	if req.InviteLimitThreshold0 != nil {
		query += `, invite_limit_threshold_0 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold0)
		argIdx++
	}
	if req.InviteLimitThreshold1 != nil {
		query += `, invite_limit_threshold_1 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold1)
		argIdx++
	}
	if req.InviteLimitThreshold2 != nil {
		query += `, invite_limit_threshold_2 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold2)
		argIdx++
	}
	if req.InviteLimitThreshold3 != nil {
		query += `, invite_limit_threshold_3 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold3)
		argIdx++
	}
	if req.InviteLimitThreshold4 != nil {
		query += `, invite_limit_threshold_4 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold4)
		argIdx++
	}
	if req.InviteLimitThreshold5 != nil {
		query += `, invite_limit_threshold_5 = $` + strconv.Itoa(argIdx)
		args = append(args, *req.InviteLimitThreshold5)
		argIdx++
	}
	query += ` WHERE id = 1`

	_, err := s.pg.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx)
}

func (s *ConfigService) GetInviteLimit(ctx context.Context, userID int64) (int, error) {
	var trustScore float64
	err := s.pg.QueryRow(ctx, `SELECT trust_score FROM users WHERE id = $1`, userID).Scan(&trustScore)
	if err != nil {
		return 0, err
	}

	cfg, err := s.Get(ctx)
	if err != nil {
		return 0, err
	}

	var threshold int
	switch {
	case trustScore >= 5.0:
		threshold = cfg.InviteLimitThreshold5
	case trustScore >= 2.0:
		threshold = cfg.InviteLimitThreshold4
	case trustScore >= 0.5:
		threshold = cfg.InviteLimitThreshold3
	case trustScore >= 0.3:
		threshold = cfg.InviteLimitThreshold2
	case trustScore >= 0.1:
		threshold = cfg.InviteLimitThreshold1
	default:
		threshold = cfg.InviteLimitThreshold0
	}

	if threshold < 0 {
		return -1, nil
	}
	return threshold, nil
}

func (s *ConfigService) GenerateInvite(ctx context.Context, userID int64) (*model.UserInvite, error) {
	limit, err := s.GetInviteLimit(ctx, userID)
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		return nil, model.ErrForbidden
	}

	if limit > 0 {
		count, err := s.InviteCountThisMonth(ctx, userID)
		if err != nil {
			return nil, err
		}
		if count >= limit {
			return nil, model.ErrForbidden
		}
	}

	code := genInviteCode()
	invite := &model.UserInvite{}
	err = s.pg.QueryRow(ctx,
		`INSERT INTO user_invites (inviter_id, code) VALUES ($1, $2)
		 RETURNING id, inviter_id, code, used_by, used_at, created_at`,
		userID, code,
	).Scan(&invite.ID, &invite.InviterID, &invite.Code, &invite.UsedBy, &invite.UsedAt, &invite.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invite, nil
}

func (s *ConfigService) InviteCountThisMonth(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.pg.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_invites
		 WHERE inviter_id = $1
		 AND created_at >= date_trunc('month', NOW())`,
		userID,
	).Scan(&count)
	return count, err
}

func (s *ConfigService) ListMyInvites(ctx context.Context, userID int64) ([]*model.UserInvite, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, inviter_id, code, used_by, used_at, created_at
		 FROM user_invites WHERE inviter_id = $1
		 ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*model.UserInvite
	for rows.Next() {
		invite := &model.UserInvite{}
		if err := rows.Scan(&invite.ID, &invite.InviterID, &invite.Code, &invite.UsedBy, &invite.UsedAt, &invite.CreatedAt); err != nil {
			return nil, err
		}
		if invite.UsedBy != nil && *invite.UsedBy == 0 {
			invite.UsedBy = nil
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

func genInviteCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
