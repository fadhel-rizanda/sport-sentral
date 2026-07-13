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

type VenueUseCase interface {
	Create(ctx context.Context, req dto.CreateVenueRequest) (*dto.VenueResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.VenueResponse, error)
	List(ctx context.Context, req dto.ListVenuesRequest) (*dto.ListVenuesResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateVenueRequest) (*dto.VenueResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type venueUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.VenueRepository
}

func NewVenueUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.VenueRepository,
) VenueUseCase {
	return &venueUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
	}
}

func (uc *venueUseCase) Create(ctx context.Context, req dto.CreateVenueRequest) (*dto.VenueResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return nil, err
	}

	venueID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	venue := &entity.Venue{
		ID:                       venueID,
		Name:                     req.Name,
		Description:              req.Description,
		StreetAddress:            req.StreetAddress,
		Notes:                    req.Notes,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		AdministrativeDivisionID: req.AdminDivisionID,
		OwnerID:                  req.OwnerID,
		StatusID:                 req.StatusID,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	if err := uc.repo.Create(ctx, venue); err != nil {
		return nil, apperr.Internal(err)
	}

	resVenue, err := uc.repo.GetByID(ctx, venueID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToVenueResponse(resVenue), nil
}

func (uc *venueUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.VenueResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.read"); err != nil {
		return nil, err
	}

	venue, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("venue")
		}
		return nil, apperr.Internal(err)
	}

	return mapper.ToVenueResponse(venue), nil
}

func (uc *venueUseCase) List(ctx context.Context, req dto.ListVenuesRequest) (*dto.ListVenuesResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.read"); err != nil {
		return nil, err
	}

	filters := repository.VenueFilters{
		OwnerID:  req.OwnerID,
		StatusID: req.StatusID,
		Search:   req.Search,
	}

	venues, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.VenueResponse, len(venues))
	for i, v := range venues {
		result[i] = mapper.ToVenueResponse(v)
	}

	return &dto.ListVenuesResponse{
		Venues:   result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *venueUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateVenueRequest) (*dto.VenueResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return nil, err
	}

	venue, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("venue")
		}
		return nil, apperr.Internal(err)
	}

	if req.Name != nil {
		venue.Name = *req.Name
	}
	if req.Description != nil {
		venue.Description = *req.Description
	}
	if req.StreetAddress != nil {
		venue.StreetAddress = *req.StreetAddress
	}
	if req.Notes != nil {
		venue.Notes = req.Notes
	}
	if req.Latitude != nil {
		venue.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		venue.Longitude = req.Longitude
	}
	if req.StatusID != nil {
		venue.StatusID = *req.StatusID
	}
	venue.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, venue); err != nil {
		return nil, apperr.Internal(err)
	}

	resVenue, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToVenueResponse(resVenue), nil
}

func (uc *venueUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "court.manage"); err != nil {
		return err
	}

	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("venue")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}

	return nil
}
