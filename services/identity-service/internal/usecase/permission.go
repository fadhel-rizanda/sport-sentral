package usecase

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/dto"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/mapper"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	sharedgrpc "microservice-golang/shared/pkg/grpc"
)

type PermissionUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error)
	List(ctx context.Context, req dto.ListPermissionRequest) (*dto.ListPermissionResponse, error)
	Create(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	Update(ctx context.Context, req dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteRequest) error
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

type permissionUseCase struct {
	permissionRepo repository.PermissionRepository
	publisher      PermissionEventPublisher
}

func NewPermissionUseCase(permissionRepo repository.PermissionRepository, publisher PermissionEventPublisher) PermissionUseCase {
	return &permissionUseCase{
		permissionRepo: permissionRepo,
		publisher:      publisher,
	}
}

func (uc *permissionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error) {
	if err := uc.validatePermission(ctx, "permission", "read"); err != nil {
		return nil, err
	}
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("permission")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToPermissionResponse(permission), nil
}

func (uc *permissionUseCase) List(ctx context.Context, req dto.ListPermissionRequest) (*dto.ListPermissionResponse, error) {
	if err := uc.validatePermission(ctx, "permission", "read"); err != nil {
		return nil, err
	}
	permissionList, total, err := uc.permissionRepo.List(ctx, req.RoleId, req.Page, req.PageSize)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("permission")
		}
		return nil, apperr.Internal(err)
	}
	result := make([]*dto.PermissionResponse, len(permissionList))
	for i, permission := range permissionList {
		result[i] = mapper.ToPermissionResponse(permission)
	}
	return &dto.ListPermissionResponse{
		Permissions: result,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (uc *permissionUseCase) Create(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	if err := uc.validatePermission(ctx, "permission", "create"); err != nil {
		return nil, err
	}
	permission := &entity.Permission{
		Action:      req.Action,
		Resource:    req.Resource,
		Description: req.Description,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
		Slug:        req.Slug,
	}
	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_permissions_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_CREATED, permission)
	if err := uc.publisher.PublishPermissionCreated(ctx, event); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToPermissionResponse(permission), nil
}

func (uc *permissionUseCase) Update(ctx context.Context, req dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	if err := uc.validatePermission(ctx, "permission", "update"); err != nil {
		return nil, err
	}
	permission, err := uc.permissionRepo.GetByID(ctx, req.ID)
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
	if req.Slug != nil {
		permission.Slug = *req.Slug
	}
	if req.Description != nil || req.Action != nil || req.Resource != nil || req.Slug != nil {
		permission.UpdatedByID = req.UpdatedByID
	}

	if err := uc.permissionRepo.Update(ctx, permission); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_permissions_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_UPDATED, permission)
	if err := uc.publisher.PublishPermissionUpdated(ctx, event); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToPermissionResponse(permission), nil
}

func (uc *permissionUseCase) SoftDelete(ctx context.Context, req dto.DeleteRequest) error {
	if err := uc.validatePermission(ctx, "permission", "delete"); err != nil {
		return err
	}
	permission, err := uc.permissionRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("permission")
		}
		return apperr.Internal(err)
	}

	permission.DeletedByID = &req.DeletedByID
	permission.DeletedAt.Time = time.Now()
	permission.DeletedAt.Valid = true

	if err := uc.permissionRepo.Update(ctx, permission); err != nil {
		return apperr.Internal(err)
	}

	// Event build
	event := uc.buildEvent(rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_DELETED, permission)
	if err := uc.publisher.PublishPermissionDeleted(ctx, event); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *permissionUseCase) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false, apperr.InvalidArgument("invalid user id")
	}
	return uc.permissionRepo.CheckPermission(ctx, uID, resource, action)
}

func (uc *permissionUseCase) validatePermission(ctx context.Context, resource, action string) error {
	userID, err := sharedgrpc.ExtractUserID(ctx)
	if err != nil {
		return err
	}
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return err
	}
	if activeRole == "platform_admin" {
		return nil
	}
	has, err := uc.permissionRepo.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		return err
	}
	if !has {
		return apperr.Forbidden("insufficient permissions")
	}
	return nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (uc *permissionUseCase) buildEvent(
	eventType rbacv1.PermissionEventType,
	p *entity.Permission,
) *rbacv1.PermissionEvent {
	evtID, _ := uuid.NewV7()

	evt := &rbacv1.PermissionEvent{
		EventId:            evtID.String(),
		EventType:          eventType,
		OccurredAt:         timestamppb.Now(),
		PermissionId:       p.ID.String(),
		PermissionAction:   p.Action,
		PermissionResource: p.Resource,
		PermissionSlug:     p.Slug,
	}
	if p.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(p.DeletedAt.Time)
	}

	return evt
}
