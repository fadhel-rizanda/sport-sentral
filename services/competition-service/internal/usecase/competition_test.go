package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/repository"
	"microservice-golang/services/competition-service/internal/usecase"
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

type MockCompetitionRepository struct {
	GetByIDFunc     func(ctx context.Context, id uuid.UUID) (*entity.Competition, error)
	ListFunc        func(ctx context.Context, filters repository.CompetitionFilters, page, pageSize int) ([]*entity.Competition, int64, error)
	CreateFunc      func(ctx context.Context, comp *entity.Competition) error
	UpdateFunc      func(ctx context.Context, comp *entity.Competition) error
	DeleteFunc      func(ctx context.Context, id uuid.UUID) error
	CreateAdminFunc func(ctx context.Context, admin *entity.CompetitionAdmin) error
	GetAdminFunc    func(ctx context.Context, competitionID, userID uuid.UUID) (*entity.CompetitionAdmin, error)
	DeleteAdminFunc func(ctx context.Context, competitionID, userID uuid.UUID) error
}

func (m *MockCompetitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Competition, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockCompetitionRepository) List(ctx context.Context, filters repository.CompetitionFilters, page, pageSize int) ([]*entity.Competition, int64, error) {
	return m.ListFunc(ctx, filters, page, pageSize)
}
func (m *MockCompetitionRepository) Create(ctx context.Context, comp *entity.Competition) error {
	return m.CreateFunc(ctx, comp)
}
func (m *MockCompetitionRepository) Update(ctx context.Context, comp *entity.Competition) error {
	return m.UpdateFunc(ctx, comp)
}
func (m *MockCompetitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockCompetitionRepository) CreateAdmin(ctx context.Context, admin *entity.CompetitionAdmin) error {
	return m.CreateAdminFunc(ctx, admin)
}
func (m *MockCompetitionRepository) GetAdmin(ctx context.Context, competitionID, userID uuid.UUID) (*entity.CompetitionAdmin, error) {
	return m.GetAdminFunc(ctx, competitionID, userID)
}
func (m *MockCompetitionRepository) DeleteAdmin(ctx context.Context, competitionID, userID uuid.UUID) error {
	return m.DeleteAdminFunc(ctx, competitionID, userID)
}

type MockAcademyAdminRepository struct {
	GetByUserIDFunc func(ctx context.Context, userID uuid.UUID) (*entity.AcademyAdmin, error)
}

func (m *MockAcademyAdminRepository) Upsert(ctx context.Context, admin entity.AcademyAdmin) error {
	return nil
}
func (m *MockAcademyAdminRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockAcademyAdminRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyAdmin, error) {
	return nil, nil
}
func (m *MockAcademyAdminRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AcademyAdmin, error) {
	return m.GetByUserIDFunc(ctx, userID)
}

type MockAcademyBranchRepository struct {
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error)
}

func (m *MockAcademyBranchRepository) Upsert(ctx context.Context, branch entity.AcademyBranch) error {
	return nil
}
func (m *MockAcademyBranchRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockAcademyBranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyBranch, error) {
	return m.GetByIDFunc(ctx, id)
}

type MockRoleRepository struct {
	GetBySlugFunc func(ctx context.Context, slug string) (*entity.Role, error)
}

func (m *MockRoleRepository) Upsert(ctx context.Context, role entity.Role) error { return nil }
func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	return nil, nil
}
func (m *MockRoleRepository) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}
func (m *MockRoleRepository) GetBySlug(ctx context.Context, slug string) (*entity.Role, error) {
	return m.GetBySlugFunc(ctx, slug)
}

// ---- Context Helper ----

func withActiveRole(role string) context.Context {
	md := metadata.New(map[string]string{
		"active-role": role,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

// ---- Tests ----

func TestCompetitionUseCase_Create(t *testing.T) {
	adminID := uuid.New()
	sportID := uuid.New()
	statusID := uuid.New()
	tierID := uuid.New()

	req := dto.CreateCompetitionRequest{
		Name:      "Championship 2026",
		SportID:   sportID,
		TierID:    tierID,
		StatusID:  statusID,
		StartDate: time.Now(),
	}

	t.Run("success_as_platform_admin", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockCompetitionRepository{}
		academyAdminRepo := &MockAcademyAdminRepository{}
		academyBranchRepo := &MockAcademyBranchRepository{}
		roleRepo := &MockRoleRepository{}

		uc := usecase.NewCompetitionUseCase(permRepo, repo, academyAdminRepo, academyBranchRepo, roleRepo)
		ctx := withActiveRole("platform_admin")

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.CreateFunc = func(ctx context.Context, c *entity.Competition) error {
			return nil
		}

		roleRepo.GetBySlugFunc = func(ctx context.Context, slug string) (*entity.Role, error) {
			return &entity.Role{ID: uuid.New(), Slug: slug}, nil
		}

		repo.CreateAdminFunc = func(ctx context.Context, admin *entity.CompetitionAdmin) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Competition, error) {
			return &entity.Competition{
				ID:       id,
				Name:     req.Name,
				SportID:  req.SportID,
				StatusID: req.StatusID,
			}, nil
		}

		res, err := uc.Create(ctx, adminID, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Championship 2026" {
			t.Errorf("expected Championship 2026, got %s", res.Name)
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		uc := usecase.NewCompetitionUseCase(permRepo, nil, nil, nil, nil)
		ctx := withActiveRole("platform_admin")

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return apperr.Forbidden("forbidden")
		}

		_, err := uc.Create(ctx, adminID, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsForbidden(err) {
			t.Errorf("expected forbidden, got %v", err)
		}
	})

	t.Run("conflict_name", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockCompetitionRepository{}
		academyAdminRepo := &MockAcademyAdminRepository{}
		academyBranchRepo := &MockAcademyBranchRepository{}
		roleRepo := &MockRoleRepository{}

		uc := usecase.NewCompetitionUseCase(permRepo, repo, academyAdminRepo, academyBranchRepo, roleRepo)
		ctx := withActiveRole("platform_admin")

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.CreateFunc = func(ctx context.Context, c *entity.Competition) error {
			return &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "idx_competition_name",
			}
		}

		_, err := uc.Create(ctx, adminID, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict, got %v", err)
		}
	})
}

func TestCompetitionUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockCompetitionRepository{}
		uc := usecase.NewCompetitionUseCase(permRepo, repo, nil, nil, nil)
		ctx := withActiveRole("user")

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, compID uuid.UUID) (*entity.Competition, error) {
			return &entity.Competition{
				ID:   compID,
				Name: "Get Comp",
			}, nil
		}

		res, err := uc.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Name != "Get Comp" {
			t.Errorf("expected Get Comp, got %s", res.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		permRepo := &MockPermissionRepository{}
		repo := &MockCompetitionRepository{}
		uc := usecase.NewCompetitionUseCase(permRepo, repo, nil, nil, nil)
		ctx := withActiveRole("user")

		permRepo.ValidateFunc = func(ctx context.Context, slug string) error {
			return nil
		}

		repo.GetByIDFunc = func(ctx context.Context, compID uuid.UUID) (*entity.Competition, error) {
			return nil, gorm.ErrRecordNotFound
		}

		_, err := uc.GetByID(ctx, id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found, got %v", err)
		}
	})
}
