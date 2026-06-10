package nats

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/sport-service/internal/usecase/replicated"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type MetaSubscriber struct {
	statusSyncUC *replicated.SyncStatusUseCase
	tagSyncUC    *replicated.SyncTagUseCase
	nats         *messaging.Client
	logger       *zap.Logger
	durableName  string
}

func NewMetaSubscriber(
	statusSyncUC *replicated.SyncStatusUseCase,
	tagSyncUC *replicated.SyncTagUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *MetaSubscriber {
	return &MetaSubscriber{
		statusSyncUC: statusSyncUC,
		tagSyncUC:    tagSyncUC,
		nats:         nats,
		logger:       logger,
		durableName:  durableName,
	}
}

func (s *MetaSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case events.SubjectStatusCreated, events.SubjectStatusUpdated, events.SubjectStatusDeleted:
		var evt metav1.StatusEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid status event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.statusSyncUC.Sync(ctx, &evt)

	case events.SubjectTagCreated, events.SubjectTagUpdated, events.SubjectTagDeleted:
		var evt metav1.TagEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid tag event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.tagSyncUC.Sync(ctx, &evt)

	default:
		s.logger.Warn("unknown subject in meta stream – skipping", zap.String("subject", subject))
		return nil
	}
}

func (s *MetaSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.MetaStreamName,
		s.durableName,
		[]string{
			events.SubjectStatusCreated,
			events.SubjectStatusUpdated,
			events.SubjectStatusDeleted,
			events.SubjectTagCreated,
			events.SubjectTagUpdated,
			events.SubjectTagDeleted,
		},
		s.handleMessage,
	)
}
