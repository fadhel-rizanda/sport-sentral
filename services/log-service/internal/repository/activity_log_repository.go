package repository

import (
	"context"
	"errors"
	"microservice-golang/services/log-service/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityLogFilter struct {
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	ResourceID   *string
	StartTime    *time.Time
	EndTime      *time.Time
	Page         int
	Limit        int
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *entity.ActivityLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error)
	List(ctx context.Context, filter ActivityLogFilter) ([]entity.ActivityLog, int64, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(ctx context.Context, log *entity.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *activityLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
	var log entity.ActivityLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *activityLogRepository) List(ctx context.Context, filter ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.ActivityLog{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil && *filter.Action != "" {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.ResourceType != nil && *filter.ResourceType != "" {
		query = query.Where("resource_type = ?", *filter.ResourceType)
	}
	if filter.ResourceID != nil && *filter.ResourceID != "" {
		query = query.Where("resource_id = ?", *filter.ResourceID)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var logs []entity.ActivityLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
