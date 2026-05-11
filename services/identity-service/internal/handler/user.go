package handler

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
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
	res, err := h.uc.Create(ctx, usecase.CreateUserRequest{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		Password: req.Password,
		RoleName: req.RoleName,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.CreateUserResponse{User: toProtoUser(res)}, nil
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

	return &userv1.GetUserResponse{User: toProtoUser(res)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	res, err := h.uc.Update(ctx, usecase.UpdateUserRequest{
		ID:       id,
		FullName: req.FullName,
		Username: req.Username,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.UpdateUserResponse{User: toProtoUser(res)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	if err := h.uc.SoftDelete(ctx, usecase.DeleteUserRequest{
		ID:       id,
		Password: req.Password,
	}); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &userv1.DeleteUserResponse{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	res, err := h.uc.List(ctx, usecase.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	users := make([]*userv1.User, len(res.Users))
	for i, u := range res.Users {
		users[i] = toProtoUser(u)
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toProtoUser(u *usecase.UserResponse) *userv1.User {
	var verifiedAt *timestamppb.Timestamp
	if u.VerifiedAt != nil {
		verifiedAt = timestamppb.New(*u.VerifiedAt)
	}

	profiles := make([]*userv1.UserRole, len(u.Roles))
	for i, p := range u.Roles {
		profiles[i] = &userv1.UserRole{
			RoleId:     p.ID.String(),
			RoleName:   p.Name,
			IsActive:   p.IsActive,
			StatusName: p.StatusName,
			StatusId:   p.StatusID.String(),
		}
	}

	res := &userv1.User{
		Id:             u.ID.String(),
		Email:          u.Email,
		Username:       u.Username,
		FullName:       u.FullName,
		StatusName:     u.StatusName,
		StatusId:       u.StatusID.String(),
		ActiveRoleName: u.ActiveRoleName,
		ActiveRoleId:   u.ActiveRoleID.String(),
		Roles:          profiles,
		CreatedAt:      timestamppb.New(u.CreatedAt),
		UpdatedAt:      timestamppb.New(u.UpdatedAt),
		VerifiedAt:     verifiedAt,
	}
	if u.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*u.DeletedAt)
	}
	return res
}
