package usecase

import (
	"context"

	"github.com/google/uuid"

	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/repository"
	replicatedRepo "microservice-golang/services/competition-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type CompetitionBranchUseCase interface {
	Create(ctx context.Context, adminID, competitionID uuid.UUID, req dto.CreateBranchRequest) (*dto.BranchResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.BranchResponse, error)
	ListByCompetition(ctx context.Context, competitionID uuid.UUID) ([]*dto.BranchResponse, error)
	Update(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateBranchRequest) (*dto.BranchResponse, error)
	Delete(ctx context.Context, adminID, id uuid.UUID) error
}

type competitionBranchUseCase struct {
	permissionRepo   replicatedRepo.PermissionRepository
	repo             repository.CompetitionBranchRepository
	compRepo         repository.CompetitionRepository
	academyAdminRepo replicatedRepo.AcademyAdminRepository
}

func NewCompetitionBranchUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.CompetitionBranchRepository,
	compRepo repository.CompetitionRepository,
	academyAdminRepo replicatedRepo.AcademyAdminRepository,
) CompetitionBranchUseCase {
	return &competitionBranchUseCase{
		permissionRepo:   permissionRepo,
		repo:             repo,
		compRepo:         compRepo,
		academyAdminRepo: academyAdminRepo,
	}
}

func (uc *competitionBranchUseCase) Create(ctx context.Context, adminID, competitionID uuid.UUID, req dto.CreateBranchRequest) (*dto.BranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.create"); err != nil {
		return nil, err
	}
	if err := validateCompetitionAdmin(ctx, adminID, competitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	branch := &entity.CompetitionBranch{
		ID:             id,
		CompetitionID:  competitionID,
		ParentBranchID: req.ParentBranchID,
		Name:           req.Name,
		Description:    req.Description,
		StatusID:       req.StatusID,
		CreatedByID:    adminID,
		UpdatedByID:    adminID,
	}

	if err := uc.repo.Create(ctx, branch); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_branch_name") {
			return nil, apperr.Conflict("a branch with this name already exists in this competition")
		}
		return nil, apperr.Internal(err)
	}

	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToBranchResponse(res), nil
}

func (uc *competitionBranchUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.BranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("competition branch")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToBranchResponse(res), nil
}

func (uc *competitionBranchUseCase) ListByCompetition(ctx context.Context, competitionID uuid.UUID) ([]*dto.BranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	filters := repository.CompetitionBranchFilters{
		CompetitionID: &competitionID,
	}

	list, _, err := uc.repo.List(ctx, filters, 1, 1000)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	res := make([]*dto.BranchResponse, len(list))
	for i, b := range list {
		res[i] = mapper.ToBranchResponse(b)
	}
	return res, nil
}

func (uc *competitionBranchUseCase) Update(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateBranchRequest) (*dto.BranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.update"); err != nil {
		return nil, err
	}
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("competition branch")
		}
		return nil, apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return nil, err
	}

	updated := false
	if req.Name != nil {
		branch.Name = *req.Name
		updated = true
	}
	if req.Description != nil {
		branch.Description = *req.Description
		updated = true
	}
	if req.StatusID != nil {
		branch.StatusID = *req.StatusID
		updated = true
	}
	if req.ParentBranchID != nil {
		branch.ParentBranchID = req.ParentBranchID
		updated = true
	}

	if updated {
		branch.UpdatedByID = adminID
		if err := uc.repo.Update(ctx, branch); err != nil {
			if postgres.IsUniqueConstraint(err, "idx_branch_name") {
				return nil, apperr.Conflict("a branch with this name already exists in this competition")
			}
			return nil, apperr.Internal(err)
		}
	}

	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToBranchResponse(res), nil
}

func (uc *competitionBranchUseCase) Delete(ctx context.Context, adminID, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.delete"); err != nil {
		return err
	}
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("competition branch")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
