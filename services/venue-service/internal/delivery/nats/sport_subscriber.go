package nats

import (
	"context"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/venue-service/internal/usecase/replicated"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type SportSubscriber struct {
	sportSyncUC *replicated.SyncSportUseCase
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewSportSubscriber(
	sportSyncUC *replicated.SyncSportUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *SportSubscriber {
	return &SportSubscriber{
		sportSyncUC: sportSyncUC,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *SportSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case events.SubjectSportCreated, events.SubjectSportUpdated, events.SubjectSportDeleted:
		var evt sportv1.SportEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid sport event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.sportSyncUC.Sync(ctx, &evt)

	default:
		s.logger.Warn("unknown subject in sport stream – skipping", zap.String("subject", subject))
		return nil
	}
}

func (s *SportSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.SportStreamName,
		s.durableName,
		[]string{
			events.SubjectSportCreated,
			events.SubjectSportUpdated,
			events.SubjectSportDeleted,
		},
		s.handleMessage,
	)
}
