package usecase

import (
	"context"
	"errors"
	"fmt"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/dto"
	"microservice-golang/services/sport-service/internal/entity"
	"microservice-golang/services/sport-service/internal/mapper"
	"microservice-golang/services/sport-service/internal/repository"
	"microservice-golang/services/sport-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SportUseCase interface {
	Create(ctx context.Context, req dto.CreateSportRequest) (*dto.SportResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.SportResponse, error)
	List(ctx context.Context, req dto.ListSportsRequest) (*dto.ListSportsResponse, error)
	Update(ctx context.Context, req dto.UpdateSportRequest) (*dto.SportResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Config
	GetConfig(ctx context.Context, sportID uuid.UUID) (*dto.SportConfigResponse, error)
	UpdateConfig(ctx context.Context, req dto.UpdateSportConfigRequest) (*dto.SportConfigResponse, error)
}

type sportUseCase struct {
	repo       repository.SportRepository
	statusRepo replicated.StatusRepository
	tagRepo    replicated.TagRepository
	publisher  SportEventPublisher
}

func NewSportUseCase(
	repo repository.SportRepository,
	statusRepo replicated.StatusRepository,
	tagRepo replicated.TagRepository,
	publisher SportEventPublisher,
) SportUseCase {
	return &sportUseCase{
		repo:       repo,
		statusRepo: statusRepo,
		tagRepo:    tagRepo,
		publisher:  publisher,
	}
}

func (uc *sportUseCase) Create(ctx context.Context, req dto.CreateSportRequest) (*dto.SportResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Validate status exists
	if _, err := uc.statusRepo.GetByID(ctx, req.StatusID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}

	// Validate tier tag exists
	if _, err := uc.tagRepo.GetByID(ctx, req.TierTagID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("tier tag")
		}
		return nil, apperr.Internal(err)
	}

	sport := &entity.Sport{
		ID:               id,
		Name:             req.Name,
		Slug:             req.Slug,
		Description:      req.Description,
		IconAttachmentID: req.IconAttachmentID,
		StatusID:         req.StatusID,
		TierTagID:        req.TierTagID,
		CreatedByID:      req.CreatedByID,
		UpdatedByID:      req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, sport); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_sports_name") {
			return nil, apperr.Conflict("sport name already exists")
		}
		if postgres.IsUniqueConstraint(err, "uni_sports_slug") {
			return nil, apperr.Conflict("sport slug already exists")
		}
		return nil, apperr.Internal(fmt.Errorf("failed to create sport: %w", err))
	}

	resSport, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	evt := buildSportEvent(sportv1.SportEventType_SPORT_EVENT_TYPE_CREATED, resSport)
	if err := uc.publisher.PublishSportCreated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToSportResponse(resSport), nil
}

func (uc *sportUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.SportResponse, error) {
	sport, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("sport")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToSportResponse(sport), nil
}

func (uc *sportUseCase) List(ctx context.Context, req dto.ListSportsRequest) (*dto.ListSportsResponse, error) {
	filters := repository.SportFilters{
		TierTagID: req.TierTagID,
		StatusID:  req.StatusID,
		Search:    req.Search,
	}

	sports, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.SportResponse, len(sports))
	for i, s := range sports {
		result[i] = mapper.ToSportResponse(s)
	}

	return &dto.ListSportsResponse{
		Sports:   result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *sportUseCase) Update(ctx context.Context, req dto.UpdateSportRequest) (*dto.SportResponse, error) {
	sport, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("sport")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.Name != nil {
		sport.Name = *req.Name
		updated = true
	}
	if req.Description != nil {
		sport.Description = *req.Description
		updated = true
	}
	if req.IconAttachmentID != nil {
		sport.IconAttachmentID = req.IconAttachmentID
		updated = true
	}
	if req.TierTagID != nil {
		// Validate tier tag
		if _, err := uc.tagRepo.GetByID(ctx, *req.TierTagID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperr.NotFound("tier tag")
			}
			return nil, apperr.Internal(err)
		}
		sport.TierTagID = *req.TierTagID
		updated = true
	}
	if req.StatusID != nil {
		// Validate status
		if _, err := uc.statusRepo.GetByID(ctx, *req.StatusID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperr.NotFound("status")
			}
			return nil, apperr.Internal(err)
		}
		sport.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		sport.UpdatedByID = req.UpdatedByID
		if err := uc.repo.Update(ctx, sport); err != nil {
			if postgres.IsUniqueConstraint(err, "uni_sports_name") {
				return nil, apperr.Conflict("sport name already exists")
			}
			if postgres.IsUniqueConstraint(err, "uni_sports_slug") {
				return nil, apperr.Conflict("sport slug already exists")
			}
			return nil, apperr.Internal(err)
		}
	}

	resSport, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	evt := buildSportEvent(sportv1.SportEventType_SPORT_EVENT_TYPE_UPDATED, resSport)
	if err := uc.publisher.PublishSportUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToSportResponse(resSport), nil
}

