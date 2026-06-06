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

type SyncTagUseCase struct {
	tagRepo replicated.TagRepository
	logger  *zap.Logger
}

func NewSyncTagUseCase(
	tagRepo replicated.TagRepository,
	logger *zap.Logger,
) *SyncTagUseCase {
	return &SyncTagUseCase{
		tagRepo: tagRepo,
		logger:  logger,
	}
}

func (uc *SyncTagUseCase) upsert(ctx context.Context, evt *metav1.TagEvent) error {
	if evt.TagId == "" || evt.TagName == "" || evt.TagSlug == "" {
		uc.logger.Warn("incomplete tag event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.TagId)
	if err != nil {
		uc.logger.Warn("invalid tag_id – skipping",
			zap.String("tag_id", evt.TagId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.tagRepo.Upsert(ctx, entity.Tag{
		ID:   id,
		Type: evt.TagType,
		Name: evt.TagName,
		Slug: evt.TagSlug,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert tag replicas: %w", err))
	}

	uc.logger.Info("tag replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("tag_id", evt.TagId),
	)

	return nil
}

func (uc *SyncTagUseCase) delete(ctx context.Context, evt *metav1.TagEvent) error {
	if evt.TagId == "" {
		uc.logger.Warn("missing tag_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.TagId)
	if err != nil {
		uc.logger.Warn("invalid tag_id – skipping",
			zap.String("tag_id", evt.TagId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.tagRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("tag already deleted", zap.String("tag_id", evt.TagId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete tag replicas: %w", err))
	}

	uc.logger.Info("tag replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("tag_id", evt.TagId),
	)

	return nil
}

func (uc *SyncTagUseCase) Sync(ctx context.Context, evt *metav1.TagEvent) error {
	uc.logger.Info("processing tag event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("tag_id", evt.TagId),
	)

	switch evt.EventType {
	case metav1.TagEventType_TAG_EVENT_TYPE_CREATED,
		metav1.TagEventType_TAG_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case metav1.TagEventType_TAG_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown tag event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
