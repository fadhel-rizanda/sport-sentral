package handler

import (
	"context"

	"google.golang.org/grpc"
	authv1 "microservice-golang/gen/auth/v1"
	"microservice-golang/services/identity-service/internal/usecase"
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
	res, err := h.uc.Login(ctx, usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &authv1.LoginResponse{
		Tokens: &authv1.TokenPair{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken,
			ExpiresAt:    res.ExpiresAt.Unix(),
		},
		UserId:         res.UserID.String(),
		Email:          res.Email,
		Username:       res.Username,
		ActiveRoleName: res.ActiveRoleName,
		ActiveRoleId:   res.ActiveRoleID.String(),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.uc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	res, err := h.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &authv1.RefreshTokenResponse{
		Tokens: &authv1.TokenPair{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken,
			ExpiresAt:    res.ExpiresAt.Unix(),
		},
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := h.uc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &authv1.ValidateTokenResponse{
		UserId:         claims.UserID,
		Email:          claims.Email,
		Username:       claims.Username,
		RoleIds:        claims.Roles,
		ActiveRoleName: claims.ActiveRoleName,
		ActiveRoleId:   claims.ActiveRoleID.String(),
	}, nil
}
