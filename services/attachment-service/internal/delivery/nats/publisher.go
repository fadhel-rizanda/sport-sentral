package nats

import (
	"context"

	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type AttachmentEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewAttachmentEventPublisher(nats *messaging.Client, logger *zap.Logger) *AttachmentEventPublisher {
	return &AttachmentEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *AttachmentEventPublisher) PublishAttachmentCreated(ctx context.Context, evt *attachmentv1.AttachmentCreatedEvent) error {
	if p.nats == nil {
		return nil
	}
	if err := p.nats.Publish(ctx, events.SubjectAttachmentCreated, evt); err != nil {
		p.logger.Error("failed to publish attachment created event",
			zap.String("attachment_id", evt.AttachmentId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("attachment created event published", zap.String("attachment_id", evt.AttachmentId))
	return nil
}

func (p *AttachmentEventPublisher) PublishAttachmentUpdated(ctx context.Context, evt *attachmentv1.AttachmentUpdatedEvent) error {
	if p.nats == nil {
		return nil
	}
	if err := p.nats.Publish(ctx, events.SubjectAttachmentUpdated, evt); err != nil {
		p.logger.Error("failed to publish attachment updated event",
			zap.String("attachment_id", evt.AttachmentId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("attachment updated event published", zap.String("attachment_id", evt.AttachmentId))
	return nil
}

func (p *AttachmentEventPublisher) PublishAttachmentDeleted(ctx context.Context, evt *attachmentv1.AttachmentDeletedEvent) error {
	if p.nats == nil {
		return nil
	}
	if err := p.nats.Publish(ctx, events.SubjectAttachmentDeleted, evt); err != nil {
		p.logger.Error("failed to publish attachment deleted event",
			zap.String("attachment_id", evt.AttachmentId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("attachment deleted event published", zap.String("attachment_id", evt.AttachmentId))
	return nil
}
