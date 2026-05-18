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

type statusWithUsers struct {
	entity.Status
	CreatedByEmail    string `gorm:"column:created_by_email"`
	CreatedByUsername string `gorm:"column:created_by_username"`
	CreatedByFullName string `gorm:"column:created_by_full_name"`
	UpdatedByEmail    string `gorm:"column:updated_by_email"`
	UpdatedByUsername string `gorm:"column:updated_by_username"`
	UpdatedByFullName string `gorm:"column:updated_by_full_name"`
	DeletedByEmail    string `gorm:"column:deleted_by_email"`
	DeletedByUsername string `gorm:"column:deleted_by_username"`
	DeletedByFullName string `gorm:"column:deleted_by_full_name"`
}

func (r *statusRepository) GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error) {
	var row statusWithUsers

	err := r.db.WithContext(ctx).
		Table("statuses").
		Select(`
			statuses.*,
			cb.email AS created_by_email,
			cb.username AS created_by_username,
			cb.full_name AS created_by_full_name,
			ub.email AS updated_by_email,
			ub.username AS updated_by_username,
			ub.full_name AS updated_by_full_name,
			db.email AS deleted_by_email,
			db.username AS deleted_by_username,
			db.full_name AS deleted_by_full_name
		`).
		Joins("LEFT JOIN user_caches cb ON cb.id = statuses.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = statuses.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = statuses.deleted_by_id").
		Where("type = ? AND name = ?", statusType, name).
		First(&row).Error

	if err != nil {
		return nil, err
	}

	row.Status.CreatedBy = &entity.UserCache{
		ID:       row.CreatedByID,
		Email:    row.CreatedByEmail,
		Username: row.CreatedByUsername,
		FullName: row.CreatedByFullName,
	}
	row.Status.UpdatedBy = &entity.UserCache{
		ID:       row.UpdatedByID,
		Email:    row.UpdatedByEmail,
		Username: row.UpdatedByUsername,
		FullName: row.UpdatedByFullName,
	}
	if row.DeletedByID != nil && row.DeletedByEmail != "" {
		row.Status.DeletedBy = &entity.UserCache{
			ID:       *row.DeletedByID,
			Email:    row.DeletedByEmail,
			Username: row.DeletedByUsername,
			FullName: row.DeletedByFullName,
		}
	}

	return &row.Status, nil
}

func (r *statusRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
	var row statusWithUsers

	err := r.db.WithContext(ctx).
		Table("statuses").
		Select(`
			statuses.*,
			cb.email AS created_by_email,
			cb.username AS created_by_username,
			cb.full_name AS created_by_full_name,
			ub.email AS updated_by_email,
			ub.username AS updated_by_username,
			ub.full_name AS updated_by_full_name,
			db.email AS deleted_by_email,
			db.username AS deleted_by_username,
			db.full_name AS deleted_by_full_name
		`).
		Joins("LEFT JOIN user_caches cb ON cb.id = statuses.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = statuses.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = statuses.deleted_by_id").
		Where("statuses.id = ?", id).
		First(&row).Error

	if err != nil {
		return nil, err
	}

	row.Status.CreatedBy = &entity.UserCache{
		ID:       row.CreatedByID,
		Email:    row.CreatedByEmail,
		Username: row.CreatedByUsername,
		FullName: row.CreatedByFullName,
	}
	row.Status.UpdatedBy = &entity.UserCache{
		ID:       row.UpdatedByID,
		Email:    row.UpdatedByEmail,
		Username: row.UpdatedByUsername,
		FullName: row.UpdatedByFullName,
	}
	if row.DeletedByID != nil && row.DeletedByEmail != "" {
		row.Status.DeletedBy = &entity.UserCache{
			ID:       *row.DeletedByID,
			Email:    row.DeletedByEmail,
			Username: row.DeletedByUsername,
			FullName: row.DeletedByFullName,
		}
	}

	return &row.Status, nil
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
			if loadErr := r.loadUserCaches(ctx, items); loadErr == nil {
				return items, total, nil
			}
		}
	}

	listQuery := r.db.WithContext(ctx).
		Table("statuses").
		Select(`
          statuses.*,
          cb.email AS created_by_email,
          cb.username AS created_by_username,
          cb.full_name AS created_by_full_name,
          ub.email AS updated_by_email,
          ub.username AS updated_by_username,
          ub.full_name AS updated_by_full_name,
          db.email AS deleted_by_email,
          db.username AS deleted_by_username,
          db.full_name AS deleted_by_full_name
       `).
		Joins("LEFT JOIN user_caches cb ON cb.id = statuses.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = statuses.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = statuses.deleted_by_id")

	if statusType != nil && *statusType != "" {
		listQuery = listQuery.Where("statuses.type = ?", *statusType)
	}

	var rows []statusWithUsers
	err = listQuery.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*entity.Status, len(rows))
	for i := range rows {
		rows[i].Status.CreatedBy = &entity.UserCache{
			ID:       rows[i].CreatedByID,
			Email:    rows[i].CreatedByEmail,
			Username: rows[i].CreatedByUsername,
			FullName: rows[i].CreatedByFullName,
		}
		rows[i].Status.UpdatedBy = &entity.UserCache{
			ID:       rows[i].UpdatedByID,
			Email:    rows[i].UpdatedByEmail,
			Username: rows[i].UpdatedByUsername,
			FullName: rows[i].UpdatedByFullName,
		}
		if rows[i].DeletedByID != nil && rows[i].DeletedByEmail != "" {
			rows[i].Status.DeletedBy = &entity.UserCache{
				ID:       *rows[i].DeletedByID,
				Email:    rows[i].DeletedByEmail,
				Username: rows[i].DeletedByUsername,
				FullName: rows[i].DeletedByFullName,
			}
		}
		items[i] = &rows[i].Status
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

func (r *statusRepository) loadUserCaches(ctx context.Context, statuses []*entity.Status) error {
	if len(statuses) == 0 {
		return nil
	}

	userIDMap := make(map[uuid.UUID]bool)
	for _, status := range statuses {
		userIDMap[status.CreatedByID] = true
		userIDMap[status.UpdatedByID] = true
		if status.DeletedByID != nil {
			userIDMap[*status.DeletedByID] = true
		}
	}

	userIDs := make([]uuid.UUID, 0, len(userIDMap))
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		return nil
	}

	var users []entity.UserCache
	if err := r.db.WithContext(ctx).
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return err
	}

	userMap := make(map[uuid.UUID]*entity.UserCache)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	for _, status := range statuses {
		if user, ok := userMap[status.CreatedByID]; ok {
			status.CreatedBy = user
		}
		if user, ok := userMap[status.UpdatedByID]; ok {
			status.UpdatedBy = user
		}
		if status.DeletedByID != nil {
			if user, ok := userMap[*status.DeletedByID]; ok {
				status.DeletedBy = user
			}
		}
	}

	return nil
}
