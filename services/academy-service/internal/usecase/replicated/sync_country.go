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

type SyncCountryUseCase struct {
	countryRepo replicated.CountryRepository
	logger      *zap.Logger
}

func NewSyncCountryUseCase(
	countryRepo replicated.CountryRepository,
	logger *zap.Logger,
) *SyncCountryUseCase {
	return &SyncCountryUseCase{
		countryRepo: countryRepo,
		logger:      logger,
	}
}

func (uc *SyncCountryUseCase) upsert(ctx context.Context, evt *metav1.CountryEvent) error {
	if evt.CountryId == "" || evt.CountryName == "" {
		uc.logger.Warn("incomplete country event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.CountryId)
	if err != nil {
		uc.logger.Warn("invalid country_id – skipping",
			zap.String("country_id", evt.CountryId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.countryRepo.Upsert(ctx, entity.Country{
		ID:           id,
		Name:         evt.CountryName,
		ISOAlpha2:    evt.CountryIsoAlpha_2,
		ISOAlpha3:    evt.CountryIsoAlpha_3,
		PhoneCode:    evt.CountryPhoneCode,
		CurrencyCode: evt.CountryCurrencyCode,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert country replicas: %w", err))
	}

	uc.logger.Info("country replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("country_id", evt.CountryId),
	)

	return nil
}

func (uc *SyncCountryUseCase) delete(ctx context.Context, evt *metav1.CountryEvent) error {
	if evt.CountryId == "" {
		uc.logger.Warn("missing country_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.CountryId)
	if err != nil {
		uc.logger.Warn("invalid country_id – skipping",
			zap.String("country_id", evt.CountryId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.countryRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("country already deleted", zap.String("country_id", evt.CountryId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete country replicas: %w", err))
	}

	uc.logger.Info("country replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("country_id", evt.CountryId),
	)

	return nil
}

func (uc *SyncCountryUseCase) Sync(ctx context.Context, evt *metav1.CountryEvent) error {
	uc.logger.Info("processing country event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("country_id", evt.CountryId),
	)

	switch evt.EventType {
	case metav1.CountryEventType_COUNTRY_EVENT_TYPE_CREATED,
		metav1.CountryEventType_COUNTRY_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case metav1.CountryEventType_COUNTRY_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown country event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
