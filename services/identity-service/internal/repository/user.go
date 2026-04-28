package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"microservice-golang/services/identity-service/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("UserRoles.Status").
		Preload("UserRoles.Role").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("UserRoles.Status").
		Preload("UserRoles.Role").
		First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("UserRoles.Status").
		Preload("UserRoles.Role").
		First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("UserRoles.Status").
		Preload("UserRoles.Role").
		Offset(offset).Limit(pageSize).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
