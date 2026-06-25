package replicated

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AcademyAdminRepository interface {
	Upsert(ctx context.Context, admin entity.AcademyAdmin) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyAdmin, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AcademyAdmin, error)
}

type academyAdminRepository struct {
	db *gorm.DB
}

func NewAcademyAdminRepository(db *gorm.DB) AcademyAdminRepository {
	return &academyAdminRepository{
		db: db,
	}
}

func (r *academyAdminRepository) Upsert(ctx context.Context, admin entity.AcademyAdmin) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"academy_id", "branch_id", "user_id", "role_id", "deleted_at"}),
		}).
		Create(&admin).Error
}

func (r *academyAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AcademyAdmin{}, "id = ?", id).Error
}

func (r *academyAdminRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyAdmin, error) {
	var admin entity.AcademyAdmin
	err := r.db.WithContext(ctx).
		Preload("Academy").
		Preload("Branch").
		Preload("User").
		Preload("Role").
		First(&admin, "id = ?", id).Error
	return &admin, err
}

func (r *academyAdminRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AcademyAdmin, error) {
	var admin entity.AcademyAdmin
	err := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("User").
		Preload("Role").
		Preload("Academy").
		First(&admin, "user_id = ?", userID).Error
	return &admin, err
}
