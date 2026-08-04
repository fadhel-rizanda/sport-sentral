package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/scout-service/internal/entity"
)

type ScoutRepository interface {
	// Scout Profile
	CreateScout(ctx context.Context, scout *entity.Scout) error
	GetScoutByUserID(ctx context.Context, userID uuid.UUID) (*entity.Scout, error)
	GetScoutByID(ctx context.Context, id uuid.UUID) (*entity.Scout, error)
	UpdateScout(ctx context.Context, scout *entity.Scout) error

	// Watchlist
	AddToWatchlist(ctx context.Context, entry *entity.WatchlistEntry) error
	RemoveFromWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID) error
	GetWatchlistEntry(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error)
	ListWatchlist(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error)
	UpdateWatchlistEntry(ctx context.Context, entry *entity.WatchlistEntry) error

	// Athlete Profiles & Leaderboards
	ListAthleteProfiles(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error)
	GetAthleteProfile(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error)
	UpsertAthleteProfile(ctx context.Context, profile *entity.AthleteProfile) error
	GetLeaderboard(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error)

	// Activity Log
	CreateActivityLog(ctx context.Context, log *entity.ScoutActivityLog) error
	ListActivityLogs(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error)
}

type scoutRepository struct {
	db *gorm.DB
}

func NewScoutRepository(db *gorm.DB) ScoutRepository {
	return &scoutRepository{db: db}
}

func (r *scoutRepository) CreateScout(ctx context.Context, scout *entity.Scout) error {
	return r.db.WithContext(ctx).Create(scout).Error
}

func (r *scoutRepository) GetScoutByUserID(ctx context.Context, userID uuid.UUID) (*entity.Scout, error) {
	var scout entity.Scout
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&scout).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get scout by user id: %w", err)
	}
	return &scout, nil
}

func (r *scoutRepository) GetScoutByID(ctx context.Context, id uuid.UUID) (*entity.Scout, error) {
	var scout entity.Scout
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&scout).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get scout by id: %w", err)
	}
	return &scout, nil
}

func (r *scoutRepository) UpdateScout(ctx context.Context, scout *entity.Scout) error {
	return r.db.WithContext(ctx).Save(scout).Error
}

func (r *scoutRepository) AddToWatchlist(ctx context.Context, entry *entity.WatchlistEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *scoutRepository) RemoveFromWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("scout_id = ? AND athlete_id = ?", scoutID, athleteID).
		Delete(&entity.WatchlistEntry{}).Error
}

func (r *scoutRepository) GetWatchlistEntry(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error) {
	var entry entity.WatchlistEntry
	err := r.db.WithContext(ctx).
		Preload("AthleteProfile").
		Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get watchlist entry: %w", err)
	}
	return &entry, nil
}

func (r *scoutRepository) ListWatchlist(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var entries []entity.WatchlistEntry
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.WatchlistEntry{}).Where("scout_id = ?", scoutID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("AthleteProfile").
		Offset(offset).
		Limit(limit).
		Order("added_at DESC").
		Find(&entries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list watchlist: %w", err)
	}
	return entries, total, nil
}

func (r *scoutRepository) UpdateWatchlistEntry(ctx context.Context, entry *entity.WatchlistEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *scoutRepository) ListAthleteProfiles(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&entity.AthleteProfile{})

	if sportID != nil {
		query = query.Where("sport_id = ?", *sportID)
	}
	if minAge != nil {
		query = query.Where("age >= ?", *minAge)
	}
	if maxAge != nil {
		query = query.Where("age <= ?", *maxAge)
	}
	if level != nil && *level != "" {
		query = query.Where("highest_competition_level = ?", *level)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var profiles []entity.AthleteProfile
	err := query.Preload("Athlete").
		Offset(offset).
		Limit(limit).
		Order("leaderboard_score DESC").
		Find(&profiles).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list athlete profiles: %w", err)
	}
	return profiles, total, nil
}

func (r *scoutRepository) GetAthleteProfile(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error) {
	var profile entity.AthleteProfile
	query := r.db.WithContext(ctx).Where("athlete_id = ?", athleteID)
	if sportID != nil {
		query = query.Where("sport_id = ?", *sportID)
	}
	err := query.Preload("Athlete").First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get athlete profile: %w", err)
	}
	return &profile, nil
}

func (r *scoutRepository) UpsertAthleteProfile(ctx context.Context, profile *entity.AthleteProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *scoutRepository) GetLeaderboard(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&entity.LeaderboardEntry{}).
		Where("sport_id = ?", sportID)

	if periodTagID != nil {
		query = query.Where("period_tag_id = ?", *periodTagID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entries []entity.LeaderboardEntry
	err := query.Preload("AthleteProfile").
		Offset(offset).
		Limit(limit).
		Order("rank ASC").
		Find(&entries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("get leaderboard: %w", err)
	}
	return entries, total, nil
}

func (r *scoutRepository) CreateActivityLog(ctx context.Context, log *entity.ScoutActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *scoutRepository) ListActivityLogs(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&entity.ScoutActivityLog{}).Where("scout_id = ?", scoutID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []entity.ScoutActivityLog
	err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list activity logs: %w", err)
	}
	return logs, total, nil
}
