package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	userv1 "microservice-golang/gen/user/v1"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/jwt"
)

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	ValidateToken(ctx context.Context, accessToken string) (*jwt.Claims, error)
}

type authUseCase struct {
	userClient userv1.UserInternalServiceClient
	jwtManager *jwt.Manager
	redis      *redis.Client
	refreshTTL time.Duration
}

func NewAuthUseCase(
	userClient userv1.UserInternalServiceClient,
	jwtManager *jwt.Manager,
	redis *redis.Client,
	refreshTTL time.Duration,
) AuthUseCase {
	return &authUseCase{
		userClient: userClient,
		jwtManager: jwtManager,
		redis:      redis,
		refreshTTL: refreshTTL,
	}
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	resp, err := uc.userClient.GetUserByEmailInternal(ctx, &userv1.GetUserByEmailInternalRequest{
		Email: email,
	})
	if err != nil {
		return nil, apperr.Unauthorized("invalid email or password")
	}

	user := resp.User

	if user.VerifiedAt == nil {
		return nil, apperr.Unauthorized("account not verified")
	}

	if !checkPassword(user.HashedPassword, password) {
		return nil, apperr.Unauthorized("invalid email or password")
	}

	roleIds := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roleIds[i] = r.Id
	}
	tokens, err := uc.jwtManager.GenerateTokenPair(user.Id, user.Email, user.Username, roleIds)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	key := refreshTokenKey(user.Id, tokens.RefreshToken)
	if err := uc.redis.Set(ctx, key, "1", uc.refreshTTL).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	return &LoginResponse{
		Tokens:   tokens,
		UserID:   user.Id,
		Email:    user.Email,
		Username: user.Username,
		Roles:    user.Roles,
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

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
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

	// Hapus token lama — rotation pattern
	if err := uc.redis.Del(ctx, key).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	tokens, err := uc.jwtManager.GenerateTokenPair(claims.UserID, claims.Email, claims.Username, claims.Roles)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	newKey := refreshTokenKey(claims.UserID, tokens.RefreshToken)
	if err := uc.redis.Set(ctx, newKey, "1", uc.refreshTTL).Err(); err != nil {
		return nil, apperr.Internal(err)
	}

	return tokens, nil
}

func (uc *authUseCase) ValidateToken(ctx context.Context, accessToken string) (*jwt.Claims, error) {
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

func checkPassword(hashedPassword, plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword)) == nil
}
