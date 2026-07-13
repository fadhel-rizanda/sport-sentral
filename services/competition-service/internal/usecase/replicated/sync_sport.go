package replicated

import (
	"context"
	"errors"
	"fmt"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"
	"microservice-golang/shared/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncSportUseCase struct {
	sportRepo replicated.SportRepository
	logger    *zap.Logger
}

func NewSyncSportUseCase(
	sportRepo replicated.SportRepository,
	logger *zap.Logger,
) *SyncSportUseCase {
	return &SyncSportUseCase{
		sportRepo: sportRepo,
		logger:    logger,
	}
}

func (uc *SyncSportUseCase) upsert(ctx context.Context, evt *sportv1.SportEvent) error {
	if evt.SportId == "" || evt.Name == "" || evt.Slug == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.SportId)
	if err != nil {
		uc.logger.Warn("invalid sport_id – skipping",
			zap.String("sport_id", evt.SportId),
			zap.Error(err),
		)
		return nil
	}

	var iconID *uuid.UUID
	if evt.IconAttachmentId != nil && *evt.IconAttachmentId != "" {
		parsedIconID, err := uuid.Parse(*evt.IconAttachmentId)
		if err != nil {
			uc.logger.Warn("invalid icon_attachment_id – skipping",
				zap.String("icon_attachment_id", *evt.IconAttachmentId),
				zap.Error(err),
			)
		} else {
			iconID = &parsedIconID
		}
	}

	var regulatorID *uuid.UUID
	if evt.RegulatorId != nil && *evt.RegulatorId != "" {
		parsedRegulatorID, err := uuid.Parse(*evt.RegulatorId)
		if err != nil {
			uc.logger.Warn("invalid regulator_id – skipping",
				zap.String("regulator_id", *evt.RegulatorId),
				zap.Error(err),
			)
		} else {
			regulatorID = &parsedRegulatorID
		}
	}

	stats := make([]entity.SportStat, 0, len(evt.GetStats()))
	for _, statEvt := range evt.GetStats() {
		statID, err := uuid.Parse(statEvt.GetId())
		if err != nil {
			uc.logger.Warn("invalid stat_id – skipping stat",
				zap.String("stat_id", statEvt.GetId()),
				zap.Error(err),
			)
			continue
		}
		statTypeTagID, err := uuid.Parse(statEvt.GetStatTypeTagId())
		if err != nil {
			uc.logger.Warn("invalid stat_type_tag_id – skipping stat",
				zap.String("stat_type_tag_id", statEvt.GetStatTypeTagId()),
				zap.Error(err),
			)
			continue
		}
		stats = append(stats, entity.SportStat{
			ID:                statID,
			SportID:           id,
			StatTypeTagID:     statTypeTagID,
			AggregationMethod: utils.AggregationMethodToString(statEvt.GetAggregationMethod()),
		})
	}

	if err := uc.sportRepo.Upsert(ctx, entity.Sport{
		ID:               id,
		Name:             evt.Name,
		Slug:             evt.Slug,
		IconAttachmentID: iconID,
		IsVerified:       evt.IsVerified,
		RegulatorID:      regulatorID,
		Tier:             evt.Tier,
		Stats:            stats,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert sport replicas: %w", err))
	}

	uc.logger.Info("sport replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("sport_id", evt.SportId),
	)

	return nil
}

func (uc *SyncSportUseCase) delete(ctx context.Context, evt *sportv1.SportEvent) error {
	if evt.SportId == "" {
		uc.logger.Warn("missing sport_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.SportId)
	if err != nil {
		uc.logger.Warn("invalid sport_id – skipping",
			zap.String("sport_id", evt.SportId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.sportRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("sport already deleted", zap.String("sport_id", evt.SportId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete sport replicas: %w", err))
	}

	uc.logger.Info("sport replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("sport_id", evt.SportId),
	)

	return nil
}

func (uc *SyncSportUseCase) Sync(ctx context.Context, evt *sportv1.SportEvent) error {
	uc.logger.Info("processing sport event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("sport_id", evt.SportId),
	)

	switch evt.EventType {
	case sportv1.SportEventType_SPORT_EVENT_TYPE_CREATED,
		sportv1.SportEventType_SPORT_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case sportv1.SportEventType_SPORT_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
