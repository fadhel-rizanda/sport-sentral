package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type RoleHandler struct {
	rbacv1.UnimplementedRBACServiceServer
	roleUC       usecase.RoleUseCase
	permissionUC usecase.PermissionUseCase
}

func NewRoleHandler(roleUC usecase.RoleUseCase, permissionUC usecase.PermissionUseCase) *RoleHandler {
	return &RoleHandler{
		roleUC:       roleUC,
		permissionUC: permissionUC,
	}
}

func (h *RoleHandler) RegisterGRPC(s *grpc.Server) {
	rbacv1.RegisterRBACServiceServer(s, h)
}

// ─── Role CRUD ────────────────────────────────────────────────────────────────

func (h *RoleHandler) CreateRole(ctx context.Context, req *rbacv1.CreateRoleRequest) (*rbacv1.CreateRoleResponse, error) {
	role, err := h.roleUC.CreateRole(ctx, req.Name, req.Description, req.PermissionIds)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.CreateRoleResponse{Role: toProtoRole(role)}, nil
}

func (h *RoleHandler) GetRole(ctx context.Context, req *rbacv1.GetRoleRequest) (*rbacv1.GetRoleResponse, error) {
	role, err := h.roleUC.GetRole(ctx, req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.GetRoleResponse{Role: toProtoRole(role)}, nil
}

func (h *RoleHandler) UpdateRole(ctx context.Context, req *rbacv1.UpdateRoleRequest) (*rbacv1.UpdateRoleResponse, error) {
	role, err := h.roleUC.UpdateRole(ctx, req.Id, req.Name, req.Description, req.PermissionIds)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.UpdateRoleResponse{Role: toProtoRole(role)}, nil
}

func (h *RoleHandler) DeleteRole(ctx context.Context, req *rbacv1.DeleteRoleRequest) (*rbacv1.DeleteRoleResponse, error) {
	if err := h.roleUC.DeleteRole(ctx, req.Id); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.DeleteRoleResponse{}, nil
}

func (h *RoleHandler) ListRoles(ctx context.Context, req *rbacv1.ListRolesRequest) (*rbacv1.ListRolesResponse, error) {
	roles, total, err := h.roleUC.ListRoles(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	protoRoles := make([]*rbacv1.Role, len(roles))
	for i, r := range roles {
		protoRoles[i] = toProtoRole(r)
	}

	return &rbacv1.ListRolesResponse{
		Roles:    protoRoles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ─── Permission CRUD ──────────────────────────────────────────────────────────

func (h *RoleHandler) CreatePermission(ctx context.Context, req *rbacv1.CreatePermissionRequest) (*rbacv1.CreatePermissionResponse, error) {
	permission, err := h.permissionUC.CreatePermission(ctx, req.Resource, req.Action, req.Description)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.CreatePermissionResponse{Permission: toProtoPermission(permission)}, nil
}

func (h *RoleHandler) GetPermission(ctx context.Context, req *rbacv1.GetPermissionRequest) (*rbacv1.GetPermissionResponse, error) {
	permission, err := h.permissionUC.GetPermission(ctx, req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.GetPermissionResponse{Permission: toProtoPermission(permission)}, nil
}

func (h *RoleHandler) DeletePermission(ctx context.Context, req *rbacv1.DeletePermissionRequest) (*rbacv1.DeletePermissionResponse, error) {
	if err := h.permissionUC.DeletePermission(ctx, req.Id); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.DeletePermissionResponse{}, nil
}

func (h *RoleHandler) ListPermissions(ctx context.Context, req *rbacv1.ListPermissionsRequest) (*rbacv1.ListPermissionsResponse, error) {
	permissions, total, err := h.permissionUC.ListPermissions(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	protoPermissions := make([]*rbacv1.Permission, len(permissions))
	for i, p := range permissions {
		protoPermissions[i] = toProtoPermission(p)
	}

	return &rbacv1.ListPermissionsResponse{
		Permissions: protoPermissions,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// ─── Assignment ───────────────────────────────────────────────────────────────

func (h *RoleHandler) AssignPermissionToRole(ctx context.Context, req *rbacv1.AssignPermissionToRoleRequest) (*rbacv1.AssignPermissionToRoleResponse, error) {
	if err := h.roleUC.AssignPermission(ctx, req.RoleId, req.PermissionIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.AssignPermissionToRoleResponse{}, nil
}

func (h *RoleHandler) RevokePermissionFromRole(ctx context.Context, req *rbacv1.RevokePermissionFromRoleRequest) (*rbacv1.RevokePermissionFromRoleResponse, error) {
	if err := h.roleUC.RevokePermission(ctx, req.RoleId, req.PermissionIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.RevokePermissionFromRoleResponse{}, nil
}

func (h *RoleHandler) AssignRolesToUser(ctx context.Context, req *rbacv1.AssignRolesToUserRequest) (*rbacv1.AssignRolesToUserResponse, error) {
	if err := h.roleUC.AssignRolesToUser(ctx, req.UserId, req.RoleIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.AssignRolesToUserResponse{}, nil
}

func (h *RoleHandler) RemoveRolesFromUser(ctx context.Context, req *rbacv1.RemoveRolesFromUserRequest) (*rbacv1.RemoveRolesFromUserResponse, error) {
	if err := h.roleUC.RemoveRolesFromUser(ctx, req.UserId, req.RoleIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.RemoveRolesFromUserResponse{}, nil
}

// ─── Authorization Check ──────────────────────────────────────────────────────

func (h *RoleHandler) CheckPermission(ctx context.Context, req *rbacv1.CheckPermissionRequest) (*rbacv1.CheckPermissionResponse, error) {
	allowed, err := h.permissionUC.CheckPermission(ctx, req.UserId, req.Resource, req.Action)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.CheckPermissionResponse{Allowed: allowed}, nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func toProtoRole(r *entity.Role) *rbacv1.Role {
	protoPermissions := make([]*rbacv1.Permission, len(r.Permissions))
	for i, p := range r.Permissions {
		protoPermissions[i] = toProtoPermission(p)
	}
	return &rbacv1.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		Permissions: protoPermissions,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}
