package repository

import (
	"context"
	"microservice-golang/services/sport-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SportRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error)
	List(ctx context.Context, filters SportFilters, page, pageSize int) ([]*entity.Sport, int64, error)
	Create(ctx context.Context, sport *entity.Sport) error
	Update(ctx context.Context, sport *entity.Sport) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Config
	GetConfigBySportID(ctx context.Context, sportID uuid.UUID) (*entity.SportConfig, error)
	UpsertConfig(ctx context.Context, config *entity.SportConfig) error
}

type SportFilters struct {
	TierTagID *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
}

type sportRepository struct {
	db *gorm.DB
}

func NewSportRepository(db *gorm.DB) SportRepository {
	return &sportRepository{
		db: db,
	}
}

func (r *sportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error) {
	var sport entity.Sport
	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("TierTag").
		Preload("Config.ParticipantTypeTag").
		Preload("Config.Stats.StatTypeTag").
		Preload("ActiveRegulator").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&sport, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sport, nil
}

func (r *sportRepository) List(ctx context.Context, filters SportFilters, page, pageSize int) ([]*entity.Sport, int64, error) {
	var sports []*entity.Sport
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&entity.Sport{})

	if filters.TierTagID != nil {
		query = query.Where("tier_tag_id = ?", filters.TierTagID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}
	if filters.Search != nil && *filters.Search != "" {
		search := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", search, search)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Status").
		Preload("TierTag").
		Preload("Config.ParticipantTypeTag").
		Preload("Config.Stats.StatTypeTag").
		Preload("ActiveRegulator").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&sports).Error

	return sports, total, err
}

func (r *sportRepository) Create(ctx context.Context, sport *entity.Sport) error {
	return r.db.WithContext(ctx).Create(sport).Error
}

func (r *sportRepository) Update(ctx context.Context, sport *entity.Sport) error {
	return r.db.WithContext(ctx).Save(sport).Error
}

func (r *sportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Sport{}, "id = ?", id).Error
}

func (r *sportRepository) GetConfigBySportID(ctx context.Context, sportID uuid.UUID) (*entity.SportConfig, error) {
	var config entity.SportConfig
	err := r.db.WithContext(ctx).
		Preload("ParticipantTypeTag").
		Preload("Stats.StatTypeTag").
		First(&config, "sport_id = ?", sportID).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *sportRepository) UpsertConfig(ctx context.Context, config *entity.SportConfig) error {
	if err := r.db.WithContext(ctx).Omit("Stats", "ParticipantTypeTag", "Sport").Save(config).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(config).Association("Stats").Replace(config.Stats)
}
