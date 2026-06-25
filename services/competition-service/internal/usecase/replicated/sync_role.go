package replicated

import (
	"context"
	"errors"
	"fmt"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/competition-service/internal/entity"
	rplatd "microservice-golang/services/competition-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncRoleUseCase struct {
	roleRepo rplatd.RoleRepository
	logger   *zap.Logger
}

func NewSyncRoleUseCase(
	roleRepo rplatd.RoleRepository,
	logger *zap.Logger,
) *SyncRoleUseCase {
	return &SyncRoleUseCase{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

func (uc *SyncRoleUseCase) upsert(ctx context.Context, evt *rbacv1.RoleEvent) error {
	if evt.RoleId == "" || evt.RoleName == "" || evt.RoleSlug == "" {
		uc.logger.Warn("incomplete role event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.RoleId)
	if err != nil {
		uc.logger.Warn("invalid role_id – skipping",
			zap.String("role_id", evt.RoleId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.roleRepo.Upsert(ctx, entity.Role{
		ID:   id,
		Name: evt.RoleName,
		Slug: evt.RoleSlug,
	}); err != nil {
		return apperr.Internal(fmt.Errorf("upsert role replicas: %w", err))
	}

	// Sync many-to-many permissions
	permissionIDs := make([]uuid.UUID, 0, len(evt.PermissionIds))
	for _, permIDStr := range evt.PermissionIds {
		permID, err := uuid.Parse(permIDStr)
		if err != nil {
			uc.logger.Warn("invalid permission_id in role event – skipping this permission ID",
				zap.String("permission_id", permIDStr),
				zap.Error(err),
			)
			continue
		}
		permissionIDs = append(permissionIDs, permID)
	}

	if err := uc.roleRepo.ReplacePermissions(ctx, id, permissionIDs); err != nil {
		return apperr.Internal(fmt.Errorf("replace role permission replicas: %w", err))
	}

	uc.logger.Info("role replicas and permissions synced",
		zap.String("event_id", evt.EventId),
		zap.String("role_id", evt.RoleId),
		zap.Int("permissions_count", len(permissionIDs)),
	)

	return nil
}

func (uc *SyncRoleUseCase) delete(ctx context.Context, evt *rbacv1.RoleEvent) error {
	if evt.RoleId == "" {
		uc.logger.Warn("missing role_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.RoleId)
	if err != nil {
		uc.logger.Warn("invalid role_id – skipping",
			zap.String("role_id", evt.RoleId),
			zap.Error(err),
		)
		return nil
	}

	// Clear associations first
	if err := uc.roleRepo.ReplacePermissions(ctx, id, []uuid.UUID{}); err != nil {
		return apperr.Internal(fmt.Errorf("clear role permissions before delete: %w", err))
	}

	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("role already deleted", zap.String("role_id", evt.RoleId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete role replicas: %w", err))
	}

	uc.logger.Info("role replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("role_id", evt.RoleId),
	)

	return nil
}

func (uc *SyncRoleUseCase) Sync(ctx context.Context, evt *rbacv1.RoleEvent) error {
	uc.logger.Info("processing role event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("role_id", evt.RoleId),
	)

	switch evt.EventType {
	case rbacv1.RoleEventType_ROLE_EVENT_TYPE_CREATED,
		rbacv1.RoleEventType_ROLE_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case rbacv1.RoleEventType_ROLE_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown role event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
