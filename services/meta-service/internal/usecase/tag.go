package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type TagUseCase interface {
	Create(ctx context.Context, req CreateTagRequest) (*TagResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*TagResponse, error)
	GetByTypeAndName(ctx context.Context, tagType string, name string) (*TagResponse, error)
	ListByType(ctx context.Context, tagType string) ([]*TagResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateTagRequest) (*TagResponse, error)
	SoftDelete(ctx context.Context, req DeleteTagRequest) error
	HardDelete(ctx context.Context, req DeleteTagRequest) error
}

type tagUseCase struct {
	repo repository.TagRepository
}

func NewTagUseCase(repo repository.TagRepository) TagUseCase {
	return &tagUseCase{repo: repo}
}

func (uc *tagUseCase) Create(ctx context.Context, req CreateTagRequest) (*TagResponse, error) {
	var slug string
	if req.Slug == nil {
		slug = slugify(req.Name)
	} else {
		slug = *req.Slug
	}

	tag := &entity.Tag{
		ID:        uuid.New(),
		Type:      req.Type,
		Name:      req.Name,
		Slug:      slug,
		CreatedBy: req.CreatedBy,
	}

	if err := uc.repo.Create(ctx, tag); err != nil {
		if postgres.IsUniqueConstraint(err, "tags_type_name_key") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}

	return toTagResponse(tag), nil
}

func (uc *tagUseCase) GetByID(ctx context.Context, id uuid.UUID) (*TagResponse, error) {
	tag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tag")
		}
		return nil, apperr.Internal(err)
	}
	return toTagResponse(tag), nil
}

func (uc *tagUseCase) GetByTypeAndName(ctx context.Context, tagType string, name string) (*TagResponse, error) {
	tags, err := uc.repo.GetByTypeAndName(ctx, tagType, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tag")
		}
		return nil, apperr.Internal(err)
	}
	return toTagResponse(tags), nil
}

func (uc *tagUseCase) ListByType(ctx context.Context, tagType string) ([]*TagResponse, error) {
	tags, err := uc.repo.ListByType(ctx, tagType)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	result := make([]*TagResponse, len(tags))
	for i, t := range tags {
		result[i] = toTagResponse(t)
	}
	return result, nil
}

func (uc *tagUseCase) Update(ctx context.Context, id uuid.UUID, req UpdateTagRequest) (*TagResponse, error) {
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
		tag.UpdatedBy = req.UpdatedBy
	}

	if err := uc.repo.Update(ctx, *tag); err != nil {
		if postgres.IsUniqueConstraint(err, "tags_type_name_key") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}
	return toTagResponse(tag), nil
}

func (uc *tagUseCase) SoftDelete(ctx context.Context, req DeleteTagRequest) error {
	if err := uc.repo.SoftDelete(ctx, req.ID, req.DeletedBy); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("tag")
		}
		return apperr.Internal(err)
	}
	return nil
}

func (uc *tagUseCase) HardDelete(ctx context.Context, req DeleteTagRequest) error {
	if err := uc.repo.HardDelete(ctx, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("tag")
		}
		return apperr.Internal(err)
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
