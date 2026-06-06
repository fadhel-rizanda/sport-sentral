package replicated

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdministrativeDivisionRepository interface {
	Upsert(ctx context.Context, ad entity.AdministrativeDivision) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AdministrativeDivision, error)
}

type administrativeDivisionRepository struct {
	db *gorm.DB
}

func NewAdministrativeDivisionRepository(db *gorm.DB) AdministrativeDivisionRepository {
	return &administrativeDivisionRepository{
		db: db,
	}
}

func (r *administrativeDivisionRepository) Upsert(ctx context.Context, ad entity.AdministrativeDivision) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"country_id", "parent_id", "name", "level", "postal_code", "deleted_at"}),
		}).
		Create(&ad).Error
}

func (r *administrativeDivisionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AdministrativeDivision{}, "id = ?", id).Error
}

func (r *administrativeDivisionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AdministrativeDivision, error) {
	var ad entity.AdministrativeDivision
	err := r.db.WithContext(ctx).Preload("Country").Preload("Parent").First(&ad, "id = ?", id).Error
	return &ad, err
}