func (uc *sportUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	sport, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("sport")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}

	sport.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	evt := buildSportEvent(sportv1.SportEventType_SPORT_EVENT_TYPE_DELETED, sport)
	if err := uc.publisher.PublishSportDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *sportUseCase) GetConfig(ctx context.Context, sportID uuid.UUID) (*dto.SportConfigResponse, error) {
	config, err := uc.repo.GetConfigBySportID(ctx, sportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("sport configuration")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToSportConfigResponse(config), nil
}

func (uc *sportUseCase) UpdateConfig(ctx context.Context, req dto.UpdateSportConfigRequest) (*dto.SportConfigResponse, error) {
	// Verify sport exists
	if _, err := uc.repo.GetByID(ctx, req.SportID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("sport")
		}
		return nil, apperr.Internal(err)
	}

	// Verify participant type tag exists
	if _, err := uc.tagRepo.GetByID(ctx, req.ParticipantTypeTagID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("participant type tag")
		}
		return nil, apperr.Internal(err)
	}

	// Verify stat tags exist
	statTags := make([]entity.Tag, len(req.StatTagIDs))
	for i, tagID := range req.StatTagIDs {
		tag, err := uc.tagRepo.GetByID(ctx, tagID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperr.NotFound(fmt.Sprintf("stat tag with id %s", tagID))
			}
			return nil, apperr.Internal(err)
		}
		statTags[i] = *tag
	}

	// Get existing config or construct a new one
	config, err := uc.repo.GetConfigBySportID(ctx, req.SportID)
	var configID uuid.UUID
	var isNew bool
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newID, err := uuid.NewV7()
			if err != nil {
				return nil, apperr.Internal(err)
			}
			configID = newID
			config = &entity.SportConfig{
				ID:        configID,
				SportID:   req.SportID,
				CreatedAt: time.Now(),
			}
			isNew = true
		} else {
			return nil, apperr.Internal(err)
		}
	} else {
		configID = config.ID
	}

	config.ParticipantTypeTagID = req.ParticipantTypeTagID
	config.MinRosterSize = req.MinRosterSize
	config.MaxRosterSize = req.MaxRosterSize
	config.TypicalRosterSize = req.TypicalRosterSize
	config.RulesURL = req.RulesURL
	config.Description = req.Description
	config.StatTags = statTags
	config.UpdatedAt = time.Now()

	// GORM schema doesn't have CreatedByID / UpdatedByID for SportConfig, but let's make sure it operates properly.
	if err := uc.repo.UpsertConfig(ctx, config); err != nil {
		return nil, apperr.Internal(fmt.Errorf("failed to save config: %w", err))
	}

	// Re-fetch to return fully populated config
	resConfig, err := uc.repo.GetConfigBySportID(ctx, req.SportID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if isNew {
		// Log config creation or do other side effects
	}

	return mapper.ToSportConfigResponse(resConfig), nil
}
