package repository

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademyBranchRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error)
	List(ctx context.Context, filters AcademyBranchFilters, page, pageSize int) ([]*entity.AcademyBranch, int64, error)
	Create(ctx context.Context, branch *entity.AcademyBranch) error
	Update(ctx context.Context, branch *entity.AcademyBranch) error
	Delete(ctx context.Context, id uuid.UUID) error
	//GetPrimaryAddress(ctx context.Context, branchID uuid.UUID) (*entity.AcademyBranchAddress, error)
	//GetAllAddresses(ctx context.Context, branchID uuid.UUID) ([]*entity.AcademyBranchAddress, error)
}

type AcademyBranchFilters struct {
	HoldingID *uuid.UUID
	SportID   *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
}

type academyBranchRepository struct {
	db *gorm.DB
}

func NewAcademyBranchRepository(db *gorm.DB) AcademyBranchRepository {
	return &academyBranchRepository{
		db: db,
	}
}

func (r *academyBranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error) {
	var branch entity.AcademyBranch
	err := r.db.WithContext(ctx).
		Preload("Holding").
		Preload("Sport").
		Preload("Status").
		Preload("Addresses.AdministrativeDivision.Country").
		Preload("Addresses.AdministrativeDivision.Parent").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&branch, "id = ?", id).Error
	return &branch, err
}

func (r *academyBranchRepository) List(ctx context.Context, filters AcademyBranchFilters, page, pageSize int) ([]*entity.AcademyBranch, int64, error) {
	var branches []*entity.AcademyBranch
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")

	// Apply filters
	if filters.HoldingID != nil {
		query = query.Where("holding_id = ?", *filters.HoldingID)
	}
	if filters.SportID != nil {
		query = query.Where("sport_id = ?", *filters.SportID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", *filters.StatusID)
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where("name ILIKE ?", "%"+*filters.Search+"%")
	}

	// Count total
	if err := query.Session(&gorm.Session{}).Model(&entity.AcademyBranch{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := query.
		Preload("Holding").
		Preload("Sport").
		Preload("Status").
		Preload("Addresses", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true).
				Preload("AdministrativeDivision.Country")
		}).
		Preload("CreatedBy").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&branches).Error

	return branches, total, err
}

func (r *academyBranchRepository) Create(ctx context.Context, branch *entity.AcademyBranch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *academyBranchRepository) Update(ctx context.Context, branch *entity.AcademyBranch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *academyBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.AcademyBranch{}, "id = ?", id).Error
}
