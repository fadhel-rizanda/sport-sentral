package usecase

import (
	"microservice-golang/services/identity-service/internal/entity"
)

func ToUserResponse(user *entity.User) *UserResponse {
	roles := make([]UserRoleSimpleResponse, len(user.UserRoles))
	var activeRole RoleSimpleResponse

	for i, ur := range user.UserRoles {
		permissions := make([]PermissionSimpleResponse, len(ur.Role.Permissions))
		for j, p := range ur.Role.Permissions {
			permissions[j] = PermissionSimpleResponse{
				ID:       p.ID,
				Resource: p.Resource,
				Action:   p.Action,
				Slug:     p.Slug,
			}
		}

		roleSimple := RoleSimpleResponse{
			ID:          ur.Role.ID,
			Name:        ur.Role.Name,
			Slug:        ur.Role.Slug,
			Permissions: permissions,
		}

		roles[i] = UserRoleSimpleResponse{
			Role: roleSimple,
			Status: StatusSimpleResponse{
				ID:   ur.Status.ID,
				Type: ur.Status.Type,
				Name: ur.Status.Name,
				Slug: ur.Status.Slug,
			},
			IsActive: ur.IsActive,
		}

		if ur.IsActive {
			activeRole = roleSimple
		}
	}

	res := &UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Status: StatusSimpleResponse{
			ID:   user.Status.ID,
			Type: user.Status.Type,
			Name: user.Status.Name,
			Slug: user.Status.Slug,
		},
		ActiveRole: activeRole,
		Roles:      roles,
		VerifiedAt: user.VerifiedAt,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	if user.DeletedAt.Valid {
		res.DeletedAt = &user.DeletedAt.Time
	}

	return res
}

func ToRoleResponse(role *entity.Role) *RoleResponse {
	permissions := make([]PermissionSimpleResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = PermissionSimpleResponse{
			ID:       p.ID,
			Resource: p.Resource,
			Action:   p.Action,
			Slug:     p.Slug,
		}
	}

	res := &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Slug:        role.Slug,
		Permissions: permissions,
		CreatedBy: UserSimpleResponse{
			ID:       role.CreatedByID,
			Email:    role.CreatedBy.Email,
			Username: role.CreatedBy.Username,
			FullName: role.CreatedBy.FullName,
		},
		UpdatedBy: UserSimpleResponse{
			ID:       role.UpdatedByID,
			Email:    role.UpdatedBy.Email,
			Username: role.UpdatedBy.Username,
			FullName: role.UpdatedBy.FullName,
		},
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}

	if role.DeletedAt.Valid {
		res.DeletedAt = &role.DeletedAt.Time
		if role.DeletedByID != nil {
			res.DeletedBy = &UserSimpleResponse{
				ID:       *role.DeletedByID,
				Email:    role.DeletedBy.Email,
				Username: role.DeletedBy.Username,
				FullName: role.DeletedBy.FullName,
			}
		}
	}

	return res
}

func ToPermissionResponse(permission *entity.Permission) *PermissionResponse {
	res := &PermissionResponse{
		ID:          permission.ID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		Slug:        permission.Slug,
		CreatedBy: UserSimpleResponse{
			ID:       permission.CreatedByID,
			Email:    permission.CreatedBy.Email,
			Username: permission.CreatedBy.Username,
			FullName: permission.CreatedBy.FullName,
		},
		UpdatedBy: UserSimpleResponse{
			ID:       permission.UpdatedByID,
			Email:    permission.UpdatedBy.Email,
			Username: permission.UpdatedBy.Username,
			FullName: permission.UpdatedBy.FullName,
		},
		CreatedAt: permission.CreatedAt,
		UpdatedAt: permission.UpdatedAt,
	}

	if permission.DeletedAt.Valid {
		res.DeletedAt = &permission.DeletedAt.Time
		if permission.DeletedByID != nil {
			res.DeletedBy = &UserSimpleResponse{
				ID:       *permission.DeletedByID,
				Email:    permission.DeletedBy.Email,
				Username: permission.DeletedBy.Username,
				FullName: permission.DeletedBy.FullName,
			}
		}
	}

	return res
}
