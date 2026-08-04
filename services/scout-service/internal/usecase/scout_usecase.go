package usecase

import (
	"context"
	"fmt"
	"time"

	"microservice-golang/services/scout-service/internal/entity"
	"microservice-golang/services/scout-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
)

type ScoutUsecase interface {
	CreateScoutProfile(ctx context.Context, userID uuid.UUID, orgName, bio string, avatarID *uuid.UUID) (*entity.Scout, error)
	GetScoutProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.Scout, error)
	UpdateScoutProfile(ctx context.Context, scoutID uuid.UUID, orgName, bio *string, avatarID *uuid.UUID) (*entity.Scout, error)
	// TODO: Feature gating and subscription tier management is delegated to RBAC permissions in identity-service.

	AddToWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID, notes string, priorityTagID *uuid.UUID) (*entity.WatchlistEntry, error)
	RemoveFromWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID) error
	ListWatchlist(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error)
	UpdateWatchlistEntry(ctx context.Context, entryID, scoutID uuid.UUID, notes *string, priorityTagID *uuid.UUID) (*entity.WatchlistEntry, error)

	ListAthleteProfiles(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error)
	GetAthleteProfile(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error)
	GetLeaderboard(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error)

	// TODO: Feature Log/audit-trail is delegated to log-service
	LogScoutActivity(ctx context.Context, scoutID uuid.UUID, action string, resourceID *uuid.UUID, resourceType string) error
	ListActivityLogs(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error)
}

type scoutUsecase struct {
	repo repository.ScoutRepository
}

func NewScoutUsecase(repo repository.ScoutRepository) ScoutUsecase {
	return &scoutUsecase{repo: repo}
}

func (u *scoutUsecase) CreateScoutProfile(ctx context.Context, userID uuid.UUID, orgName, bio string, avatarID *uuid.UUID) (*entity.Scout, error) {
	existing, err := u.repo.GetScoutByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("scout profile")
	}

	scout := &entity.Scout{
		ID:                 uuid.New(),
		UserID:             userID,
		OrganizationName:   orgName,
		Bio:                bio,
		AvatarAttachmentID: avatarID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := u.repo.CreateScout(ctx, scout); err != nil {
		return nil, apperr.Internal(fmt.Errorf("create scout profile: %w", err))
	}

	return scout, nil
}

func (u *scoutUsecase) GetScoutProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.Scout, error) {
	scout, err := u.repo.GetScoutByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if scout == nil {
		return nil, apperr.NotFound("scout profile")
	}
	return scout, nil
}

func (u *scoutUsecase) UpdateScoutProfile(ctx context.Context, scoutID uuid.UUID, orgName, bio *string, avatarID *uuid.UUID) (*entity.Scout, error) {
	scout, err := u.repo.GetScoutByID(ctx, scoutID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if scout == nil {
		return nil, apperr.NotFound("scout profile")
	}

	if orgName != nil {
		scout.OrganizationName = *orgName
	}
	if bio != nil {
		scout.Bio = *bio
	}
	if avatarID != nil {
		scout.AvatarAttachmentID = avatarID
	}
	scout.UpdatedAt = time.Now()

	if err := u.repo.UpdateScout(ctx, scout); err != nil {
		return nil, apperr.Internal(fmt.Errorf("update scout profile: %w", err))
	}

	return scout, nil
}

func (u *scoutUsecase) AddToWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID, notes string, priorityTagID *uuid.UUID) (*entity.WatchlistEntry, error) {
	tagID := uuid.Nil
	if priorityTagID != nil {
		tagID = *priorityTagID
	}

	entry := &entity.WatchlistEntry{
		ID:            uuid.New(),
		ScoutID:       scoutID,
		AthleteID:     athleteID,
		Notes:         notes,
		PriorityTagID: tagID,
		AddedAt:       time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.repo.AddToWatchlist(ctx, entry); err != nil {
		return nil, apperr.Internal(fmt.Errorf("add to watchlist: %w", err))
	}

	return u.repo.GetWatchlistEntry(ctx, entry.ID)
}

func (u *scoutUsecase) RemoveFromWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID) error {
	if err := u.repo.RemoveFromWatchlist(ctx, scoutID, athleteID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (u *scoutUsecase) ListWatchlist(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error) {
	entries, total, err := u.repo.ListWatchlist(ctx, scoutID, page, limit)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return entries, total, nil
}

func (u *scoutUsecase) UpdateWatchlistEntry(ctx context.Context, entryID, scoutID uuid.UUID, notes *string, priorityTagID *uuid.UUID) (*entity.WatchlistEntry, error) {
	entry, err := u.repo.GetWatchlistEntry(ctx, entryID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if entry == nil || entry.ScoutID != scoutID {
		return nil, apperr.NotFound("watchlist entry")
	}

	if notes != nil {
		entry.Notes = *notes
	}
	if priorityTagID != nil {
		entry.PriorityTagID = *priorityTagID
	}
	entry.UpdatedAt = time.Now()

	if err := u.repo.UpdateWatchlistEntry(ctx, entry); err != nil {
		return nil, apperr.Internal(fmt.Errorf("update watchlist entry: %w", err))
	}

	return entry, nil
}

func (u *scoutUsecase) ListAthleteProfiles(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error) {
	profiles, total, err := u.repo.ListAthleteProfiles(ctx, sportID, minAge, maxAge, level, page, limit)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return profiles, total, nil
}

func (u *scoutUsecase) GetAthleteProfile(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error) {
	profile, err := u.repo.GetAthleteProfile(ctx, athleteID, sportID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile == nil {
		return nil, apperr.NotFound("athlete profile")
	}
	return profile, nil
}

func (u *scoutUsecase) GetLeaderboard(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error) {
	entries, total, err := u.repo.GetLeaderboard(ctx, sportID, periodTagID, page, limit)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return entries, total, nil
}

func (u *scoutUsecase) LogScoutActivity(ctx context.Context, scoutID uuid.UUID, action string, resourceID *uuid.UUID, resourceType string) error {
	log := &entity.ScoutActivityLog{
		ID:           uuid.New(),
		ScoutID:      scoutID,
		Action:       action,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CreatedAt:    time.Now(),
	}
	if err := u.repo.CreateActivityLog(ctx, log); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (u *scoutUsecase) ListActivityLogs(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error) {
	logs, total, err := u.repo.ListActivityLogs(ctx, scoutID, page, limit)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return logs, total, nil
}
