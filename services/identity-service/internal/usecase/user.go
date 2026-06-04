package usecase

import (
	"context"
	"fmt"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/identity-service/internal/dto"
	"microservice-golang/services/identity-service/internal/mapper"
	"microservice-golang/shared/pkg/constants"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/mailer"
	"microservice-golang/shared/pkg/redisclient"
	"microservice-golang/shared/pkg/token"
)

type UserUseCase interface {
	Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*dto.UserResponse, error)
	Update(ctx context.Context, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteUserRequest) error
	List(ctx context.Context, req dto.ListRequest) (*dto.ListUsersResponse, error)
	SendVerifyEmail(ctx context.Context, email string) error
	VerifyAccount(ctx context.Context, tokenStr string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, tokenStr, newPassword string) error

	AssignRolesToUser(ctx context.Context, userID string, roleIDs []string) error
	RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error
}

type userUseCase struct {
	db              *gorm.DB
	userRepo        repository.UserRepository
	userRoleRepo    repository.UserRoleRepository
	roleRepo        repository.RoleRepository
	mailer          *mailer.Mailer
	redis           redisclient.Client
	AppURL          string
	logger          *zap.Logger
	statusCacheRepo repository.StatusRepository
	publisher       UserEventPublisher
}

func NewUserUseCase(
	db *gorm.DB,
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
	mailer *mailer.Mailer,
	redis redisclient.Client,
	AppURL string,
	logger *zap.Logger,
	statusCacheRepo repository.StatusRepository,
	publisher UserEventPublisher,
) UserUseCase {
	return &userUseCase{
		db:              db,
		userRepo:        userRepo,
		userRoleRepo:    userRoleRepo,
		roleRepo:        roleRepo,
		redis:           redis,
		mailer:          mailer,
		AppURL:          AppURL,
		logger:          logger,
		statusCacheRepo: statusCacheRepo,
		publisher:       publisher,
	}
}

func (uc *userUseCase) Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, apperr.NotFound("role")
	}

	if entity.RolesAdminAssignOnly[role.Name] {
		return nil, apperr.Forbidden("cannot self-register with this role")
	}

	userStatusName := constants.StatusActive
	if entity.RolesPendingApproval[role.Name] {
		userStatusName = constants.StatusPending
	}

	userRoleStatusName := constants.StatusActive
	if entity.RolesPendingApproval[role.Name] {
		userRoleStatusName = constants.StatusPending
	}

	statusCache, err := uc.statusCacheRepo.GetByTypeAndName(ctx, constants.StatusTypeUser, userStatusName)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if statusCache == nil {
		return nil, apperr.Internal(fmt.Errorf("status cache not found"))
	}

	roleStatusCache, err := uc.statusCacheRepo.GetByTypeAndName(ctx, constants.StatusTypeUserRole, userRoleStatusName)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if roleStatusCache == nil {
		return nil, apperr.Internal(fmt.Errorf("role status cache not found"))
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	user := &entity.User{
		ID:             id,
		Email:          req.Email,
		Username:       req.Username,
		FullName:       req.FullName,
		HashedPassword: string(hashed),
		StatusID:       statusCache.ID,
	}

	if err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txUserRoleRepo := repository.NewUserRoleRepository(tx)

		if err := txUserRepo.Create(ctx, user); err != nil {
			if postgres.IsUniqueConstraint(err, "idx_users_email") {
				return apperr.Conflict("email")
			}
			if postgres.IsUniqueConstraint(err, "idx_users_username") {
				return apperr.Conflict("username")
			}
			return apperr.Internal(err)
		}

		if err := txUserRoleRepo.Add(ctx, &entity.UserRole{
			UserID:   user.ID,
			RoleID:   role.ID,
			IsActive: true,
			StatusID: roleStatusCache.ID,
		}); err != nil {
			return apperr.Internal(err)
		}

		return nil
	}); err != nil {
		return nil, apperr.Internal(err)
	}

	go func() {
		if err := uc.sendVerifyEmailInternal(context.Background(), user); err != nil {
			uc.logger.Error("failed to send verify email", zap.Error(err))
		}
	}()

	return uc.GetByID(ctx, user.ID)
}

func (uc *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToUserResponse(user), nil
}

func (uc *userUseCase) GetByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToUserResponse(user), nil
}

