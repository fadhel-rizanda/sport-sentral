package replicated

import (
	"context"
	"errors"
	"fmt"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/services/venue-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncUserUseCase struct {
	userRepo replicated.UserRepository
	logger   *zap.Logger
}

func NewSyncUserUseCase(
	userRepo replicated.UserRepository,
	logger *zap.Logger,
) *SyncUserUseCase {
	return &SyncUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *SyncUserUseCase) upsert(ctx context.Context, evt *userv1.UserEvent) error {
	if evt.UserId == "" || evt.UserEmail == "" || evt.UserUsername == "" {
		uc.logger.Warn("incomplete user event – skipping", zap.String("event_id", evt.EventId))
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
	roleID, err := uuid.Parse(evt.UserActiveRoleId)
	if err != nil {
		uc.logger.Warn("invalid active_role_id – skipping",
			zap.String("active_role_id", evt.UserActiveRoleId),
			zap.Error(err),
		)
		return nil
	}
	statusID, err := uuid.Parse(evt.UserStatusId)
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
		ActiveRoleID: roleID,
		StatusID:     statusID,
	}
	if evt.DeletedAt != nil {
		user.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.userRepo.Upsert(ctx, user); err != nil {
		return apperr.Internal(fmt.Errorf("upsert user replicas: %w", err))
	}

	uc.logger.Info("user replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *SyncUserUseCase) delete(ctx context.Context, evt *userv1.UserEvent) error {
	if evt.UserId == "" {
		uc.logger.Warn("missing user_id – skipping", zap.String("event_id", evt.EventId))
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
		return apperr.Internal(fmt.Errorf("delete user replicas: %w", err))
	}

	uc.logger.Info("user replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("user_id", evt.UserId),
	)

	return nil
}

func (uc *SyncUserUseCase) Sync(ctx context.Context, evt *userv1.UserEvent) error {
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
		uc.logger.Warn("unknown user event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
