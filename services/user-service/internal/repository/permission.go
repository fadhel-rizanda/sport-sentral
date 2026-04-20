package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/user-service/internal/entity"
	pgerr "microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *entity.Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error)
	Update(ctx context.Context, permission *entity.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error)

	AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error

	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

type gormPermissionRepo struct {
	db *gorm.DB
}

func NewGormPermissionRepository(db *gorm.DB) PermissionRepository {
	return &gormPermissionRepo{db: db}
}

func (r *gormPermissionRepo) Create(ctx context.Context, permission *entity.Permission) error {
	result := r.db.WithContext(ctx).Create(permission)

	if result.Error != nil {
		if pgerr.IsUniqueConstraint(result.Error, "permissions_resource_action_key") {
			return apperr.Conflict("permission already exists")
		}
		return apperr.Internal(result.Error)
	}

	return nil
}

func (r *gormPermissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	var permission entity.Permission
	result := r.db.WithContext(ctx).First(&permission, "id = ?", id)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("permission not found")
		}
		return nil, apperr.Internal(result.Error)
	}

	return &permission, nil
}

func (r *gormPermissionRepo) Update(ctx context.Context, permission *entity.Permission) error {
	result := r.db.WithContext(ctx).
		Model(permission).
		Updates(map[string]interface{}{
			"resource":    permission.Resource,
			"action":      permission.Action,
			"description": permission.Description,
		})

	if result.Error != nil {
		return apperr.Internal(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperr.NotFound("permission not found")
	}

	return nil
}

func (r *gormPermissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Permission{}, "id = ?", id)

	if result.Error != nil {
		return apperr.Internal(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperr.NotFound("permission not found")
	}

	return nil
}

func (r *gormPermissionRepo) List(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error) {
	var permissions []*entity.Permission
	var total int64

	result := r.db.WithContext(ctx).
		Model(&entity.Permission{}).
		Count(&total).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&permissions)

	if result.Error != nil {
		return nil, 0, apperr.Internal(result.Error)
	}

	return permissions, total, nil
}

func (r *gormPermissionRepo) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	role := &entity.Role{ID: roleID}
	permission := &entity.Permission{ID: permissionID}

	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permission); err != nil {
		if pgerr.IsUniqueConstraint(err, "role_permissions_pkey") {
			return apperr.Conflict("permission already assigned to role")
		}
		return apperr.Internal(err)
	}

	return nil
}

func (r *gormPermissionRepo) RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	role := &entity.Role{ID: roleID}
	permission := &entity.Permission{ID: permissionID}

	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permission); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (r *gormPermissionRepo) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Joins("JOIN users ON users.id = user_roles.user_id").
		Where("users.id = ?", userID).
		Where("permissions.resource = ?", resource).
		Where("permissions.action = ?", action).
		Where("users.deleted_at IS NULL").
		Where("roles.deleted_at IS NULL").
		Where("permissions.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, apperr.Internal(err)
	}

	return count > 0, nil
}
