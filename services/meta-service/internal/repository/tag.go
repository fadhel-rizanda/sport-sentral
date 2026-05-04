package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
	"time"
)

type TagRepository interface {
	Create(ctx context.Context, entity *entity.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Tag, error)
	ListByType(ctx context.Context, statusType string) ([]*entity.Tag, error)
	Update(ctx context.Context, entity entity.Tag) error
	SoftDelete(ctx context.Context, id, deletedBy uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, entity *entity.Tag) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error
	return &tag, err
}

func (r *tagRepository) GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).First(&tag, "type = ? AND name = ?", statusType, name).Error
	return &tag, err
}

func (r *tagRepository) ListByType(ctx context.Context, statusType string) ([]*entity.Tag, error) {
	var tags []*entity.Tag
	err := r.db.WithContext(ctx).Find(&tags, "type = ?", statusType).Error
	return tags, err
}

func (r *tagRepository) Update(ctx context.Context, entity entity.Tag) error {
	return r.db.WithContext(ctx).Model(&entity).Updates(entity).Error
}

func (r *tagRepository) SoftDelete(ctx context.Context, id, deletedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Tag{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_by": deletedBy,
			"deleted_at": gorm.DeletedAt{Time: time.Now(), Valid: true},
		}).Error
}

func (r *tagRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&entity.Tag{}).Error
}
