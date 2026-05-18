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

type tagWithUsers struct {
	entity.Tag
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

func (r *tagRepository) GetByTypeAndName(ctx context.Context, tagType, name string) (*entity.Tag, error) {
	var row tagWithUsers

	err := r.db.WithContext(ctx).
		Table("tags").
		Select(`
			tags.*,
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
		Joins("LEFT JOIN user_caches cb ON cb.id = tags.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = tags.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = tags.deleted_by_id").
		Where("type = ? AND name = ?", tagType, name).
		First(&row).Error

	if err != nil {
		return nil, err
	}

	row.Tag.CreatedBy = &entity.UserCache{
		ID:       row.CreatedByID,
		Email:    row.CreatedByEmail,
		Username: row.CreatedByUsername,
		FullName: row.CreatedByFullName,
	}
	row.Tag.UpdatedBy = &entity.UserCache{
		ID:       row.UpdatedByID,
		Email:    row.UpdatedByEmail,
		Username: row.UpdatedByUsername,
		FullName: row.UpdatedByFullName,
	}
	if row.DeletedByID != nil && row.DeletedByEmail != "" {
		row.Tag.DeletedBy = &entity.UserCache{
			ID:       *row.DeletedByID,
			Email:    row.DeletedByEmail,
			Username: row.DeletedByUsername,
			FullName: row.DeletedByFullName,
		}
	}

	return &row.Tag, nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	var row tagWithUsers

	err := r.db.WithContext(ctx).
		Table("tags").
		Select(`
			tags.*,
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
		Joins("LEFT JOIN user_caches cb ON cb.id = tags.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = tags.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = tags.deleted_by_id").
		Where("tags.id = ?", id).
		First(&row).Error

	if err != nil {
		return nil, err
	}

	row.Tag.CreatedBy = &entity.UserCache{
		ID:       row.CreatedByID,
		Email:    row.CreatedByEmail,
		Username: row.CreatedByUsername,
		FullName: row.CreatedByFullName,
	}
	row.Tag.UpdatedBy = &entity.UserCache{
		ID:       row.UpdatedByID,
		Email:    row.UpdatedByEmail,
		Username: row.UpdatedByUsername,
		FullName: row.UpdatedByFullName,
	}
	if row.DeletedByID != nil && row.DeletedByEmail != "" {
		row.Tag.DeletedBy = &entity.UserCache{
			ID:       *row.DeletedByID,
			Email:    row.DeletedByEmail,
			Username: row.DeletedByUsername,
			FullName: row.DeletedByFullName,
		}
	}

	return &row.Tag, nil
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

	key := fmt.Sprintf("tags:list:%s:p%d:s%d", typeKey, page, pageSize)

	cached, err := r.redis.Get(ctx, key)
	if err == nil && cached != "" {
		var items []*entity.Tag
		if jsonErr := json.Unmarshal([]byte(cached), &items); jsonErr == nil {
			if loadErr := r.loadUserCaches(ctx, items); loadErr == nil {
				return items, total, nil
			}
		}
	}

	listQuery := r.db.WithContext(ctx).
		Table("tags").
		Select(`
			tags.*,
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
		Joins("LEFT JOIN user_caches cb ON cb.id = tags.created_by_id").
		Joins("LEFT JOIN user_caches ub ON ub.id = tags.updated_by_id").
		Joins("LEFT JOIN user_caches db ON db.id = tags.deleted_by_id")

	if tagType != nil && *tagType != "" {
		listQuery = listQuery.Where("tags.type = ?", *tagType)
	}

	var rows []tagWithUsers
	err = listQuery.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*entity.Tag, len(rows))
	for i := range rows {
		rows[i].Tag.CreatedBy = &entity.UserCache{
			ID:       rows[i].CreatedByID,
			Email:    rows[i].CreatedByEmail,
			Username: rows[i].CreatedByUsername,
			FullName: rows[i].CreatedByFullName,
		}
		rows[i].Tag.UpdatedBy = &entity.UserCache{
			ID:       rows[i].UpdatedByID,
			Email:    rows[i].UpdatedByEmail,
			Username: rows[i].UpdatedByUsername,
			FullName: rows[i].UpdatedByFullName,
		}
		if rows[i].DeletedByID != nil && rows[i].DeletedByEmail != "" {
			rows[i].Tag.DeletedBy = &entity.UserCache{
				ID:       *rows[i].DeletedByID,
				Email:    rows[i].DeletedByEmail,
				Username: rows[i].DeletedByUsername,
				FullName: rows[i].DeletedByFullName,
			}
		}
		items[i] = &rows[i].Tag
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

func (r *tagRepository) loadUserCaches(ctx context.Context, tags []*entity.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	userIDMap := make(map[uuid.UUID]bool)
	for _, tag := range tags {
		userIDMap[tag.CreatedByID] = true
		userIDMap[tag.UpdatedByID] = true
		if tag.DeletedByID != nil {
			userIDMap[*tag.DeletedByID] = true
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

	for _, tag := range tags {
		if user, ok := userMap[tag.CreatedByID]; ok {
			tag.CreatedBy = user
		}
		if user, ok := userMap[tag.UpdatedByID]; ok {
			tag.UpdatedBy = user
		}
		if tag.DeletedByID != nil {
			if user, ok := userMap[*tag.DeletedByID]; ok {
				tag.DeletedBy = user
			}
		}
	}

	return nil
}
