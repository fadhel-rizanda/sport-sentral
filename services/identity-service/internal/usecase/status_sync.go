package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
)

type StatusSyncUseCase struct {
	cacheRepo repository.StatusCacheRepository
	logger    *zap.Logger
}

func NewStatusSyncUseCase(
	cacheRepo repository.StatusCacheRepository,
	logger *zap.Logger,
) *StatusSyncUseCase {
	return &StatusSyncUseCase{
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

func (uc *StatusSyncUseCase) upsertCache(ctx context.Context, evt *metav1.StatusEvent) error {
	if evt.StatusId == "" || evt.StatusName == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.StatusId)
	if err != nil {
		uc.logger.Warn("invalid status_id – skipping",
			zap.String("status_id", evt.StatusId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.cacheRepo.Upsert(ctx, entity.StatusCache{
		ID:   id,
		Type: evt.StatusType,
		Name: evt.StatusName,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert status cache: %w", err))
	}

	uc.logger.Info("status cache upserted",
		zap.String("event_id", evt.EventId),
		zap.String("status_id", evt.StatusId),
	)

	return nil
}

func (uc *StatusSyncUseCase) deleteCache(ctx context.Context, evt *metav1.StatusEvent) error {
	if evt.StatusId == "" {
		uc.logger.Warn("missing status_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.StatusId)
	if err != nil {
		uc.logger.Warn("invalid status_id – skipping",
			zap.String("status_id", evt.StatusId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.cacheRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("status already deleted", zap.String("status_id", evt.StatusId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete status cache: %w", err))
	}

	uc.logger.Info("status cache deleted",
		zap.String("event_id", evt.EventId),
		zap.String("status_id", evt.StatusId),
	)

	return nil
}

func (uc *StatusSyncUseCase) SyncStatus(ctx context.Context, evt *metav1.StatusEvent) error {
	uc.logger.Info("processing status event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("status_id", evt.StatusId),
	)

	switch evt.EventType {
	case metav1.StatusEventType_STATUS_EVENT_TYPE_CREATED,
		metav1.StatusEventType_STATUS_EVENT_TYPE_UPDATED:
		return uc.upsertCache(ctx, evt)

	case metav1.StatusEventType_STATUS_EVENT_TYPE_DELETED:
		return uc.deleteCache(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
