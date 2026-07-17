package model

import "time"

type SiteConfig struct {
	ID                    int       `json:"id"`
	RegistrationMode      string    `json:"registration_mode"`
	InviteLimitThreshold0 int       `json:"invite_limit_threshold_0"`
	InviteLimitThreshold1 int       `json:"invite_limit_threshold_1"`
	InviteLimitThreshold2 int       `json:"invite_limit_threshold_2"`
	InviteLimitThreshold3 int       `json:"invite_limit_threshold_3"`
	InviteLimitThreshold4 int       `json:"invite_limit_threshold_4"`
	InviteLimitThreshold5 int       `json:"invite_limit_threshold_5"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type UserInvite struct {
	ID        int64      `json:"id"`
	InviterID int64      `json:"inviter_id"`
	Code      string     `json:"code"`
	UsedBy    *int64     `json:"used_by,omitempty"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type UpdateConfigRequest struct {
	RegistrationMode      string `json:"registration_mode"`
	InviteLimitThreshold0 *int   `json:"invite_limit_threshold_0,omitempty"`
	InviteLimitThreshold1 *int   `json:"invite_limit_threshold_1,omitempty"`
	InviteLimitThreshold2 *int   `json:"invite_limit_threshold_2,omitempty"`
	InviteLimitThreshold3 *int   `json:"invite_limit_threshold_3,omitempty"`
	InviteLimitThreshold4 *int   `json:"invite_limit_threshold_4,omitempty"`
	InviteLimitThreshold5 *int   `json:"invite_limit_threshold_5,omitempty"`
}
