package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	replicatedRepo "microservice-golang/services/academy-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	"microservice-golang/shared/pkg/constants"
	apperr "microservice-golang/shared/pkg/errors"
)

type AcademyBranchUseCase interface {
	Create(ctx context.Context, req dto.CreateAcademyBranchRequest) (*dto.AcademyBranchResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyBranchResponse, error)
	List(ctx context.Context, req dto.ListAcademyBranchesRequest) (*dto.ListAcademyBranchesResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyBranchRequest) (*dto.AcademyBranchResponse, error)
	Delete(ctx context.Context, req dto.DeleteAcademyBranchRequest) error
}

type academyBranchUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.AcademyBranchRepository
	publisher      AcademyBranchEventPublisher
}

func NewAcademyBranchUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.AcademyBranchRepository,
	publisher AcademyBranchEventPublisher,
) AcademyBranchUseCase {
	return &academyBranchUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		publisher:      publisher,
	}
}

func (uc *academyBranchUseCase) Create(ctx context.Context, req dto.CreateAcademyBranchRequest) (*dto.AcademyBranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyCreate); err != nil {
		return nil, err
	}
	branchID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	addressID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	address := entity.AcademyBranchAddress{
		ID:                       addressID,
		BranchID:                 branchID,
		IsPrimary:                true,
		StreetAddress:            req.StreetAddress,
		Notes:                    req.Notes,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		AdministrativeDivisionID: req.AdminDivisionID,
	}

	branch := &entity.AcademyBranch{
		ID:          branchID,
		HoldingID:   req.HoldingID,
		SportID:     req.SportID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		StatusID:    req.StatusID,
		Addresses:   []entity.AcademyBranchAddress{address},
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, branch); err != nil {
		return nil, apperr.Internal(err)
	}

	// Fetch holding fully preloaded to return it
	resBranch, err := uc.repo.GetByID(ctx, branchID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	_ = uc.publisher.PublishBranchCreated(ctx, &academyv1.AcademyBranchEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_CREATED,
		OccurredAt: timestamppb.New(time.Now()),
		BranchId:   resBranch.ID.String(),
		HoldingId:  resBranch.HoldingID.String(),
		SportId:    resBranch.SportID.String(),
		Name:       resBranch.Name,
		StatusId:   resBranch.StatusID.String(),
	})

	return mapper.ToAcademyBranchResponse(resBranch), nil
}

func (uc *academyBranchUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyBranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy branch")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyBranchResponse(branch), nil
}

func (uc *academyBranchUseCase) List(ctx context.Context, req dto.ListAcademyBranchesRequest) (*dto.ListAcademyBranchesResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	filters := repository.AcademyBranchFilters{
		HoldingID: req.HoldingID,
		SportID:   req.SportID,
		StatusID:  req.StatusID,
		Search:    req.Search,
	}

	branches, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AcademyBranchResponse, len(branches))
	for i, b := range branches {
		result[i] = mapper.ToAcademyBranchResponse(b)
	}

	return &dto.ListAcademyBranchesResponse{
		Branches: result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *academyBranchUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyBranchRequest) (*dto.AcademyBranchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyUpdate); err != nil {
		return nil, err
	}
	branch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy branch")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.Name != nil {
		branch.Name = *req.Name
		updated = true
	}
	if req.Email != nil {
		branch.Email = *req.Email
		updated = true
	}
	if req.PhoneNumber != nil {
		branch.PhoneNumber = *req.PhoneNumber
		updated = true
	}
	if req.StatusID != nil {
		branch.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		branch.UpdatedByID = req.UpdatedByID
		if err := uc.repo.Update(ctx, branch); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	resBranch, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	_ = uc.publisher.PublishBranchUpdated(ctx, &academyv1.AcademyBranchEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_UPDATED,
		OccurredAt: timestamppb.New(time.Now()),
		BranchId:   resBranch.ID.String(),
		HoldingId:  resBranch.HoldingID.String(),
		SportId:    resBranch.SportID.String(),
		Name:       resBranch.Name,
		StatusId:   resBranch.StatusID.String(),
	})

	return mapper.ToAcademyBranchResponse(resBranch), nil
}

func (uc *academyBranchUseCase) Delete(ctx context.Context, req dto.DeleteAcademyBranchRequest) error {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyDelete); err != nil {
		return err
	}
	// Check if exists
	branch, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy branch")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	_ = uc.publisher.PublishBranchDeleted(ctx, &academyv1.AcademyBranchEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyBranchEventType_ACADEMY_BRANCH_EVENT_TYPE_DELETED,
		OccurredAt: timestamppb.New(time.Now()),
		BranchId:   branch.ID.String(),
		HoldingId:  branch.HoldingID.String(),
		SportId:    branch.SportID.String(),
		Name:       branch.Name,
		StatusId:   branch.StatusID.String(),
	})

	return nil
}
