package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/user-service/internal/entity"
	apperr "microservice-golang/shared/pkg/errors"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByIDWithPermissions(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)

	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RevokePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
}

type gormRoleRepo struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepo{db: db}
}

func (r *gormRoleRepo) Create(ctx context.Context, role *entity.Role) error {
	result := r.db.WithContext(ctx).Create(role)

	if result.Error != nil {
		if isDuplicateError(result.Error) {
			return apperr.Conflict("role name already exists")
		}
		return apperr.Internal(result.Error)
	}

	return nil
}

func (r *gormRoleRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	result := r.db.WithContext(ctx).First(&role, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("role not found")
		}
		return nil, apperr.Internal(result.Error)
	}

	return &role, nil
}

func (r *gormRoleRepo) GetByIDWithPermissions(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("role not found")
		}
		return nil, apperr.Internal(result.Error)
	}

	return &role, nil
}

func (r *gormRoleRepo) Update(ctx context.Context, role *entity.Role) error {
	result := r.db.WithContext(ctx).
		Model(role).
		Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
		})

	if result.Error != nil {
		if isDuplicateError(result.Error) {
			return apperr.Conflict("role name already exists")
		}
		return apperr.Internal(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperr.NotFound("role not found")
	}

	return nil
}

func (r *gormRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Role{}, "id = ?", id)

	if result.Error != nil {
		return apperr.Internal(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperr.NotFound("role not found")
	}

	return nil
}

func (r *gormRoleRepo) List(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.Role{}).Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&roles)

	if result.Error != nil {
		return nil, 0, apperr.Internal(result.Error)
	}

	return roles, total, nil
}

func (r *gormRoleRepo) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID

	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permissions); err != nil {
		if isDuplicateError(err) {
			return apperr.Conflict("one or more permissions already assigned to role")
		}
		return apperr.Internal(err)
	}

	return nil
}

func (r *gormRoleRepo) RevokePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID

	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permissions); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (r *gormRoleRepo) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID

	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (r *gormRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	var role entity.Role

	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("role not found")
		}
		return nil, apperr.Internal(err)
	}

	return role.Permissions, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toPermissionRefs(ids []uuid.UUID) []entity.Permission {
	permissions := make([]entity.Permission, len(ids))
	for i, id := range ids {
		permissions[i] = entity.Permission{}
		permissions[i].ID = id
	}
	return permissions
}
