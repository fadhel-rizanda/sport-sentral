package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error

	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RevokePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []*entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Offset(offset).Limit(pageSize).
		Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permissions)
}

func (r *roleRepository) RevokePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permissions)
}

func (r *roleRepository) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	permissions := toPermissionRefs(permissionIDs)
	role := &entity.Role{}
	role.ID = roleID
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions)
}

func (r *roleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
		return nil, err
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
