package replicated

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AcademyHoldingRepository interface {
	Upsert(ctx context.Context, holding entity.AcademyHolding) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error)
}

type academyHoldingRepository struct {
	db *gorm.DB
}

func NewAcademyHoldingRepository(db *gorm.DB) AcademyHoldingRepository {
	return &academyHoldingRepository{
		db: db,
	}
}

func (r *academyHoldingRepository) Upsert(ctx context.Context, holding entity.AcademyHolding) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"description",
				"email",
				"phone_number",
				"image_attachment_id",
				"status_id",
				"updated_at",
				"deleted_at",
			}),
		}).
		Create(&holding).Error
}

func (r *academyHoldingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AcademyHolding{}, "id = ?", id).Error
}

func (r *academyHoldingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error) {
	var holding entity.AcademyHolding
	err := r.db.WithContext(ctx).First(&holding, "id = ?", id).Error
	return &holding, err
}
