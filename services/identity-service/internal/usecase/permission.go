package usecase

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	"time"
)

type PermissionUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*PermissionResponse, error)
	List(ctx context.Context, req ListPermissionRequest) (*ListPermissionResponse, error)
	Create(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error)
	Update(ctx context.Context, req UpdatePermissionRequest) (*PermissionResponse, error)
	SoftDelete(ctx context.Context, req DeleteRequest) error
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

type permissionUseCase struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionUseCase(permissionRepo repository.PermissionRepository) PermissionUseCase {
	return &permissionUseCase{
		permissionRepo: permissionRepo,
	}
}

func (p *permissionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*PermissionResponse, error) {
	permission, err := p.permissionRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("permission")
		}
		return nil, apperr.Internal(err)
	}
	return ToPermissionResponse(permission), nil
}

func (p *permissionUseCase) List(ctx context.Context, req ListPermissionRequest) (*ListPermissionResponse, error) {
	permissionList, total, err := p.permissionRepo.List(ctx, req.RoleId, req.Page, req.PageSize)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("permission")
		}
		return nil, apperr.Internal(err)
	}
	result := make([]*PermissionResponse, total)
	for i, permission := range permissionList {
		result[i] = ToPermissionResponse(permission)
	}
	return &ListPermissionResponse{
		Permissions: result,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (p *permissionUseCase) Create(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error) {
	permission := &entity.Permission{
		Action:      req.Action,
		Resource:    req.Resource,
		Description: req.Description,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}
	if err := p.permissionRepo.Create(ctx, permission); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_permissions_action_resource") {
			return nil, apperr.Conflict("action and resource combination")
		}
		return nil, apperr.Internal(err)
	}
	return ToPermissionResponse(permission), nil
}

func (p *permissionUseCase) Update(ctx context.Context, req UpdatePermissionRequest) (*PermissionResponse, error) {
	permission, err := p.permissionRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("permission")
		}
		return nil, apperr.Internal(err)
	}
	if req.Action != nil {
		permission.Action = *req.Action
	}
	if req.Resource != nil {
		permission.Resource = *req.Resource
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.Description != nil || req.Action != nil || req.Resource != nil {
		permission.UpdatedByID = req.UpdatedByID
	}

	if err := p.permissionRepo.Update(ctx, permission); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_permissions_action_resource") {
			return nil, apperr.Conflict("action and resource combination")
		}
		return nil, apperr.Internal(err)
	}
	return ToPermissionResponse(permission), nil
}

func (p *permissionUseCase) SoftDelete(ctx context.Context, req DeleteRequest) error {
	permission, err := p.permissionRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("permission")
		}
		return apperr.Internal(err)
	}

	permission.DeletedByID = &req.DeletedByID
	permission.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := p.permissionRepo.Update(ctx, permission); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (p *permissionUseCase) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false, apperr.InvalidArgument("invalid user id")
	}
	return p.permissionRepo.CheckPermission(ctx, uID, resource, action)
}
