package handler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userv1.UnimplementedUserInternalServiceServer
	uc usecase.UserUseCase
}

func NewUserHandler(uc usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) RegisterGRPC(s *grpc.Server) {
	userv1.RegisterUserServiceServer(s, h)
	userv1.RegisterUserInternalServiceServer(s, h)
}

// ─── UserService ──────────────────────────────────────────────────────────────

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := h.uc.CreateUser(ctx, req.Email, req.Username, req.FullName, req.Password)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.CreateUserResponse{User: toProto(user)}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := h.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.GetUserResponse{User: toProto(user)}, nil
}

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *userv1.GetUserByEmailRequest) (*userv1.GetUserByEmailResponse, error) {
	user, err := h.uc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.GetUserByEmailResponse{User: toProto(user)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	user, err := h.uc.UpdateUser(ctx, req.Id, req.FullName, req.Username)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.UpdateUserResponse{User: toProto(user)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if err := h.uc.DeleteUser(ctx, req.Id); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.DeleteUserResponse{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, total, err := h.uc.ListUsers(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	protoUsers := make([]*userv1.User, len(users))
	for i, u := range users {
		protoUsers[i] = toProto(u)
	}

	return &userv1.ListUsersResponse{
		Users:    protoUsers,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ─── UserInternalService ──────────────────────────────────────────────────────

func (h *UserHandler) GetUserByEmailInternal(ctx context.Context, req *userv1.GetUserByEmailInternalRequest) (*userv1.GetUserByEmailInternalResponse, error) {
	user, err := h.uc.GetUserByEmailInternal(ctx, req.Email)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &userv1.GetUserByEmailInternalResponse{
		User: toProtoInternal(user),
	}, nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────
func toProto(u *entity.User) *userv1.User {
	return &userv1.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		Username:  u.Username,
		FullName:  u.FullName,
		Roles:     toProtoRoles(u.Roles),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func toProtoInternal(u *entity.User) *userv1.UserInternal {
	return &userv1.UserInternal{
		Id:             u.ID.String(),
		Email:          u.Email,
		Username:       u.Username,
		FullName:       u.FullName,
		Roles:          toProtoRoles(u.Roles),
		HashedPassword: u.HashedPassword,
	}
}
