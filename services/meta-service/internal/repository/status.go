package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/shared/pkg/redisclient"
)

const (
	listByTypeTTL    = 5 * time.Minute
	listByTypePrefix = "status:list:"
)

func listByTypeKey(statusType string) string {
	return listByTypePrefix + statusType
}

type StatusRepository interface {
	Create(ctx context.Context, entity entity.Status) error
	GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error)
	List(ctx context.Context, statusType *string, page, pageSize int) ([]*entity.Status, int64, error)
	Update(ctx context.Context, entity entity.Status) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type statusRepository struct {
	db    *gorm.DB
	redis redisclient.Client
}

func NewStatusRepository(db *gorm.DB, redis redisclient.Client) StatusRepository {
	return &statusRepository{db: db, redis: redis}
}

func (r *statusRepository) Create(ctx context.Context, e entity.Status) error {
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return err
	}
	r.invalidateListCache(ctx, e.Type)
	return nil
}

func (r *statusRepository) GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error) {
	var status entity.Status
	err := r.db.WithContext(ctx).
		Where("type = ? AND name = ?", statusType, name).
		First(&status).Error
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

func (r *statusRepository) List(ctx context.Context, statusType *string, page, pageSize int) ([]*entity.Status, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Status{})

	typeKey := "all"
	if statusType != nil && *statusType != "" {
		typeKey = *statusType
		query = query.Where("type = ?", *statusType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	key := fmt.Sprintf("statuses:list:%s:p%d:s%d", typeKey, page, pageSize)

	cached, err := r.redis.Get(ctx, key)
	if err == nil && cached != "" {
		var items []*entity.Status
		if jsonErr := json.Unmarshal([]byte(cached), &items); jsonErr == nil {
			return items, total, nil
		}
	}

	var items []*entity.Status
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

func (r *statusRepository) Update(ctx context.Context, e entity.Status) error {
	if err := r.db.WithContext(ctx).Model(&e).Updates(e).Error; err != nil {
		return err
	}
	r.invalidateListCache(ctx, e.Type)
	return nil
}

func (r *statusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	statusType, err := r.getTypeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("hard delete: resolve type for cache invalidation: %w", err)
	}

	if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&entity.Status{}).Error; err != nil {
		return err
	}

	r.invalidateListCache(ctx, statusType)
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (r *statusRepository) invalidateListCache(ctx context.Context, statusType string) {
	_ = r.redis.Del(ctx, listByTypeKey(statusType))
}

func (r *statusRepository) getTypeByID(ctx context.Context, id uuid.UUID) (string, error) {
	var s entity.Status
	err := r.db.WithContext(ctx).
		Select("type").
		Where("id = ?", id).
		First(&s).Error
	return s.Type, err
}
