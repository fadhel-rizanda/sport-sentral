package replicated

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AcademyBranchRepository interface {
	Upsert(ctx context.Context, branch entity.AcademyBranch) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error)
}

type academyBranchRepository struct {
	db *gorm.DB
}

func NewAcademyBranchRepository(db *gorm.DB) AcademyBranchRepository {
	return &academyBranchRepository{
		db: db,
	}
}

func (r *academyBranchRepository) Upsert(ctx context.Context, branch entity.AcademyBranch) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"holding_id",
				"sport_id",
				"name",
				"status_id",
				"updated_at",
				"deleted_at",
			}),
		}).
		Create(&branch).Error
}

func (r *academyBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AcademyBranch{}, "id = ?", id).Error
}

func (r *academyBranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error) {
	var branch entity.AcademyBranch
	err := r.db.WithContext(ctx).First(&branch, "id = ?", id).Error
	return &branch, err
}
