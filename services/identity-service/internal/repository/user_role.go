package repository

import (
	"context"

	"microservice-golang/services/identity-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoleRepository interface {
	Add(ctx context.Context, userRole *entity.UserRole) error
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserRole, error)
	GetByUserIDAndRoleID(ctx context.Context, userID, roleID uuid.UUID) (*entity.UserRole, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserRole, error)
	Update(ctx context.Context, userRole *entity.UserRole) error
	SetActive(ctx context.Context, userID, roleID uuid.UUID) error
	Delete(ctx context.Context, userID, roleID uuid.UUID) error
}

type userRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) Add(ctx context.Context, userRole *entity.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *userRoleRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserRole, error) {
	var userRole entity.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		First(&userRole, "user_id = ? AND is_active = true", userID).Error
	if err != nil {
		return nil, err
	}

	if err := r.loadStatus(ctx, &userRole); err != nil {
		return nil, err
	}

	return &userRole, nil
}

func (r *userRoleRepository) GetByUserIDAndRoleID(ctx context.Context, userID, roleID uuid.UUID) (*entity.UserRole, error) {
	var userRole entity.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		First(&userRole, "user_id = ? AND role_id = ?", userID, roleID).Error
	if err != nil {
		return nil, err
	}

	if err := r.loadStatus(ctx, &userRole); err != nil {
		return nil, err
	}

	return &userRole, nil
}

func (r *userRoleRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserRole, error) {
	var userRoles []*entity.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		Find(&userRoles, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	for i := range userRoles {
		if err := r.loadStatus(ctx, userRoles[i]); err != nil {
			return nil, err
		}
	}

	return userRoles, nil
}

func (r *userRoleRepository) Update(ctx context.Context, userRole *entity.UserRole) error {
	return r.db.WithContext(ctx).Save(userRole).Error
}

func (r *userRoleRepository) SetActive(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.UserRole{}).
			Where("user_id = ?", userID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.UserRole{}).
			Where("user_id = ? AND role_id = ?", userID, roleID).
			Update("is_active", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *userRoleRepository) Delete(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&entity.UserRole{}).Error
}

func (r *userRoleRepository) loadStatus(ctx context.Context, userRole *entity.UserRole) error {
	var statusCache entity.Status
	err := r.db.WithContext(ctx).
		Table("replicated_statuses").
		Where("id = ?", userRole.StatusID).
		First(&statusCache).Error

	if err != nil {
		return err
	}

	userRole.Status = statusCache
	return nil
}
