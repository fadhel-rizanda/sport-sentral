package usecase

import (
	"github.com/google/uuid"
	"microservice-golang/services/identity-service/internal/entity"
)

func ToUserResponse(user *entity.User) *UserResponse {
	roles := make([]UserRoleResponse, len(user.UserRoles))
	activeRoleName := ""
	var activeRoleID uuid.UUID

	for i, ur := range user.UserRoles {
		roles[i] = UserRoleResponse{
			ID:         ur.RoleID,
			Name:       ur.Role.Name,
			IsActive:   ur.IsActive,
			StatusName: ur.Status.Name,
			StatusID:   ur.Status.ID,
		}
		if ur.IsActive {
			activeRoleName = ur.Role.Name
			activeRoleID = ur.Role.ID
		}
	}

	res := &UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		FullName:       user.FullName,
		StatusName:     user.Status.Name,
		StatusID:       user.Status.ID,
		ActiveRoleName: activeRoleName,
		ActiveRoleID:   activeRoleID,
		Roles:          roles,
		VerifiedAt:     user.VerifiedAt,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
	if user.DeletedAt.Valid {
		res.DeletedAt = &user.DeletedAt.Time
	}
	return res
}

func (u *UserResponse) RoleIDs() []string {
	ids := make([]string, len(u.Roles))
	for i, p := range u.Roles {
		ids[i] = p.ID.String()
	}
	return ids
}

func ToRoleResponse(role *entity.Role) *RoleResponse {
	permissions := make([]PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {

		permissions[i] = PermissionResponse{
			ID:          p.ID,
			Action:      p.Action,
			Resource:    p.Resource,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			CreatedByID: p.CreatedByID,
			UpdatedByID: p.UpdatedByID,
			Description: p.Description,
		}
		if p.DeletedAt.Valid {
			permissions[i].DeletedAt = &p.DeletedAt.Time
			permissions[i].DeletedByID = p.DeletedByID
		}
	}

	res := &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		CreatedByID: role.CreatedByID,
		UpdatedByID: role.UpdatedByID,
	}
	if res.DeletedAt != nil {
		res.DeletedAt = &role.DeletedAt.Time
		res.DeletedByID = role.DeletedByID
	}
	return res
}

func ToPermissionResponse(permission *entity.Permission) *PermissionResponse {
	res := &PermissionResponse{
		ID:          permission.ID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
		CreatedByID: permission.CreatedByID,
		UpdatedByID: permission.UpdatedByID,
	}
	if res.DeletedAt != nil {
		res.DeletedAt = &permission.DeletedAt.Time
		res.DeletedByID = permission.DeletedByID
	}
	return res
}
