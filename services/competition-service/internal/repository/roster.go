package repository

import (
	"context"
	"microservice-golang/services/competition-service/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RosterRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Roster, error)
	List(ctx context.Context, filters RosterFilters, page, pageSize int) ([]*entity.Roster, int64, error)

	Create(ctx context.Context, roster *entity.Roster) error
	Update(ctx context.Context, roster *entity.Roster) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetMembers(ctx context.Context, rosterID uuid.UUID) ([]*entity.RosterMember, error)
	GetMemberByID(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) (*entity.RosterMember, error)
	AddMember(ctx context.Context, member *entity.RosterMember) error
	RemoveMember(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID, kickedByID uuid.UUID, reason string) error
	DeleteMember(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) error
	CheckAthleteInRoster(ctx context.Context, rosterID, athleteID uuid.UUID) (bool, error)
}

type RosterFilters struct {
	CompetitionID *uuid.UUID
	BranchID      *uuid.UUID
	StatusID      *uuid.UUID
	TagID         *uuid.UUID
	Search        *string
}

type rosterRepository struct {
	db *gorm.DB
}

func NewRosterRepository(db *gorm.DB) RosterRepository {
	return &rosterRepository{
		db: db,
	}
}

func (r *rosterRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Roster, error) {
	var roster entity.Roster
	err := r.db.WithContext(ctx).
		Preload("AcademyBranch.Holding").
		Preload("AcademyBranch.Sport").
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("removed_at ASC NULLS FIRST").Order("added_at ASC")
		}).
		Preload("Members.Athlete").
		Preload("Members.Position").
		Preload("Members.Status").
		Preload("Tag").
		First(&roster, "id = ?", id).Error
	return &roster, err
}

func (r *rosterRepository) List(ctx context.Context, filters RosterFilters, page, pageSize int) ([]*entity.Roster, int64, error) {
	var rosters []*entity.Roster
	var total int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Roster{})

	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}
	if filters.CompetitionID != nil {
		query = query.Where("competition_id = ?", filters.CompetitionID)
	}
	if filters.BranchID != nil {
		query = query.Where("academy_branch_id = ?", filters.BranchID)
	}
	if filters.TagID != nil {
		query = query.Where("tag_id = ?", filters.TagID)
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
		Select("rosters.*, (SELECT COUNT(*) FROM roster_members WHERE roster_members.roster_id = rosters.id AND roster_members.removed_at IS NULL) as member_count").
		Preload("AcademyBranch.Holding").
		Preload("AcademyBranch.Sport").
		Preload("Tag").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rosters).Error

	return rosters, total, err
}

func (r *rosterRepository) Create(ctx context.Context, roster *entity.Roster) error {
	return r.db.WithContext(ctx).Create(roster).Error
}

func (r *rosterRepository) Update(ctx context.Context, roster *entity.Roster) error {
	return r.db.WithContext(ctx).Save(roster).Error
}

func (r *rosterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Roster{}, id).Error
}

func (r *rosterRepository) GetMembers(ctx context.Context, rosterID uuid.UUID) ([]*entity.RosterMember, error) {
	var members []*entity.RosterMember

	err := r.db.WithContext(ctx).
		Where("roster_id = ?", rosterID).
		Preload("Athlete").
		Preload("Position").
		Preload("Status").
		Preload("AddedBy").
		Preload("RemovedBy").
		Order("removed_at ASC NULLS FIRST").
		Order("added_at ASC").
		Find(&members).Error

	return members, err
}

func (r *rosterRepository) GetMemberByID(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) (*entity.RosterMember, error) {
	var member entity.RosterMember
	err := r.db.WithContext(ctx).
		Preload("Athlete").
		Preload("Position").
		Preload("Status").
		Preload("AddedBy").
		Preload("RemovedBy").
		First(&member, "id = ? AND roster_id = ?", memberID, rosterID).Error

	return &member, err
}

func (r *rosterRepository) AddMember(ctx context.Context, member *entity.RosterMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *rosterRepository) RemoveMember(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID, kickedByID uuid.UUID, reason string) error {
	now := time.Now()

	return r.db.WithContext(ctx).Model(&entity.RosterMember{}).
		Where("id = ? AND roster_id = ? AND removed_at IS NULL", memberID, rosterID).
		Updates(map[string]interface{}{
			"removed_at":     &now,
			"removed_by_id":  kickedByID,
			"removal_reason": reason,
		}).Error
}

func (r *rosterRepository) DeleteMember(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND roster_id = ?", memberID, rosterID).Delete(&entity.RosterMember{}).Error
}

func (r *rosterRepository) CheckAthleteInRoster(ctx context.Context, rosterID, athleteID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.RosterMember{}).
		Where("roster_id = ? AND athlete_id = ? AND removed_at IS NULL", rosterID, athleteID).
		Count(&count).Error

	return count > 0, err
}
