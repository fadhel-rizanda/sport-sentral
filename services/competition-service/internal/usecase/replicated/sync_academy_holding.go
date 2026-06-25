package replicated

import (
	"context"
	"errors"
	"fmt"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncAcademyHoldingUseCase struct {
	holdingRepo replicated.AcademyHoldingRepository
	logger      *zap.Logger
}

func NewSyncAcademyHoldingUseCase(
	holdingRepo replicated.AcademyHoldingRepository,
	logger *zap.Logger,
) *SyncAcademyHoldingUseCase {
	return &SyncAcademyHoldingUseCase{
		holdingRepo: holdingRepo,
		logger:      logger,
	}
}

func (uc *SyncAcademyHoldingUseCase) upsert(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	if evt.HoldingId == "" || evt.Name == "" || evt.StatusId == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.HoldingId)
	if err != nil {
		uc.logger.Warn("invalid holding_id – skipping",
			zap.String("holding_id", evt.HoldingId),
			zap.Error(err),
		)
		return nil
	}

	statusID, err := uuid.Parse(evt.StatusId)
	if err != nil {
		uc.logger.Warn("invalid status_id – skipping",
			zap.String("status_id", evt.StatusId),
			zap.Error(err),
		)
		return nil
	}

	var imageID *uuid.UUID
	if evt.ImageAttachmentId != nil && *evt.ImageAttachmentId != "" {
		parsedImageID, err := uuid.Parse(*evt.ImageAttachmentId)
		if err != nil {
			uc.logger.Warn("invalid image_attachment_id – skipping",
				zap.String("image_attachment_id", *evt.ImageAttachmentId),
				zap.Error(err),
			)
		} else {
			imageID = &parsedImageID
		}
	}

	holding := entity.AcademyHolding{
		ID:                id,
		Name:              evt.Name,
		Description:       evt.Description,
		Email:             evt.Email,
		PhoneNumber:       evt.PhoneNumber,
		ImageAttachmentID: imageID,
		StatusID:          statusID,
	}

	if evt.DeletedAt != nil {
		holding.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.holdingRepo.Upsert(ctx, holding); err != nil {
		return apperr.Internal(fmt.Errorf("upsert academy holding replicas: %w", err))
	}

	uc.logger.Info("academy holding replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("holding_id", evt.HoldingId),
	)

	return nil
}

func (uc *SyncAcademyHoldingUseCase) delete(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	if evt.HoldingId == "" {
		uc.logger.Warn("missing holding_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.HoldingId)
	if err != nil {
		uc.logger.Warn("invalid holding_id – skipping",
			zap.String("holding_id", evt.HoldingId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.holdingRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("academy holding already deleted", zap.String("holding_id", evt.HoldingId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete academy holding replicas: %w", err))
	}

	uc.logger.Info("academy holding replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("holding_id", evt.HoldingId),
	)

	return nil
}

func (uc *SyncAcademyHoldingUseCase) Sync(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	uc.logger.Info("processing academy holding event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("holding_id", evt.HoldingId),
	)

	switch evt.EventType {
	case academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_CREATED,
		academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case academyv1.AcademyHoldingEventType_ACADEMY_HOLDING_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
