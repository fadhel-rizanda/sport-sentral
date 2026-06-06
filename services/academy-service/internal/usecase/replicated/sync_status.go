package replicated

import (
	"context"
	"errors"
	"fmt"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncStatusUseCase struct {
	statusRepo replicated.StatusRepository
	logger     *zap.Logger
}

func NewSyncStatusUseCase(
	statusRepo replicated.StatusRepository,
	logger *zap.Logger,
) *SyncStatusUseCase {
	return &SyncStatusUseCase{
		statusRepo: statusRepo,
		logger:     logger,
	}
}

func (uc *SyncStatusUseCase) upsert(ctx context.Context, evt *metav1.StatusEvent) error {
	if evt.StatusId == "" || evt.StatusName == "" || evt.StatusSlug == "" {
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

	if err := uc.statusRepo.Upsert(ctx, entity.Status{
		ID:   id,
		Type: evt.StatusType,
		Name: evt.StatusName,
		Slug: evt.StatusSlug,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert status replicas: %w", err))
	}

	uc.logger.Info("status replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("status_id", evt.StatusId),
	)

	return nil
}

func (uc *SyncStatusUseCase) delete(ctx context.Context, evt *metav1.StatusEvent) error {
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

	if err := uc.statusRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("status already deleted", zap.String("status_id", evt.StatusId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete status replicas: %w", err))
	}

	uc.logger.Info("status replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("status_id", evt.StatusId),
	)

	return nil
}

func (uc *SyncStatusUseCase) Sync(ctx context.Context, evt *metav1.StatusEvent) error {
	uc.logger.Info("processing status event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("status_id", evt.StatusId),
	)

	switch evt.EventType {
	case metav1.StatusEventType_STATUS_EVENT_TYPE_CREATED,
		metav1.StatusEventType_STATUS_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case metav1.StatusEventType_STATUS_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
