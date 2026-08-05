package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	logv1 "microservice-golang/gen/log/v1"
	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/handler"
	"microservice-golang/services/log-service/internal/repository"
)

type mockAuditUC struct {
	createAuditLogFunc  func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error)
	getAuditLogByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)
	listAuditLogsFunc   func(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error)
	getLogStatsFunc     func(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error)
}

func (m *mockAuditUC) CreateAuditLog(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
	if m.createAuditLogFunc != nil {
		return m.createAuditLogFunc(ctx, log)
	}
	return log, nil
}

func (m *mockAuditUC) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	if m.getAuditLogByIDFunc != nil {
		return m.getAuditLogByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAuditUC) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	if m.listAuditLogsFunc != nil {
		return m.listAuditLogsFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockAuditUC) GetLogStats(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
	if m.getLogStatsFunc != nil {
		return m.getLogStatsFunc(ctx, serviceName, startTime, endTime)
	}
	return nil, nil
}

type mockActivityUC struct {
	createActivityLogFunc  func(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error)
	getActivityLogByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error)
	listActivityLogsFunc   func(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error)
}

func (m *mockActivityUC) CreateActivityLog(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error) {
	if m.createActivityLogFunc != nil {
		return m.createActivityLogFunc(ctx, log)
	}
	return log, nil
}

func (m *mockActivityUC) GetActivityLogByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
	if m.getActivityLogByIDFunc != nil {
		return m.getActivityLogByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockActivityUC) ListActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
	if m.listActivityLogsFunc != nil {
		return m.listActivityLogsFunc(ctx, filter)
	}
	return nil, 0, nil
}

func TestLogHandler_CreateAuditLog(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		auditMock := &mockAuditUC{
			createAuditLogFunc: func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
				log.ID = uuid.New()
				log.CreatedAt = time.Now()
				return log, nil
			},
		}

		h := handler.NewLogHandler(auditMock, nil)
		req := &logv1.CreateAuditLogRequest{
			ServiceName: "identity-service",
			Module:      "auth",
			Action:      "LOGIN",
			EntityType:  "User",
			Status:      "SUCCESS",
			Severity:    "INFO",
		}

		res, err := h.CreateAuditLog(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "identity-service", res.Log.ServiceName)
	})

	t.Run("usecase error", func(t *testing.T) {
		auditMock := &mockAuditUC{
			createAuditLogFunc: func(ctx context.Context, log *entity.AuditLog) (*entity.AuditLog, error) {
				return nil, errors.New("usecase error")
			},
		}

		h := handler.NewLogHandler(auditMock, nil)
		res, err := h.CreateAuditLog(ctx, &logv1.CreateAuditLogRequest{ServiceName: "test"})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestLogHandler_GetAuditLog(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("invalid uuid", func(t *testing.T) {
		h := handler.NewLogHandler(nil, nil)
		res, err := h.GetAuditLog(ctx, &logv1.GetAuditLogRequest{Id: "invalid-uuid"})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("not found", func(t *testing.T) {
		auditMock := &mockAuditUC{
			getAuditLogByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
				return nil, nil
			},
		}
		h := handler.NewLogHandler(auditMock, nil)
		res, err := h.GetAuditLog(ctx, &logv1.GetAuditLogRequest{Id: testID.String()})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		auditMock := &mockAuditUC{
			getAuditLogByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
				return &entity.AuditLog{ID: testID, ServiceName: "venue-service", CreatedAt: time.Now()}, nil
			},
		}
		h := handler.NewLogHandler(auditMock, nil)
		res, err := h.GetAuditLog(ctx, &logv1.GetAuditLogRequest{Id: testID.String()})
		assert.NoError(t, err)
		assert.Equal(t, testID.String(), res.Log.Id)
	})
}

func TestLogHandler_ListAuditLogs(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		auditMock := &mockAuditUC{
			listAuditLogsFunc: func(ctx context.Context, filter repository.AuditLogFilter) ([]entity.AuditLog, int64, error) {
				return []entity.AuditLog{
					{ID: uuid.New(), ServiceName: "sport-service", CreatedAt: time.Now()},
				}, 1, nil
			},
		}

		h := handler.NewLogHandler(auditMock, nil)
		res, err := h.ListAuditLogs(ctx, &logv1.ListAuditLogsRequest{
			Page:      1,
			Limit:     10,
			StartTime: timestamppb.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.TotalCount)
		assert.Len(t, res.Logs, 1)
	})
}

func TestLogHandler_CreateActivityLog(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid request (missing uuid)", func(t *testing.T) {
		h := handler.NewLogHandler(nil, nil)
		res, err := h.CreateActivityLog(ctx, &logv1.CreateActivityLogRequest{
			UserId: "invalid-uuid",
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		userUUID := uuid.New()
		activityMock := &mockActivityUC{
			createActivityLogFunc: func(ctx context.Context, log *entity.ActivityLog) (*entity.ActivityLog, error) {
				log.ID = uuid.New()
				log.CreatedAt = time.Now()
				return log, nil
			},
		}

		h := handler.NewLogHandler(nil, activityMock)
		res, err := h.CreateActivityLog(ctx, &logv1.CreateActivityLogRequest{
			UserId:       userUUID.String(),
			Action:       "VIEW_VENUE",
			ResourceType: "Venue",
			Description:  "Viewed venue",
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestLogHandler_GetLogStats(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		auditMock := &mockAuditUC{
			getLogStatsFunc: func(ctx context.Context, serviceName *string, startTime, endTime *time.Time) (*entity.LogStats, error) {
				return &entity.LogStats{
					TotalAuditLogs:    10,
					TotalActivityLogs: 5,
					CountBySeverity:   []entity.SeverityCount{{Severity: "INFO", Count: 10}},
				}, nil
			},
		}

		h := handler.NewLogHandler(auditMock, nil)
		res, err := h.GetLogStats(ctx, &logv1.GetLogStatsRequest{})
		assert.NoError(t, err)
		assert.Equal(t, int64(10), res.TotalAuditLogs)
		assert.Equal(t, int64(5), res.TotalActivityLogs)
	})
}
