package replicated

import (
	"context"
	"errors"
	"fmt"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/repository/replicated"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncAcademyAdminUseCase struct {
	adminRepo replicated.AcademyAdminRepository
	logger    *zap.Logger
}

func NewSyncAcademyAdminUseCase(
	adminRepo replicated.AcademyAdminRepository,
	logger *zap.Logger,
) *SyncAcademyAdminUseCase {
	return &SyncAcademyAdminUseCase{
		adminRepo: adminRepo,
		logger:    logger,
	}
}

func (uc *SyncAcademyAdminUseCase) upsert(ctx context.Context, evt *academyv1.AcademyAdminEvent) error {
	if evt.AdminId == "" || evt.AcademyId == "" || evt.UserId == "" || evt.RoleId == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.AdminId)
	if err != nil {
		uc.logger.Warn("invalid admin_id – skipping",
			zap.String("admin_id", evt.AdminId),
			zap.Error(err),
		)
		return nil
	}

	academyID, err := uuid.Parse(evt.AcademyId)
	if err != nil {
		uc.logger.Warn("invalid academy_id – skipping",
			zap.String("academy_id", evt.AcademyId),
			zap.Error(err),
		)
		return nil
	}

	var branchID *uuid.UUID
	if evt.BranchId != nil && *evt.BranchId != "" {
		parsedBranchID, err := uuid.Parse(*evt.BranchId)
		if err != nil {
			uc.logger.Warn("invalid branch_id – skipping",
				zap.String("branch_id", *evt.BranchId),
				zap.Error(err),
			)
		} else {
			branchID = &parsedBranchID
		}
	}

	userID, err := uuid.Parse(evt.UserId)
	if err != nil {
		uc.logger.Warn("invalid user_id – skipping",
			zap.String("user_id", evt.UserId),
			zap.Error(err),
		)
		return nil
	}

	roleID, err := uuid.Parse(evt.RoleId)
	if err != nil {
		uc.logger.Warn("invalid role_id – skipping",
			zap.String("role_id", evt.RoleId),
			zap.Error(err),
		)
		return nil
	}

	admin := entity.AcademyAdmin{
		ID:        id,
		AcademyID: academyID,
		BranchID:  branchID,
		UserID:    userID,
		RoleID:    roleID,
	}

	if evt.DeletedAt != nil {
		admin.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.adminRepo.Upsert(ctx, admin); err != nil {
		return apperr.Internal(fmt.Errorf("upsert academy admin replicas: %w", err))
	}

	uc.logger.Info("academy admin replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("admin_id", evt.AdminId),
	)

	return nil
}

func (uc *SyncAcademyAdminUseCase) delete(ctx context.Context, evt *academyv1.AcademyAdminEvent) error {
	if evt.AdminId == "" {
		uc.logger.Warn("missing admin_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.AdminId)
	if err != nil {
		uc.logger.Warn("invalid admin_id – skipping",
			zap.String("admin_id", evt.AdminId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.adminRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("academy admin already deleted", zap.String("admin_id", evt.AdminId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete academy admin replicas: %w", err))
	}

	uc.logger.Info("academy admin replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("admin_id", evt.AdminId),
	)

	return nil
}

func (uc *SyncAcademyAdminUseCase) Sync(ctx context.Context, evt *academyv1.AcademyAdminEvent) error {
	uc.logger.Info("processing academy admin event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("admin_id", evt.AdminId),
	)

	switch evt.EventType {
	case academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_ASSIGNED:
		return uc.upsert(ctx, evt)

	case academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_REVOKED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
