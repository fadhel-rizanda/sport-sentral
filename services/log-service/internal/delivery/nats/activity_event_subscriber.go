package nats

import (
	"context"
	"microservice-golang/services/log-service/internal/usecase"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type ActivityEventSubscriber struct {
	syncUC      usecase.NatsLogSyncUsecase
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewActivityEventSubscriber(
	syncUC usecase.NatsLogSyncUsecase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *ActivityEventSubscriber {
	return &ActivityEventSubscriber{
		syncUC:      syncUC,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *ActivityEventSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	s.logger.Info("received activity event message", zap.String("subject", subject))
	return s.syncUC.ProcessActivityEvent(ctx, subject, data)
}

func (s *ActivityEventSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.LogStreamName,
		s.durableName,
		[]string{
			events.SubjectActivityLogCreated,
		},
		s.handleMessage,
	)
}
