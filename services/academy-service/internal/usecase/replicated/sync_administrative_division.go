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

type SyncAdministrativeDivisionUseCase struct {
	adminDivRepo replicated.AdministrativeDivisionRepository
	logger       *zap.Logger
}

func NewSyncAdministrativeDivisionUseCase(
	adminDivRepo replicated.AdministrativeDivisionRepository,
	logger *zap.Logger,
) *SyncAdministrativeDivisionUseCase {
	return &SyncAdministrativeDivisionUseCase{
		adminDivRepo: adminDivRepo,
		logger:       logger,
	}
}

func (uc *SyncAdministrativeDivisionUseCase) upsert(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	if evt.AdministrativeDivisionId == "" || evt.AdministrativeDivisionName == "" || evt.AdministrativeDivisionCountryId == "" {
		uc.logger.Warn("incomplete admin division event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.AdministrativeDivisionId)
	if err != nil {
		uc.logger.Warn("invalid administrative_division_id – skipping",
			zap.String("administrative_division_id", evt.AdministrativeDivisionId),
			zap.Error(err),
		)
		return nil
	}

	countryID, err := uuid.Parse(evt.AdministrativeDivisionCountryId)
	if err != nil {
		uc.logger.Warn("invalid country_id – skipping",
			zap.String("country_id", evt.AdministrativeDivisionCountryId),
			zap.Error(err),
		)
		return nil
	}

	var parentID *uuid.UUID
	if evt.AdministrativeDivisionParentId != "" {
		pID, err := uuid.Parse(evt.AdministrativeDivisionParentId)
		if err != nil {
			uc.logger.Warn("invalid parent_id – skipping",
				zap.String("parent_id", evt.AdministrativeDivisionParentId),
				zap.Error(err),
			)
			return nil
		}
		parentID = &pID
	}

	if err := uc.adminDivRepo.Upsert(ctx, entity.AdministrativeDivision{
		ID:         id,
		CountryID:  countryID,
		ParentID:   parentID,
		Name:       evt.AdministrativeDivisionName,
		Level:      evt.AdministrativeDivisionLevel,
		PostalCode: evt.AdministrativeDivisionPostalCode,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert administrative division replicas: %w", err))
	}

	uc.logger.Info("administrative division replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("administrative_division_id", evt.AdministrativeDivisionId),
	)

	return nil
}

func (uc *SyncAdministrativeDivisionUseCase) delete(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	if evt.AdministrativeDivisionId == "" {
		uc.logger.Warn("missing administrative_division_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.AdministrativeDivisionId)
	if err != nil {
		uc.logger.Warn("invalid administrative_division_id – skipping",
			zap.String("administrative_division_id", evt.AdministrativeDivisionId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.adminDivRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("administrative division already deleted", zap.String("administrative_division_id", evt.AdministrativeDivisionId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete administrative division replicas: %w", err))
	}

	uc.logger.Info("administrative division replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("administrative_division_id", evt.AdministrativeDivisionId),
	)

	return nil
}

func (uc *SyncAdministrativeDivisionUseCase) Sync(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	uc.logger.Info("processing administrative division event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("administrative_division_id", evt.AdministrativeDivisionId),
	)

	switch evt.EventType {
	case metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_CREATED,
		metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case metav1.AdministrativeDivisionEventType_ADMINISTRATIVE_DIVISION_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown administrative division event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
