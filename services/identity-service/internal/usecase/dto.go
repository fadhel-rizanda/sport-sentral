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

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
	UserID         uuid.UUID
	Email          string
	Username       string
	ActiveRoleName string
	ActiveRoleID   uuid.UUID
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
	RoleName string
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
	ID             uuid.UUID
	Email          string
	Username       string
	FullName       string
	StatusName     string
	StatusID       uuid.UUID
	ActiveRoleName string
	ActiveRoleID   uuid.UUID
	Roles          []UserRoleResponse
	VerifiedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type UserRoleResponse struct {
	ID         uuid.UUID
	Name       string
	IsActive   bool
	StatusName string
	StatusID   uuid.UUID
}

type ListUsersResponse struct {
	Users    []*UserResponse
	Total    int64
	Page     int
	PageSize int
}

// ─── Profile ──────────────────────────────────────────────────────────────────

type ApplyProfileRequest struct {
	UserID   uuid.UUID
	RoleName string
}

type ToggleProfileRequest struct {
	UserID   uuid.UUID
	RoleName string
}

type ApproveProfileRequest struct {
	UserID   uuid.UUID
	RoleName string
}

// ─── Role ─────────────────────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Name          string
	Description   string
	CreatedByID   uuid.UUID
	PermissionIDs []*uuid.UUID
}

type UpdateRoleRequest struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	UpdatedByID uuid.UUID
}

type RoleResponse struct {
	ID          uuid.UUID
	Name        string
	Description string
	Permissions []PermissionResponse
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedByID uuid.UUID
	UpdatedByID uuid.UUID
	DeletedByID *uuid.UUID
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

type PermissionResponse struct {
	ID          uuid.UUID
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedByID uuid.UUID
	UpdatedByID uuid.UUID
	DeletedByID *uuid.UUID
}

type ListPermissionResponse struct {
	Permissions []*PermissionResponse
	Total       int64
	Page        int
	PageSize    int
}

type CreatePermissionRequest struct {
	Resource    string
	Action      string
	Description string
	CreatedByID uuid.UUID
}

type UpdatePermissionRequest struct {
	ID          uuid.UUID
	Resource    *string
	Action      *string
	Description *string
	UpdatedByID uuid.UUID
}
