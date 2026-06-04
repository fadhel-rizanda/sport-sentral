package nats

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type PermissionEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewPermissionEventPublisher(nats *messaging.Client, logger *zap.Logger) *PermissionEventPublisher {
	return &PermissionEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *PermissionEventPublisher) PublishPermissionCreated(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectPermissionCreated, evt); err != nil {
		p.logger.Error("failed to publish permission created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
	}
	p.logger.Debug("permission created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *PermissionEventPublisher) PublishPermissionUpdated(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectPermissionUpdated, evt); err != nil {
		p.logger.Error("failed to publish permission updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("permission updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *PermissionEventPublisher) PublishPermissionDeleted(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectPermissionDeleted, evt); err != nil {
		p.logger.Error("failed to publish permission deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("permission deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
