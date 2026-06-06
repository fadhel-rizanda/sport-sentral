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

type AcademyHoldingHandler struct {
	academyv1.UnimplementedAcademyHoldingServiceServer
	uc usecase.AcademyHoldingUseCase
}

func NewAcademyHoldingHandler(uc usecase.AcademyHoldingUseCase) *AcademyHoldingHandler {
	return &AcademyHoldingHandler{uc: uc}
}

func (h *AcademyHoldingHandler) RegisterGRPC(s *grpc.Server) {
	academyv1.RegisterAcademyHoldingServiceServer(s, h)
}

func (h *AcademyHoldingHandler) CreateAcademyHolding(ctx context.Context, req *academyv1.CreateAcademyHoldingRequest) (*academyv1.CreateAcademyHoldingResponse, error) {
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

	var imageAttachmentID *uuid.UUID
	if req.ImageAttachmentId != nil && *req.ImageAttachmentId != "" {
		id, err := uuid.Parse(*req.ImageAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid image attachment id"))
		}
		imageAttachmentID = &id
	}

	dtoReq := dto.CreateAcademyHoldingRequest{
		Name:              req.GetName(),
		Description:       req.GetDescription(),
		Email:             req.GetEmail(),
		PhoneNumber:       req.GetPhoneNumber(),
		ImageAttachmentID: imageAttachmentID,
		StatusID:          statusID,
		StreetAddress:     req.GetStreetAddress(),
		Notes:             req.Notes,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		AdminDivisionID:   adminDivisionID,
		CreatedByID:       createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.CreateAcademyHoldingResponse{
		Holding: mapper.ToProtoAcademyHolding(res),
	}, nil
}

func (h *AcademyHoldingHandler) GetAcademyHolding(ctx context.Context, req *academyv1.GetAcademyHoldingRequest) (*academyv1.GetAcademyHoldingResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetAcademyHoldingResponse{
		Holding: mapper.ToProtoAcademyHolding(res),
	}, nil
}

func (h *AcademyHoldingHandler) ListAcademyHoldings(ctx context.Context, req *academyv1.ListAcademyHoldingsRequest) (*academyv1.ListAcademyHoldingsResponse, error) {
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

	dtoReq := dto.ListAcademyHoldingsRequest{
		StatusID: statusID,
		Search:   req.Search,
		Page:     page,
		PageSize: pageSize,
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	holdings := make([]*academyv1.AcademyHolding, len(res.Holdings))
	for i, holding := range res.Holdings {
		holdings[i] = mapper.ToProtoAcademyHolding(holding)
	}

	return &academyv1.ListAcademyHoldingsResponse{
		Holdings: holdings,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *AcademyHoldingHandler) UpdateAcademyHolding(ctx context.Context, req *academyv1.UpdateAcademyHoldingRequest) (*academyv1.UpdateAcademyHoldingResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	var imageAttachmentID *uuid.UUID
	if req.ImageAttachmentId != nil && *req.ImageAttachmentId != "" {
		imgID, err := uuid.Parse(*req.ImageAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid image attachment id"))
		}
		imageAttachmentID = &imgID
	}

	dtoReq := dto.UpdateAcademyHoldingRequest{
		ID:                id,
		Name:              req.Name,
		Description:       req.Description,
		Email:             req.Email,
		PhoneNumber:       req.PhoneNumber,
		ImageAttachmentID: imageAttachmentID,
		UpdatedByID:       updatedByID,
	}

	res, err := h.uc.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.UpdateAcademyHoldingResponse{
		Holding: mapper.ToProtoAcademyHolding(res),
	}, nil
}

func (h *AcademyHoldingHandler) DeleteAcademyHolding(ctx context.Context, req *academyv1.DeleteAcademyHoldingRequest) (*academyv1.DeleteAcademyHoldingResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid holding id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.DeleteAcademyHoldingRequest{
		ID:          id,
		DeletedByID: deletedByID,
	}

	err = h.uc.Delete(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.DeleteAcademyHoldingResponse{
		Id: req.GetId(),
	}, nil
}
