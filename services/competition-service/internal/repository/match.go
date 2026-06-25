package repository

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Match, error)
	List(ctx context.Context, filters MatchFilters, page, pageSize int) ([]*entity.Match, int64, error)
	Create(ctx context.Context, match *entity.Match) error
	Update(ctx context.Context, match *entity.Match) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetParticipants(ctx context.Context, matchID uuid.UUID) ([]*entity.MatchParticipant, error)
	GetParticipantByID(ctx context.Context, matchID, participantID uuid.UUID) (*entity.MatchParticipant, error)
	AddParticipant(ctx context.Context, participant *entity.MatchParticipant) error
	UpdateParticipant(ctx context.Context, participant *entity.MatchParticipant) error
	DeleteParticipant(ctx context.Context, matchID, participantID uuid.UUID) error

	CreateScoreLog(ctx context.Context, scoreLog *entity.MatchParticipantScoreLog) error
	GetScoreLogs(ctx context.Context, participantID uuid.UUID) ([]*entity.MatchParticipantScoreLog, error)
}

type MatchFilters struct {
	BranchID        *uuid.UUID
	Status          *string
	ScheduledBefore *time.Time
	ScheduledAfter  *time.Time
	Search          *string
}

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{
		db: db,
	}
}

func (r *matchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Match, error) {
	var match entity.Match
	err := r.db.WithContext(ctx).
		Preload("Branch.Competition").
		Preload("Stats").
		Preload("Participants.Roster").
		Preload("Participants.FormatTag").
		Preload("Participants.ResultTag").
		First(&match, "id = ?", id).Error
	return &match, err
}

func (r *matchRepository) List(ctx context.Context, filters MatchFilters, page, pageSize int) ([]*entity.Match, int64, error) {
	var matches []*entity.Match
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Match{})

	if filters.BranchID != nil {
		query = query.Where("branch_id = ?", filters.BranchID)
	}
	if filters.Status != nil && *filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ScheduledBefore != nil {
		query = query.Where("scheduled_at <= ?", filters.ScheduledBefore)
	}
	if filters.ScheduledAfter != nil {
		query = query.Where("scheduled_at >= ?", filters.ScheduledAfter)
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where("location ILIKE ? OR referee ILIKE ?", "%"+*filters.Search+"%", "%"+*filters.Search+"%")
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Branch.Competition").
		Preload("Participants.Roster").
		Preload("Participants.FormatTag").
		Preload("Participants.ResultTag").
		Order("scheduled_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&matches).Error

	return matches, total, err
}

func (r *matchRepository) Create(ctx context.Context, match *entity.Match) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *matchRepository) Update(ctx context.Context, match *entity.Match) error {
	return r.db.WithContext(ctx).Save(match).Error
}

func (r *matchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Match{}, "id = ?", id).Error
}

func (r *matchRepository) GetParticipants(ctx context.Context, matchID uuid.UUID) ([]*entity.MatchParticipant, error) {
	var participants []*entity.MatchParticipant
	err := r.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		Preload("Roster").
		Preload("FormatTag").
		Preload("ResultTag").
		Find(&participants).Error
	return participants, err
}

func (r *matchRepository) GetParticipantByID(ctx context.Context, matchID, participantID uuid.UUID) (*entity.MatchParticipant, error) {
	var participant entity.MatchParticipant
	err := r.db.WithContext(ctx).
		First(&participant, "id = ? AND match_id = ?", participantID, matchID).Error
	return &participant, err
}

func (r *matchRepository) AddParticipant(ctx context.Context, participant *entity.MatchParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *matchRepository) UpdateParticipant(ctx context.Context, participant *entity.MatchParticipant) error {
	return r.db.WithContext(ctx).Save(participant).Error
}

func (r *matchRepository) DeleteParticipant(ctx context.Context, matchID, participantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND match_id = ?", participantID, matchID).
		Delete(&entity.MatchParticipant{}).Error
}

func (r *matchRepository) CreateScoreLog(ctx context.Context, scoreLog *entity.MatchParticipantScoreLog) error {
	return r.db.WithContext(ctx).Create(scoreLog).Error
}

func (r *matchRepository) GetScoreLogs(ctx context.Context, participantID uuid.UUID) ([]*entity.MatchParticipantScoreLog, error) {
	var logs []*entity.MatchParticipantScoreLog
	err := r.db.WithContext(ctx).
		Where("match_participant_id = ?", participantID).
		Order("updated_at DESC").
		Find(&logs).Error
	return logs, err
}
