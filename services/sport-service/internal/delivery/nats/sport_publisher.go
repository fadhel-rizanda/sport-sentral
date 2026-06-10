package nats

import (
	"context"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type SportEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewSportEventPublisher(nats *messaging.Client, logger *zap.Logger) *SportEventPublisher {
	return &SportEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *SportEventPublisher) PublishSportCreated(ctx context.Context, evt *sportv1.SportEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectSportCreated, evt); err != nil {
		p.logger.Error("failed to publish sport created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("sport created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *SportEventPublisher) PublishSportUpdated(ctx context.Context, evt *sportv1.SportEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectSportUpdated, evt); err != nil {
		p.logger.Error("failed to publish sport updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("sport updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *SportEventPublisher) PublishSportDeleted(ctx context.Context, evt *sportv1.SportEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectSportDeleted, evt); err != nil {
		p.logger.Error("failed to publish sport deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("sport deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
