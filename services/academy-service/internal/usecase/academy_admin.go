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

type AcademyAdminUseCase interface {
	Create(ctx context.Context, req dto.CreateAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyAdminResponse, error)
	List(ctx context.Context, req dto.ListAcademyAdminsRequest) (*dto.ListAcademyAdminsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	Delete(ctx context.Context, req dto.DeleteAcademyAdminRequest) error
	GetByUser(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (*dto.AcademyAdminResponse, error)
	CheckUserIsAdmin(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (bool, error)

	Assign(ctx context.Context, req dto.AssignAcademyAdminRequest) (*dto.AcademyAdminResponse, error)
	Revoke(ctx context.Context, req dto.RevokeAcademyAdminRequest) error
}

type academyAdminUseCase struct {
	permissionRepo replicatedRepo.PermissionRepository
	repo           repository.AcademyAdminRepository
	publisher      AcademyAdminEventPublisher
}

func NewAcademyAdminUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.AcademyAdminRepository,
	publisher AcademyAdminEventPublisher,
) AcademyAdminUseCase {
	return &academyAdminUseCase{
		permissionRepo: permissionRepo,
		repo:           repo,
		publisher:      publisher,
	}
}

func (uc *academyAdminUseCase) Create(ctx context.Context, req dto.CreateAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyManage); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	admin := &entity.AcademyAdmin{
		ID:          id,
		AcademyID:   req.AcademyID,
		BranchID:    req.BranchID,
		UserID:      req.UserID,
		RoleID:      req.RoleID,
		CreatedByID: req.CreatedByID,
		UpdatedByID: req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var branchIDStr *string
	if resAdmin.BranchID != nil {
		str := resAdmin.BranchID.String()
		branchIDStr = &str
	}
	_ = uc.publisher.PublishAdminAssigned(ctx, &academyv1.AcademyAdminEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_ASSIGNED,
		OccurredAt: timestamppb.New(time.Now()),
		AdminId:    resAdmin.ID.String(),
		AcademyId:  resAdmin.AcademyID.String(),
		BranchId:   branchIDStr,
		UserId:     resAdmin.UserID.String(),
		RoleId:     resAdmin.RoleID.String(),
	})

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.AcademyAdminResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	admin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyAdminResponse(admin), nil
}

func (uc *academyAdminUseCase) List(ctx context.Context, req dto.ListAcademyAdminsRequest) (*dto.ListAcademyAdminsResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	filters := repository.AdminFilters{
		AdminScope: repository.AdminScope{
			HoldingID: req.AcademyID,
			BranchID:  req.BranchID,
		},
		StatusID: req.StatusID,
		Search:   req.Search,
	}

	admins, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.AcademyAdminResponse, len(admins))
	for i, a := range admins {
		result[i] = mapper.ToAcademyAdminResponse(a)
	}

	return &dto.ListAcademyAdminsResponse{
		Admins:   result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *academyAdminUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyManage); err != nil {
		return nil, err
	}
	admin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}

	admin.RoleID = req.RoleID
	admin.UpdatedByID = req.UpdatedByID

	if err := uc.repo.Update(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var branchIDStr *string
	if resAdmin.BranchID != nil {
		str := resAdmin.BranchID.String()
		branchIDStr = &str
	}
	_ = uc.publisher.PublishAdminAssigned(ctx, &academyv1.AcademyAdminEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_ASSIGNED,
		OccurredAt: timestamppb.New(time.Now()),
		AdminId:    resAdmin.ID.String(),
		AcademyId:  resAdmin.AcademyID.String(),
		BranchId:   branchIDStr,
		UserId:     resAdmin.UserID.String(),
		RoleId:     resAdmin.RoleID.String(),
	})

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) Delete(ctx context.Context, req dto.DeleteAcademyAdminRequest) error {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyManage); err != nil {
		return err
	}
	admin, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy admin")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var branchIDStr *string
	if admin.BranchID != nil {
		str := admin.BranchID.String()
		branchIDStr = &str
	}
	_ = uc.publisher.PublishAdminRevoked(ctx, &academyv1.AcademyAdminEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_REVOKED,
		OccurredAt: timestamppb.New(time.Now()),
		AdminId:    admin.ID.String(),
		AcademyId:  admin.AcademyID.String(),
		BranchId:   branchIDStr,
		UserId:     admin.UserID.String(),
		RoleId:     admin.RoleID.String(),
	})

	return nil
}

func (uc *academyAdminUseCase) GetByUser(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (*dto.AcademyAdminResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyRead); err != nil {
		return nil, err
	}
	admin, err := uc.repo.GetByUser(ctx, scope, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("academy admin")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToAcademyAdminResponse(admin), nil
}

func (uc *academyAdminUseCase) CheckUserIsAdmin(ctx context.Context, scope repository.AdminScope, userID uuid.UUID) (bool, error) {
	isAdmin, err := uc.repo.CheckUserIsAdmin(ctx, scope, userID)
	if err != nil {
		return false, apperr.Internal(err)
	}
	return isAdmin, nil
}

func (uc *academyAdminUseCase) Assign(ctx context.Context, req dto.AssignAcademyAdminRequest) (*dto.AcademyAdminResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyManage); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	now := time.Now()
	admin := &entity.AcademyAdmin{
		ID:           id,
		AcademyID:    req.AcademyID,
		BranchID:     req.BranchID,
		UserID:       req.UserID,
		RoleID:       req.RoleID,
		CreatedByID:  req.AssignedByID,
		UpdatedByID:  req.AssignedByID,
		ApprovedAt:   &now,
		ApprovedByID: &req.AssignedByID,
	}

	if err := uc.repo.Assign(ctx, admin); err != nil {
		return nil, apperr.Internal(err)
	}

	resAdmin, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var branchIDStr *string
	if resAdmin.BranchID != nil {
		str := resAdmin.BranchID.String()
		branchIDStr = &str
	}
	_ = uc.publisher.PublishAdminAssigned(ctx, &academyv1.AcademyAdminEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_ASSIGNED,
		OccurredAt: timestamppb.New(time.Now()),
		AdminId:    resAdmin.ID.String(),
		AcademyId:  resAdmin.AcademyID.String(),
		BranchId:   branchIDStr,
		UserId:     resAdmin.UserID.String(),
		RoleId:     resAdmin.RoleID.String(),
	})

	return mapper.ToAcademyAdminResponse(resAdmin), nil
}

func (uc *academyAdminUseCase) Revoke(ctx context.Context, req dto.RevokeAcademyAdminRequest) error {
	if err := uc.permissionRepo.Validate(ctx, constants.PermissionAcademyManage); err != nil {
		return err
	}
	admin, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("academy admin")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Revoke(ctx, req.ID, req.RevokedByID); err != nil {
		return apperr.Internal(err)
	}

	// Publish event
	evtID, _ := uuid.NewV7()
	var branchIDStr *string
	if admin.BranchID != nil {
		str := admin.BranchID.String()
		branchIDStr = &str
	}
	_ = uc.publisher.PublishAdminRevoked(ctx, &academyv1.AcademyAdminEvent{
		EventId:    evtID.String(),
		EventType:  academyv1.AcademyAdminEventType_ACADEMY_ADMIN_EVENT_TYPE_REVOKED,
		OccurredAt: timestamppb.New(time.Now()),
		AdminId:    admin.ID.String(),
		AcademyId:  admin.AcademyID.String(),
		BranchId:   branchIDStr,
		UserId:     admin.UserID.String(),
		RoleId:     admin.RoleID.String(),
	})

	return nil
}
