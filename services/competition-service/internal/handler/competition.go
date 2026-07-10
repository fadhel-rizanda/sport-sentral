package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type CompetitionHandler struct {
	competitionv1.UnimplementedCompetitionServiceServer
	uc usecase.CompetitionUseCase
}

func NewCompetitionHandler(uc usecase.CompetitionUseCase) *CompetitionHandler {
	return &CompetitionHandler{uc: uc}
}

func (h *CompetitionHandler) RegisterGRPC(s *grpc.Server) {
	competitionv1.RegisterCompetitionServiceServer(s, h)
}

func extractAdminID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, apperr.Unauthorized("missing metadata")
	}
	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		userIDs = md.Get("x-user-id")
	}
	if len(userIDs) == 0 {
		return uuid.Nil, apperr.Unauthorized("user id is required in metadata")
	}
	id, err := uuid.Parse(userIDs[0])
	if err != nil {
		return uuid.Nil, apperr.InvalidArgument("invalid user id in metadata")
	}
	return id, nil
}

func (h *CompetitionHandler) CreateCompetition(ctx context.Context, req *competitionv1.CreateCompetitionRequest) (*competitionv1.CreateCompetitionResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	var hostAcademyBranchID *uuid.UUID
	if req.HostAcademyBranchId != nil && *req.HostAcademyBranchId != "" {
		hID, err := uuid.Parse(*req.HostAcademyBranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid host academy branch id"))
		}
		hostAcademyBranchID = &hID
	}

	tierID, err := uuid.Parse(req.GetTierId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	var endDate *time.Time
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	if req.StartDate == nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("start date is required"))
	}

	dtoReq := dto.CreateCompetitionRequest{
		SportID:             sportID,
		HostAcademyBranchID: hostAcademyBranchID,
		Name:                req.GetName(),
		Description:         req.GetDescription(),
		TierID:              tierID,
		StartDate:           req.StartDate.AsTime(),
		EndDate:             endDate,
		StatusID:            statusID,
	}

	res, err := h.uc.Create(ctx, adminID, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.CreateCompetitionResponse{
		Competition: mapper.ToProtoCompetition(res),
	}, nil
}

func (h *CompetitionHandler) GetCompetition(ctx context.Context, req *competitionv1.GetCompetitionRequest) (*competitionv1.GetCompetitionResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.GetCompetitionResponse{
		Competition: mapper.ToProtoCompetition(res),
	}, nil
}

func (h *CompetitionHandler) ListCompetitions(ctx context.Context, req *competitionv1.ListCompetitionsRequest) (*competitionv1.ListCompetitionsResponse, error) {
	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		id, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &id
	}

	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		id, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
		}
		sportID = &id
	}

	var tierID *uuid.UUID
	if req.TierId != nil && *req.TierId != "" {
		id, err := uuid.Parse(*req.TierId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier id"))
		}
		tierID = &id
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

	filters := dto.CompetitionFilters{
		BranchID: branchID,
		SportID:  sportID,
		TierID:   tierID,
		StatusID: statusID,
		Search:   req.Search,
	}

	res, total, err := h.uc.List(ctx, filters, page, pageSize)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	competitions := make([]*competitionv1.Competition, len(res))
	for i, c := range res {
		competitions[i] = mapper.ToProtoCompetition(c)
	}

	return &competitionv1.ListCompetitionsResponse{
		Competitions: competitions,
		Total:        total,
		Page:         int32(page),
		PageSize:     int32(pageSize),
	}, nil
}

func (h *CompetitionHandler) UpdateCompetition(ctx context.Context, req *competitionv1.UpdateCompetitionRequest) (*competitionv1.UpdateCompetitionResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		sID, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
		}
		sportID = &sID
	}

	var hostAcademyBranchID *uuid.UUID
	if req.HostAcademyBranchId != nil {
		if *req.HostAcademyBranchId != "" {
			hID, err := uuid.Parse(*req.HostAcademyBranchId)
			if err != nil {
				return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid host academy branch id"))
			}
			hostAcademyBranchID = &hID
		} else {
			nilID := uuid.Nil
			hostAcademyBranchID = &nilID
		}
	}

	var tierID *uuid.UUID
	if req.TierId != nil && *req.TierId != "" {
		tID, err := uuid.Parse(*req.TierId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier id"))
		}
		tierID = &tID
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		sID, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &sID
	}

	var startDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}

	var endDate *time.Time
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	dtoReq := dto.UpdateCompetitionRequest{
		SportID:             sportID,
		HostAcademyBranchID: hostAcademyBranchID,
		Name:                req.Name,
		Description:         req.Description,
		TierID:              tierID,
		StartDate:           startDate,
		EndDate:             endDate,
		StatusID:            statusID,
	}

	res, err := h.uc.Update(ctx, adminID, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateCompetitionResponse{
		Competition: mapper.ToProtoCompetition(res),
	}, nil
}

func (h *CompetitionHandler) DeleteCompetition(ctx context.Context, req *competitionv1.DeleteCompetitionRequest) (*competitionv1.DeleteCompetitionResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	err = h.uc.Delete(ctx, adminID, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteCompetitionResponse{
		Id: req.GetId(),
	}, nil
}

func (h *CompetitionHandler) UpdateCompetitionStatus(ctx context.Context, req *competitionv1.UpdateCompetitionStatusRequest) (*competitionv1.UpdateCompetitionStatusResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	err = h.uc.UpdateStatus(ctx, adminID, id, statusID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateCompetitionStatusResponse{
		Id:       req.GetId(),
		StatusId: req.GetStatusId(),
	}, nil
}
