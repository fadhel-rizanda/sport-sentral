package nats

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type AdministrativeDivisionEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewAdministrativeDivisionPublisher(nats *messaging.Client, logger *zap.Logger) *AdministrativeDivisionEventPublisher {
	return &AdministrativeDivisionEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *AdministrativeDivisionEventPublisher) PublishAdministrativeDivisionCreated(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAdministrativeDivisionCreated, evt); err != nil {
		p.logger.Error("failed to publish administrative division created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("administrative division created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AdministrativeDivisionEventPublisher) PublishAdministrativeDivisionUpdated(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAdministrativeDivisionUpdated, evt); err != nil {
		p.logger.Error("failed to publish administrative division updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("administrative division updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AdministrativeDivisionEventPublisher) PublishAdministrativeDivisionDeleted(ctx context.Context, evt *metav1.AdministrativeDivisionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAdministrativeDivisionDeleted, evt); err != nil {
		p.logger.Error("failed to publish administrative division deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("administrative division deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
