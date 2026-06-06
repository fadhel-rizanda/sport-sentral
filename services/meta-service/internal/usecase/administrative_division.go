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

type AdministrativeDivisionUseCase interface {
	Create(ctx context.Context, req dto.CreateAdministrativeDivisionRequest) (*dto.AdministrativeDivisionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AdministrativeDivisionResponse, error)
	List(ctx context.Context, req dto.ListAdministrativeDivisionsRequest) (*dto.ListAdministrativeDivisionsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAdministrativeDivisionRequest) (*dto.AdministrativeDivisionResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteAdministrativeDivisionRequest) error
	HardDelete(ctx context.Context, req dto.DeleteAdministrativeDivisionRequest) error
}

type administrativeDivisionUseCase struct {
	repo      repository.AdministrativeDivisionRepository
	publisher AdministrativeDivisionEventPublisher
}

func NewAdministrativeDivisionUseCase(
	repo repository.AdministrativeDivisionRepository,
	publisher AdministrativeDivisionEventPublisher,
) AdministrativeDivisionUseCase {
	return &administrativeDivisionUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *administrativeDivisionUseCase) Create(ctx context.Context, req dto.CreateAdministrativeDivisionRequest) (*dto.AdministrativeDivisionResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	entity := &entity.AdministrativeDivision{
		ID:          id,
		CountryID:   req.CountryID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Level:       req.Level,
		PostalCode:  req.PostalCode,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, entity); err != nil {
		return nil, apperr.Internal(err)
	}

	// Refetch to preload relations
	entityWithRelations, err := uc.repo.GetByID(ctx, id)
	if err == nil {
		entity = entityWithRelations
	}

	evt := uc.buildEvent(
		metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_CREATED,
		entity,
	)
	if err := uc.publisher.PublishAdministrativeDivisionCreated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAdministrativeDivisionResponse(entity), nil
}

func (uc *administrativeDivisionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AdministrativeDivisionResponse, error) {
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("administrative division")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAdministrativeDivisionResponse(entity), nil
}

func (uc *administrativeDivisionUseCase) List(ctx context.Context, req dto.ListAdministrativeDivisionsRequest) (*dto.ListAdministrativeDivisionsResponse, error) {
	filters := repository.AdministrativeDivisionFilters{
		Name:       req.Name,
		Level:      req.Level,
		PostalCode: req.PostalCode,
		CountryID:  req.CountryID,
		ParentID:   req.ParentID,
	}

	entities, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AdministrativeDivisionResponse, len(entities))
	for i, e := range entities {
		result[i] = mapper.ToAdministrativeDivisionResponse(e)
	}

	return &dto.ListAdministrativeDivisionsResponse{
		AdministrativeDivisions: result,
		Total:                   total,
		Page:                    req.Page,
		PageSize:                req.PageSize,
	}, nil
}

func (uc *administrativeDivisionUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAdministrativeDivisionRequest) (*dto.AdministrativeDivisionResponse, error) {
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("administrative division")
		}
		return nil, apperr.Internal(err)
	}

	hasChanges := false
	if req.Name != nil {
		entity.Name = *req.Name
		hasChanges = true
	}
	if req.Level != nil {
		entity.Level = *req.Level
		hasChanges = true
	}
	if req.PostalCode != nil {
		entity.PostalCode = *req.PostalCode
		hasChanges = true
	}
	if req.CountryID != nil {
		entity.CountryID = *req.CountryID
		hasChanges = true
	}
	if req.ParentID != nil {
		entity.ParentID = req.ParentID
		hasChanges = true
	}

	if hasChanges {
		entity.UpdatedByID = req.UpdatedByID
	}

	if err := uc.repo.Update(ctx, entity); err != nil {
		return nil, apperr.Internal(err)
	}

	// Refetch to preload relations
	entityWithRelations, err := uc.repo.GetByID(ctx, id)
	if err == nil {
		entity = entityWithRelations
	}

	evt := uc.buildEvent(
		metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_UPDATED,
		entity,
	)
	if err := uc.publisher.PublishAdministrativeDivisionUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAdministrativeDivisionResponse(entity), nil
}

func (uc *administrativeDivisionUseCase) SoftDelete(ctx context.Context, req dto.DeleteAdministrativeDivisionRequest) error {
	entity, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("administrative division")
		}
		return apperr.Internal(err)
	}

	entity.DeletedByID = &req.DeletedByID
	entity.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := uc.repo.Update(ctx, entity); err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildEvent(
		metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_DELETED,
		entity,
	)
	if err := uc.publisher.PublishAdministrativeDivisionDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *administrativeDivisionUseCase) HardDelete(ctx context.Context, req dto.DeleteAdministrativeDivisionRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("administrative division")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *administrativeDivisionUseCase) buildEvent(
	eventType metav1.AdministrativeDivisionEventType,
	ad *entity.AdministrativeDivision,
) *metav1.AdministrativeDivisionEvent {
	evtID, _ := uuid.NewV7()

	evt := &metav1.AdministrativeDivisionEvent{
		EventId:                          evtID.String(),
		EventType:                        eventType,
		OccurredAt:                       timestamppb.Now(),
		AdministrativeDivisionId:         ad.ID.String(),
		AdministrativeDivisionName:       ad.Name,
		AdministrativeDivisionLevel:      ad.Level,
		AdministrativeDivisionPostalCode: ad.PostalCode,
		AdministrativeDivisionCountryId:  ad.CountryID.String(),
	}

	if ad.ParentID != nil {
		evt.AdministrativeDivisionParentId = ad.ParentID.String()
	}

	if ad.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(ad.DeletedAt.Time)
	}

	return evt
}
