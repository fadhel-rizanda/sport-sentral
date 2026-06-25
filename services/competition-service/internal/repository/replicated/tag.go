package replicated

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagRepository interface {
	Upsert(ctx context.Context, tag entity.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{
		db: db,
	}
}

func (r *tagRepository) Upsert(ctx context.Context, tag entity.Tag) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"type",
				"name",
				"slug",
				"deleted_at",
			}),
		}).
		Create(&tag).Error
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Tag{}, "id = ?", id).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error
	return &tag, err
}
