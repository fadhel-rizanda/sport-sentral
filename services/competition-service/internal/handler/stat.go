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

type StatHandler struct {
	competitionv1.UnimplementedStatServiceServer
	uc usecase.StatUseCase
}

func NewStatHandler(uc usecase.StatUseCase) *StatHandler {
	return &StatHandler{uc: uc}
}

func (h *StatHandler) RegisterGRPC(s *grpc.Server) {
	competitionv1.RegisterStatServiceServer(s, h)
}

func (h *StatHandler) RecordMatchStat(ctx context.Context, req *competitionv1.RecordMatchStatRequest) (*competitionv1.RecordMatchStatResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	matchID, err := uuid.Parse(req.GetMatchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	statTypeID, err := uuid.Parse(req.GetStatTypeId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid stat type id"))
	}

	dtoReq := dto.RecordStatRequest{
		AthleteID:  athleteID,
		StatTypeID: statTypeID,
		Value:      req.GetValue(),
	}

	res, err := h.uc.RecordMatchStat(ctx, adminID, matchID, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.RecordMatchStatResponse{
		Stat: mapper.ToProtoMatchStat(res),
	}, nil
}

func (h *StatHandler) UpdateMatchStat(ctx context.Context, req *competitionv1.UpdateMatchStatRequest) (*competitionv1.UpdateMatchStatResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid stat id"))
	}

	err = h.uc.UpdateMatchStat(ctx, adminID, id, req.GetValue())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateMatchStatResponse{
		Id:    req.GetId(),
		Value: req.GetValue(),
	}, nil
}

func (h *StatHandler) DeleteMatchStat(ctx context.Context, req *competitionv1.DeleteMatchStatRequest) (*competitionv1.DeleteMatchStatResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid stat id"))
	}

	err = h.uc.DeleteMatchStat(ctx, adminID, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteMatchStatResponse{
		Id: req.GetId(),
	}, nil
}

func (h *StatHandler) GetMatchStats(ctx context.Context, req *competitionv1.GetMatchStatsRequest) (*competitionv1.GetMatchStatsResponse, error) {
	matchID, err := uuid.Parse(req.GetMatchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid match id"))
	}

	res, err := h.uc.GetMatchStats(ctx, matchID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	stats := make([]*competitionv1.MatchStat, len(res))
	for i, s := range res {
		stats[i] = mapper.ToProtoMatchStat(s)
	}

	return &competitionv1.GetMatchStatsResponse{
		Stats: stats,
	}, nil
}

func (h *StatHandler) GetAthleteAggregate(ctx context.Context, req *competitionv1.GetAthleteAggregateRequest) (*competitionv1.GetAthleteAggregateResponse, error) {
	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	competitionID, err := uuid.Parse(req.GetCompetitionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	res, err := h.uc.GetAthleteAggregate(ctx, athleteID, competitionID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	aggregates := make([]*competitionv1.AthleteStatsAggregate, len(res))
	for i, a := range res {
		aggregates[i] = mapper.ToProtoAthleteStatsAggregate(a)
	}

	return &competitionv1.GetAthleteAggregateResponse{
		Aggregates: aggregates,
	}, nil
}

func (h *StatHandler) RecalculateAggregate(ctx context.Context, req *competitionv1.RecalculateAggregateRequest) (*competitionv1.RecalculateAggregateResponse, error) {
	adminID, err := extractAdminID(ctx)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	competitionID, err := uuid.Parse(req.GetCompetitionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	err = h.uc.RecalculateAggregate(ctx, adminID, athleteID, competitionID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.RecalculateAggregateResponse{
		AthleteId:     req.GetAthleteId(),
		CompetitionId: req.GetCompetitionId(),
	}, nil
}
