package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/repository"
	"microservice-golang/services/log-service/internal/usecase"
)

type mockAuditUsecaseForSync struct {
	createAuditLogFunc func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error)
}

func (m *mockAuditUsecaseForSync) CreateAuditLog(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
	if m.createAuditLogFunc != nil {
		return m.createAuditLogFunc(ctx, log)
	}
	return log, nil
}

func (m *mockAuditUsecaseForSync) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditUsecaseForSync) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (m *mockAuditUsecaseForSync) GetLogStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
	return nil, nil
}

type mockActivityUsecaseForSync struct {
	createActivityLogFunc func(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error)
}

func (m *mockActivityUsecaseForSync) CreateActivityLog(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error) {
	if m.createActivityLogFunc != nil {
		return m.createActivityLogFunc(ctx, log)
	}
	return log, nil
}

func (m *mockActivityUsecaseForSync) GetActivityLogByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
	return nil, nil
}

func (m *mockActivityUsecaseForSync) ListActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
	return nil, 0, nil
}

func TestNatsLogSyncUsecase_ProcessSystemEvent(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully process event with json payload containing id", func(t *testing.T) {
		testEntityID := uuid.New().String()
		payload := []byte(`{"id":"` + testEntityID + `", "name":"Tournament A"}`)

		auditMock := &mockAuditUsecaseForSync{
			createAuditLogFunc: func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
				assert.Equal(t, entity.ServiceNameNatsBus, log.ServiceName)
				assert.Equal(t, "competition.created", log.Action)
				assert.Equal(t, testEntityID, *log.EntityID)
				assert.Nil(t, log.UserID)
				return log, nil
			},
		}

		syncUC := usecase.NewNatsLogSyncUseCase(auditMock, nil, logger)
		err := syncUC.ProcessSystemEvent(ctx, "competition.created", payload)
		assert.NoError(t, err)
	})

	t.Run("successfully process event with json payload containing id and user_id", func(t *testing.T) {
		testEntityID := uuid.New().String()
		testUserID := uuid.New()
		payload := []byte(`{"id":"` + testEntityID + `", "user_id":"` + testUserID.String() + `", "name":"Tournament B"}`)

		auditMock := &mockAuditUsecaseForSync{
			createAuditLogFunc: func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
				assert.Equal(t, testEntityID, *log.EntityID)
				assert.NotNil(t, log.UserID)
				assert.Equal(t, testUserID, *log.UserID)
				return log, nil
			},
		}

		syncUC := usecase.NewNatsLogSyncUseCase(auditMock, nil, logger)
		err := syncUC.ProcessSystemEvent(ctx, "competition.created", payload)
		assert.NoError(t, err)
	})

	t.Run("returns error when audit creation fails", func(t *testing.T) {
		auditMock := &mockAuditUsecaseForSync{
			createAuditLogFunc: func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
				return nil, errors.New("audit creation error")
			},
		}

		syncUC := usecase.NewNatsLogSyncUseCase(auditMock, nil, logger)
		err := syncUC.ProcessSystemEvent(ctx, "user.created", []byte(`{}`))
		assert.Error(t, err)
	})
}

func TestNatsLogSyncUsecase_ProcessActivityEvent(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	testUserID := uuid.New()

	t.Run("successfully process activity event payload", func(t *testing.T) {
		payload := []byte(`{"user_id":"` + testUserID.String() + `", "action":"SEARCH_VENUE", "resource_type":"Venue", "description":"User searched for venue"}`)

		activityMock := &mockActivityUsecaseForSync{
			createActivityLogFunc: func(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error) {
				assert.Equal(t, testUserID, log.UserID)
				assert.Equal(t, "SEARCH_VENUE", log.Action)
				assert.Equal(t, "Venue", log.ResourceType)
				return log, nil
			},
		}

		syncUC := usecase.NewNatsLogSyncUseCase(nil, activityMock, logger)
		err := syncUC.ProcessActivityEvent(ctx, "log.activity.created", payload)
		assert.NoError(t, err)
	})

	t.Run("skips when payload missing user_id", func(t *testing.T) {
		syncUC := usecase.NewNatsLogSyncUseCase(nil, nil, logger)
		err := syncUC.ProcessActivityEvent(ctx, "log.activity.created", []byte(`{}`))
		assert.NoError(t, err)
	})

	t.Run("skips when user_id is invalid UUID", func(t *testing.T) {
		syncUC := usecase.NewNatsLogSyncUseCase(nil, nil, logger)
		err := syncUC.ProcessActivityEvent(ctx, "log.activity.created", []byte(`{"user_id":"invalid-uuid"}`))
		assert.NoError(t, err)
	})
}
