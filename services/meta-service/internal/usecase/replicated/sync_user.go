package replicated

import (
	"context"
	"errors"
	"fmt"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserSyncUseCase struct {
	userRepo replicated.UserRepository
	logger   *zap.Logger
}

func NewUserSyncUseCase(
	userRepo replicated.UserRepository,
	logger *zap.Logger,
) *UserSyncUseCase {
	return &UserSyncUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *UserSyncUseCase) upsert(ctx context.Context, evt *userv1.UserEvent) error {
	if evt.UserId == "" || evt.UserEmail == "" || evt.UserUsername == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.UserId)
	if err != nil {
		uc.logger.Warn("invalid user_id – skipping",
			zap.String("user_id", evt.UserId),
			zap.Error(err),
		)
		return nil
	}
	roleId, err := uuid.Parse(evt.UserActiveRoleId)
	if err != nil {
		uc.logger.Warn("invalid active_role_id – skipping",
			zap.String("active_role_id", evt.UserActiveRoleId),
			zap.Error(err),
		)
		return nil
	}
	statusId, err := uuid.Parse(evt.UserStatusId)
	if err != nil {
		uc.logger.Warn("invalid status_id – skipping",
			zap.String("status_id", evt.UserStatusId),
			zap.Error(err),
		)
		return nil
	}

	user := entity.User{
		ID:           id,
		Email:        evt.UserEmail,
		Username:     evt.UserUsername,
		FullName:     evt.UserFullName,
		ActiveRoleID: roleId,
		StatusID:     statusId,
	}
	if evt.DeletedAt != nil {
		user.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.userRepo.Upsert(ctx, user); err != nil {
		return apperr.Internal(fmt.Errorf("upsert user: %w", err))
	}

	uc.logger.Info("user upserted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *UserSyncUseCase) delete(ctx context.Context, evt *userv1.UserEvent) error {
	if evt.UserId == "" || evt.UserEmail == "" || evt.UserUsername == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.UserId)
	if err != nil {
		uc.logger.Warn("invalid user_id – skipping",
			zap.String("user_id", evt.UserId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("user already deleted", zap.String("user_id", evt.UserId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete user: %w", err))
	}

	uc.logger.Info("user deleted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *UserSyncUseCase) Sync(ctx context.Context, evt *userv1.UserEvent) error {
	uc.logger.Info("processing user event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("user_id", evt.UserId),
	)

	switch evt.EventType {
	case userv1.UserEventType_USER_EVENT_TYPE_CREATED,
		userv1.UserEventType_USER_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case userv1.UserEventType_USER_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
