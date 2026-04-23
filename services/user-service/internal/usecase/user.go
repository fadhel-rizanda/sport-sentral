package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/mailer"
	"microservice-golang/shared/pkg/token"
	"time"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, email, username, fullName, password string) (*entity.User, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByEmailInternal(ctx context.Context, email string) (*entity.User, error) // return full entity termasuk HashedPassword
	UpdateUser(ctx context.Context, id, fullName, username string) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)
	SendVerifyEmail(ctx context.Context, email string) error
	VerifyAccount(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type userUseCase struct {
	repo   repository.UserRepository
	redis  *redis.Client
	mailer *mailer.Mailer
	appURL string
}

func NewUserUseCase(
	repo repository.UserRepository,
	redis *redis.Client,
	mailer *mailer.Mailer,
	appURL string,
) UserUseCase {
	return &userUseCase{
		repo:   repo,
		redis:  redis,
		mailer: mailer,
		appURL: appURL,
	}
}

func (uc *userUseCase) CreateUser(ctx context.Context, email, username, fullName, password string) (*entity.User, error) {
	user, err := entity.NewUser(email, username, fullName, password)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := uc.SendVerifyEmail(ctx, user.ID.String()); err != nil {
		logger.Get().Warn("failed to send verification email",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	return user, nil
}

func (uc *userUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid user id")
	}

	return uc.repo.GetByID(ctx, uid)
}

func (uc *userUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return uc.repo.GetByEmail(ctx, email)
}

func (uc *userUseCase) GetUserByEmailInternal(ctx context.Context, email string) (*entity.User, error) {
	return uc.repo.GetByEmailWithRoles(ctx, email)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, id, fullName, username string) (*entity.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid user id")
	}

	user, err := uc.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	user.Update(fullName, username)

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	return uc.repo.Delete(ctx, uid)
}

func (uc *userUseCase) ListUsers(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.List(ctx, page, pageSize)
}

func (uc *userUseCase) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.VerifiedAt != nil {
		return apperr.Conflict("account already verified")
	}

	oldToken, err := uc.redis.Get(ctx, verifyTokenKey(user.ID.String())).Result()
	if err == nil {
		uc.redis.Del(ctx, verifyUserKey(oldToken))
	}

	t, err := token.Generate(32)
	if err != nil {
		return apperr.Internal(err)
	}

	uc.redis.Set(ctx, verifyUserKey(t), user.ID.String(), 24*time.Hour)
	uc.redis.Set(ctx, verifyTokenKey(user.ID.String()), t, 24*time.Hour)

	//	NOTE:
	// 	biar jalan dibackground sesimpel tambahin go func(){}()
	// 	dan didalamnya dibuat context baru dengan timeoutnya agar tidak mati saat parent sudah cancel
	// 	beserta defernya agar tidak leak
	//	selama proses i/o bound dia bakal maksimal, kalo cpu bound tidak akan ngaruh soalnya dia akan rebutan cpu/thread si osnya
	go func() {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		verifyURL := fmt.Sprintf("%s/verify?token=%s", uc.appURL, t)
		body := mailer.VerifyAccountBody(user.Username, verifyURL)

		if err := uc.mailer.Send(user.Email, "Verify your account", body); err != nil {
			logger.Get().Warn("failed to send verify email",
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
			return
		}

		logger.Get().Info("verification email sent", zap.String("user_id", user.ID.String()))
	}()

	return nil
}

func (uc *userUseCase) VerifyAccount(ctx context.Context, token string) error {
	userID, err := uc.redis.Get(ctx, verifyUserKey(token)).Result()
	if err != nil {
		return apperr.InvalidArgument("invalid or expired token")
	}

	uid, _ := uuid.Parse(userID)
	user, err := uc.repo.GetByID(ctx, uid)
	if err != nil {
		return err
	}

	if user.VerifiedAt != nil {
		return apperr.Conflict("account already verified")
	}

	user.Verify()
	if err := uc.repo.Update(ctx, user); err != nil {
		return err
	}

	uc.redis.Del(ctx, verifyUserKey(token))
	uc.redis.Del(ctx, verifyTokenKey(userID))

	logger.Get().Info("account verified", zap.String("user_id", userID))
	return nil
}

func (uc *userUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if user.VerifiedAt == nil {
		return apperr.NotFound("account not found")
	}

	userID := user.ID.String()

	oldToken, err := uc.redis.Get(ctx, resetTokenKey(userID)).Result()
	if err == nil {
		uc.redis.Del(ctx, resetUserKey(oldToken))
	}

	t, err := token.Generate(32)
	if err != nil {
		return apperr.Internal(err)
	}

	uc.redis.Set(ctx, resetUserKey(t), userID, time.Hour)
	uc.redis.Set(ctx, resetTokenKey(userID), t, time.Hour)

	go func() {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resetURL := fmt.Sprintf("%s/reset?token=%s", uc.appURL, t)
		body := mailer.ForgotPasswordBody(user.Username, resetURL)

		if err := uc.mailer.Send(user.Email, "Reset your password", body); err != nil {
			logger.Get().Warn("failed to send reset email",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			return
		}

		logger.Get().Info("reset email sent", zap.String("user_id", userID))
	}()

	return nil
}

func (uc *userUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := uc.redis.Get(ctx, resetUserKey(token)).Result()
	if err != nil {
		return apperr.InvalidArgument("invalid or expired token")
	}

	uid, _ := uuid.Parse(userID)
	user, err := uc.repo.GetByID(ctx, uid)
	if err != nil {
		return err
	}

	if err := user.UpdatePassword(newPassword); err != nil {
		return apperr.Internal(err)
	}

	if err := uc.repo.UpdatePassword(ctx, user); err != nil {
		return err
	}

	uc.redis.Del(ctx, resetUserKey(token))
	uc.redis.Del(ctx, resetTokenKey(userID))

	logger.Get().Info("reset email sent", zap.String("user_id", userID))
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
func verifyTokenKey(userID string) string {
	return fmt.Sprintf("verify:token:%s", userID)
}
func resetTokenKey(userID string) string {
	return fmt.Sprintf("reset:token:%s", userID)
}
func verifyUserKey(token string) string {
	return fmt.Sprintf("verify:user:%s", token)
}
func resetUserKey(token string) string {
	return fmt.Sprintf("reset:user:%s", token)
}
