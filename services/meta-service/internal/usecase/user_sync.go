package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
)

type UserSyncUseCase struct {
	cacheRepo repository.UserCacheRepository
	logger    *zap.Logger
}

func NewUserSyncUseCase(
	cacheRepo repository.UserCacheRepository,
	logger *zap.Logger,
) *UserSyncUseCase {
	return &UserSyncUseCase{
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

func (uc *UserSyncUseCase) upsertCache(ctx context.Context, evt *userv1.UserEvent) error {
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

	userCache := entity.UserCache{
		ID:             id,
		Email:          evt.UserEmail,
		Username:       evt.UserUsername,
		FullName:       evt.UserFullName,
		ActiveRoleID:   roleId,
		ActiveRoleName: evt.UserActiveRoleName,
		StatusID:       statusId,
	}
	if evt.DeletedAt != nil {
		userCache.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.cacheRepo.Upsert(ctx, userCache); err != nil {
		return apperr.Internal(fmt.Errorf("upsert user cache: %w", err))
	}

	uc.logger.Info("user cache upserted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *UserSyncUseCase) deleteCache(ctx context.Context, evt *userv1.UserEvent) error {
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

	if err := uc.cacheRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("user already deleted", zap.String("user_id", evt.UserId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete user cache: %w", err))
	}

	uc.logger.Info("user cache deleted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *UserSyncUseCase) SyncUser(ctx context.Context, evt *userv1.UserEvent) error {
	uc.logger.Info("processing user event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("user_id", evt.UserId),
	)

	switch evt.EventType {
	case userv1.UserEventType_USER_EVENT_TYPE_CREATED,
		userv1.UserEventType_USER_EVENT_TYPE_UPDATED:
		return uc.upsertCache(ctx, evt)

	case userv1.UserEventType_USER_EVENT_TYPE_DELETED:
		return uc.deleteCache(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
