package replicated

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SportRepository interface {
	Upsert(ctx context.Context, s entity.Sport) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error)
}

type sportRepository struct {
	db *gorm.DB
}

func NewSportRepository(db *gorm.DB) SportRepository {
	return &sportRepository{
		db: db,
	}
}

func (r *sportRepository) Upsert(ctx context.Context, s entity.Sport) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "slug", "icon_attachment_id", "is_verified", "regulator_id", "tier"}),
		}).
		Create(&s).Error
}

func (r *sportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Sport{}, "id = ?", id).Error
}

func (r *sportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error) {
	var s entity.Sport
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}
