package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"microservice-golang/services/scout-service/internal/entity"
	"microservice-golang/services/scout-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type MockScoutRepository struct {
	CreateScoutFunc          func(ctx context.Context, scout *entity.Scout) error
	GetScoutByUserIDFunc     func(ctx context.Context, userID uuid.UUID) (*entity.Scout, error)
	GetScoutByIDFunc         func(ctx context.Context, id uuid.UUID) (*entity.Scout, error)
	UpdateScoutFunc          func(ctx context.Context, scout *entity.Scout) error
	AddToWatchlistFunc       func(ctx context.Context, entry *entity.WatchlistEntry) error
	RemoveFromWatchlistFunc  func(ctx context.Context, scoutID, athleteID uuid.UUID) error
	GetWatchlistEntryFunc    func(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error)
	ListWatchlistFunc        func(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error)
	UpdateWatchlistEntryFunc func(ctx context.Context, entry *entity.WatchlistEntry) error
	ListAthleteProfilesFunc  func(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error)
	GetAthleteProfileFunc    func(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error)
	UpsertAthleteProfileFunc func(ctx context.Context, profile *entity.AthleteProfile) error
	GetLeaderboardFunc       func(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error)
	CreateActivityLogFunc    func(ctx context.Context, log *entity.ScoutActivityLog) error
	ListActivityLogsFunc     func(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error)
}

func (m *MockScoutRepository) CreateScout(ctx context.Context, scout *entity.Scout) error {
	if m.CreateScoutFunc != nil {
		return m.CreateScoutFunc(ctx, scout)
	}
	return nil
}

