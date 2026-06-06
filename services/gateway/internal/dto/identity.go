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
	Role     RoleSimpleResponse   `json:"role"`
	Status   StatusSimpleResponse `json:"status"`
	IsActive bool                 `json:"is_active"`
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
