package replicated

import (
	"context"
	"errors"
	"fmt"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncPermissionUseCase struct {
	permissionRepo replicated.PermissionRepository
	logger         *zap.Logger
}

func NewSyncPermissionUseCase(
	permissionRepo replicated.PermissionRepository,
	logger *zap.Logger,
) *SyncPermissionUseCase {
	return &SyncPermissionUseCase{
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

func (uc *SyncPermissionUseCase) upsert(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	if evt.PermissionId == "" || evt.PermissionResource == "" || evt.PermissionAction == "" || evt.PermissionSlug == "" {
		uc.logger.Warn("incomplete permission event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.PermissionId)
	if err != nil {
		uc.logger.Warn("invalid permission_id – skipping",
			zap.String("permission_id", evt.PermissionId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.permissionRepo.Upsert(ctx, entity.Permission{
		ID:       id,
		Resource: evt.PermissionResource,
		Action:   evt.PermissionAction,
		Slug:     evt.PermissionSlug,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert permission replicas: %w", err))
	}

	uc.logger.Info("permission replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("permission_id", evt.PermissionId),
	)

	return nil
}

func (uc *SyncPermissionUseCase) delete(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	if evt.PermissionId == "" {
		uc.logger.Warn("missing permission_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.PermissionId)
	if err != nil {
		uc.logger.Warn("invalid permission_id – skipping",
			zap.String("permission_id", evt.PermissionId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.permissionRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("permission already deleted", zap.String("permission_id", evt.PermissionId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete permission replicas: %w", err))
	}

	uc.logger.Info("permission replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("permission_id", evt.PermissionId),
	)

	return nil
}

func (uc *SyncPermissionUseCase) Sync(ctx context.Context, evt *rbacv1.PermissionEvent) error {
	uc.logger.Info("processing permission event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("permission_id", evt.PermissionId),
	)

	switch evt.EventType {
	case rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_CREATED,
		rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case rbacv1.PermissionEventType_PERMISSION_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown permission event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
