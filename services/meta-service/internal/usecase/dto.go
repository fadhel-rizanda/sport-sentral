package usecase

import (
	"time"

	"github.com/google/uuid"
)

// ─── Statuses ───────────────────────────────────────────────────────────────────

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
	CreatedByID uuid.UUID
}

type UpdateStatusRequest struct {
	Type        *string
	Name        *string
	UpdatedByID uuid.UUID
}

type DeleteStatusRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

type StatusResponse struct {
	ID          uuid.UUID
	Type        string
	Name        string
	CreatedByID uuid.UUID
	UpdatedByID uuid.UUID
	DeletedByID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
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
	Slug        *string
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
	ID          uuid.UUID
	Type        string
	Name        string
	Slug        string
	CreatedByID uuid.UUID
	UpdatedByID uuid.UUID
	DeletedByID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
