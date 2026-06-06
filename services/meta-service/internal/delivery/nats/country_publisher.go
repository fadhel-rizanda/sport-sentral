package nats

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type CountryEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewCountryEventPublisher(nats *messaging.Client, logger *zap.Logger) *CountryEventPublisher {
	return &CountryEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *CountryEventPublisher) PublishCountryCreated(ctx context.Context, evt *metav1.CountryEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectCountryCreated, evt); err != nil {
		p.logger.Error("failed to publish country created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("country created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *CountryEventPublisher) PublishCountryUpdated(ctx context.Context, evt *metav1.CountryEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectCountryUpdated, evt); err != nil {
		p.logger.Error("failed to publish country updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("country updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *CountryEventPublisher) PublishCountryDeleted(ctx context.Context, evt *metav1.CountryEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectCountryDeleted, evt); err != nil {
		p.logger.Error("failed to publish country deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("country deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
