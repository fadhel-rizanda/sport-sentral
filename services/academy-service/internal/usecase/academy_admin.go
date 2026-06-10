package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type AcademyAdminUseCase interface {
	Create(ctx context.Context, req dto.CreateAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyAdminResponse, error)
	List(ctx context.Context, req dto.ListAcademyAdminsRequest) (*dto.ListAcademyAdminsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	Delete(ctx context.Context, req dto.DeleteAcademyAdminRequest) error
	GetByUser(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (*dto.AcademyAdminResponse, error)
	CheckUserIsAdmin(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (bool, error)

	Assign(ctx context.Context, req dto.AssignAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	Revoke(ctx context.Context, req dto.RevokeAcademyAdminRequest) error
}

type academyAdminUseCase struct {
	repo repository.AcademyAdminRepository
}

func NewAcademyAdminUseCase(repo repository.AcademyAdminRepository) AcademyAdminUseCase {
	return &academyAdminUseCase{
		repo: repo,
	}
}

func (uc *academyAdminUseCase) Create(ctx context.Context, req dto.CreateAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	admin := &entity.AcademyAdmin{
		ID:          id,
		AcademyID:   req.AcademyID,
		BranchID:    req.BranchID,
		UserID:      req.UserID,
		RoleID:      req.RoleID,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyAdminResponse, error) {
	admin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyAdminResponse(admin), nil
}

func (uc *academyAdminUseCase) List(ctx context.Context, req dto.ListAcademyAdminsRequest) (*dto.ListAcademyAdminsResponse, error) {
	filters := repository.AdminFilters{
		AdminScope: repository.AdminScope{
			HoldingID: req.AcademyID,
			BranchID:  req.BranchID,
		},
		StatusID: req.StatusID,
		Search:   req.Search,
	}

	admins, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AcademyAdminResponse, len(admins))
	for i, a := range admins {
		result[i] = mapper.ToAcademyAdminResponse(a)
	}

	return &dto.ListAcademyAdminsResponse{
		Admins:   result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *academyAdminUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	admin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}

	admin.RoleID = req.RoleID
	admin.UpdatedByID = req.UpdatedByID

	if err := uc.repo.Update(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) Delete(ctx context.Context, req dto.DeleteAcademyAdminRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy admin")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *academyAdminUseCase) GetByUser(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (*dto.AcademyAdminResponse, error) {
	admin, err := uc.repo.GetByUser(ctx, scope, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyAdminResponse(admin), nil
}

func (uc *academyAdminUseCase) CheckUserIsAdmin(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (bool, error) {
	isAdmin, err := uc.repo.CheckUserIsAdmin(ctx, scope, userID)
	if err != nil {
		return false, apperr.Internal(err)
	}
	return isAdmin, nil
}

func (uc *academyAdminUseCase) Assign(ctx context.Context, req dto.AssignAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	now := time.Now()
	admin := &entity.AcademyAdmin{
		ID:           id,
		AcademyID:    req.AcademyID,
		BranchID:     req.BranchID,
		UserID:       req.UserID,
		RoleID:       req.RoleID,
		CreatedByID:  req.AssignedByID,
		UpdatedByID:  req.AssignedByID,
		ApprovedAt:   &now,
		ApprovedByID: &req.AssignedByID,
	}

	if err := uc.repo.Assign(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) Revoke(ctx context.Context, req dto.RevokeAcademyAdminRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy admin")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Revoke(ctx, req.ID, req.RevokedByID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
