package handler

import (
	"context"
	"github.com/google/uuid"

	"google.golang.org/grpc"

	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type UserInternalHandler struct {
	userv1.UnimplementedUserInternalServiceServer
	uc usecase.UserUseCase
}

func NewUserInternalHandler(uc usecase.UserUseCase) *UserInternalHandler {
	return &UserInternalHandler{uc: uc}
}

func (h *UserInternalHandler) RegisterGRPC(s *grpc.Server) {
	userv1.RegisterUserInternalServiceServer(s, h)
}

func (h *UserInternalHandler) GetUserByEmailInternal(ctx context.Context, req *userv1.GetUserByEmailInternalRequest) (*userv1.GetUserByEmailInternalResponse, error) {
	user, err := h.uc.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.GetUserByEmailInternalResponse{
		User: toProtoUserInternal(user),
	}, nil
}

func (h *UserInternalHandler) GetUserByIDInternal(ctx context.Context, req *userv1.GetUserByIDInternalRequest) (*userv1.GetUserByIDInternalResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	user, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.GetUserByIDInternalResponse{
		User: toProtoUserInternal(user),
	}, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toProtoUserInternal(u *usecase.UserResponse) *userv1.UserInternal {
	return &userv1.UserInternal{
		Id:            u.ID.String(),
		Email:         u.Email,
		Username:      u.Username,
		FullName:      u.FullName,
		Status:        u.Status,
		ActiveProfile: u.ActiveProfile,
		RoleIds:       u.RoleIDs(),
	}
}
