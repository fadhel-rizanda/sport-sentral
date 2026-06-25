package nats

import (
	"context"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type AcademyEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewAcademyEventPublisher(nats *messaging.Client, logger *zap.Logger) *AcademyEventPublisher {
	return &AcademyEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *AcademyEventPublisher) PublishHoldingCreated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyHoldingCreated, evt); err != nil {
		p.logger.Error("failed to publish academy holding created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy holding created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishHoldingUpdated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyHoldingUpdated, evt); err != nil {
		p.logger.Error("failed to publish academy holding updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy holding updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishHoldingDeleted(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyHoldingDeleted, evt); err != nil {
		p.logger.Error("failed to publish academy holding deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy holding deleted event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishBranchCreated(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyBranchCreated, evt); err != nil {
		p.logger.Error("failed to publish academy branch created event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy branch created event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishBranchUpdated(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyBranchUpdated, evt); err != nil {
		p.logger.Error("failed to publish academy branch updated event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy branch updated event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishBranchDeleted(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyBranchDeleted, evt); err != nil {
		p.logger.Error("failed to publish academy branch deleted event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy branch deleted event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishAdminAssigned(ctx context.Context, evt *academyv1.AcademyAdminEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyAdminCreated, evt); err != nil {
		p.logger.Error("failed to publish academy admin assigned event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy admin assigned event published", zap.String("event_id", evt.EventId))
	return nil
}

func (p *AcademyEventPublisher) PublishAdminRevoked(ctx context.Context, evt *academyv1.AcademyAdminEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAcademyAdminDeleted, evt); err != nil {
		p.logger.Error("failed to publish academy admin revoked event",
			zap.String("event_id", evt.EventId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("academy admin revoked event published", zap.String("event_id", evt.EventId))
	return nil
}
