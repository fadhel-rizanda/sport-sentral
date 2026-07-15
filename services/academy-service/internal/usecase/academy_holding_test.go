package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/repository"
	"microservice-golang/services/academy-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

// ---- Mocks ----

type MockPermissionRepository struct {
	ValidateFunc func(ctx context.Context, permissionSlug string) error
}

func (m *MockPermissionRepository) Upsert(ctx context.Context, p entity.Permission) error {
	return nil
}
func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	return nil, nil
}
func (m *MockPermissionRepository) Validate(ctx context.Context, permissionSlug string) error {
	return m.ValidateFunc(ctx, permissionSlug)
}

type MockAcademyHoldingRepository struct {
	GetByIDFunc       func(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error)
	ListFunc          func(ctx context.Context, filters repository.AcademyHoldingFilters, page, pageSize int) ([]*entity.AcademyHolding, int64, error)
	CreateFunc        func(ctx context.Context, holding *entity.AcademyHolding) error
	UpdateFunc        func(ctx context.Context, holding *entity.AcademyHolding) error
	DeleteFunc        func(ctx context.Context, id uuid.UUID) error
	CountBranchesFunc func(ctx context.Context, holdingID uuid.UUID) (int64, error)
}

func (m *MockAcademyHoldingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockAcademyHoldingRepository) List(ctx context.Context, filters repository.AcademyHoldingFilters, page, pageSize int) ([]*entity.AcademyHolding, int64, error) {
	return m.ListFunc(ctx, filters, page, pageSize)
}
func (m *MockAcademyHoldingRepository) Create(ctx context.Context, holding *entity.AcademyHolding) error {
	return m.CreateFunc(ctx, holding)
}
func (m *MockAcademyHoldingRepository) Update(ctx context.Context, holding *entity.AcademyHolding) error {
	return m.UpdateFunc(ctx, holding)
}
func (m *MockAcademyHoldingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockAcademyHoldingRepository) CountBranches(ctx context.Context, holdingID uuid.UUID) (int64, error) {
	return m.CountBranchesFunc(ctx, holdingID)
}

type MockAcademyHoldingEventPublisher struct {
	PublishHoldingCreatedFunc func(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
	PublishHoldingUpdatedFunc func(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
	PublishHoldingDeletedFunc func(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
}

func (m *MockAcademyHoldingEventPublisher) PublishHoldingCreated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	return m.PublishHoldingCreatedFunc(ctx, evt)
}
func (m *MockAcademyHoldingEventPublisher) PublishHoldingUpdated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	return m.PublishHoldingUpdatedFunc(ctx, evt)
}
func (m *MockAcademyHoldingEventPublisher) PublishHoldingDeleted(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	return m.PublishHoldingDeletedFunc(ctx, evt)
}

// ---- Tests ----

func TestAcademyHoldingUseCase_Create(t *testing.T) {
	adminID := uuid.New()
	statusID := uuid.New()
	adminDivisionID := uuid.New()

	req := dto.CreateAcademyHoldingRequest{
		Name:            "Academy One",
		Description:     "Top holding academy",
		Email:           "info@academyone.com",
		PhoneNumber:     "123456",
		StatusID:        statusID,
		AdminDivisionID: adminDivisionID,
		CreatedByID:     adminID,
	}

	t.Run("success", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockAcademyHoldingRepository{}
		publisher := &MockAcademyHoldingEventPublisher{}

		uc := usecase.NewAcademyHoldingUseCase(permRepo, repo, publisher)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.CreateFunc = func(ctx context.Context, h *entity.AcademyHolding) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.AcademyHolding, error) {
			return &entity.AcademyHolding{
				ID:          id,
				Name:        req.Name,
				Description: req.Description,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
				StatusID:    req.StatusID,
				Address:     &entity.AcademyHoldingAddress{ID: uuid.New()},
			}, nil
		}

		published := false
		publisher.PublishHoldingCreatedFunc = func(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
			published = true
			if evt.Name != "Academy One" {
				t.Errorf("expected published name Academy One, got %s", evt.Name)
			}
			return nil
		}

		res, err := uc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Academy One" {
			t.Errorf("expected Academy One, got %s", res.Name)
		}
		if !published {
			t.Error("expected event to be published")
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		uc := usecase.NewAcademyHoldingUseCase(permRepo, nil, nil)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return apperr.Forbidden("forbidden")
		}

		_, err := uc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsForbidden(err) {
			t.Errorf("expected forbidden, got %v", err)
		}
	})

	t.Run("unique_constraint_name_conflict", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockAcademyHoldingRepository{}
		uc := usecase.NewAcademyHoldingUseCase(permRepo, repo, nil)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.CreateFunc = func(ctx context.Context, h *entity.AcademyHolding) error {
			return &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "uni_academy_holdings_name",
			}
		}

		_, err := uc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict, got %v", err)
		}
	})
}

func TestAcademyHoldingUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockAcademyHoldingRepository{}
		uc := usecase.NewAcademyHoldingUseCase(permRepo, repo, nil)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, holdingID uuid.UUID) (*entity.AcademyHolding, error) {
			return &entity.AcademyHolding{
				ID:   holdingID,
				Name: "Get Academy",
			}, nil
		}

		res, err := uc.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Name != "Get Academy" {
			t.Errorf("expected Get Academy, got %s", res.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockAcademyHoldingRepository{}
		uc := usecase.NewAcademyHoldingUseCase(permRepo, repo, nil)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, holdingID uuid.UUID) (*entity.AcademyHolding, error) {
			return nil, gorm.ErrRecordNotFound
		}

		_, err := uc.GetByID(context.Background(), id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found, got %v", err)
		}
	})
}
