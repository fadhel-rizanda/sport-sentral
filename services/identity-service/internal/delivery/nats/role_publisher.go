package nats

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type RoleEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewRoleEventPublisher(nats *messaging.Client, logger *zap.Logger) *RoleEventPublisher {
	return &RoleEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *RoleEventPublisher) PublishRoleCreated(ctx context.Context, evt *rbacv1.RoleEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectRoleCreated, evt); err != nil {
		p.logger.Error("failed to publish role created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
	}
	p.logger.Debug("role created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *RoleEventPublisher) PublishRoleUpdated(ctx context.Context, evt *rbacv1.RoleEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectRoleUpdated, evt); err != nil {
		p.logger.Error("failed to publish role updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("role updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *RoleEventPublisher) PublishRoleDeleted(ctx context.Context, evt *rbacv1.RoleEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectRoleDeleted, evt); err != nil {
		p.logger.Error("failed to publish role deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("role deleted event published", zap.String("event_id", evt.EventId))
	return nil
}
