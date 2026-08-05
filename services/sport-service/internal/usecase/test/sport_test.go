package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/dto"
	"microservice-golang/services/sport-service/internal/entity"
	"microservice-golang/services/sport-service/internal/repository"
	"microservice-golang/services/sport-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

// ---- Mocks ----

type MockSportRepository struct {
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*entity.Sport, error)
	ListFunc               func(ctx context.Context, filters repository.SportFilters, page, pageSize int) ([]*entity.Sport, int64, error)
	CreateFunc             func(ctx context.Context, sport *entity.Sport) error
	UpdateFunc             func(ctx context.Context, sport *entity.Sport) error
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	GetConfigBySportIDFunc func(ctx context.Context, sportID uuid.UUID) (*entity.SportConfig, error)
	UpsertConfigFunc       func(ctx context.Context, config *entity.SportConfig) error
}

func (m *MockSportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Sport, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockSportRepository) List(ctx context.Context, filters repository.SportFilters, page, pageSize int) ([]*entity.Sport, int64, error) {
	return m.ListFunc(ctx, filters, page, pageSize)
}
func (m *MockSportRepository) Create(ctx context.Context, sport *entity.Sport) error {
	return m.CreateFunc(ctx, sport)
}
func (m *MockSportRepository) Update(ctx context.Context, sport *entity.Sport) error {
	return m.UpdateFunc(ctx, sport)
}
func (m *MockSportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockSportRepository) GetConfigBySportID(ctx context.Context, sportID uuid.UUID) (*entity.SportConfig, error) {
	return m.GetConfigBySportIDFunc(ctx, sportID)
}
func (m *MockSportRepository) UpsertConfig(ctx context.Context, config *entity.SportConfig) error {
	return m.UpsertConfigFunc(ctx, config)
}

type MockStatusRepository struct {
	UpsertFunc           func(ctx context.Context, s entity.Status) error
	DeleteFunc           func(ctx context.Context, id uuid.UUID) error
	GetByIDFunc          func(ctx context.Context, id uuid.UUID) (*entity.Status, error)
	GetByTypeAndNameFunc func(ctx context.Context, t string, name string) (*entity.Status, error)
}

func (m *MockStatusRepository) Upsert(ctx context.Context, s entity.Status) error {
	return m.UpsertFunc(ctx, s)
}
func (m *MockStatusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockStatusRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockStatusRepository) GetByTypeAndName(ctx context.Context, t string, name string) (*entity.Status, error) {
	return m.GetByTypeAndNameFunc(ctx, t, name)
}

type MockTagRepository struct {
	UpsertFunc  func(ctx context.Context, t entity.Tag) error
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
}

func (m *MockTagRepository) Upsert(ctx context.Context, t entity.Tag) error {
	return m.UpsertFunc(ctx, t)
}
func (m *MockTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	return m.GetByIDFunc(ctx, id)
}

type MockSportEventPublisher struct {
	PublishSportCreatedFunc func(ctx context.Context, evt *sportv1.SportEvent) error
	PublishSportUpdatedFunc func(ctx context.Context, evt *sportv1.SportEvent) error
	PublishSportDeletedFunc func(ctx context.Context, evt *sportv1.SportEvent) error
}

func (m *MockSportEventPublisher) PublishSportCreated(ctx context.Context, evt *sportv1.SportEvent) error {
	return m.PublishSportCreatedFunc(ctx, evt)
}
func (m *MockSportEventPublisher) PublishSportUpdated(ctx context.Context, evt *sportv1.SportEvent) error {
	return m.PublishSportUpdatedFunc(ctx, evt)
}
func (m *MockSportEventPublisher) PublishSportDeleted(ctx context.Context, evt *sportv1.SportEvent) error {
	return m.PublishSportDeletedFunc(ctx, evt)
}

// ---- Context Helper ----

