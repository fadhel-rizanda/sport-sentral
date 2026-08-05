package repository

import (
	"context"
	"errors"
	"microservice-golang/services/log-service/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogFilter struct {
	ServiceName *string
	Module      *string
	EntityType  *string
	EntityID    *string
	Action      *string
	UserID      *uuid.UUID
	Status      *string
	Severity    *string
	SearchQuery *string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	Limit       int
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)
	List(ctx context.Context, filter AuditLogFilter) ([]entity.AuditLog, int64, error)
	GetStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	var log entity.AuditLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *auditLogRepository) List(ctx context.Context, filter AuditLogFilter) ([]entity.AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.AuditLog{})

	if filter.ServiceName != nil && *filter.ServiceName != "" {
		query = query.Where("service_name = ?", *filter.ServiceName)
	}
	if filter.Module != nil && *filter.Module != "" {
		query = query.Where("module = ?", *filter.Module)
	}
	if filter.EntityType != nil && *filter.EntityType != "" {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}
	if filter.EntityID != nil && *filter.EntityID != "" {
		query = query.Where("entity_id = ?", *filter.EntityID)
	}
	if filter.Action != nil && *filter.Action != "" {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Severity != nil && *filter.Severity != "" {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		q := "%" + *filter.SearchQuery + "%"
		query = query.Where("action ILIKE ? OR module ILIKE ? OR entity_type ILIKE ? OR user_email ILIKE ?", q, q, q, q)
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

	var logs []entity.AuditLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *auditLogRepository) GetStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
	stats := &entity.LogStats{}

	// Total audit logs count
	auditQuery := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	if serviceName != nil && *serviceName != "" {
		auditQuery = auditQuery.Where("service_name = ?", *serviceName)
	}
	if startTime != nil {
		auditQuery = auditQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		auditQuery = auditQuery.Where("created_at <= ?", *endTime)
	}
	if err := auditQuery.Count(&stats.TotalAuditLogs).Error; err != nil {
		return nil, err
	}

	// Total activity logs count
	activityQuery := r.db.WithContext(ctx).Model(&entity.ActivityLog{})
	if startTime != nil {
		activityQuery = activityQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		activityQuery = activityQuery.Where("created_at <= ?", *endTime)
	}
	if err := activityQuery.Count(&stats.TotalActivityLogs).Error; err != nil {
		return nil, err
	}

	// Severity breakdown
	var severityCounts []entity.SeverityCount
	sevQuery := r.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Select("severity, COUNT(*) as count")
	if serviceName != nil && *serviceName != "" {
		sevQuery = sevQuery.Where("service_name = ?", *serviceName)
	}
	if startTime != nil {
		sevQuery = sevQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		sevQuery = sevQuery.Where("created_at <= ?", *endTime)
	}
	if err := sevQuery.Group("severity").Scan(&severityCounts).Error; err != nil {
		return nil, err
	}
	stats.CountBySeverity = severityCounts

	// Action breakdown
	var actionCounts []entity.ActionCount
	actQuery := r.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Select("action, COUNT(*) as count")
	if serviceName != nil && *serviceName != "" {
		actQuery = actQuery.Where("service_name = ?", *serviceName)
	}
	if startTime != nil {
		actQuery = actQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		actQuery = actQuery.Where("created_at <= ?", *endTime)
	}
	if err := actQuery.Group("action").Order("count DESC").Limit(10).Scan(&actionCounts).Error; err != nil {
		return nil, err
	}
	stats.CountByAction = actionCounts

	// Service breakdown
	var serviceCounts []entity.ServiceCount
	srvQuery := r.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Select("service_name, COUNT(*) as count")
	if startTime != nil {
		srvQuery = srvQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		srvQuery = srvQuery.Where("created_at <= ?", *endTime)
	}
	if err := srvQuery.Group("service_name").Scan(&serviceCounts).Error; err != nil {
		return nil, err
	}
	stats.CountByService = serviceCounts

	return stats, nil
}
