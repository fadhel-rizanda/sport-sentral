package usecase

import (
	"context"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/mapper"
	"time"

	"gorm.io/gorm"

	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StatusUseCase interface {
	Create(ctx context.Context, req dto.CreateStatusRequest) (*dto.StatusResponse, error)
	GetByTypeAndName(ctx context.Context, statusType, name string) (*dto.StatusResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.StatusResponse, error)
	List(ctx context.Context, req dto.ListStatusesRequest) (*dto.ListStatusesResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateStatusRequest) (*dto.StatusResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteStatusRequest) error
	HardDelete(ctx context.Context, req dto.DeleteStatusRequest) error
}

type statusUseCase struct {
	repo      repository.StatusRepository
	publisher StatusEventPublisher
}

func NewStatusUseCase(repo repository.StatusRepository, publisher StatusEventPublisher) StatusUseCase {
	return &statusUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *statusUseCase) Create(ctx context.Context, req dto.CreateStatusRequest) (*dto.StatusResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	status := &entity.Status{
		ID:          id,
		Type:        req.Type,
		Name:        req.Name,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
		Slug:        req.Slug,
	}

	if err := uc.repo.Create(ctx, status); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_statuses_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	// Build event
	evt := uc.buildEvent(
		metav1.StatusEventType_STATUS_EVENT_TYPE_CREATED,
		status,
	)
	if err := uc.publisher.PublishStatusCreated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToStatusResponse(status), nil
}

func (uc *statusUseCase) GetByTypeAndName(ctx context.Context, statusType, name string) (*dto.StatusResponse, error) {
	s, err := uc.repo.GetByTypeAndName(ctx, statusType, name)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToStatusResponse(s), nil
}

func (uc *statusUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.StatusResponse, error) {
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToStatusResponse(s), nil
}

func (uc *statusUseCase) List(ctx context.Context, req dto.ListStatusesRequest) (*dto.ListStatusesResponse, error) {
	statuses, total, err := uc.repo.List(ctx, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.StatusResponse, len(statuses))
	for i, s := range statuses {
		result[i] = mapper.ToStatusResponse(s)
	}

	return &dto.ListStatusesResponse{
		Statuses: result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *statusUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateStatusRequest) (*dto.StatusResponse, error) {
	status, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}

	if req.Name != nil {
		status.Name = *req.Name
	}
	if req.Type != nil {
		status.Type = *req.Type
	}
	if req.Slug != nil {
		status.Slug = *req.Slug
	}
	if req.Name != nil || req.Type != nil || req.Slug != nil {
		status.UpdatedByID = req.UpdatedByID
	}

	if err := uc.repo.Update(ctx, *status); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_statuses_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	evt := uc.buildEvent(
		metav1.StatusEventType_STATUS_EVENT_TYPE_UPDATED,
		status,
	)

	if err := uc.publisher.PublishStatusUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToStatusResponse(status), nil
}

func (uc *statusUseCase) SoftDelete(ctx context.Context, req dto.DeleteStatusRequest) error {
	status, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("status")
		}
		return apperr.Internal(err)
	}

	status.DeletedByID = &req.DeletedByID
	status.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := uc.repo.Update(ctx, *status); err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildEvent(
		metav1.StatusEventType_STATUS_EVENT_TYPE_DELETED,
		status,
	)

	if err := uc.publisher.PublishStatusDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *statusUseCase) HardDelete(ctx context.Context, req dto.DeleteStatusRequest) error {
	status, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("status")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	// Delete: identity-service retains the last StatusName as tombstone,
	// no event needed. If you later want cleanup, publish DELETED here too.
	_ = status
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (uc *statusUseCase) buildEvent(
	eventType metav1.StatusEventType,
	s *entity.Status,
) *metav1.StatusEvent {
	evtID, _ := uuid.NewV7()

	evt := &metav1.StatusEvent{
		EventId:    evtID.String(),
		EventType:  eventType,
		OccurredAt: timestamppb.Now(),
		StatusId:   s.ID.String(),
		StatusType: s.Type,
		StatusName: s.Name,
		StatusSlug: s.Slug,
	}
	if s.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(s.DeletedAt.Time)
	}

	return evt
}
