package usecase

import (
	"context"
	"microservice-golang/shared/pkg/constants"

	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type ProfileUseCase interface {
	ApplyProfile(ctx context.Context, req ApplyProfileRequest) error
	ToggleProfile(ctx context.Context, req ToggleProfileRequest) error
	ApproveProfile(ctx context.Context, req ApproveProfileRequest) error
	RejectProfile(ctx context.Context, req ApproveProfileRequest) error
}

type profileUseCase struct {
	userRepo        repository.UserRepository
	userRoleRepo    repository.UserRoleRepository
	roleRepo        repository.RoleRepository
	statusCacheRepo repository.StatusCacheRepository
}

func NewProfileUseCase(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
	statusCacheRepo repository.StatusCacheRepository,
) ProfileUseCase {
	return &profileUseCase{
		userRepo:        userRepo,
		userRoleRepo:    userRoleRepo,
		roleRepo:        roleRepo,
		statusCacheRepo: statusCacheRepo,
	}
}

func (uc *profileUseCase) ApplyProfile(ctx context.Context, req ApplyProfileRequest) error {
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("user")
		}
		return apperr.Internal(err)
	}

	// hanya athlete yang bisa expand profile
	activeRole, err := uc.userRoleRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return apperr.Internal(err)
	}

	allowedExpansions, canExpand := entity.RolesExpandableFrom[activeRole.Role.Name]
	if !canExpand {
		return apperr.Forbidden("your role cannot apply for additional profiles")
	}

	// cek apakah role target allowed
	allowed := false
	for _, r := range allowedExpansions {
		if r == req.RoleName {
			allowed = true
			break
		}
	}
	if !allowed {
		return apperr.Forbidden("cannot apply for this profile")
	}

	// cek sudah punya role ini belum
	targetRole, err := uc.roleRepo.GetByName(ctx, req.RoleName)
	if err != nil {
		return apperr.NotFound("role")
	}

	existing, err := uc.userRoleRepo.GetByUserIDAndRoleID(ctx, user.ID, targetRole.ID)
	if err != nil && !postgres.IsNotFound(err) {
		return apperr.Internal(err)
	}
	if existing != nil {
		return apperr.Conflict("profile already applied or active")
	}

	// scout → langsung active, lainnya → pending
	statusName := constants.StatusPending
	if req.RoleName == constants.RoleScout {
		statusName = constants.StatusActive
	}

	status, err := uc.statusCacheRepo.GetByTypeAndName(ctx, constants.StatusTypeUserRole, statusName)
	if err != nil {
		return err
	}

	userRole := &entity.UserRole{
		UserID:   user.ID,
		RoleID:   targetRole.ID,
		IsActive: false,
		StatusID: status.ID,
	}

	return uc.userRoleRepo.Add(ctx, userRole)
}

func (uc *profileUseCase) ToggleProfile(ctx context.Context, req ToggleProfileRequest) error {
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("user")
		}
		return apperr.Internal(err)
	}

	targetRole, err := uc.roleRepo.GetByName(ctx, req.RoleName)
	if err != nil {
		return apperr.NotFound("role")
	}

	userRole, err := uc.userRoleRepo.GetByUserIDAndRoleID(ctx, user.ID, targetRole.ID)
	if err != nil {
		return apperr.NotFound("profile")
	}

	// hanya bisa toggle jika status active
	if userRole.Status.Name != constants.StatusActive {
		return apperr.Forbidden("profile is not yet approved")
	}

	return uc.userRoleRepo.SetActive(ctx, user.ID, targetRole.ID)
}

func (uc *profileUseCase) ApproveProfile(ctx context.Context, req ApproveProfileRequest) error {
	targetRole, err := uc.roleRepo.GetByName(ctx, req.RoleName)
	if err != nil {
		return apperr.NotFound("role")
	}

	userRole, err := uc.userRoleRepo.GetByUserIDAndRoleID(ctx, req.UserID, targetRole.ID)
	if err != nil {
		return apperr.NotFound("profile")
	}

	if userRole.Status.Name != constants.StatusPending {
		return apperr.InvalidArgument("profile is not pending approval")
	}

	activeStatus, err := uc.statusCacheRepo.GetByTypeAndName(ctx, constants.StatusTypeUserRole, constants.StatusActive)
	if err != nil {
		return err
	}

	userRole.StatusID = activeStatus.ID
	return uc.userRoleRepo.Update(ctx, userRole)
}

func (uc *profileUseCase) RejectProfile(ctx context.Context, req ApproveProfileRequest) error {
	targetRole, err := uc.roleRepo.GetByName(ctx, req.RoleName)
	if err != nil {
		return apperr.NotFound("role")
	}

	userRole, err := uc.userRoleRepo.GetByUserIDAndRoleID(ctx, req.UserID, targetRole.ID)
	if err != nil {
		return apperr.NotFound("profile")
	}

	if userRole.Status.Name != constants.StatusPending {
		return apperr.InvalidArgument("profile is not pending approval")
	}

	// hapus userRole — rejected berarti tidak ada sama sekali
	return uc.userRoleRepo.Delete(ctx, req.UserID, targetRole.ID)
}
