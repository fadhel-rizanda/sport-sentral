package usecase

import (
	"context"
	"github.com/google/uuid"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
)

type RoleUseCase interface {
	CreateRole(ctx context.Context, name, description string, permissionIDs []string) (*entity.Role, error)
	GetRole(ctx context.Context, id string) (*entity.Role, error)
	UpdateRole(ctx context.Context, id, name, description string, permissionIDs []string) (*entity.Role, error)
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)

	AssignPermission(ctx context.Context, roleID string, permissionIDs []string) error
	RevokePermission(ctx context.Context, roleID string, permissionIDs []string) error

	AssignRolesToUser(ctx context.Context, userID string, roleIDs []string) error
	RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error
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

func (uc *roleUseCase) CreateRole(ctx context.Context, name, description string, permissionIDs []string) (*entity.Role, error) {
	role := &entity.Role{
		Name:        name,
		Description: description,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	if len(permissionIDs) > 0 {
		pids, err := parseUUIDs(permissionIDs)
		if err != nil {
			return nil, apperr.InvalidArgument("invalid permission ids")
		}

		if err := uc.validatePermissionsExist(ctx, pids); err != nil {
			return nil, err
		}

		if err := uc.roleRepo.AssignPermissions(ctx, role.ID, pids); err != nil {
			return nil, err
		}
	}

	return uc.roleRepo.GetByIDWithPermissions(ctx, role.ID)
}

func (uc *roleUseCase) GetRole(ctx context.Context, id string) (*entity.Role, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid role id")
	}
	return uc.roleRepo.GetByIDWithPermissions(ctx, uid)
}

func (uc *roleUseCase) UpdateRole(ctx context.Context, id, name, description string, permissionIDs []string) (*entity.Role, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid role id")
	}

	role, err := uc.roleRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	pids, err := parseUUIDs(permissionIDs)
	if err != nil {
		return nil, err
	}

	if len(pids) > 0 {
		if err := uc.validatePermissionsExist(ctx, pids); err != nil {
			return nil, err
		}
	}
	if err := uc.roleRepo.ReplacePermissions(ctx, role.ID, pids); err != nil {
		return nil, err
	}

	return uc.roleRepo.GetByIDWithPermissions(ctx, role.ID)
}

func (uc *roleUseCase) DeleteRole(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperr.InvalidArgument("invalid role id")
	}
	return uc.roleRepo.Delete(ctx, uid)
}

func (uc *roleUseCase) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 15
	}
	return uc.roleRepo.List(ctx, page, pageSize)
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

func (uc *roleUseCase) AssignRolesToUser(ctx context.Context, userID string, roleID []string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	rids, err := parseUUIDs(roleID)
	if err != nil {
		return apperr.InvalidArgument("invalid role id")
	}

	if _, err := uc.userRepo.GetByID(ctx, uid); err != nil {
		return err
	}

	for _, rid := range rids {
		if _, err := uc.roleRepo.GetByID(ctx, rid); err != nil {
			return err
		}
	}

	return uc.userRepo.AssignRoles(ctx, uid, rids)
}

func (uc *roleUseCase) RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	rIDs, err := parseUUIDs(roleIDs)
	if err != nil {
		return err
	}

	return uc.userRepo.RemoveRoles(ctx, uID, rIDs)
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
