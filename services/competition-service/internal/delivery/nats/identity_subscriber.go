package nats

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/competition-service/internal/usecase/replicated"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type IdentitySubscriber struct {
	userSyncUC       *replicated.SyncUserUseCase
	roleSyncUC       *replicated.SyncRoleUseCase
	permissionSyncUC *replicated.SyncPermissionUseCase
	nats             *messaging.Client
	logger           *zap.Logger
	durableName      string
}

func NewIdentitySubscriber(
	userSyncUC *replicated.SyncUserUseCase,
	roleSyncUC *replicated.SyncRoleUseCase,
	permissionSyncUC *replicated.SyncPermissionUseCase,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *IdentitySubscriber {
	return &IdentitySubscriber{
		userSyncUC:       userSyncUC,
		roleSyncUC:       roleSyncUC,
		permissionSyncUC: permissionSyncUC,
		nats:             nats,
		logger:           logger,
		durableName:      durableName,
	}
}

func (s *IdentitySubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case events.SubjectUserCreated, events.SubjectUserUpdated, events.SubjectUserDeleted:
		var evt userv1.UserEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid user event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.userSyncUC.Sync(ctx, &evt)

	case events.SubjectRoleCreated, events.SubjectRoleUpdated, events.SubjectRoleDeleted:
		var evt rbacv1.RoleEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid role event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.roleSyncUC.Sync(ctx, &evt)

	case events.SubjectPermissionCreated, events.SubjectPermissionUpdated, events.SubjectPermissionDeleted:
		var evt rbacv1.PermissionEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid permission event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}
		return s.permissionSyncUC.Sync(ctx, &evt)

	default:
		s.logger.Warn("unknown subject in identity stream – skipping", zap.String("subject", subject))
		return nil
	}
}

func (s *IdentitySubscriber) Listen(ctx context.Context) error {
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
		},
		s.handleMessage,
	)
}
