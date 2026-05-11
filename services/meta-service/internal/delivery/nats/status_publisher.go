package nats

import (
	"context"
	"go.uber.org/zap"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"
)

type StatusEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewStatusEventPublisher(nats *messaging.Client, logger *zap.Logger) *StatusEventPublisher {
	return &StatusEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *StatusEventPublisher) PublishStatusCreated(ctx context.Context, evt *metav1.StatusEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectStatusCreated, evt); err != nil {
		p.logger.Error("failed to publish status created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("status created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *StatusEventPublisher) PublishStatusUpdated(ctx context.Context, evt *metav1.StatusEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectStatusUpdated, evt); err != nil {
		p.logger.Error("failed to publish status updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("status updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *StatusEventPublisher) PublishStatusDeleted(ctx context.Context, evt *metav1.StatusEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectStatusDeleted, evt); err != nil {
		p.logger.Error("failed to publish status deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("status deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
