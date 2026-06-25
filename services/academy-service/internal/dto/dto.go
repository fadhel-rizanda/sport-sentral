package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Common Simple Responses ──────────────────────────────────────────────────

type UserSimpleResponse struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
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

type PermissionSimpleResponse struct {
	ID       uuid.UUID
	Resource string
	Action   string
	Slug     string
}

type RoleSimpleResponse struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	Permissions    []PermissionSimpleResponse
	PermissionsIDs []string
}

type CountrySimpleResponse struct {
	ID           uuid.UUID
	Name         string
	ISOAlpha2    string
	ISOAlpha3    string
	PhoneCode    string
	CurrencyCode string
}

type AdministrativeDivisionSimpleResponse struct {
	ID         uuid.UUID
	CountryID  uuid.UUID
	ParentID   *uuid.UUID
	Name       string
	Level      string
	PostalCode string
	Country    CountrySimpleResponse
	Parent     *AdministrativeDivisionSimpleResponse
}

type SportSimpleResponse struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	IconAttachmentID *uuid.UUID
	IsVerified       bool
	RegulatorID      *uuid.UUID
	Tier             string
}

type AcademyHoldingAddressResponse struct {
	ID                     uuid.UUID
	StreetAddress          string
	Notes                  *string
	Latitude               *float64
	Longitude              *float64
	AdministrativeDivision AdministrativeDivisionSimpleResponse
}

type AcademyBranchAddressResponse struct {
	ID                     uuid.UUID
	BranchID               uuid.UUID
	IsPrimary              bool
	StreetAddress          string
	Notes                  *string
	Latitude               *float64
	Longitude              *float64
	AdministrativeDivision AdministrativeDivisionSimpleResponse
}

// ─── Academy Holding ──────────────────────────────────────────────────────────

type CreateAcademyHoldingRequest struct {
	Name              string
	Description       string
	Email             string
	PhoneNumber       string
	ImageAttachmentID *uuid.UUID
	StatusID          uuid.UUID

	StreetAddress   string
	Notes           *string
	Latitude        *float64
	Longitude       *float64
	AdminDivisionID uuid.UUID

	CreatedByID uuid.UUID
}

type UpdateAcademyHoldingRequest struct {
	ID                uuid.UUID
	Name              *string
	Description       *string
	Email             *string
	PhoneNumber       *string
	ImageAttachmentID *uuid.UUID
	UpdatedByID       uuid.UUID
}

type AcademyHoldingResponse struct {
	ID                string
	Name              string
	Description       string
	Email             string
	PhoneNumber       string
	ImageAttachmentID *string
	Branches          []AcademyBranchSimpleResponse
	Address           *AcademyHoldingAddressResponse
	Status            *StatusSimpleResponse
	StatusID          string
	CreatedAt         time.Time
	UpdatedAt         time.Time

	CreatedBy UserSimpleResponse
	UpdatedBy UserSimpleResponse
	DeletedBy *UserSimpleResponse
}

type AcademyHoldingRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type AcademyHoldingSimpleResponse struct {
	ID                string
	Name              string
	Description       string
	Email             string
	PhoneNumber       string
	ImageAttachmentID *string
	BranchesCount     int64
	StatusID          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ListAcademyHoldingsRequest struct {
	StatusID *uuid.UUID
	Search   *string
	Page     int
	PageSize int
}

type ListAcademyHoldingsResponse struct {
	Holdings []*AcademyHoldingResponse
	Total    int64
	Page     int
	PageSize int
}

type DeleteAcademyHoldingRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

// ─── Academy Branches ──────────────────────────────────────────────────────────

type CreateAcademyBranchRequest struct {
	HoldingID       uuid.UUID
	SportID         uuid.UUID
	Name            string
	Email           string
	PhoneNumber     string
	StatusID        uuid.UUID
	StreetAddress   string
	Notes           *string
	Latitude        *float64
	Longitude       *float64
	AdminDivisionID uuid.UUID
	CreatedByID     uuid.UUID
}

type UpdateAcademyBranchRequest struct {
	ID          uuid.UUID
	Name        *string
	Email       *string
	PhoneNumber *string
	StatusID    *uuid.UUID
	UpdatedByID uuid.UUID
}

type AcademyBranchResponse struct {
	ID          string
	HoldingID   string
	SportID     string
	SportName   string
	Name        string
	Email       string
	PhoneNumber string
	StatusID    string
	MemberCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Holding   *AcademyHoldingSimpleResponse
	Sport     *SportSimpleResponse
	Status    *StatusSimpleResponse
	Addresses []AcademyBranchAddressResponse
	CreatedBy UserSimpleResponse
	UpdatedBy UserSimpleResponse
	DeletedBy *UserSimpleResponse
}

type AcademyBranchSimpleResponse struct {
	ID          string
	HoldingID   string
	SportID     string
	SportName   string
	Name        string
	Email       string
	PhoneNumber string
	StatusID    string
	MemberCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListAcademyBranchesRequest struct {
	HoldingID *uuid.UUID
	SportID   *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
	Page      int
	PageSize  int
}

type ListAcademyBranchesResponse struct {
	Branches []*AcademyBranchResponse
	Total    int64
	Page     int
	PageSize int
}

type DeleteAcademyBranchRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

// ─── Academy Admin ────────────────────────────────────────────────────────────

type CreateAcademyAdminRequest struct {
	AcademyID   uuid.UUID
	BranchID    *uuid.UUID
	UserID      uuid.UUID
	RoleID      uuid.UUID
	CreatedByID uuid.UUID
}

type UpdateAcademyAdminRequest struct {
	RoleID      uuid.UUID
	UpdatedByID uuid.UUID
}

type AcademyAdminResponse struct {
	ID           uuid.UUID
	AcademyID    uuid.UUID
	BranchID     *uuid.UUID
	UserID       uuid.UUID
	RoleID       uuid.UUID
	ApprovedAt   *time.Time
	ApprovedByID *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time

	Academy    *AcademyHoldingSimpleResponse
	Branch     *AcademyBranchSimpleResponse
	User       UserSimpleResponse
	Role       RoleSimpleResponse
	ApprovedBy *UserSimpleResponse
	CreatedBy  UserSimpleResponse
	UpdatedBy  UserSimpleResponse
	DeletedBy  *UserSimpleResponse
}

type ListAcademyAdminsRequest struct {
	AcademyID *uuid.UUID
	BranchID  *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
	Page      int
	PageSize  int
}

type ListAcademyAdminsResponse struct {
	Admins   []*AcademyAdminResponse
	Total    int64
	Page     int
	PageSize int
}

type DeleteAcademyAdminRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type AssignAcademyAdminRequest struct {
	AcademyID    uuid.UUID
	BranchID     *uuid.UUID
	UserID       uuid.UUID
	RoleID       uuid.UUID
	AssignedByID uuid.UUID
}

type RevokeAcademyAdminRequest struct {
	ID          uuid.UUID
	RevokedByID uuid.UUID
}

// ─── Enrollment ───────────────────────────────────────────────────────────────

type CreateEnrollmentRequest struct {
	AcademyBranchID uuid.UUID
	AthleteID       uuid.UUID
	JoinedAt        time.Time
	ExpiresAt       *time.Time
	StatusID        uuid.UUID
	CreatedByID     uuid.UUID
}

type UpdateEnrollmentRequest struct {
	LeftAt      *time.Time
	ExpiresAt   *time.Time
	StatusID    *uuid.UUID
	UpdatedByID uuid.UUID
}

type EnrollmentResponse struct {
	ID              uuid.UUID
	AcademyBranchID uuid.UUID
	AthleteID       uuid.UUID
	JoinedAt        time.Time
	LeftAt          *time.Time
	ExpiresAt       *time.Time
	ApprovedAt      *time.Time
	ApprovedByID    *uuid.UUID
	StatusID        uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	AcademyBranch *AcademyBranchSimpleResponse
	Athlete       UserSimpleResponse
	ApprovedBy    *UserSimpleResponse
	Status        StatusSimpleResponse
	CreatedBy     UserSimpleResponse
	UpdatedBy     UserSimpleResponse
	DeletedBy     *UserSimpleResponse
}

type ListEnrollmentsRequest struct {
	BranchID  *uuid.UUID
	AthleteID *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
	Page      int
	PageSize  int
}

type ListEnrollmentsResponse struct {
	Enrollments []*EnrollmentResponse
	Total       int64
	Page        int
	PageSize    int
}

type DeleteEnrollmentRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}
