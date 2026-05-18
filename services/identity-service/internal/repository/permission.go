package repository

import (
	"context"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
)

type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error)
	List(ctx context.Context, roleID *uuid.UUID, page, pageSize int) ([]*entity.Permission, int64, error)

	Create(ctx context.Context, permission *entity.Permission) error
	Update(ctx context.Context, permission *entity.Permission) error

	AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error

	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context, roleID *uuid.UUID, page, pageSize int) ([]*entity.Permission, int64, error) {
	var permissions []*entity.Permission
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Permission{})

	if roleID != nil {
		query = query.Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
			Where("rp.role_id = ?", *roleID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&permissions).Error

	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

func (r *permissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionRepository) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	role := &entity.Role{ID: roleID}
	permission := &entity.Permission{ID: permissionID}
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permission)
}

func (r *permissionRepository) RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	role := &entity.Role{ID: roleID}
	permission := &entity.Permission{ID: permissionID}
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permission)
}

func (r *permissionRepository) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
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
