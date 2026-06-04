package nats

import (
	"context"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/meta-service/internal/usecase"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type UserSubscriber struct {
	syncUC      *usecase.UserSyncUseCase
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewUserSubscriber(
	syncUC *usecase.UserSyncUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *UserSubscriber {
	return &UserSubscriber{
		syncUC:      syncUC,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *UserSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	var evt userv1.UserEvent
	if err := proto.Unmarshal(data, &evt); err != nil {
		s.logger.Warn("invalid protobuf – skipping",
			zap.String("subject", subject),
			zap.Error(err),
		)
		return nil
	}

	return s.syncUC.Sync(ctx, &evt)
}

func (s *UserSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.IdentityStreamName,
		events.MetaDurableName,
		[]string{
			events.SubjectUserCreated,
			events.SubjectUserUpdated,
			events.SubjectUserDeleted,
		},
		s.handleMessage,
	)
}
