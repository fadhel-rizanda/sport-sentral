package repository

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompetitionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Competition, error)
	List(ctx context.Context, filters CompetitionFilters, page, pageSize int) ([]*entity.Competition, int64, error)
	Create(ctx context.Context, competition *entity.Competition) error
	Update(ctx context.Context, competition *entity.Competition) error
	Delete(ctx context.Context, id uuid.UUID) error

	CreateAdmin(ctx context.Context, admin *entity.CompetitionAdmin) error
	GetAdmin(ctx context.Context, competitionID, userID uuid.UUID) (*entity.CompetitionAdmin, error)
	DeleteAdmin(ctx context.Context, competitionID, userID uuid.UUID) error
}

type CompetitionFilters struct {
	BranchID *uuid.UUID
	SportID  *uuid.UUID
	TierID   *uuid.UUID
	StatusID *uuid.UUID
	Search   *string
}

type competitionRepository struct {
	db *gorm.DB
}

func NewCompetitionRepository(db *gorm.DB) CompetitionRepository {
	return &competitionRepository{
		db: db,
	}
}

func (r *competitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Competition, error) {
	var competition entity.Competition
	err := r.db.WithContext(ctx).
		Preload("Sport").
		Preload("HostAcademy.Holding").
		Preload("HostAcademy.Sport").
		Preload("HostAcademy.Status").
		Preload("TierTag").
		Preload("Status").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("Branches").
		First(&competition, "id = ?", id).Error
	return &competition, err
}

func (r *competitionRepository) List(ctx context.Context, filters CompetitionFilters, page, pageSize int) ([]*entity.Competition, int64, error) {
	var competitions []*entity.Competition
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Competition{})

	if filters.BranchID != nil {
		query = query.Where("host_academy_branch_id = ?", filters.BranchID)
	}
	if filters.SportID != nil {
		query = query.Where("sport_id = ?", filters.SportID)
	}
	if filters.TierID != nil {
		query = query.Where("tier_id = ?", filters.TierID)
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
		Preload("Sport").
		Preload("HostAcademy.Holding").
		Preload("HostAcademy.Sport").
		Preload("HostAcademy.Status").
		Preload("TierTag").
		Preload("Status").
		Preload("CreatedBy").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&competitions).Error

	return competitions, total, err
}

func (r *competitionRepository) Create(ctx context.Context, competition *entity.Competition) error {
	return r.db.WithContext(ctx).Create(competition).Error
}

func (r *competitionRepository) Update(ctx context.Context, competition *entity.Competition) error {
	return r.db.WithContext(ctx).Save(competition).Error
}

func (r *competitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Competition{}, "id = ?", id).Error
}

func (r *competitionRepository) CreateAdmin(ctx context.Context, admin *entity.CompetitionAdmin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *competitionRepository) GetAdmin(ctx context.Context, competitionID, userID uuid.UUID) (*entity.CompetitionAdmin, error) {
	var admin entity.CompetitionAdmin
	err := r.db.WithContext(ctx).
		Where("competition_id = ? AND user_id = ?", competitionID, userID).
		First(&admin).Error
	return &admin, err
}

func (r *competitionRepository) DeleteAdmin(ctx context.Context, competitionID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("competition_id = ? AND user_id = ?", competitionID, userID).
		Delete(&entity.CompetitionAdmin{}).Error
}
