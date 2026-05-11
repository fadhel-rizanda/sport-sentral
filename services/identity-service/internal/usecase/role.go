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

type RoleUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*RoleResponse, error)
	GetByName(ctx context.Context, name string) (*RoleResponse, error)
	List(ctx context.Context, req ListRequest) (*ListRoleResponse, error)
	Create(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error)
	Update(ctx context.Context, req UpdateRoleRequest) (*RoleResponse, error)
	SoftDelete(ctx context.Context, req DeleteRequest) error

	AssignPermission(ctx context.Context, roleID string, permissionIDs []string) error
	RevokePermission(ctx context.Context, roleID string, permissionIDs []string) error
}

type roleUseCase struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
}

func NewRoleUseCase(roleRepo repository.RoleRepository, permissionRepo repository.PermissionRepository, userRepo repository.UserRepository) RoleUseCase {
	return &roleUseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
	}
}

func (uc *roleUseCase) GetByID(ctx context.Context, id uuid.UUID) (*RoleResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	return ToRoleResponse(role), nil
}

func (uc *roleUseCase) GetByName(ctx context.Context, name string) (*RoleResponse, error) {
	role, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	return ToRoleResponse(role), nil
}

func (uc *roleUseCase) List(ctx context.Context, req ListRequest) (*ListRoleResponse, error) {
	roleList, total, err := uc.roleRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role")
		}
		return nil, apperr.Internal(err)
	}
	result := make([]*RoleResponse, total)
	for i, role := range roleList {
		result[i] = ToRoleResponse(role)
	}

	return &ListRoleResponse{
		Roles:    result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *roleUseCase) Create(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_name") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}
	return ToRoleResponse(role), nil
}

func (uc *roleUseCase) Update(ctx context.Context, req UpdateRoleRequest) (*RoleResponse, error) {
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
	if req.Description != nil || req.Name != nil {
		role.UpdatedByID = req.UpdatedByID
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roles_name") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}
	return ToRoleResponse(role), nil
}

func (uc *roleUseCase) SoftDelete(ctx context.Context, req DeleteRequest) error {
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

	return uc.roleRepo.AssignPermissions(ctx, rid, pids)
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

	return uc.roleRepo.RevokePermissions(ctx, rid, pids)
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
