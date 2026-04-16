package usecase

import (
	"context"
	"github.com/google/uuid"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
)

type PermissionUseCase interface {
	CreatePermission(ctx context.Context, resource, action, description string) (*entity.Permission, error)
	GetPermission(ctx context.Context, id string) (*entity.Permission, error)
	DeletePermission(ctx context.Context, id string) error
	ListPermissions(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error)

	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

type permissionUseCase struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionUseCase(permissionRepo repository.PermissionRepository) PermissionUseCase {
	return &permissionUseCase{permissionRepo: permissionRepo}
}

func (uc *permissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*entity.Permission, error) {
	permission := entity.NewPermission(resource, action, description)

	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (uc *permissionUseCase) GetPermission(ctx context.Context, id string) (*entity.Permission, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid permission id")
	}
	return uc.permissionRepo.GetByID(ctx, uid)
}

func (uc *permissionUseCase) DeletePermission(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperr.InvalidArgument("invalid permission id")
	}
	return uc.permissionRepo.Delete(ctx, uid)
}

func (uc *permissionUseCase) ListPermissions(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.permissionRepo.List(ctx, page, pageSize)
}

func (uc *permissionUseCase) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false, apperr.InvalidArgument("invalid user id")
	}
	return uc.permissionRepo.CheckPermission(ctx, uID, resource, action)
}
