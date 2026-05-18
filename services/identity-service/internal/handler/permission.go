package handler

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type PermissionHandler struct {
	rbacv1.UnimplementedPermissionServiceServer
	uc usecase.PermissionUseCase
}

func NewPermissionHandler(uc usecase.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{uc: uc}
}

func (h *PermissionHandler) RegisterGRPC(s *grpc.Server) {
	rbacv1.RegisterPermissionServiceServer(s, h)
}

func (h *PermissionHandler) CreatePermission(ctx context.Context, req *rbacv1.CreatePermissionRequest) (*rbacv1.CreatePermissionResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}
	res, err := h.uc.Create(ctx, usecase.CreatePermissionRequest{
		Action:      req.GetAction(),
		Resource:    req.GetResource(),
		Description: req.GetDescription(),
		CreatedByID: createdBy,
		Slug:        req.Slug,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.CreatePermissionResponse{Permission: toProtoPermission(res)}, nil
}

func (h *PermissionHandler) GetPermission(ctx context.Context, req *rbacv1.GetPermissionRequest) (*rbacv1.GetPermissionResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid permission id"))
	}
	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.GetPermissionResponse{
		Permission: toProtoPermission(res),
	}, nil
}

func (h *PermissionHandler) UpdatePermission(ctx context.Context, req *rbacv1.UpdatePermissionRequest) (*rbacv1.UpdatePermissionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid permission id"))
	}

	updatedBy, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated_by id"))
	}
	res, err := h.uc.Update(ctx, usecase.UpdatePermissionRequest{
		ID:          id,
		Action:      req.Action,
		Resource:    req.Resource,
		Description: req.Description,
		UpdatedByID: updatedBy,
		Slug:        req.Slug,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.UpdatePermissionResponse{
		Permission: toProtoPermission(res),
	}, nil
}

func (h *PermissionHandler) DeletePermission(ctx context.Context, req *rbacv1.DeletePermissionRequest) (*rbacv1.DeletePermissionResponse, error) {
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
	return &rbacv1.DeletePermissionResponse{}, nil
}

func (h *PermissionHandler) ListPermissions(ctx context.Context, req *rbacv1.ListPermissionsRequest) (*rbacv1.ListPermissionsResponse, error) {
	var roleId *uuid.UUID
	if req.GetRoleId() != "" {
		id, err := uuid.Parse(*req.RoleId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
		}
		roleId = &id
	}
	res, err := h.uc.List(ctx, usecase.ListPermissionRequest{
		RoleId:   roleId,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	permissions := make([]*rbacv1.Permission, 0, res.Total)
	for _, p := range res.Permissions {
		permissions = append(permissions, toProtoPermission(p))
	}
	return &rbacv1.ListPermissionsResponse{
		Permissions: permissions,
		Page:        int32(res.Page),
		PageSize:    int32(res.PageSize),
	}, nil
}

func (h *PermissionHandler) CheckPermission(ctx context.Context, req *rbacv1.CheckPermissionRequest) (*rbacv1.CheckPermissionResponse, error) {
	allowed, err := h.uc.CheckPermission(ctx, req.GetUserId(), req.GetResource(), req.GetAction())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &rbacv1.CheckPermissionResponse{Allowed: allowed}, nil
}
