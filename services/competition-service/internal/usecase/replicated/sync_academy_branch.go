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

type SyncAcademyBranchUseCase struct {
	branchRepo replicated.AcademyBranchRepository
	logger     *zap.Logger
}

func NewSyncAcademyBranchUseCase(
	branchRepo replicated.AcademyBranchRepository,
	logger *zap.Logger,
) *SyncAcademyBranchUseCase {
	return &SyncAcademyBranchUseCase{
		branchRepo: branchRepo,
		logger:     logger,
	}
}

func (uc *SyncAcademyBranchUseCase) upsert(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	if evt.BranchId == "" || evt.HoldingId == "" || evt.SportId == "" || evt.Name == "" || evt.StatusId == "" {
		uc.logger.Warn("incomplete event – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.BranchId)
	if err != nil {
		uc.logger.Warn("invalid branch_id – skipping",
			zap.String("branch_id", evt.BranchId),
			zap.Error(err),
		)
		return nil
	}

	holdingID, err := uuid.Parse(evt.HoldingId)
	if err != nil {
		uc.logger.Warn("invalid holding_id – skipping",
			zap.String("holding_id", evt.HoldingId),
			zap.Error(err),
		)
		return nil
	}

	sportID, err := uuid.Parse(evt.SportId)
	if err != nil {
		uc.logger.Warn("invalid sport_id – skipping",
			zap.String("sport_id", evt.SportId),
			zap.Error(err),
		)
		return nil
	}

	statusID, err := uuid.Parse(evt.StatusId)
	if err != nil {
		uc.logger.Warn("invalid status_id – skipping",
			zap.String("status_id", evt.StatusId),
			zap.Error(err),
		)
		return nil
	}

	branch := entity.AcademyBranch{
		ID:        id,
		HoldingID: holdingID,
		SportID:   sportID,
		Name:      evt.Name,
		StatusID:  statusID,
	}

	if evt.DeletedAt != nil {
		branch.DeletedAt = gorm.DeletedAt{
			Time:  evt.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	if err := uc.branchRepo.Upsert(ctx, branch); err != nil {
		return apperr.Internal(fmt.Errorf("upsert academy branch replicas: %w", err))
	}

	uc.logger.Info("academy branch replicas upserted",
		zap.String("event_id", evt.EventId),
		zap.String("branch_id", evt.BranchId),
	)

	return nil
}

func (uc *SyncAcademyBranchUseCase) delete(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	if evt.BranchId == "" {
		uc.logger.Warn("missing branch_id – skipping", zap.String("event_id", evt.EventId))
		return nil
	}

	id, err := uuid.Parse(evt.BranchId)
	if err != nil {
		uc.logger.Warn("invalid branch_id – skipping",
			zap.String("branch_id", evt.BranchId),
			zap.Error(err),
		)
		return nil
	}

	if err := uc.branchRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.Debug("academy branch already deleted", zap.String("branch_id", evt.BranchId))
			return nil
		}
		return apperr.Internal(fmt.Errorf("delete academy branch replicas: %w", err))
	}

	uc.logger.Info("academy branch replicas deleted",
		zap.String("event_id", evt.EventId),
		zap.String("branch_id", evt.BranchId),
	)

	return nil
}

func (uc *SyncAcademyBranchUseCase) Sync(ctx context.Context, evt *academyv1.AcademyBranchEvent) error {
	uc.logger.Info("processing academy branch event",
		zap.String("event_id", evt.EventId),
		zap.Any("type", evt.EventType),
		zap.String("branch_id", evt.BranchId),
	)

	switch evt.EventType {
	case academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_CREATED,
		academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_UPDATED:
		return uc.upsert(ctx, evt)

	case academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_DELETED:
		return uc.delete(ctx, evt)

	default:
		uc.logger.Warn("unknown event type – skipping", zap.Any("type", evt.EventType))
		return nil
	}
}
