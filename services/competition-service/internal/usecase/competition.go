package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/repository"
	replicatedRepo "microservice-golang/services/competition-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	sharedgrpc "microservice-golang/shared/pkg/grpc"
)

type CompetitionUseCase interface {
	Create(ctx context.Context, adminID uuid.UUID, req dto.CreateCompetitionRequest) (*dto.CompetitionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CompetitionResponse, error)
	List(ctx context.Context, filters dto.CompetitionFilters, page, pageSize int) ([]*dto.CompetitionResponse, int64, error)
	Update(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateCompetitionRequest) (*dto.CompetitionResponse, error)
	Delete(ctx context.Context, adminID, id uuid.UUID) error
	UpdateStatus(ctx context.Context, adminID, id uuid.UUID, statusID uuid.UUID) error
}

type competitionUseCase struct {
	permissionRepo    replicatedRepo.PermissionRepository
	repo              repository.CompetitionRepository
	academyAdminRepo  replicatedRepo.AcademyAdminRepository
	academyBranchRepo replicatedRepo.AcademyBranchRepository
	roleRepo          replicatedRepo.RoleRepository
}

func NewCompetitionUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.CompetitionRepository,
	academyAdminRepo replicatedRepo.AcademyAdminRepository,
	academyBranchRepo replicatedRepo.AcademyBranchRepository,
	roleRepo replicatedRepo.RoleRepository,
) CompetitionUseCase {
	return &competitionUseCase{
		permissionRepo:    permissionRepo,
		repo:              repo,
		academyAdminRepo:  academyAdminRepo,
		academyBranchRepo: academyBranchRepo,
		roleRepo:          roleRepo,
	}
}

func (uc *competitionUseCase) Create(ctx context.Context, adminID uuid.UUID, req dto.CreateCompetitionRequest) (*dto.CompetitionResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.create"); err != nil {
		return nil, err
	}

	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return nil, err
	}

	// Validation based on creator role and HostAcademyBranchID presence
	if activeRole != "platform_admin" {
		if req.HostAcademyBranchID == nil || *req.HostAcademyBranchID == uuid.Nil {
			// Created by organizer (no academy branch)
			if activeRole != "organizer" {
				return nil, apperr.Forbidden("only organizers or platform admins can create competitions without an academy branch")
			}
		} else {
			// Created by academy admin (with academy branch)
			academyAdmin, err := uc.academyAdminRepo.GetByUserID(ctx, adminID)
			if err != nil {
				return nil, apperr.Forbidden("user is not an academy admin")
			}
			if academyAdmin.BranchID != nil {
				if *academyAdmin.BranchID != *req.HostAcademyBranchID {
					return nil, apperr.Forbidden("insufficient permissions for this academy branch")
				}
			} else {
				// Check if user is academy holding admin
				branch, err := uc.academyBranchRepo.GetByID(ctx, *req.HostAcademyBranchID)
				if err != nil {
					return nil, apperr.Forbidden("host academy branch not found")
				}
				if academyAdmin.AcademyID != branch.HoldingID {
					return nil, apperr.Forbidden("insufficient permissions for this academy branch")
				}
			}
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if req.EndDate != nil && req.StartDate.After(*req.EndDate) {
		return nil, apperr.InvalidArgument("start date must be before or equal to end date")
	}

	var hostBranchID *uuid.UUID
	if req.HostAcademyBranchID != nil && *req.HostAcademyBranchID != uuid.Nil {
		hostBranchID = req.HostAcademyBranchID
	}

	comp := &entity.Competition{
		ID:                  id,
		SportID:             req.SportID,
		HostAcademyBranchID: hostBranchID,
		Name:                req.Name,
		Description:         req.Description,
		TierID:              req.TierID,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		StatusID:            req.StatusID,
		CreatedByID:         adminID,
		UpdatedByID:         adminID,
	}

	// TODO need a transaction
	if err := uc.repo.Create(ctx, comp); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_competition_name") {
			return nil, apperr.Conflict("a competition with this name already exists")
		}
		return nil, apperr.Internal(err)
	}

	// Auto-assign the creator as the first competition admin
	adminRecordID, err := uuid.NewV7()
	if err == nil {
		var roleID uuid.UUID
		roleRecord, err := uc.roleRepo.GetBySlug(ctx, activeRole)
		if err == nil {
			roleID = roleRecord.ID
		}
		adminRecord := &entity.CompetitionAdmin{
			ID:            adminRecordID,
			CompetitionID: id,
			UserID:        adminID,
			RoleID:        roleID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_ = uc.repo.CreateAdmin(ctx, adminRecord)
	}

	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCompetitionResponse(res), nil
}

