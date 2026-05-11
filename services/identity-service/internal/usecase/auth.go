package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"microservice-golang/services/identity-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/jwt"
)

type AuthUseCase interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	ValidateToken(_ context.Context, accessToken string) (*jwt.Claims, error)
}

type authUseCase struct {
	userRepo        repository.UserRepository
	userRoleRepo    repository.UserRoleRepository
	jwtManager      *jwt.Manager
	redis           *redis.Client
	refreshTTL      time.Duration
	statusCacheRepo repository.StatusCacheRepository
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	jwtManager *jwt.Manager,
	redis *redis.Client,
	refreshTTL time.Duration,
	statusCacheRepo repository.StatusCacheRepository,
) AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		userRoleRepo:    userRoleRepo,
		jwtManager:      jwtManager,
		redis:           redis,
		refreshTTL:      refreshTTL,
		statusCacheRepo: statusCacheRepo,
	}
}

func (uc *authUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.Unauthorized("invalid email or password")
	}

	if user.VerifiedAt == nil {
		return nil, apperr.Unauthorized("account not verified")
	}

	status, err := uc.statusCacheRepo.GetByTypeAndName(ctx, "account", "active")
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if user.StatusID != status.ID {
		return nil, apperr.Unauthorized("account is " + status.Name)
	}

	if !user.CheckPassword(req.Password) {
		return nil, apperr.Unauthorized("invalid email or password")
	}

	activeRole, err := uc.userRoleRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	roleIDs := make([]string, len(user.UserRoles))
	for i, ur := range user.UserRoles {
		roleIDs[i] = ur.RoleID.String()
	}

	tokens, err := uc.jwtManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Username,
		roleIDs,
		activeRole.Role.Name,
		activeRole.Role.ID,
	)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	key := refreshTokenKey(user.ID.String(), tokens.RefreshToken)
	if err := uc.redis.Set(ctx, key, "1", uc.refreshTTL).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	return &LoginResponse{
		AccessToken:    tokens.AccessToken,
		RefreshToken:   tokens.RefreshToken,
		ExpiresAt:      tokens.ExpiresAt,
		UserID:         user.ID,
		Email:          user.Email,
		Username:       user.Username,
		ActiveRoleName: activeRole.Role.Name,
	}, nil
}

func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	claims, err := uc.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		return apperr.Unauthorized("invalid refresh token")
	}

	key := refreshTokenKey(claims.UserID, refreshToken)
	if err := uc.redis.Del(ctx, key).Err(); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	claims, err := uc.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, apperr.Unauthorized("invalid refresh token")
	}

	key := refreshTokenKey(claims.UserID, refreshToken)
	exists, err := uc.redis.Exists(ctx, key).Result()
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if exists == 0 {
		return nil, apperr.Unauthorized("refresh token expired or revoked")
	}

	if err := uc.redis.Del(ctx, key).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	tokens, err := uc.jwtManager.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Username,
		claims.Roles,
		claims.ActiveRoleName,
		claims.ActiveRoleID,
	)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	newKey := refreshTokenKey(claims.UserID, tokens.RefreshToken)
	if err := uc.redis.Set(ctx, newKey, "1", uc.refreshTTL).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	return &RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

func (uc *authUseCase) ValidateToken(_ context.Context, accessToken string) (*jwt.Claims, error) {
	claims, err := uc.jwtManager.ValidateAccess(accessToken)
	if err != nil {
		return nil, apperr.Unauthorized("invalid or expired token")
	}
	return claims, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func refreshTokenKey(userID, token string) string {
	return fmt.Sprintf("refresh:%s:%s", userID, token)
}