func (m *MockScoutRepository) GetScoutByUserID(ctx context.Context, userID uuid.UUID) (*entity.Scout, error) {
	if m.GetScoutByUserIDFunc != nil {
		return m.GetScoutByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockScoutRepository) GetScoutByID(ctx context.Context, id uuid.UUID) (*entity.Scout, error) {
	if m.GetScoutByIDFunc != nil {
		return m.GetScoutByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockScoutRepository) UpdateScout(ctx context.Context, scout *entity.Scout) error {
	if m.UpdateScoutFunc != nil {
		return m.UpdateScoutFunc(ctx, scout)
	}
	return nil
}

func (m *MockScoutRepository) AddToWatchlist(ctx context.Context, entry *entity.WatchlistEntry) error {
	if m.AddToWatchlistFunc != nil {
		return m.AddToWatchlistFunc(ctx, entry)
	}
	return nil
}

func (m *MockScoutRepository) RemoveFromWatchlist(ctx context.Context, scoutID, athleteID uuid.UUID) error {
	if m.RemoveFromWatchlistFunc != nil {
		return m.RemoveFromWatchlistFunc(ctx, scoutID, athleteID)
	}
	return nil
}

func (m *MockScoutRepository) GetWatchlistEntry(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error) {
	if m.GetWatchlistEntryFunc != nil {
		return m.GetWatchlistEntryFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockScoutRepository) ListWatchlist(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.WatchlistEntry, int64, error) {
	if m.ListWatchlistFunc != nil {
		return m.ListWatchlistFunc(ctx, scoutID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockScoutRepository) UpdateWatchlistEntry(ctx context.Context, entry *entity.WatchlistEntry) error {
	if m.UpdateWatchlistEntryFunc != nil {
		return m.UpdateWatchlistEntryFunc(ctx, entry)
	}
	return nil
}

func (m *MockScoutRepository) ListAthleteProfiles(ctx context.Context, sportID *uuid.UUID, minAge, maxAge *int32, level *string, page, limit int) ([]entity.AthleteProfile, int64, error) {
	if m.ListAthleteProfilesFunc != nil {
		return m.ListAthleteProfilesFunc(ctx, sportID, minAge, maxAge, level, page, limit)
	}
	return nil, 0, nil
}

func (m *MockScoutRepository) GetAthleteProfile(ctx context.Context, athleteID uuid.UUID, sportID *uuid.UUID) (*entity.AthleteProfile, error) {
	if m.GetAthleteProfileFunc != nil {
		return m.GetAthleteProfileFunc(ctx, athleteID, sportID)
	}
	return nil, nil
}

func (m *MockScoutRepository) UpsertAthleteProfile(ctx context.Context, profile *entity.AthleteProfile) error {
	if m.UpsertAthleteProfileFunc != nil {
		return m.UpsertAthleteProfileFunc(ctx, profile)
	}
	return nil
}

func (m *MockScoutRepository) GetLeaderboard(ctx context.Context, sportID uuid.UUID, periodTagID *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error) {
	if m.GetLeaderboardFunc != nil {
		return m.GetLeaderboardFunc(ctx, sportID, periodTagID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockScoutRepository) CreateActivityLog(ctx context.Context, log *entity.ScoutActivityLog) error {
	if m.CreateActivityLogFunc != nil {
		return m.CreateActivityLogFunc(ctx, log)
	}
	return nil
}

func (m *MockScoutRepository) ListActivityLogs(ctx context.Context, scoutID uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error) {
	if m.ListActivityLogsFunc != nil {
		return m.ListActivityLogsFunc(ctx, scoutID, page, limit)
	}
	return nil, 0, nil
}

func TestScoutUseCase_CreateScoutProfile(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetScoutByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*entity.Scout, error) {
				return nil, nil
			},
			CreateScoutFunc: func(ctx context.Context, scout *entity.Scout) error {
				return nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		scout, err := uc.CreateScoutProfile(context.Background(), userID, "Global Scout Org", "Bio text", nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if scout.OrganizationName != "Global Scout Org" {
			t.Errorf("expected organization Global Scout Org, got %s", scout.OrganizationName)
		}
	})

	t.Run("conflict_existing_profile", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetScoutByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*entity.Scout, error) {
				return &entity.Scout{ID: uuid.New(), UserID: uid}, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		_, err := uc.CreateScoutProfile(context.Background(), userID, "Global Scout Org", "Bio text", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict error, got %v", err)
		}
	})
}

func TestScoutUseCase_GetScoutProfileByUserID(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetScoutByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*entity.Scout, error) {
				return &entity.Scout{ID: uuid.New(), UserID: uid, OrganizationName: "Org Alpha"}, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		scout, err := uc.GetScoutProfileByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if scout.OrganizationName != "Org Alpha" {
			t.Errorf("expected Org Alpha, got %s", scout.OrganizationName)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetScoutByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*entity.Scout, error) {
				return nil, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		_, err := uc.GetScoutProfileByUserID(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestScoutUseCase_WatchlistOperations(t *testing.T) {
	scoutID := uuid.New()
	athleteID := uuid.New()
	entryID := uuid.New()

	t.Run("add_to_watchlist_success", func(t *testing.T) {
		repo := &MockScoutRepository{
			AddToWatchlistFunc: func(ctx context.Context, entry *entity.WatchlistEntry) error {
				return nil
			},
			GetWatchlistEntryFunc: func(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error) {
				return &entity.WatchlistEntry{
					ID:        id,
					ScoutID:   scoutID,
					AthleteID: athleteID,
					Notes:     "Top prospect",
				}, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		entry, err := uc.AddToWatchlist(context.Background(), scoutID, athleteID, "Top prospect", nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if entry.Notes != "Top prospect" {
			t.Errorf("expected notes 'Top prospect', got %s", entry.Notes)
		}
	})

	t.Run("remove_from_watchlist_success", func(t *testing.T) {
		repo := &MockScoutRepository{
			RemoveFromWatchlistFunc: func(ctx context.Context, sid, aid uuid.UUID) error {
				return nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		err := uc.RemoveFromWatchlist(context.Background(), scoutID, athleteID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("update_watchlist_entry_not_found", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetWatchlistEntryFunc: func(ctx context.Context, id uuid.UUID) (*entity.WatchlistEntry, error) {
				return nil, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		_, err := uc.UpdateWatchlistEntry(context.Background(), entryID, scoutID, nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestScoutUseCase_AthleteProfilesAndLeaderboard(t *testing.T) {
	athleteID := uuid.New()
	sportID := uuid.New()

	t.Run("get_athlete_profile_success", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetAthleteProfileFunc: func(ctx context.Context, aid uuid.UUID, sid *uuid.UUID) (*entity.AthleteProfile, error) {
				return &entity.AthleteProfile{
					ID:        uuid.New(),
					AthleteID: aid,
					Age:       21,
				}, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		prof, err := uc.GetAthleteProfile(context.Background(), athleteID, &sportID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if prof.Age != 21 {
			t.Errorf("expected age 21, got %d", prof.Age)
		}
	})

	t.Run("get_leaderboard_success", func(t *testing.T) {
		repo := &MockScoutRepository{
			GetLeaderboardFunc: func(ctx context.Context, sid uuid.UUID, ptag *uuid.UUID, page, limit int) ([]entity.LeaderboardEntry, int64, error) {
				return []entity.LeaderboardEntry{
					{ID: uuid.New(), Rank: 1, Score: 98.5},
				}, 1, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		entries, total, err := uc.GetLeaderboard(context.Background(), sportID, nil, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 || len(entries) != 1 || entries[0].Rank != 1 {
			t.Errorf("unexpected leaderboard output: total=%d entries=%+v", total, entries)
		}
	})
}

func TestScoutUseCase_ActivityLogs(t *testing.T) {
	scoutID := uuid.New()

	t.Run("log_and_list_activity_logs", func(t *testing.T) {
		repo := &MockScoutRepository{
			CreateActivityLogFunc: func(ctx context.Context, log *entity.ScoutActivityLog) error {
				return nil
			},
			ListActivityLogsFunc: func(ctx context.Context, sid uuid.UUID, page, limit int) ([]entity.ScoutActivityLog, int64, error) {
				return []entity.ScoutActivityLog{
					{ID: uuid.New(), ScoutID: sid, Action: "VIEW_PROFILE", CreatedAt: time.Now()},
				}, 1, nil
			},
		}

		uc := usecase.NewScoutUsecase(repo)
		err := uc.LogScoutActivity(context.Background(), scoutID, "VIEW_PROFILE", nil, "ATHLETE")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		logs, total, err := uc.ListActivityLogs(context.Background(), scoutID, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 || len(logs) != 1 || logs[0].Action != "VIEW_PROFILE" {
			t.Errorf("unexpected activity log output: total=%d logs=%+v", total, logs)
		}
	})
}
