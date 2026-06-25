package usecase

import (
	"context"
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
	sharedgrpc "microservice-golang/shared/pkg/grpc"
)

type RegulatorUseCase interface {
	Create(ctx context.Context, req dto.CreateRegulatorRequest) (*dto.RegulatorResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RegulatorResponse, error)

	// Staff Management
	AddStaff(ctx context.Context, req dto.AddStaffRequest) (*dto.RegulatorStaffResponse, error)
	RemoveStaff(ctx context.Context, req dto.RemoveStaffRequest) error

	// Sport Assignment
	AssignSport(ctx context.Context, req dto.AssignSportToRegulatorRequest) error
}

type regulatorUseCase struct {
	repo       repository.RegulatorRepository
	sportRepo  repository.SportRepository
	statusRepo replicated.StatusRepository
	tagRepo    replicated.TagRepository
	userRepo   replicated.UserRepository
	publisher  SportEventPublisher
}

func NewRegulatorUseCase(
	repo repository.RegulatorRepository,
	sportRepo repository.SportRepository,
	statusRepo replicated.StatusRepository,
	tagRepo replicated.TagRepository,
	userRepo replicated.UserRepository,
	publisher SportEventPublisher,
) RegulatorUseCase {
	return &regulatorUseCase{
		repo:       repo,
		sportRepo:  sportRepo,
		statusRepo: statusRepo,
		tagRepo:    tagRepo,
		userRepo:   userRepo,
		publisher:  publisher,
	}
}

func (uc *regulatorUseCase) Create(ctx context.Context, req dto.CreateRegulatorRequest) (*dto.RegulatorResponse, error) {
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return nil, err
	}
	if activeRole != "platform_admin" {
		return nil, apperr.Forbidden("insufficient permissions")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	addressID, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Validate status exists
	if _, err := uc.statusRepo.GetByID(ctx, req.StatusID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("status")
		}
		return nil, apperr.Internal(err)
	}

	address := &entity.RegulatorAddress{
		ID:                       addressID,
		StreetAddress:            req.StreetAddress,
		Notes:                    req.Notes,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		AdministrativeDivisionID: req.AdminDivisionID,
	}

	regulator := &entity.Regulator{
		ID:               id,
		OrganizationName: req.OrganizationName,
		Code:             req.Code,
		LogoAttachmentID: req.LogoAttachmentID,
		ContactEmail:     req.ContactEmail,
		PhoneNumber:      req.PhoneNumber,
		WebsiteURL:       req.WebsiteURL,
		StatusID:         req.StatusID,
		AddressID:        addressID,
		Address:          address,
		CreatedByID:      req.CreatedByID,
		UpdatedByID:      req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, regulator); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_regulators_code") {
			return nil, apperr.Conflict("regulator code already exists")
		}
		return nil, apperr.Internal(fmt.Errorf("failed to create regulator: %w", err))
	}

	resRegulator, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRegulatorResponse(resRegulator), nil
}

func (uc *regulatorUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.RegulatorResponse, error) {
	regulator, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("regulator")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToRegulatorResponse(regulator), nil
}

func (uc *regulatorUseCase) AddStaff(ctx context.Context, req dto.AddStaffRequest) (*dto.RegulatorStaffResponse, error) {
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return nil, err
	}
	if activeRole != "platform_admin" {
		return nil, apperr.Forbidden("insufficient permissions")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Verify regulator exists
	if _, err := uc.repo.GetByID(ctx, req.RegulatorID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("regulator")
		}
		return nil, apperr.Internal(err)
	}

	// Verify user exists (replicated_users)
	if _, err := uc.userRepo.GetByID(ctx, req.UserID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("user")
		}
		return nil, apperr.Internal(err)
	}

	// Verify role tag exists
	if _, err := uc.tagRepo.GetByID(ctx, req.RoleTagID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("role tag")
		}
		return nil, apperr.Internal(err)
	}

	staff := &entity.RegulatorStaff{
		ID:          id,
		RegulatorID: req.RegulatorID,
		UserID:      req.UserID,
		RoleTagID:   req.RoleTagID,
		JoinedAt:    time.Now(),
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.AddStaff(ctx, staff); err != nil {
		return nil, apperr.Internal(fmt.Errorf("failed to add staff: %w", err))
	}

	resStaff, err := uc.repo.GetStaffByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRegulatorStaffResponse(resStaff), nil
}

func (uc *regulatorUseCase) RemoveStaff(ctx context.Context, req dto.RemoveStaffRequest) error {
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return err
	}
	if activeRole != "platform_admin" {
		return apperr.Forbidden("insufficient permissions")
	}

	_, err = uc.repo.GetStaffByID(ctx, req.StaffID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("regulator staff")
		}
		return apperr.Internal(err)
	}

	// Delete performs a hard delete (as requested)
	if err := uc.repo.RemoveStaff(ctx, req.StaffID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *regulatorUseCase) AssignSport(ctx context.Context, req dto.AssignSportToRegulatorRequest) error {
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return err
	}
	if activeRole != "platform_admin" {
		return apperr.Forbidden("insufficient permissions")
	}

	// Verify regulator exists
	if _, err := uc.repo.GetByID(ctx, req.RegulatorID); err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("regulator")
		}
		return apperr.Internal(err)
	}

	// Verify sport exists
	sport, err := uc.sportRepo.GetByID(ctx, req.SportID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("sport")
		}
		return apperr.Internal(err)
	}

	// Assign properties
	sport.RegulatorID = &req.RegulatorID
	sport.RequiresApprovalForOfficial = req.RequiresApprovalForOfficial
	sport.RequiresApprovalForRegional = req.RequiresApprovalForRegional
	sport.UpdatedByID = req.UpdatedByID

	if err := uc.sportRepo.Update(ctx, sport); err != nil {
		return apperr.Internal(fmt.Errorf("failed to assign sport to regulator: %w", err))
	}

	resSport, err := uc.sportRepo.GetByID(ctx, req.SportID)
	if err == nil {
		evt := buildSportEvent(sportv1.SportEventType_SPORT_EVENT_TYPE_UPDATED, resSport)
		_ = uc.publisher.PublishSportUpdated(ctx, evt)
	}

	return nil
}
