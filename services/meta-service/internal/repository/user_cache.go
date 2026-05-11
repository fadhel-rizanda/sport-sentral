package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"microservice-golang/services/meta-service/internal/entity"
)

type UserCacheRepository interface {
	Upsert(ctx context.Context, user entity.UserCache) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.UserCache, error)
}

type userCacheRepository struct {
	db *gorm.DB
}

func NewUserCacheRepository(db *gorm.DB) UserCacheRepository {
	return &userCacheRepository{
		db: db,
	}
}

func (r *userCacheRepository) Upsert(ctx context.Context, user entity.UserCache) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"email",
				"username",
				"full_name",
				"active_role_id",
				"active_role_name",
				"status_id",
				"deleted_at",
			}),
		}).
		Create(&user).Error
}

func (r *userCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.UserCache{}, "id = ?", id).Error
}

func (r *userCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.UserCache, error) {
	var user entity.UserCache
	err := r.db.WithContext(ctx).
		Table("user_caches").
		Select("user_caches.*, statuses.name as status_name").
		Joins("LEFT JOIN statuses ON statuses.id = user_caches.status_id").
		Where("user_caches.id = ? AND user_caches.deleted_at IS NULL", id).
		First(&user).Error
	return user, err
}
