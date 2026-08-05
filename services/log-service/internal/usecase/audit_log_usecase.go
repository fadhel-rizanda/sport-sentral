package usecase

import (
	"context"
	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type AuditLogUsecase interface {
	CreateAuditLog(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error)
	GetAuditLogByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)
	ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error)
	GetLogStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error)
}

type auditLogUsecase struct {
	repo repository.AuditLogRepository
}

func NewAuditLogUsecase(repo repository.AuditLogRepository) AuditLogUsecase {
	return &auditLogUsecase{repo: repo}
}

func (u *auditLogUsecase) CreateAuditLog(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.Status == "" {
		log.Status = entity.StatusSuccess
	}
	if log.Severity == "" {
		log.Severity = entity.SeverityInfo
	}
	log.CreatedAt = time.Now()

	if err := u.repo.Create(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (u *auditLogUsecase) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *auditLogUsecase) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	return u.repo.List(ctx, filter)
}

func (u *auditLogUsecase) GetLogStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
	return u.repo.GetStats(ctx, serviceName, startTime, endTime)
}
