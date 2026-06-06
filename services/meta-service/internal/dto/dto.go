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

// ─── Statuses ─────────────────────────────────────────────────────────────────

type ListStatusesRequest struct {
	Type     *string
	Page     int
	PageSize int
}

type ListStatusesResponse struct {
	Statuses []*StatusResponse
	Total    int64
	Page     int
	PageSize int
}

type CreateStatusRequest struct {
	Type        string
	Name        string
	Slug        string
	CreatedByID uuid.UUID
}

type UpdateStatusRequest struct {
	Type        *string
	Name        *string
	Slug        *string
	UpdatedByID uuid.UUID
}

type DeleteStatusRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type StatusResponse struct {
	ID        uuid.UUID
	Type      string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy UserSimpleResponse
	UpdatedBy UserSimpleResponse
	DeletedBy *UserSimpleResponse
}

// ─── Tag ──────────────────────────────────────────────────────────────────────

type ListTagsRequest struct {
	Type     *string
	Page     int
	PageSize int
}

type ListTagsResponse struct {
	Tags     []*TagResponse
	Total    int64
	Page     int
	PageSize int
}

type CreateTagRequest struct {
	Type        string
	Name        string
	Slug        string
	CreatedByID uuid.UUID
}

type UpdateTagRequest struct {
	Type        *string
	Name        *string
	Slug        *string
	UpdatedByID uuid.UUID
}

type DeleteTagRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type TagResponse struct {
	ID        uuid.UUID
	Type      string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy UserSimpleResponse
	UpdatedBy UserSimpleResponse
	DeletedBy *UserSimpleResponse
}

// ─── Administrative Divisions ───────────────────────────────────────────────

type ListAdministrativeDivisionsRequest struct {
	Name       *string
	Level      *string
	PostalCode *string
	CountryID  *uuid.UUID
	ParentID   *uuid.UUID
	Page       int
	PageSize   int
}

type ListAdministrativeDivisionsResponse struct {
	AdministrativeDivisions []*AdministrativeDivisionResponse
	Total                   int64
	Page                    int
	PageSize                int
}

type CreateAdministrativeDivisionRequest struct {
	Name        string
	Level       string
	PostalCode  string
	CountryID   uuid.UUID
	ParentID    *uuid.UUID
	CreatedByID uuid.UUID
}

type UpdateAdministrativeDivisionRequest struct {
	Name        *string
	Level       *string
	PostalCode  *string
	CountryID   *uuid.UUID
	ParentID    *uuid.UUID
	UpdatedByID uuid.UUID
}

type DeleteAdministrativeDivisionRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type AdministrativeDivisionSimpleResponse struct {
	ID         uuid.UUID
	Name       string
	Level      string
	PostalCode string
	Country    CountrySimpleResponse
	ParentID   *uuid.UUID
}

type AdministrativeDivisionResponse struct {
	ID         uuid.UUID
	CountryID  uuid.UUID
	ParentID   *uuid.UUID
	Name       string
	Level      string
	PostalCode string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	CreatedBy  UserSimpleResponse
	UpdatedBy  UserSimpleResponse
	DeletedBy  *UserSimpleResponse
	Country    CountrySimpleResponse
	Parent     *AdministrativeDivisionSimpleResponse
}

// ─── Countries ─────────────────────────────────────────────────────────────

type ListCountriesRequest struct {
	Name      *string
	ISOAlpha2 *string
	ISOAlpha3 *string
	PhoneCode *string
	Currency  *string
	Page      int
	PageSize  int
}

type ListCountriesResponse struct {
	Countries []*CountryResponse
	Total     int64
	Page      int
	PageSize  int
}

type CreateCountryRequest struct {
	Name         string
	ISOAlpha2    string
	ISOAlpha3    string
	PhoneCode    string
	CurrencyCode string
	CreatedByID  uuid.UUID
}

type UpdateCountryRequest struct {
	Name         *string
	ISOAlpha2    *string
	ISOAlpha3    *string
	PhoneCode    *string
	CurrencyCode *string
	UpdatedByID  uuid.UUID
}

type DeleteCountryRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type CountrySimpleResponse struct {
	ID           uuid.UUID
	Name         string
	ISOAlpha2    string
	ISOAlpha3    string
	PhoneCode    string
	CurrencyCode string
}

type CountryResponse struct {
	ID           uuid.UUID
	Name         string
	ISOAlpha2    string
	ISOAlpha3    string
	PhoneCode    string
	CurrencyCode string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	CreatedBy    UserSimpleResponse
	UpdatedBy    UserSimpleResponse
	DeletedBy    *UserSimpleResponse
}
