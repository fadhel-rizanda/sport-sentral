package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-golang/services/venue-service/internal/dto"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/services/venue-service/internal/repository"
	"microservice-golang/services/venue-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

// ---- Mocks ----

type MockPermissionRepository struct {
	ValidateFunc func(ctx context.Context, permissionSlug string) error
}

func (m *MockPermissionRepository) Upsert(ctx context.Context, p entity.Permission) error { return nil }
func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	return nil, nil
}
func (m *MockPermissionRepository) Validate(ctx context.Context, permissionSlug string) error {
	return m.ValidateFunc(ctx, permissionSlug)
}

type MockVenueRepository struct {
	CreateFunc  func(ctx context.Context, venue *entity.Venue) error
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.Venue, error)
	ListFunc    func(ctx context.Context, filters repository.VenueFilters, page, pageSize int) ([]*entity.Venue, int64, error)
	UpdateFunc  func(ctx context.Context, venue *entity.Venue) error
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockVenueRepository) Create(ctx context.Context, venue *entity.Venue) error {
	return m.CreateFunc(ctx, venue)
}
func (m *MockVenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Venue, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockVenueRepository) List(ctx context.Context, filters repository.VenueFilters, page, pageSize int) ([]*entity.Venue, int64, error) {
	return m.ListFunc(ctx, filters, page, pageSize)
}
func (m *MockVenueRepository) Update(ctx context.Context, venue *entity.Venue) error {
	return m.UpdateFunc(ctx, venue)
}
func (m *MockVenueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}

// ---- Tests ----

func TestVenueUseCase_Create(t *testing.T) {
	statusID := uuid.New()
	ownerID := uuid.New()
	adminDivisionID := uuid.New()

	lat := -6.2
	lng := 106.8

	req := dto.CreateVenueRequest{
		Name:            "Stadion Utama",
		Description:     "Main stadium",
		StreetAddress:   "Stadion Street 1",
		Latitude:        &lat,
		Longitude:       &lng,
		AdminDivisionID: adminDivisionID,
		OwnerID:         ownerID,
		StatusID:        statusID,
	}

	t.Run("success", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockVenueRepository{}

		uc := usecase.NewVenueUseCase(permRepo, repo)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.CreateFunc = func(ctx context.Context, v *entity.Venue) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Venue, error) {
			return &entity.Venue{
				ID:            id,
				Name:          req.Name,
				Description:   req.Description,
				StreetAddress: req.StreetAddress,
				StatusID:      req.StatusID,
			}, nil
		}

		res, err := uc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Stadion Utama" {
			t.Errorf("expected Stadion Utama, got %s", res.Name)
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		uc := usecase.NewVenueUseCase(permRepo, nil)

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
}

func TestVenueUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockVenueRepository{}
		uc := usecase.NewVenueUseCase(permRepo, repo)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, venueID uuid.UUID) (*entity.Venue, error) {
			return &entity.Venue{
				ID:   venueID,
				Name: "Get Venue",
			}, nil
		}

		res, err := uc.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Name != "Get Venue" {
			t.Errorf("expected Get Venue, got %s", res.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockVenueRepository{}
		uc := usecase.NewVenueUseCase(permRepo, repo)

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, venueID uuid.UUID) (*entity.Venue, error) {
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
