package nats

import (
	"context"
	"microservice-golang/services/log-service/internal/usecase"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type SystemEventSubscriber struct {
	syncUC      usecase.NatsLogSyncUsecase
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewSystemEventSubscriber(
	syncUC usecase.NatsLogSyncUsecase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *SystemEventSubscriber {
	return &SystemEventSubscriber{
		syncUC:      syncUC,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *SystemEventSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	s.logger.Info("received nats event message", zap.String("subject", subject))
	if subject == events.SubjectActivityLogCreated {
		return s.syncUC.ProcessActivityEvent(ctx, subject, data)
	}
	return s.syncUC.ProcessSystemEvent(ctx, subject, data)
}

func (s *SystemEventSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.IdentityStreamName,
		s.durableName,
		[]string{
			events.SubjectUserCreated,
			events.SubjectUserUpdated,
			events.SubjectUserDeleted,
			events.SubjectRoleCreated,
			events.SubjectRoleUpdated,
			events.SubjectRoleDeleted,
			events.SubjectPermissionCreated,
			events.SubjectPermissionUpdated,
			events.SubjectPermissionDeleted,
			events.SubjectActivityLogCreated,
		},
		s.handleMessage,
	)
}
