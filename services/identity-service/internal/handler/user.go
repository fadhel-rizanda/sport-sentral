package handler

import (
	"context"
	"microservice-golang/services/identity-service/internal/dto"
	"microservice-golang/services/identity-service/internal/mapper"

	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	uc usecase.UserUseCase
}

func NewUserHandler(uc usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) RegisterGRPC(s *grpc.Server) {
	userv1.RegisterUserServiceServer(s, h)
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	roleId, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	res, err := h.uc.Create(ctx, dto.CreateUserRequest{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		Password: req.Password,
		RoleID:   roleId,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.CreateUserResponse{User: mapper.ToProtoUser(res)}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.GetUserResponse{User: mapper.ToProtoUser(res)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	res, err := h.uc.Update(ctx, dto.UpdateUserRequest{
		ID:       id,
		FullName: req.FullName,
		Username: req.Username,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.UpdateUserResponse{User: mapper.ToProtoUser(res)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	if err := h.uc.SoftDelete(ctx, dto.DeleteUserRequest{
		ID:       id,
		Password: req.Password,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.DeleteUserResponse{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	res, err := h.uc.List(ctx, dto.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	users := make([]*userv1.User, len(res.Users))
	for i, u := range res.Users {
		users[i] = mapper.ToProtoUser(u)
	}

	return &userv1.ListUsersResponse{
		Users:    users,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *UserHandler) SendVerifyEmail(ctx context.Context, req *userv1.SendVerifyEmailRequest) (*userv1.SendVerifyEmailResponse, error) {
	if err := h.uc.SendVerifyEmail(ctx, req.Email); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.SendVerifyEmailResponse{}, nil
}

func (h *UserHandler) VerifyAccount(ctx context.Context, req *userv1.VerifyAccountRequest) (*userv1.VerifyAccountResponse, error) {
	if err := h.uc.VerifyAccount(ctx, req.Token); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.VerifyAccountResponse{}, nil
}

func (h *UserHandler) ForgotPassword(ctx context.Context, req *userv1.ForgotPasswordRequest) (*userv1.ForgotPasswordResponse, error) {
	if err := h.uc.ForgotPassword(ctx, req.Email); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.ForgotPasswordResponse{}, nil
}

func (h *UserHandler) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.ResetPasswordResponse, error) {
	if err := h.uc.ResetPassword(ctx, req.Token, req.Password); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.ResetPasswordResponse{}, nil
}

func (h *UserHandler) AssignRolesToUser(ctx context.Context, req *userv1.AssignRolesToUserRequest) (*userv1.AssignRolesToUserResponse, error) {
	if err := h.uc.AssignRolesToUser(ctx, req.UserId, req.RoleIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.AssignRolesToUserResponse{}, nil
}

func (h *UserHandler) RemoveRolesFromUser(ctx context.Context, req *userv1.RemoveRolesFromUserRequest) (*userv1.RemoveRolesFromUserResponse, error) {
	if err := h.uc.RemoveRolesFromUser(ctx, req.UserId, req.RoleIds); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.RemoveRolesFromUserResponse{}, nil
}
