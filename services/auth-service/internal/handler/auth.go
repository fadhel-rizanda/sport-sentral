package handler

import (
	"context"

	"google.golang.org/grpc"

	authv1 "microservice-golang/gen/auth/v1"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/auth-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	uc usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) RegisterGRPC(s *grpc.Server) {
	authv1.RegisterAuthServiceServer(s, h)
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	result, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &authv1.LoginResponse{
		Tokens: &authv1.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			ExpiresAt:    result.Tokens.ExpiresAt.Unix(),
		},
		User: &userv1.User{
			Id:       result.UserID,
			Email:    result.Email,
			Username: result.Username,
			Roles:    result.Roles,
		},
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.uc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	tokens, err := h.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &authv1.RefreshTokenResponse{
		Tokens: &authv1.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    tokens.ExpiresAt.Unix(),
		},
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := h.uc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &authv1.ValidateTokenResponse{
		UserId:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		RoleIds:  claims.Roles,
	}, nil
}
