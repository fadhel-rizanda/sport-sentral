package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"microservice-golang/services/venue-service/internal/dto"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/services/venue-service/internal/mapper"
	"microservice-golang/services/venue-service/internal/repository"
	replicatedRepo "microservice-golang/services/venue-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type CourtSlotUseCase interface {
	CreateBulk(ctx context.Context, req dto.CreateCourtSlotsRequest) ([]*dto.CourtSlotResponse, error)
	List(ctx context.Context, req dto.ListCourtSlotsRequest) ([]*dto.CourtSlotResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type courtSlotUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.CourtSlotRepository
	courtRepo      repository.CourtRepository
}

func NewCourtSlotUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.CourtSlotRepository,
	courtRepo repository.CourtRepository,
) CourtSlotUseCase {
	return &courtSlotUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		courtRepo:      courtRepo,
	}
}

func (uc *courtSlotUseCase) CreateBulk(ctx context.Context, req dto.CreateCourtSlotsRequest) ([]*dto.CourtSlotResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return nil, err
	}

	court, err := uc.courtRepo.GetByID(ctx, req.CourtID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("court")
		}
		return nil, apperr.Internal(err)
	}

	slots := make([]*entity.CourtSlot, len(req.Slots))
	for i, sInput := range req.Slots {
		slotID, err := uuid.NewV7()
		if err != nil {
			return nil, apperr.Internal(err)
		}

		price := court.PricePerHour
		if sInput.Price != nil {
			price = *sInput.Price
		}

		slots[i] = &entity.CourtSlot{
			ID:        slotID,
			CourtID:   req.CourtID,
			StartTime: sInput.StartTime,
			EndTime:   sInput.EndTime,
			Price:     price,
			StatusID:  req.StatusID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	if err := uc.repo.CreateBulk(ctx, slots); err != nil {
		return nil, apperr.Internal(err)
	}

	// Fetch created slots to return fully preloaded statuses
	resSlots, err := uc.repo.List(ctx, req.CourtID, &req.StatusID, nil, nil)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Filter only the slots we just created
	createdMap := make(map[uuid.UUID]bool)
	for _, s := range slots {
		createdMap[s.ID] = true
	}

	res := make([]*dto.CourtSlotResponse, 0, len(slots))
	for _, s := range resSlots {
		if createdMap[s.ID] {
			res = append(res, mapper.ToCourtSlotResponse(s))
		}
	}

	return res, nil
}

func (uc *courtSlotUseCase) List(ctx context.Context, req dto.ListCourtSlotsRequest) ([]*dto.CourtSlotResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.read"); err != nil {
		return nil, err
	}

	slots, err := uc.repo.List(ctx, req.CourtID, req.StatusID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	res := make([]*dto.CourtSlotResponse, len(slots))
	for i, s := range slots {
		res[i] = mapper.ToCourtSlotResponse(s)
	}

	return res, nil
}

func (uc *courtSlotUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return err
	}

	slot, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("court_slot")
		}
		return apperr.Internal(err)
	}

	if slot.BookingID != nil {
		return apperr.Conflict("cannot delete a slot that is already booked")
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}

	return nil
}
