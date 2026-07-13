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
	apperr "microservice-golang/shared/pkg/errors"
)

type CourtUseCase interface {
	Create(ctx context.Context, req dto.CreateCourtRequest) (*dto.CourtResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CourtResponse, error)
	List(ctx context.Context, req dto.ListCourtsRequest) (*dto.ListCourtsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateCourtRequest) (*dto.CourtResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type courtUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.CourtRepository
	venueRepo      repository.VenueRepository
}

func NewCourtUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.CourtRepository,
	venueRepo repository.VenueRepository,
) CourtUseCase {
	return &courtUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		venueRepo:      venueRepo,
	}
}

func (uc *courtUseCase) Create(ctx context.Context, req dto.CreateCourtRequest) (*dto.CourtResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return nil, err
	}

	// Validate Venue exists
	_, err := uc.venueRepo.GetByID(ctx, req.VenueID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("venue")
		}
		return nil, apperr.Internal(err)
	}

	courtID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	court := &entity.Court{
		ID:                courtID,
		VenueID:           req.VenueID,
		Name:              req.Name,
		Description:       req.Description,
		SportID:           req.SportID,
		PricePerHour:      req.PricePerHour,
		ImageAttachmentID: req.ImageAttachment,
		StatusID:          req.StatusID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := uc.repo.Create(ctx, court); err != nil {
		return nil, apperr.Internal(err)
	}

	resCourt, err := uc.repo.GetByID(ctx, courtID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCourtResponse(resCourt), nil
}

func (uc *courtUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.CourtResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.read"); err != nil {
		return nil, err
	}

	court, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("court")
		}
		return nil, apperr.Internal(err)
	}

	return mapper.ToCourtResponse(court), nil
}

func (uc *courtUseCase) List(ctx context.Context, req dto.ListCourtsRequest) (*dto.ListCourtsResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.read"); err != nil {
		return nil, err
	}

	filters := repository.CourtFilters{
		VenueID:  req.VenueID,
		SportID:  req.SportID,
		StatusID: req.StatusID,
		Search:   req.Search,
	}

	courts, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.CourtResponse, len(courts))
	for i, c := range courts {
		result[i] = mapper.ToCourtResponse(c)
	}

	return &dto.ListCourtsResponse{
		Courts:   result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *courtUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCourtRequest) (*dto.CourtResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return nil, err
	}

	court, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("court")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.Name != nil {
		court.Name = *req.Name
		updated = true
	}
	if req.Description != nil {
		court.Description = *req.Description
		updated = true
	}
	if req.SportID != nil {
		court.SportID = *req.SportID
		updated = true
	}
	if req.PricePerHour != nil {
		court.PricePerHour = *req.PricePerHour
		updated = true
	}
	if req.ImageAttachment != nil {
		court.ImageAttachmentID = req.ImageAttachment
		updated = true
	}
	if req.StatusID != nil {
		court.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		court.UpdatedAt = time.Now()
		if err := uc.repo.Update(ctx, court); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	resCourt, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCourtResponse(resCourt), nil
}

func (uc *courtUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return err
	}

	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("court")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(fmt.Errorf("delete court: %w", err))
	}

	return nil
}
