package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"microservice-golang/services/venue-service/internal/dto"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/services/venue-service/internal/mapper"
	"microservice-golang/services/venue-service/internal/repository"
	replicatedRepo "microservice-golang/services/venue-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	"microservice-golang/shared/pkg/constants"
	"microservice-golang/shared/pkg/database"
	apperr "microservice-golang/shared/pkg/errors"
)

type BookingUseCase interface {
	Create(ctx context.Context, req dto.CreateBookingRequest) (*dto.BookingResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error)
	List(ctx context.Context, req dto.ListBookingsRequest) (*dto.ListBookingsResponse, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, req dto.UpdateBookingStatusRequest) (*dto.BookingResponse, error)
	CheckAvailability(ctx context.Context, courtID uuid.UUID, slotIDs []uuid.UUID) (bool, error)
}

type bookingUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.BookingRepository
	courtRepo      repository.CourtRepository
	slotRepo       repository.CourtSlotRepository
	statusRepo     replicatedRepo.StatusRepository
	txManager      database.TransactionManager
}

func NewBookingUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.BookingRepository,
	courtRepo repository.CourtRepository,
	slotRepo repository.CourtSlotRepository,
	statusRepo replicatedRepo.StatusRepository,
	txManager database.TransactionManager,
) BookingUseCase {
	return &bookingUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		courtRepo:      courtRepo,
		slotRepo:       slotRepo,
		statusRepo:     statusRepo,
		txManager:      txManager,
	}
}

func (uc *bookingUseCase) Create(ctx context.Context, req dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionBookingCreate); err != nil {
		return nil, err
	}

	if len(req.SlotIDs) == 0 {
		return nil, apperr.InvalidArgument("at least one slot_id is required")
	}

	_, err := uc.courtRepo.GetByID(ctx, req.CourtID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("court")
		}
		return nil, apperr.Internal(err)
	}

	bookedStatus, err := uc.statusRepo.GetByTypeAndSlug(ctx, "booking_slot", "booked")
	if err != nil {
		bookedStatus, err = uc.statusRepo.GetByTypeAndSlug(ctx, "slot", "booked")
		if err != nil {
			return nil, apperr.Internal(fmt.Errorf("could not find booked status: %w", err))
		}
	}

	var resBooking *entity.Booking

	err = uc.txManager.Run(ctx, func(txCtx context.Context) error {
		slots, err := uc.slotRepo.GetByIDsForUpdate(txCtx, req.SlotIDs)
		if err != nil {
			return err
		}

		if len(slots) != len(req.SlotIDs) {
			return apperr.NotFound("one or more slots not found")
		}

		var totalPrice int64
		for _, slot := range slots {
			if slot.CourtID != req.CourtID {
				return apperr.InvalidArgument(fmt.Sprintf("slot %s does not belong to court %s", slot.ID, req.CourtID))
			}
			if slot.BookingID != nil || slot.Status.Slug != "available" {
				return apperr.Conflict(fmt.Sprintf("slot %s is not available", slot.ID))
			}
			totalPrice += slot.Price
		}

		bookingID, err := uuid.NewV7()
		if err != nil {
			return err
		}

		booking := &entity.Booking{
			ID:            bookingID,
			CourtID:       req.CourtID,
			UserID:        req.UserID,
			TotalPrice:    totalPrice,
			StatusID:      req.StatusID,
			PaymentStatus: "UNPAID",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := uc.repo.Create(txCtx, booking); err != nil {
			return err
		}

		if err := uc.slotRepo.UpdateStatusBulk(txCtx, req.SlotIDs, bookedStatus.ID, &bookingID); err != nil {
			return err
		}

		resBooking, err = uc.repo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if apperr.IsNotFound(err) || apperr.IsInvalidArgument(err) || apperr.IsConflict(err) {
			return nil, err
		}
		return nil, apperr.Internal(err)
	}

	return mapper.ToBookingResponse(resBooking), nil
}

func (uc *bookingUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionBookingRead); err != nil {
		return nil, err
	}

	booking, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("booking")
		}
		return nil, apperr.Internal(err)
	}

	return mapper.ToBookingResponse(booking), nil
}

func (uc *bookingUseCase) List(ctx context.Context, req dto.ListBookingsRequest) (*dto.ListBookingsResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionBookingRead); err != nil {
		return nil, err
	}

	filters := repository.BookingFilters{
		CourtID:  req.CourtID,
		UserID:   req.UserID,
		StatusID: req.StatusID,
	}

	bookings, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.BookingResponse, len(bookings))
	for i, b := range bookings {
		result[i] = mapper.ToBookingResponse(b)
	}

	return &dto.ListBookingsResponse{
		Bookings: result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *bookingUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, req dto.UpdateBookingStatusRequest) (*dto.BookingResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionBookingManage); err != nil {
		return nil, err
	}

	var resBooking *entity.Booking
	err := uc.txManager.Run(ctx, func(txCtx context.Context) error {
		booking, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		bookingStatus, err := uc.statusRepo.GetByID(txCtx, req.StatusID)
		if err != nil {
			return err
		}

		if bookingStatus.Slug == "cancelled" {
			availableStatus, err := uc.statusRepo.GetByTypeAndSlug(txCtx, "booking_slot", "available")
			if err != nil {
				availableStatus, err = uc.statusRepo.GetByTypeAndSlug(txCtx, "slot", "available")
				if err != nil {
					return fmt.Errorf("could not find available status: %w", err)
				}
			}

			slotIDs := make([]uuid.UUID, len(booking.Slots))
			for i, s := range booking.Slots {
				slotIDs[i] = s.ID
			}

			if len(slotIDs) > 0 {
				if err := uc.slotRepo.UpdateStatusBulk(txCtx, slotIDs, availableStatus.ID, nil); err != nil {
					return err
				}
			}
		}

		booking.StatusID = req.StatusID
		if req.PaymentStatus != nil {
			booking.PaymentStatus = *req.PaymentStatus
		}
		booking.UpdatedAt = time.Now()

		if err := uc.repo.Update(txCtx, booking); err != nil {
			return err
		}

		resBooking, err = uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("booking")
		}
		return nil, apperr.Internal(err)
	}

	return mapper.ToBookingResponse(resBooking), nil
}

func (uc *bookingUseCase) CheckAvailability(ctx context.Context, courtID uuid.UUID, slotIDs []uuid.UUID) (bool, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionBookingRead); err != nil {
		return false, err
	}

	if len(slotIDs) == 0 {
		return false, apperr.InvalidArgument("at least one slot_id is required")
	}

	slots, err := uc.slotRepo.GetByIDs(ctx, slotIDs)
	if err != nil {
		return false, apperr.Internal(err)
	}

	if len(slots) != len(slotIDs) {
		return false, nil
	}

	for _, slot := range slots {
		if slot.CourtID != courtID {
			return false, nil
		}
		if slot.BookingID != nil || slot.Status.Slug != "available" {
			return false, nil // Slot is booked or unavailable
		}
	}

	return true, nil
}
