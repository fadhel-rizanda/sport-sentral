package usecase

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"microservice-golang/shared/pkg/redisclient"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/mailer"
	"microservice-golang/shared/pkg/token"
)

type UserUseCase interface {
	Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*UserResponse, error)
	Update(ctx context.Context, req UpdateUserRequest) (*UserResponse, error)
	SoftDelete(ctx context.Context, req DeleteUserRequest) error
	List(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
	SendVerifyEmail(ctx context.Context, email string) error
	VerifyAccount(ctx context.Context, tokenStr string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, tokenStr, newPassword string) error
}

type userUseCase struct {
	userRepo     repository.UserRepository
	userRoleRepo repository.UserRoleRepository
	roleRepo     repository.RoleRepository
	statusRepo   repository.StatusRepository
	mailer       *mailer.Mailer
	redis        redisclient.Client
	AppURL       string
	logger       *zap.Logger
}

func NewUserUseCase(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
	statusRepo repository.StatusRepository,
	mailer *mailer.Mailer,
	redis redisclient.Client,
	AppURL string,
	logger *zap.Logger,
) UserUseCase {
	return &userUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		statusRepo:   statusRepo,
		mailer:       mailer,
		redis:        redis,
		AppURL:       AppURL,
		logger:       logger,
	}
}

func (uc *userUseCase) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	// validate role — tidak boleh assign admin-only role saat register
	if entity.RolesAdminAssignOnly[req.RoleName] {
		return nil, apperr.Forbidden("cannot self-register with this role")
	}

	role, err := uc.roleRepo.GetByName(ctx, req.RoleName)
	if err != nil {
		return nil, apperr.NotFound("role")
	}

	// determine user status based on role
	userStatusName := entity.UserStatusActive
	if entity.RolesPendingApproval[req.RoleName] {
		userStatusName = entity.UserStatusPending
	}

	userStatus, err := uc.statusRepo.GetByTypeAndName(ctx, entity.StatusTypeUser, userStatusName)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	userRoleStatusName := entity.UserRoleStatusActive
	if entity.RolesPendingApproval[req.RoleName] {
		userRoleStatusName = entity.UserRoleStatusPending
	}

	userRoleStatus, err := uc.statusRepo.GetByTypeAndName(ctx, entity.StatusTypeUserRole, userRoleStatusName)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	user := &entity.User{
		ID:             uuid.New(),
		Email:          req.Email,
		Username:       req.Username,
		FullName:       req.FullName,
		HashedPassword: string(hashed),
		StatusID:       userStatus.ID,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_users_email") {
			return nil, apperr.Conflict("email")
		}
		if postgres.IsUniqueConstraint(err, "idx_users_username") {
			return nil, apperr.Conflict("username")
		}
		return nil, apperr.Internal(err)
	}

	userRole := &entity.UserRole{
		UserID:   user.ID,
		RoleID:   role.ID,
		IsActive: true,
		StatusID: userRoleStatus.ID,
	}

	if err := uc.userRoleRepo.Add(ctx, userRole); err != nil {
		return nil, apperr.Internal(err)
	}

	if err := uc.sendVerifyEmailInternal(ctx, user); err != nil {
		uc.logger.Error("failed to send verify email", zap.Error(err))
	}

	return uc.GetByID(ctx, user.ID)
}

func (uc *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}
	return ToUserResponse(user), nil
}

func (uc *userUseCase) Update(ctx context.Context, req UpdateUserRequest) (*UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.Username != nil {
		user.Username = *req.Username
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		if postgres.IsUniqueConstraint(err, "users_username_key") {
			return nil, apperr.Conflict("username")
		}
		return nil, apperr.Internal(err)
	}

	return ToUserResponse(user), nil
}

func (uc *userUseCase) SoftDelete(ctx context.Context, req DeleteUserRequest) error {
	user, err := uc.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("user")
		}
		return apperr.Internal(err)
	}

	if !user.CheckPassword(req.Password) {
		return apperr.Unauthorized("invalid email or password")
	}

	return uc.userRepo.SoftDelete(ctx, req.ID)
}

func (uc *userUseCase) List(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	users, total, err := uc.userRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = ToUserResponse(u)
	}

	return &ListUsersResponse{
		Users:    result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *userUseCase) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// jangan expose apakah email exist atau tidak
		return nil
	}

	if user.VerifiedAt != nil {
		return apperr.InvalidArgument("account already verified")
	}

	tok, err := token.Generate(18)
	if err != nil {
		return apperr.Internal(err)
	}

	key := verifyEmailKey(tok)
	if err := uc.redis.Set(ctx, key, user.ID.String(), 24*time.Hour); err != nil {
		return apperr.Internal(err)
	}

	if err := uc.sendVerifyEmailInternal(ctx, user); err != nil {
		uc.logger.Error("failed to send verify email", zap.Error(err))
	}

	return nil
}

func (uc *userUseCase) VerifyAccount(ctx context.Context, tokenStr string) error {
	key := verifyEmailKey(tokenStr)
	userIDStr, err := uc.redis.Get(ctx, key)
	if err != nil {
		return apperr.InvalidArgument("invalid or expired verification token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperr.Internal(err)
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperr.NotFound("user")
	}

	now := time.Now()
	user.VerifiedAt = &now

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return apperr.Internal(err)
	}

	uc.redis.Del(ctx, key)
	return nil
}

func (uc *userUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	tok, err := token.Generate(18)
	if err != nil {
		return apperr.Internal(err)
	}

	key := resetPasswordKey(tok)
	if err := uc.redis.Set(ctx, key, user.ID.String(), 1*time.Hour); err != nil {
		return apperr.Internal(err)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", uc.AppURL, tok)
	body := mailer.ForgotPasswordBody(user.Username, resetURL)

	go func() {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := uc.mailer.Send(user.Email, "Reset your SportCentral password", body); err != nil {
			uc.logger.Error("failed to send reset password email", zap.Error(err))
		}
	}()
	return nil
}

func (uc *userUseCase) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	key := resetPasswordKey(tokenStr)
	userIDStr, err := uc.redis.Get(ctx, key)
	if err != nil {
		return apperr.InvalidArgument("invalid or expired reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperr.Internal(err)
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperr.NotFound("user")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperr.Internal(err)
	}

	user.HashedPassword = string(hashed)
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return apperr.Internal(err)
	}

	uc.redis.Del(ctx, key)
	return nil
}

func (uc *userUseCase) GetByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}
	return ToUserResponse(user), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (uc *userUseCase) sendVerifyEmailInternal(ctx context.Context, user *entity.User) error {
	tok, err := token.Generate(18)
	if err != nil {
		return err
	}

	key := verifyEmailKey(tok)
	if err := uc.redis.Set(ctx, key, user.ID.String(), 24*time.Hour); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/verify?token=%s", uc.AppURL, tok)
	body := mailer.VerifyAccountBody(user.Username, verifyURL)

	go func() {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := uc.mailer.Send(user.Email, "Verify your SportCentral account", body); err != nil {
			uc.logger.Error("failed to send verify email", zap.Error(err))
		}
	}()
	return nil
}

func verifyEmailKey(tok string) string {
	return "verify_email:" + tok
}

func resetPasswordKey(tok string) string {
	return "reset_password:" + tok
}
