package usecase

import (
	"context"
	"errors"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/mapper"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type TagUseCase interface {
	Create(ctx context.Context, req dto.CreateTagRequest) (*dto.TagResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.TagResponse, error)
	GetByTypeAndName(ctx context.Context, tagType string, name string) (*dto.TagResponse, error)
	List(ctx context.Context, req dto.ListTagsRequest) (*dto.ListTagsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateTagRequest) (*dto.TagResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteTagRequest) error
	HardDelete(ctx context.Context, req dto.DeleteTagRequest) error
}

type tagUseCase struct {
	repo      repository.TagRepository
	publisher TagEventPublisher
}

func NewTagUseCase(repo repository.TagRepository, publisher TagEventPublisher) TagUseCase {
	return &tagUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *tagUseCase) Create(ctx context.Context, req dto.CreateTagRequest) (*dto.TagResponse, error) {
	tag := &entity.Tag{
		ID:          uuid.New(),
		Type:        req.Type,
		Name:        req.Name,
		Slug:        req.Slug,
		CreatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, tag); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_tags_name") {
			return nil, apperr.Conflict("name")
		}
		if postgres.IsUniqueConstraint(err, "uni_tags_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	// Build event
	evt := uc.buildEvent(metav1.TagEventType_TAG_EVENT_TYPE_CREATED, tag)
	if err := uc.publisher.PublishTagCreated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToTagResponse(tag), nil
}

func (uc *tagUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.TagResponse, error) {
	tag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tag")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToTagResponse(tag), nil
}

func (uc *tagUseCase) GetByTypeAndName(ctx context.Context, tagType string, name string) (*dto.TagResponse, error) {
	tags, err := uc.repo.GetByTypeAndName(ctx, tagType, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tag")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToTagResponse(tags), nil
}

func (uc *tagUseCase) List(ctx context.Context, req dto.ListTagsRequest) (*dto.ListTagsResponse, error) {
	tags, total, err := uc.repo.List(ctx, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.TagResponse, len(tags))
	for i, t := range tags {
		result[i] = mapper.ToTagResponse(t)
	}

	return &dto.ListTagsResponse{
		Tags:     result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *tagUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateTagRequest) (*dto.TagResponse, error) {
	tag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tag")
		}
		return nil, apperr.Internal(err)
	}

	if req.Type != nil {
		tag.Type = *req.Type
	}

	if req.Name != nil {
		tag.Name = *req.Name
	}

	if req.Slug != nil {
		tag.Slug = *req.Slug
	}

	if req.Type != nil || req.Name != nil || req.Slug != nil {
		tag.UpdatedByID = req.UpdatedByID
	}

	if err := uc.repo.Update(ctx, *tag); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_tags_name") {
			return nil, apperr.Conflict("name")
		}
		if postgres.IsUniqueConstraint(err, "uni_tags_slug") {
			return nil, apperr.Conflict("slug")
		}
		return nil, apperr.Internal(err)
	}

	evt := uc.buildEvent(metav1.TagEventType_TAG_EVENT_TYPE_UPDATED, tag)
	if err := uc.publisher.PublishTagUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToTagResponse(tag), nil
}

func (uc *tagUseCase) SoftDelete(ctx context.Context, req dto.DeleteTagRequest) error {
	tag, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("tag")
		}
		return apperr.Internal(err)
	}

	tag.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	tag.DeletedByID = &req.DeletedByID
	if err := uc.repo.Update(ctx, *tag); err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildEvent(metav1.TagEventType_TAG_EVENT_TYPE_DELETED, tag)
	if err := uc.publisher.PublishTagDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *tagUseCase) HardDelete(ctx context.Context, req dto.DeleteTagRequest) error {
	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("tag")
		}
		return apperr.Internal(err)
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (uc *tagUseCase) buildEvent(
	eventType metav1.TagEventType,
	t *entity.Tag,
) *metav1.TagEvent {
	evtID, _ := uuid.NewV7()

	evt := &metav1.TagEvent{
		EventId:    evtID.String(),
		EventType:  eventType,
		OccurredAt: timestamppb.Now(),
		TagId:      t.ID.String(),
		TagType:    t.Type,
		TagName:    t.Name,
		TagSlug:    t.Slug,
	}

	if t.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(t.DeletedAt.Time)
	}

	return evt
}
