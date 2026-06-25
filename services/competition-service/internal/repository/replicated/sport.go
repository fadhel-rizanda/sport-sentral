package replicated

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SportRepository interface {
	Upsert(ctx context.Context, sport entity.Sport) error
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

func (r *sportRepository) Upsert(ctx context.Context, sport entity.Sport) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"slug",
				"icon_attachment_id",
				"is_verified",
				"regulator_id",
				"tier",
			}),
		}).Omit("Stats").Create(&sport).Error; err != nil {
			return err
		}

		if err := tx.Where("sport_id = ?", sport.ID).Delete(&entity.SportStat{}).Error; err != nil {
			return err
		}

		if len(sport.Stats) > 0 {
			if err := tx.Create(&sport.Stats).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *sportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Sport{}, "id = ?", id).Error
}

func (r *sportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error) {
	var sport entity.Sport
	err := r.db.WithContext(ctx).Preload("Stats").First(&sport, "id = ?", id).Error
	return &sport, err
}
