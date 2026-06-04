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
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type RoleUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error)
	GetByName(ctx context.Context, name string) (*dto.RoleResponse, error)
	List(ctx context.Context, req dto.ListRequest) (*dto.ListRoleResponse, error)
	Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	Update(ctx context.Context, req dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteRequest) error

	AssignPermission(ctx context.Context, roleID string, permissionIDs []string) error
	RevokePermission(ctx context.Context, roleID string, permissionIDs []string) error
}

type roleUseCase struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
	logger         *zap.Logger
	publisher      RoleEventPublisher
}

func NewRoleUseCase(roleRepo repository.RoleRepository, permissionRepo repository.PermissionRepository, userRepo repository.UserRepository, logger *zap.Logger, publisher RoleEventPublisher) RoleUseCase {
	return &roleUseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
		logger:         logger,
		publisher:      publisher,
	}
}

func (uc *roleUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToRoleResponse(role), nil
}

func (uc *roleUseCase) GetByName(ctx context.Context, name string) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToRoleResponse(role), nil
}

func (uc *roleUseCase) List(ctx context.Context, req dto.ListRequest) (*dto.ListRoleResponse, error) {
	roleList, total, err := uc.roleRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	result := make([]*dto.RoleResponse, len(roleList))
	for i, role := range roleList {
		result[i] = mapper.ToRoleResponse(role)
	}

	return &dto.ListRoleResponse{
		Roles:    result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *roleUseCase) Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
		Slug:        req.Slug,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_name") {
			return nil, apperr.Conflict("name")
		}
		if postgres.IsUniqueConstraint(err, "uni_roles_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	perms := make([]uuid.UUID, len(req.PermissionIDs))
	for i, p := range req.PermissionIDs {
		if p != nil {
			perms[i] = *p
		}
	}

	if req.PermissionIDs != nil {
		err := uc.roleRepo.AssignPermissions(ctx, role.ID, perms)
		if err != nil {
			if postgres.IsUniqueConstraint(err, "idx_roles_permission_ids") {
				return nil, apperr.Conflict("permission_ids")
			}
			return nil, apperr.Internal(err)
		}
	}

	newRole, err := uc.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.RoleEventType_ROLE_EVENT_TYPE_CREATED, newRole)
	if err := uc.publisher.PublishRoleCreated(ctx, event); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRoleResponse(newRole), nil
}

func (uc *roleUseCase) Update(ctx context.Context, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Slug != nil {
		role.Slug = *req.Slug
	}
	if req.Description != nil || req.Name != nil || req.Slug != nil {
		role.UpdatedByID = req.UpdatedByID
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_name") {
			return nil, apperr.Conflict("name")
		}
		if postgres.IsUniqueConstraint(err, "uni_roles_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	perms := make([]uuid.UUID, len(req.PermissionIDs))
	for i, p := range req.PermissionIDs {
		if p != nil {
			perms[i] = *p
		}
	}

	if req.PermissionIDs != nil {
		err := uc.roleRepo.ReplacePermissions(ctx, role.ID, perms)
		if err != nil {
			if postgres.IsUniqueConstraint(err, "idx_roles_permission_ids") {
				return nil, apperr.Conflict("permission_ids")
			}
			return nil, apperr.Internal(err)
		}
	}

	newRole, err := uc.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.RoleEventType_ROLE_EVENT_TYPE_UPDATED, newRole)
	if err := uc.publisher.PublishRoleUpdated(ctx, event); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRoleResponse(newRole), nil
}

func (uc *roleUseCase) SoftDelete(ctx context.Context, req dto.DeleteRequest) error {
	role, err := uc.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("role")
		}
		return apperr.Internal(err)
	}
	role.DeletedByID = &req.DeletedByID
	role.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.RoleEventType_ROLE_EVENT_TYPE_DELETED, role)
	if err := uc.publisher.PublishRoleDeleted(ctx, event); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *roleUseCase) AssignPermission(ctx context.Context, roleID string, permissionIDs []string) error {
	rid, err := uuid.Parse(roleID)
	if err != nil {
		return apperr.InvalidArgument("invalid role id")
	}
	if _, err := uc.roleRepo.GetByID(ctx, rid); err != nil {
		return err
	}

	pids, err := parseUUIDs(permissionIDs)
	if err != nil {
		return apperr.InvalidArgument("invalid permission id")
	}
	if err := uc.validatePermissionsExist(ctx, pids); err != nil {
		return err
	}

	if err := uc.roleRepo.AssignPermissions(ctx, rid, pids); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_permission_ids") {
			return apperr.Conflict("permission_ids")
		}
		return apperr.Internal(err)
	}

	newRole, err := uc.roleRepo.GetByID(ctx, rid)
	if err != nil {
		return apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.RoleEventType_ROLE_EVENT_TYPE_UPDATED, newRole)
	if err := uc.publisher.PublishRoleUpdated(ctx, event); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *roleUseCase) RevokePermission(ctx context.Context, roleID string, permissionIDs []string) error {
	rid, err := uuid.Parse(roleID)
	if err != nil {
		return apperr.InvalidArgument("invalid role id")
	}
	if _, err := uc.roleRepo.GetByID(ctx, rid); err != nil {
		return err
	}

	pids, err := parseUUIDs(permissionIDs)
	if err != nil {
		return apperr.InvalidArgument("invalid permission id")
	}
	if err := uc.validatePermissionsExist(ctx, pids); err != nil {
		return err
	}

	if err := uc.roleRepo.RevokePermissions(ctx, rid, pids); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_permission_ids") {
			return apperr.Conflict("permission_ids")
		}
		return apperr.Internal(err)
	}

	newRole, err := uc.roleRepo.GetByID(ctx, rid)
	if err != nil {
		return apperr.Internal(err)
	}

	// Build event
	event := uc.buildEvent(rbacv1.RoleEventType_ROLE_EVENT_TYPE_UPDATED, newRole)
	if err := uc.publisher.PublishRoleUpdated(ctx, event); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	uids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, apperr.InvalidArgument("invalid uuid: " + id)
		}
		uids = append(uids, uid)
	}
	return uids, nil
}

func (uc *roleUseCase) validatePermissionsExist(ctx context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if _, err := uc.permissionRepo.GetByID(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (uc *roleUseCase) buildEvent(
	eventType rbacv1.RoleEventType,
	r *entity.Role,
) *rbacv1.RoleEvent {
	evtID, _ := uuid.NewV7()

	evt := &rbacv1.RoleEvent{
		EventId:    evtID.String(),
		EventType:  eventType,
		OccurredAt: timestamppb.Now(),

		RoleId:   r.ID.String(),
		RoleName: r.Name,
		RoleSlug: r.Slug,
		PermissionIds: func() []string {
			ids := make([]string, len(r.Permissions))
			for index, permission := range r.Permissions {
				ids[index] = permission.ID.String()
			}
			return ids
		}(), // () diakhir berarti func ini langsung dijalankan
	}
	if r.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(r.DeletedAt.Time)
	}

	return evt
}
