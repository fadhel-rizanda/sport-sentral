package handler

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/dto"
	"microservice-golang/services/identity-service/internal/mapper"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type RoleHandler struct {
	rbacv1.UnimplementedRoleServiceServer
	uc usecase.RoleUseCase
}

func NewRoleHandler(uc usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{uc: uc}
}

func (h *RoleHandler) RegisterGRPC(s *grpc.Server) {
	rbacv1.RegisterRoleServiceServer(s, h)
}

func (h *RoleHandler) GetRole(ctx context.Context, req *rbacv1.GetRoleRequest) (*rbacv1.GetRoleResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid permission id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.GetRoleResponse{
		Role: mapper.ToProtoRole(res),
	}, nil
}

func (h *RoleHandler) CreateRole(ctx context.Context, req *rbacv1.CreateRoleRequest) (*rbacv1.CreateRoleResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}
	permissionIDs := make([]*uuid.UUID, len(req.PermissionIds))
	for i, permissionId := range req.PermissionIds {
		id, err := uuid.Parse(permissionId)
		if err != nil {
			return nil, apperr.ToGRPC(err)
		}
		permissionIDs[i] = &id
	}
	res, err := h.uc.Create(ctx, dto.CreateRoleRequest{
		Name:          req.GetName(),
		Description:   req.Description,
		CreatedByID:   createdBy,
		Slug:          req.Slug,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.CreateRoleResponse{Role: mapper.ToProtoRole(res)}, nil
}

func (h *RoleHandler) UpdateRole(ctx context.Context, req *rbacv1.UpdateRoleRequest) (*rbacv1.UpdateRoleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	updatedBy, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated_by id"))
	}
	permissionIDs := make([]*uuid.UUID, len(req.PermissionIds))
	for i, permissionId := range req.PermissionIds {
		id, err := uuid.Parse(permissionId)
		if err != nil {
			return nil, apperr.ToGRPC(err)
		}
		permissionIDs[i] = &id
	}
	res, err := h.uc.Update(ctx, dto.UpdateRoleRequest{
		ID:            id,
		Name:          req.Name,
		Description:   req.Description,
		UpdatedByID:   updatedBy,
		Slug:          req.Slug,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.UpdateRoleResponse{
		Role: mapper.ToProtoRole(res),
	}, nil
}

func (h *RoleHandler) DeleteRole(ctx context.Context, req *rbacv1.DeleteRoleRequest) (*rbacv1.DeleteRoleResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}
	deletedBy, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted_by id"))
	}

	if req.GetIsPermanent() {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("permanent delete is not allowed for roles"))
	} else {
		err = h.uc.SoftDelete(ctx, dto.DeleteRequest{
			ID:          id,
			DeletedByID: deletedBy,
		})
	}
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.DeleteRoleResponse{}, nil
}

func (h *RoleHandler) ListRoles(ctx context.Context, req *rbacv1.ListRolesRequest) (*rbacv1.ListRolesResponse, error) {
	res, err := h.uc.List(ctx, dto.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	roles := make([]*rbacv1.Role, 0, res.Total)
	for _, role := range res.Roles {
		roles = append(roles, mapper.ToProtoRole(role))
	}
	return &rbacv1.ListRolesResponse{
		Roles:    roles,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *RoleHandler) AssignPermissionToRole(ctx context.Context, req *rbacv1.AssignPermissionToRoleRequest) (*rbacv1.AssignPermissionToRoleResponse, error) {
	if err := h.uc.AssignPermission(ctx, req.RoleId, req.PermissionIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.AssignPermissionToRoleResponse{}, nil
}

func (h *RoleHandler) RevokePermissionFromRole(ctx context.Context, req *rbacv1.RevokePermissionFromRoleRequest) (*rbacv1.RevokePermissionFromRoleResponse, error) {
	if err := h.uc.RevokePermission(ctx, req.RoleId, req.PermissionIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.RevokePermissionFromRoleResponse{}, nil
}
