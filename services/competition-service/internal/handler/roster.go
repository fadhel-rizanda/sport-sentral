package handler

import (
	"context"

	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type RosterHandler struct {
	competitionv1.UnimplementedRosterServiceServer
	uc usecase.RosterUseCase
}

func NewRosterHandler(uc usecase.RosterUseCase) *RosterHandler {
	return &RosterHandler{uc: uc}
}

func (h *RosterHandler) RegisterGRPC(s *grpc.Server) {
	competitionv1.RegisterRosterServiceServer(s, h)
}

func (h *RosterHandler) CreateRoster(ctx context.Context, req *competitionv1.CreateRosterRequest) (*competitionv1.CreateRosterResponse, error) {
	branchID, err := uuid.Parse(req.GetAcademyBranchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	if req.CompetitionId == nil || *req.CompetitionId == "" {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("competition id is required"))
	}
	competitionID, err := uuid.Parse(*req.CompetitionId)
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
	}

	tagID, err := uuid.Parse(req.GetTagId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tag id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	dtoReq := dto.CreateRosterRequest{
		AcademyBranchID: branchID,
		CompetitionID:   competitionID,
		Name:            req.GetName(),
		TagID:           tagID,
		StatusID:        statusID,
		MaxSize:         int(req.GetMaxSize()),
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.CreateRosterResponse{
		Roster: mapper.ToProtoRoster(res),
	}, nil
}

func (h *RosterHandler) GetRoster(ctx context.Context, req *competitionv1.GetRosterRequest) (*competitionv1.GetRosterResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.GetRosterResponse{
		Roster: mapper.ToProtoRoster(res),
	}, nil
}

func (h *RosterHandler) ListRosters(ctx context.Context, req *competitionv1.ListRostersRequest) (*competitionv1.ListRostersResponse, error) {
	var competitionID *uuid.UUID
	if req.CompetitionId != nil && *req.CompetitionId != "" {
		id, err := uuid.Parse(*req.CompetitionId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
		}
		competitionID = &id
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

	var tagID *uuid.UUID
	if req.TagId != nil && *req.TagId != "" {
		id, err := uuid.Parse(*req.TagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tag id"))
		}
		tagID = &id
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}

	dtoReq := dto.ListRostersRequest{
		CompetitionID: competitionID,
		BranchID:      branchID,
		StatusID:      statusID,
		TagID:         tagID,
		Search:        req.Search,
		Page:          page,
		PageSize:      pageSize,
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	rosters := make([]*competitionv1.Roster, len(res.Rosters))
	for i, r := range res.Rosters {
		rosters[i] = mapper.ToProtoRoster(r)
	}

	return &competitionv1.ListRostersResponse{
		Rosters:  rosters,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *RosterHandler) UpdateRoster(ctx context.Context, req *competitionv1.UpdateRosterRequest) (*competitionv1.UpdateRosterResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	var competitionID *uuid.UUID
	if req.CompetitionId != nil && *req.CompetitionId != "" {
		cID, err := uuid.Parse(*req.CompetitionId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid competition id"))
		}
		competitionID = &cID
	}

	var tagID *uuid.UUID
	if req.TagId != nil && *req.TagId != "" {
		tID, err := uuid.Parse(*req.TagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tag id"))
		}
		tagID = &tID
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		sID, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &sID
	}

	var maxSize *int
	if req.MaxSize != nil {
		val := int(*req.MaxSize)
		maxSize = &val
	}

	dtoReq := dto.UpdateRosterRequest{
		CompetitionID: competitionID,
		Name:          req.Name,
		TagID:         tagID,
		StatusID:      statusID,
		MaxSize:       maxSize,
	}

	res, err := h.uc.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.UpdateRosterResponse{
		Roster: mapper.ToProtoRoster(res),
	}, nil
}

func (h *RosterHandler) DeleteRoster(ctx context.Context, req *competitionv1.DeleteRosterRequest) (*competitionv1.DeleteRosterResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.DeleteRosterRequest{
		ID:          id,
		DeletedByID: deletedByID,
	}

	err = h.uc.Delete(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteRosterResponse{
		Id: req.GetId(),
	}, nil
}

func (h *RosterHandler) GetRosterMembers(ctx context.Context, req *competitionv1.GetRosterMembersRequest) (*competitionv1.GetRosterMembersResponse, error) {
	rosterID, err := uuid.Parse(req.GetRosterId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	res, err := h.uc.GetMembers(ctx, rosterID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	members := make([]*competitionv1.RosterMember, len(res))
	for i, m := range res {
		members[i] = mapper.ToProtoRosterMember(m)
	}

	return &competitionv1.GetRosterMembersResponse{
		Members: members,
	}, nil
}

func (h *RosterHandler) GetRosterMember(ctx context.Context, req *competitionv1.GetRosterMemberRequest) (*competitionv1.GetRosterMemberResponse, error) {
	rosterID, err := uuid.Parse(req.GetRosterId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	memberID, err := uuid.Parse(req.GetMemberId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid member id"))
	}

	res, err := h.uc.GetMemberByID(ctx, rosterID, memberID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.GetRosterMemberResponse{
		Member: mapper.ToProtoRosterMember(*res),
	}, nil
}

func (h *RosterHandler) AddRosterMember(ctx context.Context, req *competitionv1.AddRosterMemberRequest) (*competitionv1.AddRosterMemberResponse, error) {
	rosterID, err := uuid.Parse(req.GetRosterId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	positionID, err := uuid.Parse(req.GetPositionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid position id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	addedByID, err := uuid.Parse(req.GetAddedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid added by id"))
	}

	var jerseyNumber *int
	if req.JerseyNumber != nil {
		val := int(*req.JerseyNumber)
		jerseyNumber = &val
	}

	dtoReq := dto.AddRosterMemberRequest{
		RosterID:     rosterID,
		AthleteID:    athleteID,
		JerseyNumber: jerseyNumber,
		PositionID:   positionID,
		StatusID:     statusID,
		AddedByID:    addedByID,
	}

	res, err := h.uc.AddMember(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.AddRosterMemberResponse{
		Member: mapper.ToProtoRosterMember(*res),
	}, nil
}

func (h *RosterHandler) RemoveRosterMember(ctx context.Context, req *competitionv1.RemoveRosterMemberRequest) (*competitionv1.RemoveRosterMemberResponse, error) {
	rosterID, err := uuid.Parse(req.GetRosterId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	memberID, err := uuid.Parse(req.GetMemberId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid member id"))
	}

	removedByID, err := uuid.Parse(req.GetRemovedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid removed by id"))
	}

	dtoReq := dto.RemoveRosterMemberRequest{
		RemovedByID:   removedByID,
		RemovalReason: req.GetRemovalReason(),
	}

	err = h.uc.RemoveMember(ctx, rosterID, memberID, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.RemoveRosterMemberResponse{
		MemberId: req.GetMemberId(),
	}, nil
}

func (h *RosterHandler) DeleteRosterMember(ctx context.Context, req *competitionv1.DeleteRosterMemberRequest) (*competitionv1.DeleteRosterMemberResponse, error) {
	rosterID, err := uuid.Parse(req.GetRosterId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid roster id"))
	}

	memberID, err := uuid.Parse(req.GetMemberId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid member id"))
	}

	err = h.uc.DeleteMember(ctx, rosterID, memberID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &competitionv1.DeleteRosterMemberResponse{
		MemberId: req.GetMemberId(),
	}, nil
}
