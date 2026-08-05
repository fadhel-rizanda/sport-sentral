package test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"microservice-golang/services/log-service/internal/entity"
	"microservice-golang/services/log-service/internal/repository"
	"microservice-golang/services/log-service/internal/usecase"
)

type mockActivityLogRepository struct {
	createFunc  func(ctx context.Context, log *entity.ActivityLog) error
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error)
	listFunc    func(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error)
}

func (m *mockActivityLogRepository) Create(ctx context.Context, log *entity.ActivityLog) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

func (m *mockActivityLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockActivityLogRepository) List(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func TestActivityLogUsecase_CreateActivityLog(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockActivityLogRepository{
			createFunc: func(ctx context.Context, log *entity.ActivityLog) error {
				assert.NotEqual(t, uuid.Nil, log.ID)
				assert.False(t, log.CreatedAt.IsZero())
				return nil
			},
		}

		uc := usecase.NewActivityLogUsecase(repo)
		input := &entity.ActivityLog{
			UserID:       uuid.New(),
			Action:       "VIEW_COURT",
			ResourceType: "Venue",
			Description:  "User viewed court detail",
		}

		res, err := uc.CreateActivityLog(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockActivityLogRepository{
			createFunc: func(ctx context.Context, log *entity.ActivityLog) error {
				return errors.New("db error")
			},
		}

		uc := usecase.NewActivityLogUsecase(repo)
		res, err := uc.CreateActivityLog(ctx, &entity.ActivityLog{})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestActivityLogUsecase_GetActivityLogByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &entity.ActivityLog{
			ID:     testID,
			Action: "SEARCH_ACADEMY",
		}
		repo := &mockActivityLogRepository{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.ActivityLog, error) {
				return expected, nil
			},
		}

		uc := usecase.NewActivityLogUsecase(repo)
		res, err := uc.GetActivityLogByID(ctx, testID)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestActivityLogUsecase_ListActivityLogs(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []entity.ActivityLog{
			{ID: uuid.New(), Action: "BOOK_COURT"},
		}
		repo := &mockActivityLogRepository{
			listFunc: func(ctx context.Context, filter repository.ActivityLogFilter) ([]entity.ActivityLog, int64, error) {
				return expected, 1, nil
			},
		}

		uc := usecase.NewActivityLogUsecase(repo)
		logs, total, err := uc.ListActivityLogs(ctx, repository.ActivityLogFilter{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
	})
}
