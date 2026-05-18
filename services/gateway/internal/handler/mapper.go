package handler

import (
	commonv1 "microservice-golang/gen/common/v1" // Package yang sebenarnya
	metav1 "microservice-golang/gen/meta/v1"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
	"time"
)

// ─── Helper Converters ────────────────────────────────────────────────────────

func toUserSimpleResponse(u *commonv1.UserSimple) UserSimpleResponse {
	if u == nil {
		return UserSimpleResponse{}
	}
	return UserSimpleResponse{
		ID:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func toStatusSimpleResponse(s *commonv1.StatusSimple) StatusSimpleResponse {
	if s == nil {
		return StatusSimpleResponse{}
	}
	return StatusSimpleResponse{
		ID:   s.Id,
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func toTagSimpleResponse(t *commonv1.TagSimple) TagSimpleResponse {
	if t == nil {
		return TagSimpleResponse{}
	}
	return TagSimpleResponse{
		ID:   t.Id,
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func toPermissionSimpleResponse(p *commonv1.PermissionSimple) PermissionSimpleResponse {
	if p == nil {
		return PermissionSimpleResponse{}
	}
	return PermissionSimpleResponse{
		ID:       p.Id,
		Resource: p.Resource,
		Action:   p.Action,
		Slug:     p.Slug,
	}
}

func toRoleSimpleResponse(r *commonv1.RoleSimple) RoleSimpleResponse {
	if r == nil {
		return RoleSimpleResponse{}
	}

	permissions := make([]PermissionSimpleResponse, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = toPermissionSimpleResponse(p)
	}

	return RoleSimpleResponse{
		ID:          r.Id,
		Name:        r.Name,
		Slug:        r.Slug,
		Permissions: permissions,
	}
}

// ─── Main Converters ──────────────────────────────────────────────────────────

func toPermissionResponse(permission *rbacv1.Permission) PermissionResponse {
	res := PermissionResponse{
		ID:          permission.Id,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		Slug:        permission.Slug,
		CreatedBy:   toUserSimpleResponse(permission.CreatedBy),
		CreatedAt:   permission.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   permission.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if permission.UpdatedBy != nil {
		updatedBy := toUserSimpleResponse(permission.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if permission.DeletedAt != nil {
		deletedAt := permission.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if permission.DeletedBy != nil {
			deletedBy := toUserSimpleResponse(permission.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func toRoleResponse(role *rbacv1.Role) RoleResponse {
	permissions := make([]PermissionSimpleResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = toPermissionSimpleResponse(p)
	}

	res := RoleResponse{
		ID:          role.Id,
		Name:        role.Name,
		Description: role.Description,
		Slug:        role.Slug,
		Permissions: permissions,
		CreatedBy:   toUserSimpleResponse(role.CreatedBy),
		CreatedAt:   role.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   role.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if role.UpdatedBy != nil {
		updatedBy := toUserSimpleResponse(role.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if role.DeletedAt != nil {
		deletedAt := role.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if role.DeletedBy != nil {
			deletedBy := toUserSimpleResponse(role.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func toStatusResponse(status *metav1.Status) StatusResponse {
	res := StatusResponse{
		ID:        status.Id,
		Type:      status.Type,
		Name:      status.Name,
		Slug:      status.Slug,
		CreatedBy: toUserSimpleResponse(status.CreatedBy),
		CreatedAt: status.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt: status.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if status.UpdatedBy != nil {
		updatedBy := toUserSimpleResponse(status.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if status.DeletedAt != nil {
		deletedAt := status.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if status.DeletedBy != nil {
			deletedBy := toUserSimpleResponse(status.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func toTagResponse(tag *metav1.Tag) TagResponse {
	res := TagResponse{
		ID:        tag.Id,
		Type:      tag.Type,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedBy: toUserSimpleResponse(tag.CreatedBy),
		CreatedAt: tag.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt: tag.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	if tag.UpdatedBy != nil {
		updatedBy := toUserSimpleResponse(tag.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if tag.DeletedAt != nil {
		deletedAt := tag.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if tag.DeletedBy != nil {
			deletedBy := toUserSimpleResponse(tag.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func toUserResponse(u *userv1.User) UserResponse {
	roles := make([]UserRoleSimpleResponse, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = UserRoleSimpleResponse{
			Role:     toRoleSimpleResponse(r.Role),
			Status:   toStatusSimpleResponse(r.Status),
			IsActive: r.IsActive,
		}
	}

	var verifiedAt *string
	if u.VerifiedAt != nil {
		t := u.VerifiedAt.AsTime().UTC().Format(time.RFC3339)
		verifiedAt = &t
	}

	res := UserResponse{
		ID:         u.Id,
		Email:      u.Email,
		Username:   u.Username,
		FullName:   u.FullName,
		Status:     toStatusSimpleResponse(u.Status),
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
