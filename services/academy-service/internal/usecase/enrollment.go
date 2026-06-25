package usecase

import (
	"context"

	"github.com/google/uuid"

	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	replicatedRepo "microservice-golang/services/academy-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type EnrollmentUseCase interface {
	Create(ctx context.Context, req dto.CreateEnrollmentRequest) (*dto.EnrollmentResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.EnrollmentResponse, error)
	GetByBranchAndAthlete(ctx context.Context, branchID, athleteID uuid.UUID) (*dto.EnrollmentResponse, error)
	List(ctx context.Context, req dto.ListEnrollmentsRequest) (*dto.ListEnrollmentsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateEnrollmentRequest) (*dto.EnrollmentResponse, error)
	Delete(ctx context.Context, req dto.DeleteEnrollmentRequest) error
}

type enrollmentUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.EnrollmentRepository
}

func NewEnrollmentUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.EnrollmentRepository,
) EnrollmentUseCase {
	return &enrollmentUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
	}
}

func (uc *enrollmentUseCase) Create(ctx context.Context, req dto.CreateEnrollmentRequest) (*dto.EnrollmentResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "academy.manage"); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	enrollment := &entity.Enrollment{
		ID:              id,
		AcademyBranchID: req.AcademyBranchID,
		AthleteID:       req.AthleteID,
		JoinedAt:        req.JoinedAt,
		ExpiresAt:       req.ExpiresAt,
		StatusID:        req.StatusID,
		CreatedByID:     req.CreatedByID,
		UpdatedByID:     req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, enrollment); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_academy_athlete") {
			return nil, apperr.Conflict("athlete is already enrolled in this branch")
		}
		return nil, apperr.Internal(err)
	}

	resEnrollment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToEnrollmentResponse(resEnrollment), nil
}

func (uc *enrollmentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.EnrollmentResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "academy.read"); err != nil {
		return nil, err
	}
	enrollment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("enrollment")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToEnrollmentResponse(enrollment), nil
}

func (uc *enrollmentUseCase) GetByBranchAndAthlete(ctx context.Context, branchID, athleteID uuid.UUID) (*dto.EnrollmentResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "academy.read"); err != nil {
		return nil, err
	}
	enrollment, err := uc.repo.GetByBranchAndAthlete(ctx, branchID, athleteID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("enrollment")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToEnrollmentResponse(enrollment), nil
}

func (uc *enrollmentUseCase) List(ctx context.Context, req dto.ListEnrollmentsRequest) (*dto.ListEnrollmentsResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "academy.read"); err != nil {
		return nil, err
	}
	filters := repository.EnrollmentFilters{
		BranchID:  req.BranchID,
		AthleteID: req.AthleteID,
		StatusID:  req.StatusID,
		Search:    req.Search,
	}

	enrollments, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.EnrollmentResponse, len(enrollments))
	for i, e := range enrollments {
		result[i] = mapper.ToEnrollmentResponse(e)
	}

	return &dto.ListEnrollmentsResponse{
		Enrollments: result,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (uc *enrollmentUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEnrollmentRequest) (*dto.EnrollmentResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "academy.manage"); err != nil {
		return nil, err
	}
	enrollment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("enrollment")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.LeftAt != nil {
		enrollment.LeftAt = req.LeftAt
		updated = true
	}
	if req.ExpiresAt != nil {
		enrollment.ExpiresAt = req.ExpiresAt
		updated = true
	}
	if req.StatusID != nil {
		enrollment.StatusID = *req.StatusID
		updated = true
	}

	if updated {
		enrollment.UpdatedByID = req.UpdatedByID
		if err := uc.repo.Update(ctx, enrollment); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	resEnrollment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToEnrollmentResponse(resEnrollment), nil
}

func (uc *enrollmentUseCase) Delete(ctx context.Context, req dto.DeleteEnrollmentRequest) error {
	if err := uc.permissionRepo.Validate(ctx, "academy.manage"); err != nil {
		return err
	}
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("enrollment")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
