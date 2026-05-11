package handler

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
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
		Role: toProtoRole(res),
	}, nil
}

func (h *RoleHandler) CreateRole(ctx context.Context, req *rbacv1.CreateRoleRequest) (*rbacv1.CreateRoleResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}
	res, err := h.uc.Create(ctx, usecase.CreateRoleRequest{
		Name:        req.GetName(),
		Description: req.Description,
		CreatedByID: createdBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.CreateRoleResponse{Role: toProtoRole(res)}, nil
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

	res, err := h.uc.Update(ctx, usecase.UpdateRoleRequest{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		UpdatedByID: updatedBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.UpdateRoleResponse{
		Role: toProtoRole(res),
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
		err = h.uc.SoftDelete(ctx, usecase.DeleteRequest{
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
	res, err := h.uc.List(ctx, usecase.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	roles := make([]*rbacv1.Role, 0, res.Total)
	for _, role := range res.Roles {
		roles = append(roles, toProtoRole(role))
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toProtoRole(r *usecase.RoleResponse) *rbacv1.Role {
	permissions := make([]*rbacv1.Permission, len(r.Permissions))
	for i, p := range r.Permissions {
		createdAt := timestamppb.New(p.CreatedAt)
		updatedAt := timestamppb.New(p.UpdatedAt)

		permissions[i] = &rbacv1.Permission{
			Id:          p.ID.String(),
			Action:      p.Action,
			Resource:    p.Resource,
			Description: p.Description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			CreatedById: p.CreatedByID.String(),
			UpdatedById: p.UpdatedByID.String(),
		}
		if p.DeletedAt != nil {
			deletedByID := p.DeletedByID.String()
			permissions[i].DeletedAt = timestamppb.New(*p.DeletedAt)
			permissions[i].DeletedById = &deletedByID
		}
	}

	createdAt := timestamppb.New(r.CreatedAt)
	updatedAt := timestamppb.New(r.UpdatedAt)

	res := &rbacv1.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Permissions: permissions,
		Description: r.Description,
		CreatedById: r.CreatedByID.String(),
		UpdatedById: r.UpdatedByID.String(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	if r.DeletedAt != nil {
		deletedByID := r.DeletedByID.String()
		res.DeletedAt = timestamppb.New(*r.DeletedAt)
		res.DeletedById = &deletedByID
	}
	return res
}
