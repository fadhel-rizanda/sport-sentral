package nats

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/identity-service/internal/usecase"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type StatusSubscriber struct {
	syncUC      *usecase.StatusUseCase
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewStatusSubscriber(
	syncUC *usecase.StatusUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *StatusSubscriber {
	return &StatusSubscriber{
		syncUC:      syncUC,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *StatusSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	var evt metav1.StatusEvent
	if err := proto.Unmarshal(data, &evt); err != nil {
		s.logger.Warn("invalid protobuf – skipping",
			zap.String("subject", subject),
			zap.Error(err),
		)
		return nil
	}

	return s.syncUC.Sync(ctx, &evt)
}

func (s *StatusSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.MetaStreamName,
		events.IdentityDurableName,
		[]string{
			events.SubjectStatusCreated,
			events.SubjectStatusUpdated,
			events.SubjectStatusDeleted,
		},
		s.handleMessage,
	)
}
