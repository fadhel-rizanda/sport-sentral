package repository

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademyHoldingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error)
	List(ctx context.Context, filters AcademyHoldingFilters, page, pageSize int) ([]*entity.AcademyHolding, int64, error)

	Create(ctx context.Context, holding *entity.AcademyHolding) error
	Update(ctx context.Context, holding *entity.AcademyHolding) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountBranches(ctx context.Context, holdingID uuid.UUID) (int64, error)
}

type AcademyHoldingFilters struct {
	StatusID *uuid.UUID
	Search   *string
}

type academyHoldingRepository struct {
	db *gorm.DB
}

func NewAcademyHoldingRepository(db *gorm.DB) AcademyHoldingRepository {
	return &academyHoldingRepository{
		db: db,
	}
}

func (r *academyHoldingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error) {
	var holding entity.AcademyHolding
	err := r.db.WithContext(ctx).
		Preload("Address.AdministrativeDivision.Country").
		Preload("Address.AdministrativeDivision.Parent").
		Preload("Status").
		Preload("Branches").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&holding, "id = ?", id).Error

	return &holding, err
}

func (r *academyHoldingRepository) List(ctx context.Context, filters AcademyHoldingFilters, page, pageSize int) ([]*entity.AcademyHolding, int64, error) {
	var holdings []*entity.AcademyHolding
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&entity.AcademyHolding{})

	// Apply filters
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

	// Count total
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := query.
		Preload("Address").
		Preload("Status").
		Preload("Branches").
		Preload("CreatedBy").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&holdings).Error

	return holdings, total, err
}

func (r *academyHoldingRepository) Create(ctx context.Context, holding *entity.AcademyHolding) error {
	return r.db.WithContext(ctx).Create(holding).Error
}

func (r *academyHoldingRepository) Update(ctx context.Context, holding *entity.AcademyHolding) error {
	return r.db.WithContext(ctx).Save(holding).Error
}

func (r *academyHoldingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.AcademyHolding{}, "id = ?", id).Error
}

func (r *academyHoldingRepository) CountBranches(ctx context.Context, holdingID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.AcademyBranch{}).
		Where("holding_id = ?", holdingID).
		Count(&count).Error
	return count, err
}
