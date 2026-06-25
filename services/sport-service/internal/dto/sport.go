package dto

import (
	"time"

	"github.com/google/uuid"
)

type SportResponse struct {
	ID                          uuid.UUID
	Name                        string
	Slug                        string
	Description                 string
	IconAttachmentID            *uuid.UUID
	Status                      *StatusSimpleResponse
	TierTag                     *TagSimpleResponse
	ActiveRegulator             *RegulatorSimpleResponse
	Config                      *SportConfigResponse
	RequiresApprovalForOfficial bool
	RequiresApprovalForRegional bool
	TotalCompetitions           int32
	TotalAthletes               int32
	TotalAcademies              int32
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	CreatedByID                 uuid.UUID
	UpdatedByID                 uuid.UUID
	DeletedByID                 *uuid.UUID
}

type SportConfigResponse struct {
	ID                 uuid.UUID
	SportID            uuid.UUID
	Stats              []SportStatResponse
	ParticipantTypeTag *TagSimpleResponse
	MinRosterSize      int32
	MaxRosterSize      int32
	TypicalRosterSize  int32
	RulesURL           *string
	Description        string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SportStatResponse struct {
	ID                uuid.UUID
	SportID           uuid.UUID
	StatTypeTagID     uuid.UUID
	StatTypeTag       *TagSimpleResponse
	AggregationMethod string
}

type CreateSportRequest struct {
	Name             string
	Slug             string
	Description      string
	IconAttachmentID *uuid.UUID
	TierTagID        uuid.UUID
	StatusID         uuid.UUID
	CreatedByID      uuid.UUID
}

type ListSportsRequest struct {
	TierTagID *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
	Page      int
	PageSize  int
}

type ListSportsResponse struct {
	Sports   []*SportResponse
	Total    int64
	Page     int
	PageSize int
}

type UpdateSportRequest struct {
	ID               uuid.UUID
	Name             *string
	Description      *string
	IconAttachmentID *uuid.UUID
	TierTagID        *uuid.UUID
	StatusID         *uuid.UUID
	UpdatedByID      uuid.UUID
}

type SportStatConfigInput struct {
	StatTypeTagID     uuid.UUID
	AggregationMethod string
}

type UpdateSportConfigRequest struct {
	SportID              uuid.UUID
	Stats                []SportStatConfigInput
	ParticipantTypeTagID uuid.UUID
	MinRosterSize        int32
	MaxRosterSize        int32
	TypicalRosterSize    int32
	RulesURL             *string
	Description          string
	UpdatedByID          uuid.UUID
}

type StatusSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type TagSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type UserSimpleResponse struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
}

type RegulatorSimpleResponse struct {
	ID               uuid.UUID
	OrganizationName string
	Code             string
	LogoAttachmentID *uuid.UUID
	ContactEmail     string
	PhoneNumber      *string
	WebsiteURL       *string
	Status           *StatusSimpleResponse
}
