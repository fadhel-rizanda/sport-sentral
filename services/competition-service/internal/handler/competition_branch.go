package handler

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type CompetitionBranchHandler struct {
	competitionv1.UnimplementedCompetitionBranchServiceServer
	uc usecase.CompetitionBranchUseCase
}

func NewCompetitionBranchHandler(uc usecase.CompetitionBranchUseCase) *CompetitionBranchHandler {
	return &CompetitionBranchHandler{uc: uc}
}

func (h *CompetitionBranchHandler) RegisterGRPC(s *grpc.Server) {
	competitionv1.RegisterCompetitionBranchServiceServer(s, h)
}

func (h *CompetitionBranchHandler) CreateBranch(ctx context.Context, req *competitionv1.CreateBranchRequest) (*competitionv1.CreateBranchResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	competitionID, err := uuid.Parse(req.GetCompetitionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	var parentBranchID *uuid.UUID
	if req.ParentBranchId != nil && *req.ParentBranchId != "" {
		pbID, err := uuid.Parse(*req.ParentBranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent branch id"))
		}
		parentBranchID = &pbID
	}

	dtoReq := dto.CreateBranchRequest{
		Name:           req.GetName(),
		Description:    req.GetDescription(),
		StatusID:       statusID,
		ParentBranchID: parentBranchID,
	}

	res, err := h.uc.Create(ctx, adminID, competitionID, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.CreateBranchResponse{
		Branch: mapper.ToProtoBranch(res),
	}, nil
}

func (h *CompetitionBranchHandler) GetBranch(ctx context.Context, req *competitionv1.GetBranchRequest) (*competitionv1.GetBranchResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.GetBranchResponse{
		Branch: mapper.ToProtoBranch(res),
	}, nil
}

func (h *CompetitionBranchHandler) ListBranchesByCompetition(ctx context.Context, req *competitionv1.ListBranchesByCompetitionRequest) (*competitionv1.ListBranchesByCompetitionResponse, error) {
	competitionID, err := uuid.Parse(req.GetCompetitionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	res, err := h.uc.ListByCompetition(ctx, competitionID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	branches := make([]*competitionv1.Branch, len(res))
	for i, b := range res {
		branches[i] = mapper.ToProtoBranch(b)
	}

	return &competitionv1.ListBranchesByCompetitionResponse{
		Branches: branches,
	}, nil
}

func (h *CompetitionBranchHandler) UpdateBranch(ctx context.Context, req *competitionv1.UpdateBranchRequest) (*competitionv1.UpdateBranchResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		sID, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &sID
	}

	var parentBranchID *uuid.UUID
	if req.ParentBranchId != nil && *req.ParentBranchId != "" {
		pbID, err := uuid.Parse(*req.ParentBranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent branch id"))
		}
		parentBranchID = &pbID
	}

	dtoReq := dto.UpdateBranchRequest{
		Name:           req.Name,
		Description:    req.Description,
		StatusID:       statusID,
		ParentBranchID: parentBranchID,
	}

	res, err := h.uc.Update(ctx, adminID, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateBranchResponse{
		Branch: mapper.ToProtoBranch(res),
	}, nil
}

func (h *CompetitionBranchHandler) DeleteBranch(ctx context.Context, req *competitionv1.DeleteBranchRequest) (*competitionv1.DeleteBranchResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	err = h.uc.Delete(ctx, adminID, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteBranchResponse{
		Id: req.GetId(),
	}, nil
}
