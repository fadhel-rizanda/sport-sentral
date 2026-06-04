package mapper

import (
	commonv1 "microservice-golang/gen/common/v1"
	metav1 "microservice-golang/gen/meta/v1"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/identity-service/internal/dto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── Helper Converters ────────────────────────────────────────────────────────

func ToProtoUserSimple(u dto.UserSimpleResponse) *commonv1.UserSimple {
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToProtoStatusSimple(s dto.StatusSimpleResponse) *commonv1.StatusSimple {
	return &commonv1.StatusSimple{
		Id:   s.ID.String(),
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToProtoTagSimple(t dto.TagSimpleResponse) *metav1.TagSimple {
	return &metav1.TagSimple{
		Id:   t.ID.String(),
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToProtoPermissionSimple(p dto.PermissionSimpleResponse) *commonv1.PermissionSimple {
	return &commonv1.PermissionSimple{
		Id:       p.ID.String(),
		Resource: p.Resource,
		Action:   p.Action,
		Slug:     p.Slug,
	}
}

func ToProtoRoleSimple(r dto.RoleSimpleResponse) *commonv1.RoleSimple {
	permissions := make([]*commonv1.PermissionSimple, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = ToProtoPermissionSimple(p)
	}

	return &commonv1.RoleSimple{
		Id:          r.ID.String(),
		Name:        r.Name,
		Slug:        r.Slug,
		Permissions: permissions,
	}
}

// ─── Main Converters ──────────────────────────────────────────────────────────

func ToProtoPermission(r *dto.PermissionResponse) *rbacv1.Permission {
	res := &rbacv1.Permission{
		Id:          r.ID.String(),
		Action:      r.Action,
		Resource:    r.Resource,
		Description: r.Description,
		Slug:        r.Slug,
		CreatedBy:   ToProtoUserSimple(r.CreatedBy),
		UpdatedBy:   ToProtoUserSimple(r.UpdatedBy),
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}

	if r.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*r.DeletedAt)
		if r.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*r.DeletedBy)
		}
	}

	return res
}

func ToProtoRole(r *dto.RoleResponse) *rbacv1.Role {
	permissions := make([]*commonv1.PermissionSimple, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = ToProtoPermissionSimple(p)
	}

	res := &rbacv1.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		Slug:        r.Slug,
		Permissions: permissions,
		CreatedBy:   ToProtoUserSimple(r.CreatedBy),
		UpdatedBy:   ToProtoUserSimple(r.UpdatedBy),
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}

	if r.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*r.DeletedAt)
		if r.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*r.DeletedBy)
		}
	}

	return res
}

func ToProtoUser(u *dto.UserResponse) *userv1.User {
	var verifiedAt *timestamppb.Timestamp
	if u.VerifiedAt != nil {
		verifiedAt = timestamppb.New(*u.VerifiedAt)
	}

	roles := make([]*userv1.UserRole, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = &userv1.UserRole{
			Role:     ToProtoRoleSimple(r.Role),
			Status:   ToProtoStatusSimple(r.Status),
			IsActive: r.IsActive,
		}
	}

	res := &userv1.User{
		Id:         u.ID.String(),
		Email:      u.Email,
		Username:   u.Username,
		FullName:   u.FullName,
		Status:     ToProtoStatusSimple(u.Status),
		Roles:      roles,
		CreatedAt:  timestamppb.New(u.CreatedAt),
		UpdatedAt:  timestamppb.New(u.UpdatedAt),
		VerifiedAt: verifiedAt,
	}

	if u.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*u.DeletedAt)
	}

	return res
}

func ToProtoUserInternal(u *dto.UserResponse) *userv1.UserInternal {
	roleIds := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roleIds[i] = r.Role.ID.String()
	}

	return &userv1.UserInternal{
		Id:         u.ID.String(),
		Email:      u.Email,
		Username:   u.Username,
		FullName:   u.FullName,
		ActiveRole: ToProtoRoleSimple(u.ActiveRole),
		Status:     ToProtoStatusSimple(u.Status),
	}
}
