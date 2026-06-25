package repository

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MatchStatRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.MatchStat, error)
	List(ctx context.Context, filters MatchStatFilters, page, pageSize int) ([]*entity.MatchStat, int64, error)
	Create(ctx context.Context, stat *entity.MatchStat) error
	Update(ctx context.Context, stat *entity.MatchStat) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Aggregate operations
	GetAggregate(ctx context.Context, competitionID, branchID, athleteID, statTypeID uuid.UUID) (*entity.AthleteStatsAggregate, error)
	SaveAggregate(ctx context.Context, aggregate *entity.AthleteStatsAggregate) error
	ListAggregates(ctx context.Context, filters AggregateFilters, page, pageSize int) ([]*entity.AthleteStatsAggregate, int64, error)
}

type MatchStatFilters struct {
	MatchID    *uuid.UUID
	AthleteID  *uuid.UUID
	StatTypeID *uuid.UUID
}

type AggregateFilters struct {
	CompetitionID *uuid.UUID
	BranchID      *uuid.UUID
	AthleteID     *uuid.UUID
	StatTypeID    *uuid.UUID
}

type matchStatRepository struct {
	db *gorm.DB
}

func NewMatchStatRepository(db *gorm.DB) MatchStatRepository {
	return &matchStatRepository{
		db: db,
	}
}

func (r *matchStatRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.MatchStat, error) {
	var stat entity.MatchStat
	err := r.db.WithContext(ctx).
		Preload("Match").
		Preload("Athlete").
		Preload("StatType").
		Preload("RecordedBy").
		First(&stat, "id = ?", id).Error
	return &stat, err
}

func (r *matchStatRepository) List(ctx context.Context, filters MatchStatFilters, page, pageSize int) ([]*entity.MatchStat, int64, error) {
	var stats []*entity.MatchStat
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.MatchStat{})

	if filters.MatchID != nil {
		query = query.Where("match_id = ?", filters.MatchID)
	}
	if filters.AthleteID != nil {
		query = query.Where("athlete_id = ?", filters.AthleteID)
	}
	if filters.StatTypeID != nil {
		query = query.Where("stat_type_id = ?", filters.StatTypeID)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Match").
		Preload("Athlete").
		Preload("StatType").
		Preload("RecordedBy").
		Offset(offset).
		Limit(pageSize).
		Find(&stats).Error

	return stats, total, err
}

func (r *matchStatRepository) Create(ctx context.Context, stat *entity.MatchStat) error {
	return r.db.WithContext(ctx).Create(stat).Error
}

func (r *matchStatRepository) Update(ctx context.Context, stat *entity.MatchStat) error {
	return r.db.WithContext(ctx).Save(stat).Error
}

func (r *matchStatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.MatchStat{}, "id = ?", id).Error
}

func (r *matchStatRepository) GetAggregate(ctx context.Context, competitionID, branchID, athleteID, statTypeID uuid.UUID) (*entity.AthleteStatsAggregate, error) {
	var aggregate entity.AthleteStatsAggregate
	err := r.db.WithContext(ctx).
		Preload("Competition").
		Preload("Branch").
		Preload("Athlete").
		Preload("StatType").
		First(&aggregate, "competition_id = ? AND branch_id = ? AND athlete_id = ? AND stat_type_id = ?", competitionID, branchID, athleteID, statTypeID).Error
	return &aggregate, err
}

func (r *matchStatRepository) SaveAggregate(ctx context.Context, aggregate *entity.AthleteStatsAggregate) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "competition_id"},
				{Name: "branch_id"},
				{Name: "athlete_id"},
				{Name: "stat_type_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"total_matches",
				"avg_value",
				"sum_value",
				"max_value",
				"min_value",
				"last_updated_at",
			}),
		}).
		Create(aggregate).Error
}

func (r *matchStatRepository) ListAggregates(ctx context.Context, filters AggregateFilters, page, pageSize int) ([]*entity.AthleteStatsAggregate, int64, error) {
	var aggregates []*entity.AthleteStatsAggregate
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.AthleteStatsAggregate{})

	if filters.CompetitionID != nil {
		query = query.Where("competition_id = ?", filters.CompetitionID)
	}
	if filters.BranchID != nil {
		query = query.Where("branch_id = ?", filters.BranchID)
	}
	if filters.AthleteID != nil {
		query = query.Where("athlete_id = ?", filters.AthleteID)
	}
	if filters.StatTypeID != nil {
		query = query.Where("stat_type_id = ?", filters.StatTypeID)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Competition").
		Preload("Branch").
		Preload("Athlete").
		Preload("StatType").
		Order("last_updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&aggregates).Error

	return aggregates, total, err
}
