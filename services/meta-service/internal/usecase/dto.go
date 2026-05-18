package usecase

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
