package dto

import (
	"time"

	"github.com/google/uuid"
)

type RegulatorResponse struct {
	ID               uuid.UUID
	OrganizationName string
	Code             string
	LogoAttachmentID *uuid.UUID
	ContactEmail     string
	PhoneNumber      *string
	WebsiteURL       *string
	Status           *StatusSimpleResponse
	Staff            []*RegulatorStaffResponse
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedByID      uuid.UUID
	UpdatedByID      uuid.UUID
	DeletedByID      *uuid.UUID
}

type RegulatorStaffResponse struct {
	ID          uuid.UUID
	RegulatorID uuid.UUID
	UserID      uuid.UUID
	User        *UserSimpleResponse
	RoleTag     *TagSimpleResponse
	JoinedAt    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedByID uuid.UUID
	UpdatedByID uuid.UUID
	DeletedByID *uuid.UUID
}

type CreateRegulatorRequest struct {
	OrganizationName string
	Code             string
	LogoAttachmentID *uuid.UUID
	ContactEmail     string
	PhoneNumber      *string
	WebsiteURL       *string
	StatusID         uuid.UUID
	StreetAddress    string
	Notes            *string
	Latitude         *float64
	Longitude        *float64
	AdminDivisionID  uuid.UUID
	CreatedByID      uuid.UUID
}

type AddStaffRequest struct {
	RegulatorID uuid.UUID
	UserID      uuid.UUID
	RoleTagID   uuid.UUID
	CreatedByID uuid.UUID
}

type RemoveStaffRequest struct {
	StaffID     uuid.UUID
	DeletedByID uuid.UUID
}

type AssignSportToRegulatorRequest struct {
	SportID                     uuid.UUID
	RegulatorID                 uuid.UUID
	RequiresApprovalForOfficial bool
	RequiresApprovalForRegional bool
	UpdatedByID                 uuid.UUID
}
