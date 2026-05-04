package usecase

import (
	"time"

	"github.com/google/uuid"
	"microservice-golang/services/meta-service/internal/entity"
)

// ─── Status ───────────────────────────────────────────────────────────────────

type CreateStatusRequest struct {
	Type      string
	Name      string
	CreatedBy uuid.UUID
}

type UpdateStatusRequest struct {
	Type      *string
	Name      *string
	UpdatedBy uuid.UUID
}

type DeleteStatusRequest struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type StatusResponse struct {
	ID        uuid.UUID
	Type      string
	Name      string
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func toStatusResponse(s *entity.Status) *StatusResponse {
	return &StatusResponse{
		ID:        s.ID,
		Type:      s.Type,
		Name:      s.Name,
		CreatedBy: s.CreatedBy,
		UpdatedBy: s.UpdatedBy,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// ─── Tag ──────────────────────────────────────────────────────────────────────

type CreateTagRequest struct {
	Type      string
	Name      string
	Slug      *string
	CreatedBy uuid.UUID
}

type UpdateTagRequest struct {
	Type      *string
	Name      *string
	Slug      *string
	UpdatedBy uuid.UUID
}

type DeleteTagRequest struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type TagResponse struct {
	ID        uuid.UUID
	Type      string
	Name      string
	Slug      string
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func toTagResponse(t *entity.Tag) *TagResponse {
	return &TagResponse{
		ID:        t.ID,
		Type:      t.Type,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedBy: t.CreatedBy,
		UpdatedBy: t.UpdatedBy,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
