package usecase

import (
	"time"

	"github.com/google/uuid"
	"microservice-golang/services/identity-service/internal/entity"
)

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken   string
	RefreshToken  string
	ExpiresAt     time.Time
	UserID        uuid.UUID
	Email         string
	Username      string
	ActiveProfile string
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

type UserResponse struct {
	ID            uuid.UUID
	Email         string
	Username      string
	FullName      string
	Status        string
	ActiveProfile string
	Profiles      []UserRoleResponse
	VerifiedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserRoleResponse struct {
	RoleID   uuid.UUID
	RoleName string
	IsActive bool
	Status   string
}

type ListUsersRequest struct {
	Page     int
	PageSize int
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func ToUserResponse(user *entity.User) *UserResponse {
	profiles := make([]UserRoleResponse, len(user.UserRoles))
	activeProfile := ""

	for i, ur := range user.UserRoles {
		profiles[i] = UserRoleResponse{
			RoleID:   ur.RoleID,
			RoleName: ur.Role.Name,
			IsActive: ur.IsActive,
			Status:   ur.Status.Name,
		}
		if ur.IsActive {
			activeProfile = ur.Role.Name
		}
	}

	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FullName:      user.FullName,
		Status:        user.Status.Name,
		ActiveProfile: activeProfile,
		Profiles:      profiles,
		VerifiedAt:    user.VerifiedAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func (u *UserResponse) RoleIDs() []string {
	ids := make([]string, len(u.Profiles))
	for i, p := range u.Profiles {
		ids[i] = p.RoleID.String()
	}
	return ids
}
