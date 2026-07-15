package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/dto"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

// ---- Mocks ----

type MockRoleRepository struct {
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByNameFunc          func(ctx context.Context, name string) (*entity.Role, error)
	ListFunc               func(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)
	CreateFunc             func(ctx context.Context, role *entity.Role) error
	UpdateFunc             func(ctx context.Context, role *entity.Role) error
	AssignPermissionsFunc  func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RevokePermissionsFunc  func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	ReplacePermissionsFunc func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	GetPermissionsFunc     func(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	return m.GetByNameFunc(ctx, name)
}
func (m *MockRoleRepository) List(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	return m.ListFunc(ctx, page, pageSize)
}
func (m *MockRoleRepository) Create(ctx context.Context, role *entity.Role) error {
	return m.CreateFunc(ctx, role)
}
func (m *MockRoleRepository) Update(ctx context.Context, role *entity.Role) error {
	return m.UpdateFunc(ctx, role)
}
func (m *MockRoleRepository) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.AssignPermissionsFunc(ctx, roleID, permissionIDs)
}
func (m *MockRoleRepository) RevokePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.RevokePermissionsFunc(ctx, roleID, permissionIDs)
}
func (m *MockRoleRepository) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.ReplacePermissionsFunc(ctx, roleID, permissionIDs)
}
func (m *MockRoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	return m.GetPermissionsFunc(ctx, roleID)
}

type MockPermissionRepository struct {
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*entity.Permission, error)
	ListFunc            func(ctx context.Context, roleID *uuid.UUID, page, pageSize int) ([]*entity.Permission, int64, error)
	CreateFunc          func(ctx context.Context, permission *entity.Permission) error
	UpdateFunc          func(ctx context.Context, permission *entity.Permission) error
	AssignToRoleFunc    func(ctx context.Context, roleID, permissionID uuid.UUID) error
	RevokeFromRoleFunc  func(ctx context.Context, roleID, permissionID uuid.UUID) error
	CheckPermissionFunc func(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockPermissionRepository) List(ctx context.Context, roleID *uuid.UUID, page, pageSize int) ([]*entity.Permission, int64, error) {
	return m.ListFunc(ctx, roleID, page, pageSize)
}
func (m *MockPermissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	return m.CreateFunc(ctx, permission)
}
func (m *MockPermissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	return m.UpdateFunc(ctx, permission)
}
func (m *MockPermissionRepository) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return m.AssignToRoleFunc(ctx, roleID, permissionID)
}
func (m *MockPermissionRepository) RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return m.RevokeFromRoleFunc(ctx, roleID, permissionID)
}
func (m *MockPermissionRepository) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	return m.CheckPermissionFunc(ctx, userID, resource, action)
}

type MockUserRepository struct {
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	return nil, 0, nil
}
func (m *MockUserRepository) AssignRoles(ctx context.Context, uid uuid.UUID, rids []uuid.UUID) error {
	return nil
}
func (m *MockUserRepository) RemoveRoles(ctx context.Context, uid uuid.UUID, rids []uuid.UUID) error {
	return nil
}

type MockRoleEventPublisher struct {
	PublishRoleCreatedFunc func(ctx context.Context, evt *rbacv1.RoleEvent) error
	PublishRoleUpdatedFunc func(ctx context.Context, evt *rbacv1.RoleEvent) error
	PublishRoleDeletedFunc func(ctx context.Context, evt *rbacv1.RoleEvent) error
}

func (m *MockRoleEventPublisher) PublishRoleCreated(ctx context.Context, evt *rbacv1.RoleEvent) error {
	return m.PublishRoleCreatedFunc(ctx, evt)
}
func (m *MockRoleEventPublisher) PublishRoleUpdated(ctx context.Context, evt *rbacv1.RoleEvent) error {
	return m.PublishRoleUpdatedFunc(ctx, evt)
}
func (m *MockRoleEventPublisher) PublishRoleDeleted(ctx context.Context, evt *rbacv1.RoleEvent) error {
	return m.PublishRoleDeletedFunc(ctx, evt)
}

// ---- Context Helper ----

func withUserMetadata(userID, activeRole string) context.Context {
	md := metadata.New(map[string]string{
		"user-id":     userID,
		"active-role": activeRole,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

// ---- Tests ----

func TestRoleUseCase_Create(t *testing.T) {
	userID := uuid.New()

	req := dto.CreateRoleRequest{
		Name:        "Super Admin",
		Slug:        "super_admin",
		Description: "Has all privileges",
		CreatedByID: userID,
	}

	t.Run("success_as_platform_admin", func(t *testing.T) {
		roleRepo := &MockRoleRepository{}
		permRepo := &MockPermissionRepository{}
		userRepo := &MockUserRepository{}
		publisher := &MockRoleEventPublisher{}
		logger := zap.NewNop()

		uc := usecase.NewRoleUseCase(roleRepo, permRepo, userRepo, logger, publisher)
		ctx := withUserMetadata(userID.String(), "platform_admin")

		roleRepo.CreateFunc = func(ctx context.Context, r *entity.Role) error {
			r.ID = uuid.New()
			return nil
		}

		roleRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
			return &entity.Role{
				ID:          id,
				Name:        req.Name,
				Slug:        req.Slug,
				Description: req.Description,
			}, nil
		}

		published := false
		publisher.PublishRoleCreatedFunc = func(ctx context.Context, evt *rbacv1.RoleEvent) error {
			published = true
			if evt.RoleName != "Super Admin" {
				t.Errorf("expected published role name Super Admin, got %s", evt.RoleName)
			}
			return nil
		}

		res, err := uc.Create(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Super Admin" {
			t.Errorf("expected Super Admin, got %s", res.Name)
		}
		if !published {
			t.Error("expected event to be published")
		}
	})

	t.Run("forbidden_missing_permission", func(t *testing.T) {
		roleRepo := &MockRoleRepository{}
		permRepo := &MockPermissionRepository{}
		userRepo := &MockUserRepository{}
		publisher := &MockRoleEventPublisher{}
		logger := zap.NewNop()

		uc := usecase.NewRoleUseCase(roleRepo, permRepo, userRepo, logger, publisher)
		ctx := withUserMetadata(userID.String(), "member")

		permRepo.CheckPermissionFunc = func(ctx context.Context, uid uuid.UUID, resource, action string) (bool, error) {
			return false, nil
		}

		_, err := uc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsForbidden(err) {
			t.Errorf("expected forbidden, got %v", err)
		}
	})

	t.Run("conflict_slug", func(t *testing.T) {
		roleRepo := &MockRoleRepository{}
		permRepo := &MockPermissionRepository{}
		userRepo := &MockUserRepository{}
		publisher := &MockRoleEventPublisher{}
		logger := zap.NewNop()

		uc := usecase.NewRoleUseCase(roleRepo, permRepo, userRepo, logger, publisher)
		ctx := withUserMetadata(userID.String(), "platform_admin")

		roleRepo.CreateFunc = func(ctx context.Context, r *entity.Role) error {
			return &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "uni_roles_slug",
			}
		}

		_, err := uc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict, got %v", err)
		}
	})
}
