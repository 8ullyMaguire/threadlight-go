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
	InstanceName          string    `json:"instance_name"`
	InstanceShortDesc     string    `json:"instance_short_description"`
	InstanceDesc          string    `json:"instance_description"`
	AdminContactEmail     string    `json:"admin_contact_email"`
	Version               string    `json:"version"`
	PrivacyPolicyURL      string    `json:"privacy_policy_url"`
	TermsURL              string    `json:"terms_url"`
	CodeOfConductURL      string    `json:"code_of_conduct_url"`
	DefederationPolicyURL string    `json:"defederation_policy_url"`
	DonationURL           string    `json:"donation_url"`
	DonateText            string    `json:"donate_text"`
	PruneAgeDays          int       `json:"prune_age_days"`
	PruneMinInteractions  int       `json:"prune_min_interactions"`
	AdminUsers            []string  `json:"admin_users,omitempty"`
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
	InstanceName          string `json:"instance_name,omitempty"`
	InstanceShortDesc     string `json:"instance_short_description,omitempty"`
	InstanceDesc          string `json:"instance_description,omitempty"`
	AdminContactEmail     string `json:"admin_contact_email,omitempty"`
	PrivacyPolicyURL      string `json:"privacy_policy_url,omitempty"`
	TermsURL              string `json:"terms_url,omitempty"`
	CodeOfConductURL      string `json:"code_of_conduct_url,omitempty"`
	DefederationPolicyURL string `json:"defederation_policy_url,omitempty"`
	DonationURL           string `json:"donation_url,omitempty"`
	DonateText            string `json:"donate_text,omitempty"`
	PruneAgeDays          *int   `json:"prune_age_days,omitempty"`
	PruneMinInteractions  *int   `json:"prune_min_interactions,omitempty"`
}

type AboutResponse struct {
	InstanceName          string   `json:"instance_name"`
	ShortDescription      string   `json:"short_description"`
	Description           string   `json:"description"`
	Version               string   `json:"version"`
	AdminContactEmail     string   `json:"admin_contact_email"`
	PrivacyPolicyURL      string   `json:"privacy_policy_url"`
	TermsURL              string   `json:"terms_url"`
	CodeOfConductURL      string   `json:"code_of_conduct_url"`
	DefederationPolicyURL string   `json:"defederation_policy_url"`
	DonationURL           string   `json:"donation_url"`
	DonateText            string   `json:"donate_text"`
	AdminUsers            []string `json:"admin_users"`
}
