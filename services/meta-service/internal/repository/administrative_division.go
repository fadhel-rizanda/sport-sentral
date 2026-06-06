package repository

import (
	"context"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/shared/pkg/redisclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdministrativeDivisionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AdministrativeDivision, error)
	List(ctx context.Context, filters AdministrativeDivisionFilters, page, pageSize int) ([]*entity.AdministrativeDivision, int64, error)
	Create(ctx context.Context, entity *entity.AdministrativeDivision) error
	Update(ctx context.Context, entity *entity.AdministrativeDivision) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdministrativeDivisionFilters struct {
	Name       *string
	Level      *string
	PostalCode *string
	CountryID  *uuid.UUID
	ParentID   *uuid.UUID
}

type administrativeDivisionRepository struct {
	db    *gorm.DB
	redis redisclient.Client
}

func NewAdministrativeDivisionRepository(db *gorm.DB, redis redisclient.Client) AdministrativeDivisionRepository {
	return &administrativeDivisionRepository{
		db:    db,
		redis: redis,
	}
}

func (r *administrativeDivisionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AdministrativeDivision, error) {
	var administrativeDivision entity.AdministrativeDivision
	err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Parent").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		First(&administrativeDivision, "id = ?", id).Error
	return &administrativeDivision, err
}

func (r *administrativeDivisionRepository) List(ctx context.Context, filters AdministrativeDivisionFilters, page, pageSize int) ([]*entity.AdministrativeDivision, int64, error) {
	var administrativeDivisions []*entity.AdministrativeDivision
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.AdministrativeDivision{})

	if filters.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filters.Name+"%")
	}
	if filters.Level != nil {
		query = query.Where("level = ?", *filters.Level)
	}
	if filters.PostalCode != nil {
		query = query.Where("postal_code = ?", *filters.PostalCode)
	}
	if filters.CountryID != nil {
		query = query.Where("country_id = ?", *filters.CountryID)
	}
	if filters.ParentID != nil {
		query = query.Where("parent_id = ?", *filters.ParentID)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Country").
		Preload("Parent").
		Preload("Parent.Country").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&administrativeDivisions).Error

	return administrativeDivisions, total, err
}

func (r *administrativeDivisionRepository) Create(ctx context.Context, entity *entity.AdministrativeDivision) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *administrativeDivisionRepository) Update(ctx context.Context, entity *entity.AdministrativeDivision) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *administrativeDivisionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.AdministrativeDivision{}, "id = ?", id).Error
}
