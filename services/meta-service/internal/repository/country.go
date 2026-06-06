package repository

import (
	"context"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/shared/pkg/redisclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CountryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Country, error)
	List(ctx context.Context, filters CountryFilters, page, pageSize int) ([]*entity.Country, int64, error)
	Create(ctx context.Context, entity *entity.Country) error
	Update(ctx context.Context, country *entity.Country) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CountryFilters struct {
	Name      *string
	ISOAlpha2 *string
	ISOAlpha3 *string
}

type countryRepository struct {
	db    *gorm.DB
	redis redisclient.Client
}

func NewCountryRepository(db *gorm.DB, redis redisclient.Client) CountryRepository {
	return &countryRepository{
		db:    db,
		redis: redis,
	}
}

func (r *countryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Country, error) {
	var country entity.Country
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		First(&country, "id = ?", id).Error
	return &country, err
}

func (r *countryRepository) List(ctx context.Context, filters CountryFilters, page, pageSize int) ([]*entity.Country, int64, error) {
	var countries []*entity.Country
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&entity.Country{})

	if filters.Name != nil {
		query = query.Where("name ILIKE ?", "%"+*filters.Name+"%")
	}
	if filters.ISOAlpha2 != nil {
		query = query.Where("iso_alpha2 = ?", *filters.ISOAlpha2)
	}
	if filters.ISOAlpha3 != nil {
		query = query.Where("iso_alpha3 = ?", *filters.ISOAlpha3)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		Offset(offset).
		Limit(pageSize).
		Find(&countries).Error

	return countries, total, err
}

func (r *countryRepository) Create(ctx context.Context, entity *entity.Country) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *countryRepository) Update(ctx context.Context, country *entity.Country) error {
	return r.db.WithContext(ctx).Save(&country).Error
}

func (r *countryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Country{}, "id = ?", id).Error
}