func (uc *competitionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.CompetitionResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("competition")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToCompetitionResponse(res), nil
}

func (uc *competitionUseCase) List(ctx context.Context, filters dto.CompetitionFilters, page, pageSize int) ([]*dto.CompetitionResponse, int64, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, 0, err
	}
	repoFilters := repository.CompetitionFilters{
		BranchID: filters.BranchID,
		SportID:  filters.SportID,
		TierID:   filters.TierID,
		StatusID: filters.StatusID,
		Search:   filters.Search,
	}

	list, total, err := uc.repo.List(ctx, repoFilters, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}

	res := make([]*dto.CompetitionResponse, len(list))
	for i, c := range list {
		res[i] = mapper.ToCompetitionResponse(c)
	}
	return res, total, nil
}

func (uc *competitionUseCase) Update(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateCompetitionRequest) (*dto.CompetitionResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.update"); err != nil {
		return nil, err
	}
	if err := validateCompetitionAdmin(ctx, adminID, id, uc.repo, uc.academyAdminRepo); err != nil {
		return nil, err
	}

	comp, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("competition")
		}
		return nil, apperr.Internal(err)
	}

	// Date range validation
	startDate := comp.StartDate
	if req.StartDate != nil {
		startDate = *req.StartDate
	}
	endDate := comp.EndDate
	if req.EndDate != nil {
		endDate = req.EndDate
	}
	if endDate != nil && startDate.After(*endDate) {
		return nil, apperr.InvalidArgument("start date must be before or equal to end date")
	}

	updated := false
	if req.SportID != nil {
		comp.SportID = *req.SportID
		updated = true
	}
	if req.HostAcademyBranchID != nil {
		if *req.HostAcademyBranchID == uuid.Nil {
			comp.HostAcademyBranchID = nil
		} else {
			comp.HostAcademyBranchID = req.HostAcademyBranchID
		}
		updated = true
	}
	if req.Name != nil {
		comp.Name = *req.Name
		updated = true
	}
	if req.Description != nil {
		comp.Description = *req.Description
		updated = true
	}
	if req.TierID != nil {
		comp.TierID = *req.TierID
		updated = true
	}
	if req.StartDate != nil {
		comp.StartDate = *req.StartDate
		updated = true
	}
	if req.EndDate != nil {
		comp.EndDate = req.EndDate
		updated = true
	}
	if req.StatusID != nil {
		comp.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		comp.UpdatedByID = adminID
		if err := uc.repo.Update(ctx, comp); err != nil {
			if postgres.IsUniqueConstraint(err, "idx_competition_name") {
				return nil, apperr.Conflict("a competition with this name already exists")
			}
			return nil, apperr.Internal(err)
		}
	}

	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCompetitionResponse(res), nil
}

func (uc *competitionUseCase) Delete(ctx context.Context, adminID, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.delete"); err != nil {
		return err
	}
	if err := validateCompetitionAdmin(ctx, adminID, id, uc.repo, uc.academyAdminRepo); err != nil {
		return err
	}

	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("competition")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *competitionUseCase) UpdateStatus(ctx context.Context, adminID, id uuid.UUID, statusID uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	if err := validateCompetitionAdmin(ctx, adminID, id, uc.repo, uc.academyAdminRepo); err != nil {
		return err
	}

	comp, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("competition")
		}
		return apperr.Internal(err)
	}

	comp.StatusID = statusID
	comp.UpdatedByID = adminID

	if err := uc.repo.Update(ctx, comp); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
