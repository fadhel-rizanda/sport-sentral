package handler

import (
	"context"
	"microservice-golang/services/identity-service/internal/dto"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type ProfileHandler struct {
	rbacv1.UnimplementedRBACServiceServer
	uc usecase.ProfileUseCase
}

func NewProfileHandler(uc usecase.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

func (h *ProfileHandler) RegisterGRPC(s *grpc.Server) {
	rbacv1.RegisterRBACServiceServer(s, h)
}

func (h *ProfileHandler) ApplyProfile(ctx context.Context, req *rbacv1.ApplyProfileRequest) (*rbacv1.ApplyProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	if err := h.uc.ApplyProfile(ctx, dto.ApplyProfileRequest{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.ApplyProfileResponse{}, nil
}

func (h *ProfileHandler) ToggleProfile(ctx context.Context, req *rbacv1.ToggleProfileRequest) (*rbacv1.ToggleProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	if err := h.uc.ToggleProfile(ctx, dto.ToggleProfileRequest{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.ToggleProfileResponse{}, nil
}

func (h *ProfileHandler) ApproveProfile(ctx context.Context, req *rbacv1.ApproveProfileRequest) (*rbacv1.ApproveProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	if err := h.uc.ApproveProfile(ctx, dto.ApproveProfileRequest{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.ApproveProfileResponse{}, nil
}

func (h *ProfileHandler) RejectProfile(ctx context.Context, req *rbacv1.RejectProfileRequest) (*rbacv1.RejectProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	if err := h.uc.RejectProfile(ctx, dto.ApproveProfileRequest{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &rbacv1.RejectProfileResponse{}, nil
}
