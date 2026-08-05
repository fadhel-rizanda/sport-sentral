package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/repository"
	"microservice-golang/services/log-service/internal/usecase"
)

type mockAuditLogRepository struct {
	createFunc   func(ctx context.Context, log *entity.AuditLog) error
	getByIDFunc  func(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)
	listFunc     func(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error)
	getStatsFunc func(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error)
}

func (m *mockAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

func (m *mockAuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAuditLogRepository) List(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockAuditLogRepository) GetStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, serviceName, startTime, endTime)
	}
	return nil, nil
}

func TestAuditLogUsecase_CreateAuditLog(t *testing.T) {
	ctx := context.Background()

	t.Run("success with default fields populating", func(t *testing.T) {
		repo := &mockAuditLogRepository{
			createFunc: func(ctx context.Context, log *entity.AuditLog) error {
				assert.NotEqual(t, uuid.Nil, log.ID)
				assert.Equal(t, entity.StatusSuccess, log.Status)
				assert.Equal(t, entity.SeverityInfo, log.Severity)
				assert.False(t, log.CreatedAt.IsZero())
				return nil
			},
		}

		uc := usecase.NewAuditLogUsecase(repo)
		input := &entity.AuditLog{
			ServiceName: "identity-service",
			Module:      "auth",
			Action:      "LOGIN",
			EntityType:  "User",
		}

		res, err := uc.CreateAuditLog(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "identity-service", res.ServiceName)
	})

	t.Run("repository error on create", func(t *testing.T) {
		repo := &mockAuditLogRepository{
			createFunc: func(ctx context.Context, log *entity.AuditLog) error {
				return errors.New("db error")
			},
		}

		uc := usecase.NewAuditLogUsecase(repo)
		input := &entity.AuditLog{
			ServiceName: "identity-service",
		}

		res, err := uc.CreateAuditLog(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestAuditLogUsecase_GetAuditLogByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &entity.AuditLog{
			ID:          testID,
			ServiceName: "venue-service",
		}
		repo := &mockAuditLogRepository{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
				assert.Equal(t, testID, id)
				return expected, nil
			},
		}

		uc := usecase.NewAuditLogUsecase(repo)
		res, err := uc.GetAuditLogByID(ctx, testID)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestAuditLogUsecase_ListAuditLogs(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedLogs := []entity.AuditLog{
			{ID: uuid.New(), ServiceName: "sport-service"},
		}
		repo := &mockAuditLogRepository{
			listFunc: func(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
				return expectedLogs, 1, nil
			},
		}

		uc := usecase.NewAuditLogUsecase(repo)
		logs, total, err := uc.ListAuditLogs(ctx, repository.AuditLogFilter{Page: 1, Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
	})
}

func TestAuditLogUsecase_GetLogStats(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedStats := &entity.LogStats{
			TotalAuditLogs:    100,
			TotalActivityLogs: 50,
		}
		repo := &mockAuditLogRepository{
			getStatsFunc: func(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
				return expectedStats, nil
			},
		}

		uc := usecase.NewAuditLogUsecase(repo)
		stats, err := uc.GetLogStats(ctx, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedStats, stats)
	})
}
