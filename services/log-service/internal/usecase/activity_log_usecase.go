package usecase

import (
	"context"
	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ActivityLogUsecase interface {
	CreateActivityLog(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error)
	GetActivityLogByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error)
	ListActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error)
}

type activityLogUsecase struct {
	repo repository.ActivityLogRepository
}

func NewActivityLogUsecase(repo repository.ActivityLogRepository) ActivityLogUsecase {
	return &activityLogUsecase{repo: repo}
}

func (u *activityLogUsecase) CreateActivityLog(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now()

	if err := u.repo.Create(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (u *activityLogUsecase) GetActivityLogByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *activityLogUsecase) ListActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
	return u.repo.List(ctx, filter)
}