func withActiveRole(role string) context.Context {
	md := metadata.New(map[string]string{
		"active-role": role,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

// ---- Tests ----

func TestSportUseCase_Create(t *testing.T) {
	statusID := uuid.New()
	tierTagID := uuid.New()
	createdByID := uuid.New()

	req := dto.CreateSportRequest{
		Name:        "Soccer",
		Slug:        "soccer",
		Description: "Beautiful game",
		StatusID:    statusID,
		TierTagID:   tierTagID,
		CreatedByID: createdByID,
	}

	t.Run("success", func(t *testing.T) {
		repo := &MockSportRepository{}
		statusRepo := &MockStatusRepository{}
		tagRepo := &MockTagRepository{}
		publisher := &MockSportEventPublisher{}

		uc := usecase.NewSportUseCase(repo, statusRepo, tagRepo, publisher)
		ctx := withActiveRole("platform_admin")

		statusRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
			if id != statusID {
				return nil, gorm.ErrRecordNotFound
			}
			return &entity.Status{ID: statusID, Name: "Active"}, nil
		}

		tagRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
			if id != tierTagID {
				return nil, gorm.ErrRecordNotFound
			}
			return &entity.Tag{ID: tierTagID, Name: "Tier 1"}, nil
		}

		repo.CreateFunc = func(ctx context.Context, s *entity.Sport) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Sport, error) {
			return &entity.Sport{
				ID:          id,
				Name:        req.Name,
				Slug:        req.Slug,
				Description: req.Description,
				StatusID:    req.StatusID,
				TierTagID:   req.TierTagID,
				Status:      &entity.Status{ID: statusID, Name: "Active"},
				TierTag:     &entity.Tag{ID: tierTagID, Name: "Tier 1"},
			}, nil
		}

		published := false
		publisher.PublishSportCreatedFunc = func(ctx context.Context, evt *sportv1.SportEvent) error {
			published = true
			if evt.Name != "Soccer" {
				t.Errorf("expected published sport name to be Soccer, got %s", evt.Name)
			}
			return nil
		}

		res, err := uc.Create(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Soccer" {
			t.Errorf("expected name to be Soccer, got %s", res.Name)
		}
		if !published {
			t.Error("expected event to be published")
		}
	})

	t.Run("unauthorized_missing_metadata", func(t *testing.T) {
		uc := usecase.NewSportUseCase(nil, nil, nil, nil)
		_, err := uc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("forbidden_role", func(t *testing.T) {
		uc := usecase.NewSportUseCase(nil, nil, nil, nil)
		ctx := withActiveRole("scout")
		_, err := uc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsForbidden(err) {
			t.Errorf("expected forbidden error, got %v", err)
		}
	})

	t.Run("status_not_found", func(t *testing.T) {
		statusRepo := &MockStatusRepository{}
		uc := usecase.NewSportUseCase(nil, statusRepo, nil, nil)
		ctx := withActiveRole("platform_admin")

		statusRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
			return nil, gorm.ErrRecordNotFound
		}

		_, err := uc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})

	t.Run("unique_constraint_conflict", func(t *testing.T) {
		repo := &MockSportRepository{}
		statusRepo := &MockStatusRepository{}
		tagRepo := &MockTagRepository{}
		publisher := &MockSportEventPublisher{}

		uc := usecase.NewSportUseCase(repo, statusRepo, tagRepo, publisher)
		ctx := withActiveRole("platform_admin")

		statusRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
			return &entity.Status{ID: statusID, Name: "Active"}, nil
		}

		tagRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
			return &entity.Tag{ID: tierTagID, Name: "Tier 1"}, nil
		}

		repo.CreateFunc = func(ctx context.Context, s *entity.Sport) error {
			return &pgconn.PgError{
				Code:           "23505", // UniqueViolation
				ConstraintName: "uni_sports_name",
			}
		}

		_, err := uc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict error, got %v", err)
		}
	})
}

func TestSportUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &MockSportRepository{}
		uc := usecase.NewSportUseCase(repo, nil, nil, nil)

		repo.GetByIDFunc = func(ctx context.Context, sportID uuid.UUID) (*entity.Sport, error) {
			return &entity.Sport{
				ID:   sportID,
				Name: "Basketball",
			}, nil
		}

		res, err := uc.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Name != "Basketball" {
			t.Errorf("expected Basketball, got %s", res.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		repo := &MockSportRepository{}
		uc := usecase.NewSportUseCase(repo, nil, nil, nil)

		repo.GetByIDFunc = func(ctx context.Context, sportID uuid.UUID) (*entity.Sport, error) {
			return nil, gorm.ErrRecordNotFound
		}

		_, err := uc.GetByID(context.Background(), id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestSportUseCase_Delete(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &MockSportRepository{}
		publisher := &MockSportEventPublisher{}
		uc := usecase.NewSportUseCase(repo, nil, nil, publisher)
		ctx := withActiveRole("platform_admin")

		repo.GetByIDFunc = func(ctx context.Context, sportID uuid.UUID) (*entity.Sport, error) {
			return &entity.Sport{
				ID:   sportID,
				Name: "Tennis",
			}, nil
		}

		repo.DeleteFunc = func(ctx context.Context, sportID uuid.UUID) error {
			return nil
		}

		published := false
		publisher.PublishSportDeletedFunc = func(ctx context.Context, evt *sportv1.SportEvent) error {
			published = true
			return nil
		}

		err := uc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !published {
			t.Error("expected delete event to be published")
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		uc := usecase.NewSportUseCase(nil, nil, nil, nil)
		ctx := withActiveRole("user")

		err := uc.Delete(ctx, id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsForbidden(err) {
			t.Errorf("expected forbidden, got %v", err)
		}
	})
}
