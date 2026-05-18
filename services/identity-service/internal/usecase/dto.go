package usecase

import (
	"time"

	"github.com/google/uuid"
)

// ─── General ──────────────────────────────────────────────────────────────────

type ListRequest struct {
	Page     int
	PageSize int
}

type DeleteRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

// ─── Common Simple Responses ──────────────────────────────────────────────────

type UserSimpleResponse struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
}

type StatusSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type TagSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type PermissionSimpleResponse struct {
	ID       uuid.UUID
	Resource string
	Action   string
	Slug     string
}

type RoleSimpleResponse struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	Permissions    []PermissionSimpleResponse
	PermissionsIDs []string
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       uuid.UUID
	Email        string
	Username     string
	FullName     string
	ActiveRole   RoleSimpleResponse
	Status       StatusSimpleResponse
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// ─── User ─────────────────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Email    string
	Username string
	FullName string
	Password string
	RoleID   uuid.UUID
}

type UpdateUserRequest struct {
	ID       uuid.UUID
	FullName *string
	Username *string
}

type DeleteUserRequest struct {
	ID       uuid.UUID
	Password string
}

type AssignRolesRequest struct {
	UserID  uuid.UUID
	RoleIds []uuid.UUID
}

type UserResponse struct {
	ID         uuid.UUID
	Email      string
	Username   string
	FullName   string
	Status     StatusSimpleResponse
	ActiveRole RoleSimpleResponse
	Roles      []UserRoleSimpleResponse
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type UserRoleSimpleResponse struct {
	Role     RoleSimpleResponse
	Status   StatusSimpleResponse
	IsActive bool
}

type ListUsersResponse struct {
	Users    []*UserResponse
	Total    int64
	Page     int
	PageSize int
}

// ─── Profile ──────────────────────────────────────────────────────────────────

type ApplyProfileRequest struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

type ToggleProfileRequest struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

type ApproveProfileRequest struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

// ─── Role ─────────────────────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Name          string
	Description   string
	Slug          string
	CreatedByID   uuid.UUID
	PermissionIDs []*uuid.UUID
}

type UpdateRoleRequest struct {
	ID            uuid.UUID
	Name          *string
	Description   *string
	Slug          *string
	UpdatedByID   uuid.UUID
	PermissionIDs []*uuid.UUID
}

type RoleResponse struct {
	ID          uuid.UUID
	Name        string
	Description string
	Slug        string
	Permissions []PermissionSimpleResponse
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   UserSimpleResponse
	UpdatedBy   UserSimpleResponse
	DeletedBy   *UserSimpleResponse
}

type ListRoleResponse struct {
	Roles    []*RoleResponse
	Total    int64
	Page     int
	PageSize int
}

// ─── Permission ─────────────────────────────────────────────────────────────────────

type ListPermissionRequest struct {
	RoleId   *uuid.UUID
	Page     int
	PageSize int
}

type CreatePermissionRequest struct {
	Resource    string
	Action      string
	Slug        string
	Description string
	CreatedByID uuid.UUID
}

type UpdatePermissionRequest struct {
	ID          uuid.UUID
	Resource    *string
	Action      *string
	Description *string
	Slug        *string
	UpdatedByID uuid.UUID
}

type PermissionResponse struct {
	ID          uuid.UUID
	Resource    string
	Action      string
	Description string
	Slug        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   UserSimpleResponse
	UpdatedBy   UserSimpleResponse
	DeletedBy   *UserSimpleResponse
}

type ListPermissionResponse struct {
	Permissions []*PermissionResponse
	Total       int64
	Page        int
	PageSize    int
}
