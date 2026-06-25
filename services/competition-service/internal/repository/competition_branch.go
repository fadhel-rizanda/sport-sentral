package repository

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompetitionBranchRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CompetitionBranch, error)
	List(ctx context.Context, filters CompetitionBranchFilters, page, pageSize int) ([]*entity.CompetitionBranch, int64, error)
	Create(ctx context.Context, branch *entity.CompetitionBranch) error
	Update(ctx context.Context, branch *entity.CompetitionBranch) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CompetitionBranchFilters struct {
	CompetitionID  *uuid.UUID
	ParentBranchID *uuid.UUID
	StatusID       *uuid.UUID
	Search         *string
}

type competitionBranchRepository struct {
	db *gorm.DB
}

func NewCompetitionBranchRepository(db *gorm.DB) CompetitionBranchRepository {
	return &competitionBranchRepository{
		db: db,
	}
}

func (r *competitionBranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CompetitionBranch, error) {
	var branch entity.CompetitionBranch
	err := r.db.WithContext(ctx).
		Preload("Competition").
		Preload("ParentBranch").
		Preload("Status").
		Preload("ChildBranches").
		Preload("Matches").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&branch, "id = ?", id).Error
	return &branch, err
}

func (r *competitionBranchRepository) List(ctx context.Context, filters CompetitionBranchFilters, page, pageSize int) ([]*entity.CompetitionBranch, int64, error) {
	var branches []*entity.CompetitionBranch
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.CompetitionBranch{})

	if filters.CompetitionID != nil {
		query = query.Where("competition_id = ?", filters.CompetitionID)
	}
	if filters.ParentBranchID != nil {
		query = query.Where("parent_branch_id = ?", filters.ParentBranchID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}
	if filters.Search != nil && *filters.Search != "" {
		if _, err := uuid.Parse(*filters.Search); err == nil {
			query = query.Where("id = ? OR name ILIKE ?", *filters.Search, "%"+*filters.Search+"%")
		} else {
			query = query.Where("name ILIKE ?", "%"+*filters.Search+"%")
		}
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Competition").
		Preload("Status").
		Preload("CreatedBy").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&branches).Error

	return branches, total, err
}

func (r *competitionBranchRepository) Create(ctx context.Context, branch *entity.CompetitionBranch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *competitionBranchRepository) Update(ctx context.Context, branch *entity.CompetitionBranch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *competitionBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.CompetitionBranch{}, "id = ?", id).Error
}
