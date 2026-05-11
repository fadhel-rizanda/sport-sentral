package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/shared/pkg/redisclient"
)

type TagRepository interface {
	Create(ctx context.Context, entity *entity.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetByTypeAndName(ctx context.Context, tagType, name string) (*entity.Tag, error)
	List(ctx context.Context, tagType *string, page, pageSize int) ([]*entity.Tag, int64, error)
	Update(ctx context.Context, e entity.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type tagRepository struct {
	db    *gorm.DB
	redis redisclient.Client
}

func NewTagRepository(db *gorm.DB, redis redisclient.Client) TagRepository {
	return &tagRepository{
		db:    db,
		redis: redis,
	}
}

func (r *tagRepository) Create(ctx context.Context, entity *entity.Tag) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	r.invalidateListCache(ctx, entity.Type)
	return nil
}

func (r *tagRepository) GetByTypeAndName(ctx context.Context, tagType, name string) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).
		Where("type = ? AND name = ?", tagType, name).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	return &tag, err
}

func (r *tagRepository) List(ctx context.Context, tagType *string, page, pageSize int) ([]*entity.Tag, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Tag{})

	typeKey := "all"
	if tagType != nil && *tagType != "" {
		typeKey = *tagType
		query = query.Where("type = ?", *tagType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	key := fmt.Sprintf("tages:list:%s:p%d:s%d", typeKey, page, pageSize)

	cached, err := r.redis.Get(ctx, key)
	if err == nil && cached != "" {
		var items []*entity.Tag
		if jsonErr := json.Unmarshal([]byte(cached), &items); jsonErr == nil {
			return items, total, nil
		}
	}

	var items []*entity.Tag
	err = query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	if data, jsonErr := json.Marshal(items); jsonErr == nil {
		_ = r.redis.Set(ctx, key, string(data), listByTypeTTL)
	}

	return items, total, nil
}

func (r *tagRepository) Update(ctx context.Context, e entity.Tag) error {
	if err := r.db.WithContext(ctx).Model(&e).Updates(e).Error; err != nil {
		return err
	}
	r.invalidateListCache(ctx, e.Type)
	return nil
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tagType, err := r.getTypeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("hard delete: resolve type for cache invalidation: %w", err)
	}

	if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&entity.Tag{}).Error; err != nil {
		return err
	}

	r.invalidateListCache(ctx, tagType)
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (r *tagRepository) invalidateListCache(ctx context.Context, tagType string) {
	_ = r.redis.Del(ctx, listByTypeKey(tagType))
}

func (r *tagRepository) getTypeByID(ctx context.Context, id uuid.UUID) (string, error) {
	var s entity.Tag
	err := r.db.WithContext(ctx).
		Select("type").
		Where("id = ?", id).
		First(&s).Error
	return s.Type, err
}
