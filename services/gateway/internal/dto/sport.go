package dto

// ─── SPORT DTOS ──────────────────────────────────────────────────────────────

type SportStatResponse struct {
	ID                string            `json:"id"`
	SportID           string            `json:"sport_id"`
	StatTypeTagID     string            `json:"stat_type_tag_id"`
	StatTypeTag       TagSimpleResponse `json:"stat_type_tag"`
	AggregationMethod string            `json:"aggregation_method"`
}

type SportConfigResponse struct {
	ID                 string              `json:"id"`
	SportID            string              `json:"sport_id"`
	Stats              []SportStatResponse `json:"stats"`
	ParticipantTypeTag TagSimpleResponse   `json:"participant_type_tag"`
	MinRosterSize      int32               `json:"min_roster_size"`
	MaxRosterSize      int32               `json:"max_roster_size"`
	TypicalRosterSize  int32               `json:"typical_roster_size"`
	RulesURL           string              `json:"rules_url"`
	Description        string              `json:"description"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
}

type SportResponse struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	Slug              string               `json:"slug"`
	Description       string               `json:"description"`
	IconAttachmentID  string               `json:"icon_attachment_id"`
	Status            StatusSimpleResponse `json:"status"`
	TierTag           TagSimpleResponse    `json:"tier_tag"`
	ActiveRegulator   *RegulatorResponse   `json:"active_regulator,omitempty"`
	Config            *SportConfigResponse `json:"config,omitempty"`
	TotalCompetitions int32                `json:"total_competitions"`
	TotalAthletes     int32                `json:"total_athletes"`
	TotalAcademies    int32                `json:"total_academies"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
}

type CreateSportRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=100"`
	Slug             string `json:"slug" validate:"required,min=2,max=100"`
	Description      string `json:"description" validate:"required,min=10"`
	IconAttachmentID string `json:"icon_attachment_id"`
	TierTagID        string `json:"tier_tag_id" validate:"required,uuid"`
	StatusID         string `json:"status_id" validate:"required,uuid"`
}

type UpdateSportRequest struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	IconAttachmentID *string `json:"icon_attachment_id"`
	TierTagID        *string `json:"tier_tag_id"`
	StatusID         *string `json:"status_id"`
}

type SportStatConfigInput struct {
	StatTypeTagID     string `json:"stat_type_tag_id" validate:"required,uuid"`
	AggregationMethod string `json:"aggregation_method" validate:"required,min=2"`
}

type UpdateSportConfigRequest struct {
	Stats                []SportStatConfigInput `json:"stats" validate:"required,dive"`
	ParticipantTypeTagID string                 `json:"participant_type_tag_id" validate:"required,uuid"`
	MinRosterSize        int32                  `json:"min_roster_size" validate:"required,gt=0"`
	MaxRosterSize        int32                  `json:"max_roster_size" validate:"required,gt=0"`
	TypicalRosterSize    int32                  `json:"typical_roster_size" validate:"required,gt=0"`
	RulesURL             *string                `json:"rules_url"`
	Description          *string                `json:"description"`
}

// ─── REGULATOR DTOS ──────────────────────────────────────────────────────────

type RegulatorStaffResponse struct {
	ID          string             `json:"id"`
	RegulatorID string             `json:"regulator_id"`
	UserID      string             `json:"user_id"`
	User        UserSimpleResponse `json:"user"`
	RoleTag     TagSimpleResponse  `json:"role_tag"`
	JoinedAt    string             `json:"joined_at"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type RegulatorResponse struct {
	ID               string                   `json:"id"`
	OrganizationName string                   `json:"organization_name"`
	Code             string                   `json:"code"`
	LogoAttachmentID string                   `json:"logo_attachment_id"`
	ContactEmail     string                   `json:"contact_email"`
	PhoneNumber      string                   `json:"phone_number"`
	WebsiteURL       string                   `json:"website_url"`
	Status           StatusSimpleResponse     `json:"status"`
	Staff            []RegulatorStaffResponse `json:"staff"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

type CreateRegulatorRequest struct {
	OrganizationName         string   `json:"organization_name" validate:"required,min=3"`
	Code                     string   `json:"code" validate:"required,min=2"`
	ContactEmail             string   `json:"contact_email" validate:"required,email"`
	PhoneNumber              *string  `json:"phone_number"`
	WebsiteURL               *string  `json:"website_url"`
	LogoAttachmentID         string   `json:"logo_attachment_id"`
	StreetAddress            string   `json:"street_address" validate:"required,min=3"`
	AddressNotes             *string  `json:"address_notes"`
	Latitude                 *float64 `json:"latitude"`
	Longitude                *float64 `json:"longitude"`
	AdministrativeDivisionID string   `json:"administrative_division_id" validate:"required,uuid"`
	StatusID                 string   `json:"status_id" validate:"required,uuid"`
}

type AddRegulatorStaffRequest struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	RoleTagID string `json:"role_tag_id" validate:"required,uuid"`
}

type AssignSportToRegulatorRequest struct {
	SportID                     string `json:"sport_id" validate:"required,uuid"`
	RegulatorID                 string `json:"regulator_id" validate:"required,uuid"`
	RequiresApprovalForOfficial bool   `json:"requires_approval_for_official"`
	RequiresApprovalForRegional bool   `json:"requires_approval_for_regional"`
}
