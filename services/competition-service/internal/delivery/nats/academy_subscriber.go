package nats

import (
	"context"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/competition-service/internal/usecase/replicated"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type AcademySubscriber struct {
	holdingSyncUC *replicated.SyncAcademyHoldingUseCase
	branchSyncUC  *replicated.SyncAcademyBranchUseCase
	adminSyncUC   *replicated.SyncAcademyAdminUseCase
	nats          *messaging.Client
	logger        *zap.Logger
	durableName   string
}

func NewAcademySubscriber(
	holdingSyncUC *replicated.SyncAcademyHoldingUseCase,
	branchSyncUC *replicated.SyncAcademyBranchUseCase,
	adminSyncUC *replicated.SyncAcademyAdminUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *AcademySubscriber {
	return &AcademySubscriber{
		holdingSyncUC: holdingSyncUC,
		branchSyncUC:  branchSyncUC,
		adminSyncUC:   adminSyncUC,
		nats:          nats,
		logger:        logger,
		durableName:   durableName,
	}
}

func (s *AcademySubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case events.SubjectAcademyHoldingCreated, events.SubjectAcademyHoldingUpdated, events.SubjectAcademyHoldingDeleted:
		var evt academyv1.AcademyHoldingEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid academy holding event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.holdingSyncUC.Sync(ctx, &evt)

	case events.SubjectAcademyBranchCreated, events.SubjectAcademyBranchUpdated, events.SubjectAcademyBranchDeleted:
		var evt academyv1.AcademyBranchEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid academy branch event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.branchSyncUC.Sync(ctx, &evt)

	case events.SubjectAcademyAdminCreated, events.SubjectAcademyAdminUpdated, events.SubjectAcademyAdminDeleted:
		var evt academyv1.AcademyAdminEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid academy admin event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.adminSyncUC.Sync(ctx, &evt)

	default:
		s.logger.Warn("unknown subject in academy stream – skipping", zap.String("subject", subject))
		return nil
	}
}

func (s *AcademySubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.AcademyStreamName,
		s.durableName,
		[]string{
			events.SubjectAcademyHoldingCreated,
			events.SubjectAcademyHoldingUpdated,
			events.SubjectAcademyHoldingDeleted,
			events.SubjectAcademyBranchCreated,
			events.SubjectAcademyBranchUpdated,
			events.SubjectAcademyBranchDeleted,
			events.SubjectAcademyAdminCreated,
			events.SubjectAcademyAdminUpdated,
			events.SubjectAcademyAdminDeleted,
		},
		s.handleMessage,
	)
}
