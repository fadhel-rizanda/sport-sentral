package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"microservice-golang/services/identity-service/internal/entity"
)

type StatusCacheRepository interface {
	Upsert(ctx context.Context, s entity.StatusCache) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.StatusCache, error)
	GetByTypeAndName(ctx context.Context, t string, name string) (*entity.StatusCache, error)
}

type statusCacheRepository struct {
	db *gorm.DB
}

func NewStatusCacheRepository(db *gorm.DB) StatusCacheRepository {
	return &statusCacheRepository{
		db: db,
	}
}

func (r *statusCacheRepository) Upsert(ctx context.Context, s entity.StatusCache) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"type", "name", "slug", "deleted_at"}),
		}).
		Create(&s).Error
}

func (r *statusCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.StatusCache{}, "id = ?", id).Error
}

func (r *statusCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.StatusCache, error) {
	var s entity.StatusCache
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *statusCacheRepository) GetByTypeAndName(ctx context.Context, t string, name string) (*entity.StatusCache, error) {
	var s entity.StatusCache
	err := r.db.WithContext(ctx).First(&s, "type = ? AND name = ?", t, name).Error
	return &s, err
}
