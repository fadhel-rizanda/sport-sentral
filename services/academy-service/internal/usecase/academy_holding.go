package usecase

import (
	"context"

	"github.com/google/uuid"

	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
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
	repo repository.AcademyHoldingRepository
}

func NewAcademyHoldingUseCase(repo repository.AcademyHoldingRepository) AcademyHoldingUseCase {
	return &academyHoldingUseCase{
		repo: repo,
	}
}

func (uc *academyHoldingUseCase) Create(ctx context.Context, req dto.CreateAcademyHoldingRequest) (*dto.AcademyHoldingResponse, error) {
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
		if postgres.IsUniqueConstraint(err, "uni_academy_holdings_name") || postgres.IsUniqueConstraint(err, "name") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}

	// Fetch holding fully preloaded to return it
	resHolding, err := uc.repo.GetByID(ctx, holdingID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyHoldingResponse(resHolding), nil
}

func (uc *academyHoldingUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyHoldingResponse, error) {
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
			if postgres.IsUniqueConstraint(err, "uni_academy_holdings_name") || postgres.IsUniqueConstraint(err, "name") {
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

	return mapper.ToAcademyHoldingResponse(resHolding), nil
}

func (uc *academyHoldingUseCase) Delete(ctx context.Context, req dto.DeleteAcademyHoldingRequest) error {
	// Check if exists
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy holding")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
