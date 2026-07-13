package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	replicatedRepo "microservice-golang/services/academy-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	"microservice-golang/shared/pkg/constants"
	apperr "microservice-golang/shared/pkg/errors"
)

type AcademyHoldingUseCase interface {
	Create(ctx context.Context, req dto.CreateAcademyHoldingRequest) (*dto.AcademyHoldingResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyHoldingResponse, error)
	List(ctx context.Context, req dto.ListAcademyHoldingsRequest) (*dto.ListAcademyHoldingsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyHoldingRequest) (*dto.AcademyHoldingResponse, error)
	Delete(ctx context.Context, req dto.DeleteAcademyHoldingRequest) error
}

type academyHoldingUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.AcademyHoldingRepository
	publisher      AcademyHoldingEventPublisher
}

func NewAcademyHoldingUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.AcademyHoldingRepository,
	publisher AcademyHoldingEventPublisher,
) AcademyHoldingUseCase {
	return &academyHoldingUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		publisher:      publisher,
	}
}

func (uc *academyHoldingUseCase) Create(ctx context.Context, req dto.CreateAcademyHoldingRequest) (*dto.AcademyHoldingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyCreate); err != nil {
		return nil, err
	}
	holdingID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	addressID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	address := &entity.AcademyHoldingAddress{
		ID:                       addressID,
		StreetAddress:            req.StreetAddress,
		Notes:                    req.Notes,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		AdministrativeDivisionID: req.AdminDivisionID,
	}

	holding := &entity.AcademyHolding{
		ID:                holdingID,
		Name:              req.Name,
		Description:       req.Description,
		Email:             req.Email,
		PhoneNumber:       req.PhoneNumber,
		ImageAttachmentID: req.ImageAttachmentID,
		StatusID:          req.StatusID,
		AddressID:         addressID,
		Address:           address,
		CreatedByID:       req.CreatedByID,
		UpdatedByID:       req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, holding); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_academy_holdings_name") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}

	// Fetch holding fully preloaded to return it
	resHolding, err := uc.repo.GetByID(ctx, holdingID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var imgID *string
	if resHolding.ImageAttachmentID != nil {
		str := resHolding.ImageAttachmentID.String()
		imgID = &str
	}
	_ = uc.publisher.PublishHoldingCreated(ctx, &academyv1.AcademyHoldingEvent{
		EventId:           evtID.String(),
		EventType:         academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_CREATED,
		OccurredAt:        timestamppb.New(time.Now()),
		HoldingId:         resHolding.ID.String(),
		Name:              resHolding.Name,
		Description:       resHolding.Description,
		Email:             resHolding.Email,
		PhoneNumber:       resHolding.PhoneNumber,
		ImageAttachmentId: imgID,
		StatusId:          resHolding.StatusID.String(),
	})

	return mapper.ToAcademyHoldingResponse(resHolding), nil
}

func (uc *academyHoldingUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyHoldingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	holding, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy holding")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyHoldingResponse(holding), nil
}

func (uc *academyHoldingUseCase) List(ctx context.Context, req dto.ListAcademyHoldingsRequest) (*dto.ListAcademyHoldingsResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	filters := repository.AcademyHoldingFilters{
		StatusID: req.StatusID,
		Search:   req.Search,
	}

	holdings, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AcademyHoldingResponse, len(holdings))
	for i, h := range holdings {
		result[i] = mapper.ToAcademyHoldingResponse(h)
	}

	return &dto.ListAcademyHoldingsResponse{
		Holdings: result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *academyHoldingUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyHoldingRequest) (*dto.AcademyHoldingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyUpdate); err != nil {
		return nil, err
	}
	holding, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy holding")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.Name != nil {
		holding.Name = *req.Name
		updated = true
	}
	if req.Description != nil {
		holding.Description = *req.Description
		updated = true
	}
	if req.Email != nil {
		holding.Email = *req.Email
		updated = true
	}
	if req.PhoneNumber != nil {
		holding.PhoneNumber = *req.PhoneNumber
		updated = true
	}
	if req.ImageAttachmentID != nil {
		holding.ImageAttachmentID = req.ImageAttachmentID
		updated = true
	}

	if updated {
		holding.UpdatedByID = req.UpdatedByID
		if err := uc.repo.Update(ctx, holding); err != nil {
			if postgres.IsUniqueConstraint(err, "uni_academy_holdings_name") {
				return nil, apperr.Conflict("name")
			}
			return nil, apperr.Internal(err)
		}
	}

	// Fetch holding fully preloaded to return it
	resHolding, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var imgID *string
	if resHolding.ImageAttachmentID != nil {
		str := resHolding.ImageAttachmentID.String()
		imgID = &str
	}
	_ = uc.publisher.PublishHoldingUpdated(ctx, &academyv1.AcademyHoldingEvent{
		EventId:           evtID.String(),
		EventType:         academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_UPDATED,
		OccurredAt:        timestamppb.New(time.Now()),
		HoldingId:         resHolding.ID.String(),
		Name:              resHolding.Name,
		Description:       resHolding.Description,
		Email:             resHolding.Email,
		PhoneNumber:       resHolding.PhoneNumber,
		ImageAttachmentId: imgID,
		StatusId:          resHolding.StatusID.String(),
	})

	return mapper.ToAcademyHoldingResponse(resHolding), nil
}

func (uc *academyHoldingUseCase) Delete(ctx context.Context, req dto.DeleteAcademyHoldingRequest) error {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyDelete); err != nil {
		return err
	}
	// Check if exists
	holding, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy holding")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var imgID *string
	if holding.ImageAttachmentID != nil {
		str := holding.ImageAttachmentID.String()
		imgID = &str
	}
	_ = uc.publisher.PublishHoldingDeleted(ctx, &academyv1.AcademyHoldingEvent{
		EventId:           evtID.String(),
		EventType:         academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_DELETED,
		OccurredAt:        timestamppb.New(time.Now()),
		HoldingId:         holding.ID.String(),
		Name:              holding.Name,
		Description:       holding.Description,
		Email:             holding.Email,
		PhoneNumber:       holding.PhoneNumber,
		ImageAttachmentId: imgID,
		StatusId:          holding.StatusID.String(),
	})

	return nil
}
