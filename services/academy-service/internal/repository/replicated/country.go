package replicated

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CountryRepository interface {
	Upsert(ctx context.Context, c entity.Country) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Country, error)
}

type countryRepository struct {
	db *gorm.DB
}

func NewCountryRepository(db *gorm.DB) CountryRepository {
	return &countryRepository{
		db: db,
	}
}

func (r *countryRepository) Upsert(ctx context.Context, c entity.Country) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "iso_alpha2", "iso_alpha3", "phone_code", "currency_code", "deleted_at"}),
		}).
		Create(&c).Error
}

func (r *countryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Country{}, "id = ?", id).Error
}

func (r *countryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Country, error) {
	var c entity.Country
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	return &c, err
}
