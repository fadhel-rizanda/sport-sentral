package handler

import (
	"context"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type AcademyBranchHandler struct {
	academyv1.UnimplementedAcademyBranchServiceServer
	uc usecase.AcademyBranchUseCase
}

func NewAcademyBranchHandler(uc usecase.AcademyBranchUseCase) *AcademyBranchHandler {
	return &AcademyBranchHandler{uc: uc}
}

func (h *AcademyBranchHandler) RegisterGRPC(s *grpc.Server) {
	academyv1.RegisterAcademyBranchServiceServer(s, h)
}

func (h *AcademyBranchHandler) CreateAcademyBranch(ctx context.Context, req *academyv1.CreateAcademyBranchRequest) (*academyv1.CreateAcademyBranchResponse, error) {
	holdingID, err := uuid.Parse(req.GetHoldingId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
	}

	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	adminDivisionID, err := uuid.Parse(req.GetAdminDivisionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid admin division id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	dtoReq := dto.CreateAcademyBranchRequest{
		HoldingID:       holdingID,
		SportID:         sportID,
		Name:            req.GetName(),
		Email:           req.GetEmail(),
		PhoneNumber:     req.GetPhoneNumber(),
		StatusID:        statusID,
		StreetAddress:   req.GetStreetAddress(),
		Notes:           req.Notes,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		AdminDivisionID: adminDivisionID,
		CreatedByID:     createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.CreateAcademyBranchResponse{
		Branch: mapper.ToProtoAcademyBranch(res),
	}, nil
}

func (h *AcademyBranchHandler) GetAcademyBranch(ctx context.Context, req *academyv1.GetAcademyBranchRequest) (*academyv1.GetAcademyBranchResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetAcademyBranchResponse{
		Branch: mapper.ToProtoAcademyBranch(res),
	}, nil
}

func (h *AcademyBranchHandler) ListAcademyBranches(ctx context.Context, req *academyv1.ListAcademyBranchesRequest) (*academyv1.ListAcademyBranchesResponse, error) {
	var holdingID *uuid.UUID
	if req.HoldingId != nil && *req.HoldingId != "" {
		id, err := uuid.Parse(*req.HoldingId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
		}
		holdingID = &id
	}

	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		id, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
		}
		sportID = &id
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

	dtoReq := dto.ListAcademyBranchesRequest{
		HoldingID: holdingID,
		SportID:   sportID,
		StatusID:  statusID,
		Search:    req.Search,
		Page:      page,
		PageSize:  pageSize,
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	branches := make([]*academyv1.AcademyBranch, len(res.Branches))
	for i, b := range res.Branches {
		branches[i] = mapper.ToProtoAcademyBranch(b)
	}

	return &academyv1.ListAcademyBranchesResponse{
		Branches: branches,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *AcademyBranchHandler) UpdateAcademyBranch(ctx context.Context, req *academyv1.UpdateAcademyBranchRequest) (*academyv1.UpdateAcademyBranchResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		sID, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &sID
	}

	dtoReq := dto.UpdateAcademyBranchRequest{
		ID:          id,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		StatusID:    statusID,
		UpdatedByID: updatedByID,
	}

	res, err := h.uc.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.UpdateAcademyBranchResponse{
		Branch: mapper.ToProtoAcademyBranch(res),
	}, nil
}

func (h *AcademyBranchHandler) DeleteAcademyBranch(ctx context.Context, req *academyv1.DeleteAcademyBranchRequest) (*academyv1.DeleteAcademyBranchResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.DeleteAcademyBranchRequest{
		ID:          id,
		DeletedByID: deletedByID,
	}

	err = h.uc.Delete(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.DeleteAcademyBranchResponse{
		Id: req.GetId(),
	}, nil
}
