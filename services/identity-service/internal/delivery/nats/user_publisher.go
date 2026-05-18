package nats

import (
	"context"
	"go.uber.org/zap"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"
)

type UserEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewUserEventPublisher(nats *messaging.Client, logger *zap.Logger) *UserEventPublisher {
	return &UserEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *UserEventPublisher) PublishUserCreated(ctx context.Context, evt *userv1.UserEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectUserCreated, evt); err != nil {
		p.logger.Error("failed to publish user created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("user created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *UserEventPublisher) PublishUserUpdated(ctx context.Context, evt *userv1.UserEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectUserUpdated, evt); err != nil {
		p.logger.Error("failed to publish user updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("user updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *UserEventPublisher) PublishUserDeleted(ctx context.Context, evt *userv1.UserEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectUserDeleted, evt); err != nil {
		p.logger.Error("failed to publish user deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("user deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
