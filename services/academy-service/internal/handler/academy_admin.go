package handler

import (
	"context"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	"microservice-golang/services/academy-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type AcademyAdminHandler struct {
	academyv1.UnimplementedAcademyAdminServiceServer
	uc usecase.AcademyAdminUseCase
}

func NewAcademyAdminHandler(uc usecase.AcademyAdminUseCase) *AcademyAdminHandler {
	return &AcademyAdminHandler{uc: uc}
}

func (h *AcademyAdminHandler) RegisterGRPC(s *grpc.Server) {
	academyv1.RegisterAcademyAdminServiceServer(s, h)
}

func (h *AcademyAdminHandler) CreateAcademyAdmin(ctx context.Context, req *academyv1.CreateAcademyAdminRequest) (*academyv1.CreateAcademyAdminResponse, error) {
	academyID, err := uuid.Parse(req.GetAcademyId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid academy id"))
	}

	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		bID, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &bID
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	dtoReq := dto.CreateAcademyAdminRequest{
		AcademyID:   academyID,
		BranchID:    branchID,
		UserID:      userID,
		RoleID:      roleID,
		CreatedByID: createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.CreateAcademyAdminResponse{
		Admin: mapper.ToProtoAcademyAdmin(res),
	}, nil
}

func (h *AcademyAdminHandler) GetAcademyAdmin(ctx context.Context, req *academyv1.GetAcademyAdminRequest) (*academyv1.GetAcademyAdminResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid admin id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetAcademyAdminResponse{
		Admin: mapper.ToProtoAcademyAdmin(res),
	}, nil
}

func (h *AcademyAdminHandler) ListAcademyAdmins(ctx context.Context, req *academyv1.ListAcademyAdminsRequest) (*academyv1.ListAcademyAdminsResponse, error) {
	var academyID *uuid.UUID
	if req.AcademyId != nil && *req.AcademyId != "" {
		id, err := uuid.Parse(*req.AcademyId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid academy id"))
		}
		academyID = &id
	}

	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		id, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &id
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		id, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &id
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}

	dtoReq := dto.ListAcademyAdminsRequest{
		AcademyID: academyID,
		BranchID:  branchID,
		StatusID:  statusID,
		Search:    req.Search,
		Page:      page,
		PageSize:  pageSize,
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	admins := make([]*academyv1.AcademyAdmin, len(res.Admins))
	for i, a := range res.Admins {
		admins[i] = mapper.ToProtoAcademyAdmin(a)
	}

	return &academyv1.ListAcademyAdminsResponse{
		Admins:   admins,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *AcademyAdminHandler) UpdateAcademyAdmin(ctx context.Context, req *academyv1.UpdateAcademyAdminRequest) (*academyv1.UpdateAcademyAdminResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid admin id"))
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	dtoReq := dto.UpdateAcademyAdminRequest{
		RoleID:      roleID,
		UpdatedByID: updatedByID,
	}

	res, err := h.uc.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.UpdateAcademyAdminResponse{
		Admin: mapper.ToProtoAcademyAdmin(res),
	}, nil
}

func (h *AcademyAdminHandler) DeleteAcademyAdmin(ctx context.Context, req *academyv1.DeleteAcademyAdminRequest) (*academyv1.DeleteAcademyAdminResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid admin id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.DeleteAcademyAdminRequest{
		ID:          id,
		DeletedByID: deletedByID,
	}

	err = h.uc.Delete(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.DeleteAcademyAdminResponse{
		Id: req.GetId(),
	}, nil
}

func (h *AcademyAdminHandler) GetAcademyAdminByUser(ctx context.Context, req *academyv1.GetAcademyAdminByUserRequest) (*academyv1.GetAcademyAdminByUserResponse, error) {
	var holdingID *uuid.UUID
	if req.HoldingId != nil && *req.HoldingId != "" {
		id, err := uuid.Parse(*req.HoldingId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
		}
		holdingID = &id
	}

	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		id, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &id
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	scope := repository.AdminScope{
		HoldingID: holdingID,
		BranchID:  branchID,
	}

	res, err := h.uc.GetByUser(ctx, scope, userID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetAcademyAdminByUserResponse{
		Admin: mapper.ToProtoAcademyAdmin(res),
	}, nil
}

func (h *AcademyAdminHandler) CheckUserIsAcademyAdmin(ctx context.Context, req *academyv1.CheckUserIsAcademyAdminRequest) (*academyv1.CheckUserIsAcademyAdminResponse, error) {
	var holdingID *uuid.UUID
	if req.HoldingId != nil && *req.HoldingId != "" {
		id, err := uuid.Parse(*req.HoldingId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
		}
		holdingID = &id
	}

	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		id, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &id
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	scope := repository.AdminScope{
		HoldingID: holdingID,
		BranchID:  branchID,
	}

	isAdmin, err := h.uc.CheckUserIsAdmin(ctx, scope, userID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.CheckUserIsAcademyAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
