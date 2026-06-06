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

type AcademyBranchUseCase interface {
	Create(ctx context.Context, req dto.CreateAcademyBranchRequest) (*dto.AcademyBranchResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyBranchResponse, error)
	List(ctx context.Context, req dto.ListAcademyBranchesRequest) (*dto.ListAcademyBranchesResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyBranchRequest) (*dto.AcademyBranchResponse, error)
	Delete(ctx context.Context, req dto.DeleteAcademyBranchRequest) error
}

type academyBranchUseCase struct {
	repo repository.AcademyBranchRepository
}

func NewAcademyBranchUseCase(repo repository.AcademyBranchRepository) AcademyBranchUseCase {
	return &academyBranchUseCase{
		repo: repo,
	}
}

func (uc *academyBranchUseCase) Create(ctx context.Context, req dto.CreateAcademyBranchRequest) (*dto.AcademyBranchResponse, error) {
	branchID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	addressID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	address := entity.AcademyBranchAddress{
		ID:                       addressID,
		BranchID:                 branchID,
		IsPrimary:                true,
		StreetAddress:            req.StreetAddress,
		Notes:                    req.Notes,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		AdministrativeDivisionID: req.AdminDivisionID,
	}

	branch := &entity.AcademyBranch{
		ID:          branchID,
		HoldingID:   req.HoldingID,
		SportID:     req.SportID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		StatusID:    req.StatusID,
		Addresses:   []entity.AcademyBranchAddress{address},
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, branch); err != nil {
		return nil, apperr.Internal(err)
	}

	// Fetch holding fully preloaded to return it
	resBranch, err := uc.repo.GetByID(ctx, branchID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyBranchResponse(resBranch), nil
}

func (uc *academyBranchUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyBranchResponse, error) {
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy branch")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyBranchResponse(branch), nil
}

func (uc *academyBranchUseCase) List(ctx context.Context, req dto.ListAcademyBranchesRequest) (*dto.ListAcademyBranchesResponse, error) {
	filters := repository.AcademyBranchFilters{
		HoldingID: req.HoldingID,
		SportID:   req.SportID,
		StatusID:  req.StatusID,
		Search:    req.Search,
	}

	branches, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AcademyBranchResponse, len(branches))
	for i, b := range branches {
		result[i] = mapper.ToAcademyBranchResponse(b)
	}

	return &dto.ListAcademyBranchesResponse{
		Branches: result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *academyBranchUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyBranchRequest) (*dto.AcademyBranchResponse, error) {
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy branch")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.Name != nil {
		branch.Name = *req.Name
		updated = true
	}
	if req.Email != nil {
		branch.Email = *req.Email
		updated = true
	}
	if req.PhoneNumber != nil {
		branch.PhoneNumber = *req.PhoneNumber
		updated = true
	}
	if req.StatusID != nil {
		branch.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		branch.UpdatedByID = req.UpdatedByID
		if err := uc.repo.Update(ctx, branch); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	resBranch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyBranchResponse(resBranch), nil
}

func (uc *academyBranchUseCase) Delete(ctx context.Context, req dto.DeleteAcademyBranchRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy branch")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
