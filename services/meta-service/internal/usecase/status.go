package usecase

import (
	"context"
	"github.com/google/uuid"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type StatusUseCase interface {
	Create(ctx context.Context, req CreateStatusRequest) (*StatusResponse, error)
	GetByTypeAndName(ctx context.Context, statusType, name string) (*StatusResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*StatusResponse, error)
	ListByType(ctx context.Context, statusType string) ([]*StatusResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateStatusRequest) (*StatusResponse, error)
	SoftDelete(ctx context.Context, req DeleteStatusRequest) error
	HardDelete(ctx context.Context, req DeleteStatusRequest) error
}

type statusUseCase struct {
	repo repository.StatusRepository
}

func NewStatusUseCase(repo repository.StatusRepository) StatusUseCase {
	return &statusUseCase{repo: repo}
}

func (uc *statusUseCase) Create(ctx context.Context, req CreateStatusRequest) (*StatusResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	status := &entity.Status{
		ID:        id,
		Type:      req.Type,
		Name:      req.Name,
		CreatedBy: req.CreatedBy,
	}

	if err := uc.repo.Create(ctx, *status); err != nil {
		if postgres.IsUniqueConstraint(err, "statuses_type_name_key") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}

	return toStatusResponse(status), nil
}

func (uc *statusUseCase) GetByTypeAndName(ctx context.Context, statusType, name string) (*StatusResponse, error) {
	s, err := uc.repo.GetByTypeAndName(ctx, statusType, name)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}
	return toStatusResponse(s), nil
}

func (uc *statusUseCase) GetByID(ctx context.Context, id uuid.UUID) (*StatusResponse, error) {
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}
	return toStatusResponse(s), nil
}

func (uc *statusUseCase) ListByType(ctx context.Context, statusType string) ([]*StatusResponse, error) {
	statuses, err := uc.repo.ListByType(ctx, statusType)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	result := make([]*StatusResponse, len(statuses))
	for i, s := range statuses {
		result[i] = toStatusResponse(s)
	}
	return result, nil
}

func (uc *statusUseCase) Update(ctx context.Context, id uuid.UUID, req UpdateStatusRequest) (*StatusResponse, error) {
	status, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}

	if req.Name != nil {
		status.Name = *req.Name
	}

	if req.Type != nil {
		status.Type = *req.Type
	}

	if req.Name != nil || req.Type != nil {
		status.UpdatedBy = req.UpdatedBy
	}

	if err := uc.repo.Update(ctx, *status); err != nil {
		if postgres.IsUniqueConstraint(err, "statuses_type_name_key") {
			return nil, apperr.Conflict("name")
		}
		return nil, apperr.Internal(err)
	}

	return toStatusResponse(status), nil
}

func (uc *statusUseCase) SoftDelete(ctx context.Context, req DeleteStatusRequest) error {
	if err := uc.repo.SoftDelete(ctx, req.ID, req.DeletedBy); err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("status")
		}
		return apperr.Internal(err)
	}
	return nil
}

func (uc *statusUseCase) HardDelete(ctx context.Context, req DeleteStatusRequest) error {
	if err := uc.repo.HardDelete(ctx, req.ID); err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("status")
		}
		return apperr.Internal(err)
	}
	return nil
}
