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

type MatchHandler struct {
	competitionv1.UnimplementedMatchServiceServer
	uc usecase.MatchUseCase
}

func NewMatchHandler(uc usecase.MatchUseCase) *MatchHandler {
	return &MatchHandler{uc: uc}
}

func (h *MatchHandler) RegisterGRPC(s *grpc.Server) {
	competitionv1.RegisterMatchServiceServer(s, h)
}

func (h *MatchHandler) CreateMatch(ctx context.Context, req *competitionv1.CreateMatchRequest) (*competitionv1.CreateMatchResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	branchID, err := uuid.Parse(req.GetBranchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	participants := make([]dto.CreateMatchParticipantRequest, len(req.GetParticipants()))
	for i, p := range req.GetParticipants() {
		rosterID, err := uuid.Parse(p.GetRosterId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid participant roster id"))
		}
		formatTagID, err := uuid.Parse(p.GetFormatTagId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid participant format tag id"))
		}
		resultTagID, err := uuid.Parse(p.GetResultTagId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid participant result tag id"))
		}

		participants[i] = dto.CreateMatchParticipantRequest{
			RosterID:    rosterID,
			FormatTagID: formatTagID,
			ResultTagID: resultTagID,
			Score:       p.GetScore(),
		}
	}

	dtoReq := dto.CreateMatchRequest{
		ScheduledAt:  req.ScheduledAt.AsTime(),
		Location:     req.GetLocation(),
		Referee:      req.GetReferee(),
		Notes:        req.GetNotes(),
		Participants: participants,
	}

	res, err := h.uc.Create(ctx, adminID, branchID, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.CreateMatchResponse{
		Match: mapper.ToProtoMatch(res),
	}, nil
}

func (h *MatchHandler) GetMatch(ctx context.Context, req *competitionv1.GetMatchRequest) (*competitionv1.GetMatchResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.GetMatchResponse{
		Match: mapper.ToProtoMatch(res),
	}, nil
}

func (h *MatchHandler) ListMatchesByBranch(ctx context.Context, req *competitionv1.ListMatchesByBranchRequest) (*competitionv1.ListMatchesByBranchResponse, error) {
	branchID, err := uuid.Parse(req.GetBranchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}

	res, total, err := h.uc.ListByBranch(ctx, branchID, page, pageSize)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	matches := make([]*competitionv1.Match, len(res))
	for i, m := range res {
		matches[i] = mapper.ToProtoMatch(m)
	}

	return &competitionv1.ListMatchesByBranchResponse{
		Matches:  matches,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (h *MatchHandler) UpdateMatchStatus(ctx context.Context, req *competitionv1.UpdateMatchStatusRequest) (*competitionv1.UpdateMatchStatusResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	err = h.uc.UpdateStatus(ctx, adminID, id, req.GetStatus())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateMatchStatusResponse{
		Id:     req.GetId(),
		Status: req.GetStatus(),
	}, nil
}

func (h *MatchHandler) UpdateMatchScore(ctx context.Context, req *competitionv1.UpdateMatchScoreRequest) (*competitionv1.UpdateMatchScoreResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	participants := make([]dto.UpdateParticipantScoreRequest, len(req.GetParticipants()))
	for i, p := range req.GetParticipants() {
		pID, err := uuid.Parse(p.GetParticipantId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid participant id"))
		}

		participants[i] = dto.UpdateParticipantScoreRequest{
			ParticipantID: pID,
			Score:         p.GetScore(),
		}
	}

	dtoReq := dto.UpdateScoreRequest{
		Participants: participants,
	}

	err = h.uc.UpdateScore(ctx, adminID, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateMatchScoreResponse{
		Id: req.GetId(),
	}, nil
}

func (h *MatchHandler) DeleteMatch(ctx context.Context, req *competitionv1.DeleteMatchRequest) (*competitionv1.DeleteMatchResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	err = h.uc.Delete(ctx, adminID, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteMatchResponse{
		Id: req.GetId(),
	}, nil
}
