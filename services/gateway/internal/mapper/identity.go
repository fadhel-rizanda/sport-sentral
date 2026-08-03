package mapper

import (
	"time"

	commonv1 "microservice-golang/gen/common/v1"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToUserSimpleResponse(u *commonv1.UserSimple) dto.UserSimpleResponse {
	if u == nil {
		return dto.UserSimpleResponse{}
	}
	return dto.UserSimpleResponse{
		ID:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToPermissionSimpleResponse(p *commonv1.PermissionSimple) dto.PermissionSimpleResponse {
	if p == nil {
		return dto.PermissionSimpleResponse{}
	}
	return dto.PermissionSimpleResponse{
		ID:       p.Id,
		Resource: p.Resource,
		Action:   p.Action,
		Slug:     p.Slug,
	}
}

func ToRoleSimpleResponse(r *commonv1.RoleSimple) dto.RoleSimpleResponse {
	if r == nil {
		return dto.RoleSimpleResponse{}
	}

	permissions := make([]dto.PermissionSimpleResponse, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = ToPermissionSimpleResponse(p)
	}

	return dto.RoleSimpleResponse{
		ID:          r.Id,
		Name:        r.Name,
		Slug:        r.Slug,
		Permissions: permissions,
	}
}

func ToPermissionResponse(permission *rbacv1.Permission) dto.PermissionResponse {
	res := dto.PermissionResponse{
		ID:          permission.Id,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		Slug:        permission.Slug,
		CreatedBy:   ToUserSimpleResponse(permission.CreatedBy),
		CreatedAt:   permission.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   permission.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if permission.UpdatedBy != nil {
		updatedBy := ToUserSimpleResponse(permission.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if permission.DeletedAt != nil {
		deletedAt := permission.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if permission.DeletedBy != nil {
			deletedBy := ToUserSimpleResponse(permission.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func ToRoleResponse(role *rbacv1.Role) dto.RoleResponse {
	permissions := make([]dto.PermissionSimpleResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = ToPermissionSimpleResponse(p)
	}

	res := dto.RoleResponse{
		ID:          role.Id,
		Name:        role.Name,
		Description: role.Description,
		Slug:        role.Slug,
		Permissions: permissions,
		CreatedBy:   ToUserSimpleResponse(role.CreatedBy),
		CreatedAt:   role.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   role.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if role.UpdatedBy != nil {
		updatedBy := ToUserSimpleResponse(role.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if role.DeletedAt != nil {
		deletedAt := role.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if role.DeletedBy != nil {
			deletedBy := ToUserSimpleResponse(role.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func ToUserResponse(u *userv1.User) dto.UserResponse {
	roles := make([]dto.UserRoleSimpleResponse, len(u.Roles))
	for i, r := range u.Roles {
		var expiredAt *string
		if r.ExpiredAt != nil {
			t := r.ExpiredAt.AsTime().UTC().Format(time.RFC3339)
			expiredAt = &t
		}
		roles[i] = dto.UserRoleSimpleResponse{
			Role:      ToRoleSimpleResponse(r.Role),
			Status:    ToStatusSimpleResponse(r.Status),
			ExpiredAt: expiredAt,
		}
	}

	var verifiedAt *string
	if u.VerifiedAt != nil {
		t := u.VerifiedAt.AsTime().UTC().Format(time.RFC3339)
		verifiedAt = &t
	}

	res := dto.UserResponse{
		ID:         u.Id,
		Email:      u.Email,
		Username:   u.Username,
		FullName:   u.FullName,
		Status:     ToStatusSimpleResponse(u.Status),
		Roles:      roles,
		VerifiedAt: verifiedAt,
		CreatedAt:  u.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:  u.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	if u.DeletedAt != nil {
		deletedAt := u.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
	}

	return res
}
