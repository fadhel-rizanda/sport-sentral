package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
	"time"
)

type StatusRepository interface {
	Create(ctx context.Context, entity entity.Status) error
	GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error)
	ListByType(ctx context.Context, statusType string) ([]*entity.Status, error)
	Update(ctx context.Context, entity entity.Status) error
	SoftDelete(ctx context.Context, id, deletedBy uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
}

type statusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{db: db}
}

func (r *statusRepository) Create(ctx context.Context, entity entity.Status) error {
	return r.db.WithContext(ctx).Create(&entity).Error
}

func (r *statusRepository) GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error) {
	var status entity.Status
	err := r.db.WithContext(ctx).Where("type = ? AND name = ?", statusType, name).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
	var status entity.Status
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&status).Error
	return &status, err
}

func (r *statusRepository) ListByType(ctx context.Context, statusType string) ([]*entity.Status, error) {
	var items []*entity.Status
	err := r.db.WithContext(ctx).Where("type = ?", statusType).Find(&items).Error
	return items, err
}

func (r *statusRepository) Update(ctx context.Context, entity entity.Status) error {
	return r.db.WithContext(ctx).Model(&entity).Updates(entity).Error
}

func (r *statusRepository) SoftDelete(ctx context.Context, id, deletedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Status{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error
}

func (r *statusRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&entity.Status{}).Error
}
