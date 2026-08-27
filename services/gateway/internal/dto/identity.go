package dto

type UserSimpleResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type PermissionSimpleResponse struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Slug     string `json:"slug"`
}

type RoleSimpleResponse struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"name"`
	Slug          string                     `json:"slug"`
	Permissions   []PermissionSimpleResponse `json:"permissions"`
	PermissionIDs []string                   `json:"permission_ids"`
}

type PermissionResponse struct {
	ID          string              `json:"id"`
	Resource    string              `json:"resource"`
	Action      string              `json:"action"`
	Description string              `json:"description"`
	Slug        string              `json:"slug"`
	CreatedBy   UserSimpleResponse  `json:"created_by"`
	UpdatedBy   *UserSimpleResponse `json:"updated_by,omitempty"`
	DeletedBy   *UserSimpleResponse `json:"deleted_by,omitempty"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	DeletedAt   *string             `json:"deleted_at,omitempty"`
}

type RoleResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Slug        string                     `json:"slug"`
	Permissions []PermissionSimpleResponse `json:"permissions"`
	CreatedBy   UserSimpleResponse         `json:"created_by"`
	UpdatedBy   *UserSimpleResponse        `json:"updated_by,omitempty"`
	DeletedBy   *UserSimpleResponse        `json:"deleted_by,omitempty"`
	CreatedAt   string                     `json:"created_at"`
	UpdatedAt   string                     `json:"updated_at"`
	DeletedAt   *string                    `json:"deleted_at,omitempty"`
}

type UserRoleSimpleResponse struct {
	Role      RoleSimpleResponse   `json:"role"`
	Status    StatusSimpleResponse `json:"status"`
	ExpiredAt *string              `json:"expired_at,omitempty"`
}

type UserResponse struct {
	ID         string                   `json:"id"`
	Email      string                   `json:"email"`
	Username   string                   `json:"username"`
	FullName   string                   `json:"full_name"`
	Status     StatusSimpleResponse     `json:"status"`
	Roles      []UserRoleSimpleResponse `json:"roles"`
	VerifiedAt *string                  `json:"verified_at,omitempty"`
	CreatedAt  string                   `json:"created_at"`
	UpdatedAt  string                   `json:"updated_at"`
	DeletedAt  *string                  `json:"deleted_at,omitempty"`
}

// ─── REQUEST DTOs ─────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"athlete@sportsentral.id"`
	Username string `json:"username" validate:"required" example:"john_doe"`
	FullName string `json:"full_name" validate:"required" example:"John Doe"`
	Password string `json:"password" validate:"required,min=8" example:"Password123!"`
	RoleID   string `json:"role_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"athlete@sportsentral.id"`
	Password string `json:"password" validate:"required" example:"Password123!"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type SendVerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email" example:"athlete@sportsentral.id"`
}

type VerifyAccountRequest struct {
	Token string `json:"token" validate:"required" example:"verification-token-uuid"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"athlete@sportsentral.id"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"athlete@sportsentral.id"`
	Username string `json:"username" validate:"required" example:"john_doe"`
	FullName string `json:"full_name" validate:"required" example:"John Doe"`
	Password string `json:"password" validate:"required,min=8" example:"Password123!"`
	RoleID   string `json:"role_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required" example:"reset-token-uuid"`
	Password string `json:"password" validate:"required,min=8" example:"NewPassword123!"`
}

type UpdateUserRequest struct {
	FullName *string `json:"full_name,omitempty" example:"John Doe Updated"`
	Username *string `json:"username,omitempty" example:"john_updated"`
}

type DeleteUserRequest struct {
	Password string `json:"password" validate:"required" example:"Password123!"`
}

type AssignRolesRequest struct {
	UserID string   `json:"user_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
	Roles  []string `json:"roles" validate:"required"`
}

type RemoveRolesRequest struct {
	UserID string   `json:"user_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
	Roles  []string `json:"roles" validate:"required"`
}

type ApplyProfileRequest struct {
	RoleID string `json:"role_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
}

type ToggleProfileRequest struct {
	RoleID string `json:"role_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
}

type ReviewProfileRequest struct {
	UserID string `json:"user_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
	RoleID string `json:"role_id" validate:"required" example:"01950000-0000-7000-8000-000000000001"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required" example:"Head Coach"`
	Description string   `json:"description" example:"Responsible for team roster and match strategy"`
	Permissions []string `json:"permissions"`
	Slug        string   `json:"slug" example:"head_coach"`
}

type UpdateRoleRequest struct {
	Name        *string  `json:"name,omitempty" example:"Head Coach Senior"`
	Description *string  `json:"description,omitempty" example:"Updated description"`
	Permissions []string `json:"permissions,omitempty"`
	Slug        *string  `json:"slug,omitempty" example:"head_coach_senior"`
}

type CreatePermissionRequest struct {
	Resource    string `json:"resource" validate:"required" example:"academy"`
	Action      string `json:"action" validate:"required" example:"create"`
	Description string `json:"description" example:"Can create new academy"`
	Slug        string `json:"slug" example:"academy:create"`
}

type UpdatePermissionRequest struct {
	Resource    *string `json:"resource,omitempty" example:"academy"`
	Action      *string `json:"action,omitempty" example:"update"`
	Description *string `json:"description,omitempty" example:"Can update academy"`
	Slug        *string `json:"slug,omitempty" example:"academy:update"`
}

// ─── AUTH RESPONSE DATA ───────────────────────────────────────────────────────

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponseData struct {
	Tokens     TokenPairResponse    `json:"tokens"`
	User       UserSimpleResponse   `json:"user"`
	Status     StatusSimpleResponse `json:"status"`
	ActiveRole RoleSimpleResponse   `json:"active_role"`
}
