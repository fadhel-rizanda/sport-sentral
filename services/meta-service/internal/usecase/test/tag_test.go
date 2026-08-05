package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

// ---- Mocks ----

type MockTagRepository struct {
	CreateFunc           func(ctx context.Context, tag *entity.Tag) error
	GetByIDFunc          func(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetByTypeAndNameFunc func(ctx context.Context, tagType, name string) (*entity.Tag, error)
	ListFunc             func(ctx context.Context, tagType *string, page, pageSize int) ([]*entity.Tag, int64, error)
	UpdateFunc           func(ctx context.Context, tag entity.Tag) error
	DeleteFunc           func(ctx context.Context, id uuid.UUID) error
}

func (m *MockTagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	return m.CreateFunc(ctx, tag)
}
func (m *MockTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockTagRepository) GetByTypeAndName(ctx context.Context, tagType, name string) (*entity.Tag, error) {
	return m.GetByTypeAndNameFunc(ctx, tagType, name)
}
func (m *MockTagRepository) List(ctx context.Context, tagType *string, page, pageSize int) ([]*entity.Tag, int64, error) {
	return m.ListFunc(ctx, tagType, page, pageSize)
}
func (m *MockTagRepository) Update(ctx context.Context, tag entity.Tag) error {
	return m.UpdateFunc(ctx, tag)
}
func (m *MockTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}

type MockTagEventPublisher struct {
	PublishTagCreatedFunc func(ctx context.Context, evt *metav1.TagEvent) error
	PublishTagUpdatedFunc func(ctx context.Context, evt *metav1.TagEvent) error
	PublishTagDeletedFunc func(ctx context.Context, evt *metav1.TagEvent) error
}

func (m *MockTagEventPublisher) PublishTagCreated(ctx context.Context, evt *metav1.TagEvent) error {
	return m.PublishTagCreatedFunc(ctx, evt)
}
func (m *MockTagEventPublisher) PublishTagUpdated(ctx context.Context, evt *metav1.TagEvent) error {
	return m.PublishTagUpdatedFunc(ctx, evt)
}
func (m *MockTagEventPublisher) PublishTagDeleted(ctx context.Context, evt *metav1.TagEvent) error {
	return m.PublishTagDeletedFunc(ctx, evt)
}

// ---- Tests ----

func TestTagUseCase_Create(t *testing.T) {
	userID := uuid.New()

	req := dto.CreateTagRequest{
		Type:        "sport_tier",
		Name:        "Tier A",
		Slug:        "tier-a",
		CreatedByID: userID,
	}

	t.Run("success", func(t *testing.T) {
		repo := &MockTagRepository{}
		publisher := &MockTagEventPublisher{}

		uc := usecase.NewTagUseCase(repo, publisher)

		repo.CreateFunc = func(ctx context.Context, tag *entity.Tag) error {
			tag.ID = uuid.New()
			return nil
		}

		published := false
		publisher.PublishTagCreatedFunc = func(ctx context.Context, evt *metav1.TagEvent) error {
			published = true
			if evt.TagName != "Tier A" {
				t.Errorf("expected published tag name Tier A, got %s", evt.TagName)
			}
			return nil
		}

		res, err := uc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.Name != "Tier A" {
			t.Errorf("expected Tier A, got %s", res.Name)
		}
		if !published {
			t.Error("expected event to be published")
		}
	})

	t.Run("conflict_slug", func(t *testing.T) {
		repo := &MockTagRepository{}
		uc := usecase.NewTagUseCase(repo, nil)

		repo.CreateFunc = func(ctx context.Context, tag *entity.Tag) error {
			return &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "uni_tags_slug",
			}
		}

		_, err := uc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsConflict(err) {
			t.Errorf("expected conflict, got %v", err)
		}
	})
}

func TestTagUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &MockTagRepository{}
		uc := usecase.NewTagUseCase(repo, nil)

		repo.GetByIDFunc = func(ctx context.Context, tagID uuid.UUID) (*entity.Tag, error) {
			return &entity.Tag{
				ID:   tagID,
				Name: "Tier B",
				Type: "sport_tier",
			}, nil
		}

		res, err := uc.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Name != "Tier B" {
			t.Errorf("expected Tier B, got %s", res.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		repo := &MockTagRepository{}
		uc := usecase.NewTagUseCase(repo, nil)

		repo.GetByIDFunc = func(ctx context.Context, tagID uuid.UUID) (*entity.Tag, error) {
			return nil, gorm.ErrRecordNotFound
		}

		_, err := uc.GetByID(context.Background(), id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperr.IsNotFound(err) {
			t.Errorf("expected not found, got %v", err)
		}
	})
}