func (uc *userUseCase) Update(ctx context.Context, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
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
		if postgres.IsUniqueConstraint(err, "idx_users_username") {
			return nil, apperr.Conflict("username")
		}
		return nil, apperr.Internal(err)
	}

	userRole, err := uc.userRoleRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	evt := uc.buildUserEvent(
		userv1.UserEventType_USER_EVENT_TYPE_UPDATED,
		user,
		&userRole.Role,
	)
	if err := uc.publisher.PublishUserUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToUserResponse(user), nil
}

func (uc *userUseCase) SoftDelete(ctx context.Context, req dto.DeleteUserRequest) error {
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

	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return apperr.Internal(err)
	}

	userRole, err := uc.userRoleRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildUserEvent(
		userv1.UserEventType_USER_EVENT_TYPE_DELETED,
		user,
		&userRole.Role,
	)
	if err := uc.publisher.PublishUserDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *userUseCase) List(ctx context.Context, req dto.ListRequest) (*dto.ListUsersResponse, error) {
	users, total, err := uc.userRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = mapper.ToUserResponse(u)
	}

	return &dto.ListUsersResponse{
		Users:    result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *userUseCase) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if user.VerifiedAt != nil {
		return apperr.InvalidArgument("account already verified")
	}

	go func() {
		if err := uc.sendVerifyEmailInternal(context.Background(), user); err != nil {
			uc.logger.Error("failed to send verify email", zap.Error(err))
		}
	}()

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

	user.Verify()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return apperr.Internal(err)
	}

	uc.redis.Del(ctx, key)

	userRole, err := uc.userRoleRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildUserEvent(
		userv1.UserEventType_USER_EVENT_TYPE_CREATED,
		user,
		&userRole.Role,
	)
	if err := uc.publisher.PublishUserCreated(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

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

func (uc *userUseCase) AssignRolesToUser(ctx context.Context, userID string, roleID []string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	rids, err := parseUUIDs(roleID)
	if err != nil {
		return apperr.InvalidArgument("invalid role id")
	}

	if _, err := uc.userRepo.GetByID(ctx, uid); err != nil {
		return apperr.Internal(err)
	}

	for _, rid := range rids {
		if _, err := uc.roleRepo.GetByID(ctx, rid); err != nil {
			return apperr.Internal(err)
		}
	}

	err = uc.userRepo.AssignRoles(ctx, uid, rids)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return apperr.Conflict("user already has one or more roles")
		}
		if postgres.IsForeignKeyViolation(err) {
			return apperr.NotFound("one or more roles not found")
		}
		return apperr.Internal(err)
	}
	return nil
}

func (uc *userUseCase) RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	rIDs, err := parseUUIDs(roleIDs)
	if err != nil {
		return apperr.Internal(err)
	}

	err = uc.userRepo.RemoveRoles(ctx, uID, rIDs)
	if err != nil {
		if postgres.IsForeignKeyViolation(err) {
			return apperr.NotFound("one or more roles not found")
		}
		return apperr.Internal(err)
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (uc *userUseCase) sendVerifyEmailInternal(ctx context.Context, user *entity.User) error {
	tok, err := token.Generate(18)
	if err != nil {
		return apperr.Internal(err)
	}

	key := verifyEmailKey(tok)
	if err := uc.redis.Set(ctx, key, user.ID.String(), 24*time.Hour); err != nil {
		return apperr.Internal(err)
	}

	verifyURL := fmt.Sprintf("%s/verify?token=%s", uc.AppURL, tok)
	body := mailer.VerifyAccountBody(user.Username, verifyURL)
	return uc.mailer.Send(user.Email, "Verify your SportCentral account", body)
}

func verifyEmailKey(tok string) string {
	return "verify_email:" + tok
}

func resetPasswordKey(tok string) string {
	return "reset_password:" + tok
}

func (uc *userUseCase) buildUserEvent(
	eventType userv1.UserEventType,
	user *entity.User,
	role *entity.Role,
) *userv1.UserEvent {
	evtID, _ := uuid.NewV7()

	evt := &userv1.UserEvent{
		EventId:          evtID.String(),
		EventType:        eventType,
		OccurredAt:       timestamppb.Now(),
		UserId:           user.ID.String(),
		UserEmail:        user.Email,
		UserUsername:     user.Username,
		UserFullName:     user.FullName,
		UserActiveRoleId: role.ID.String(),
		UserStatusId:     user.StatusID.String(),
	}

	if user.DeletedAt.Valid {
		deletedAt := user.DeletedAt
		evt.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return evt
}
