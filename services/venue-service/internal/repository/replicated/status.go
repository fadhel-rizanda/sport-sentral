package replicated

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatusRepository interface {
	Upsert(ctx context.Context, s entity.Status) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error)
	GetByTypeAndName(ctx context.Context, t string, name string) (*entity.Status, error)
	GetByTypeAndSlug(ctx context.Context, t string, slug string) (*entity.Status, error)
}

type statusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{
		db: db,
	}
}

func (r *statusRepository) Upsert(ctx context.Context, s entity.Status) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"type", "name", "slug", "deleted_at"}),
		}).
		Create(&s).Error
}

func (r *statusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Status{}, "id = ?", id).Error
}

func (r *statusRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
	var s entity.Status
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *statusRepository) GetByTypeAndName(ctx context.Context, t string, name string) (*entity.Status, error) {
	var s entity.Status
	err := r.db.WithContext(ctx).First(&s, "type = ? AND name = ?", t, name).Error
	return &s, err
}

func (r *statusRepository) GetByTypeAndSlug(ctx context.Context, t string, slug string) (*entity.Status, error) {
	var s entity.Status
	err := r.db.WithContext(ctx).First(&s, "type = ? AND slug = ?", t, slug).Error
	return &s, err
}
