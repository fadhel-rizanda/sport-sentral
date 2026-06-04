package nats

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type TagEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewTagEventPublisher(nats *messaging.Client, logger *zap.Logger) *TagEventPublisher {
	return &TagEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *TagEventPublisher) PublishTagCreated(ctx context.Context, evt *metav1.TagEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectTagCreated, evt); err != nil {
		p.logger.Error("failed to publish tag created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("tag created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *TagEventPublisher) PublishTagUpdated(ctx context.Context, evt *metav1.TagEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectTagUpdated, evt); err != nil {
		p.logger.Error("failed to publish tag updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("tag updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *TagEventPublisher) PublishTagDeleted(ctx context.Context, evt *metav1.TagEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectTagDeleted, evt); err != nil {
		p.logger.Error("failed to publish tag deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("tag deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
