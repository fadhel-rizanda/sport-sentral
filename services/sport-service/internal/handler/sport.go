package handler

import (
	"context"
	"fmt"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/dto"
	"microservice-golang/services/sport-service/internal/mapper"
	"microservice-golang/services/sport-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SportHandler struct {
	sportv1.UnimplementedSportServiceServer
	uc usecase.SportUseCase
}

func NewSportHandler(uc usecase.SportUseCase) *SportHandler {
	return &SportHandler{uc: uc}
}

func (h *SportHandler) RegisterGRPC(s *grpc.Server) {
	sportv1.RegisterSportServiceServer(s, h)
}

func (h *SportHandler) GetSport(ctx context.Context, req *sportv1.GetSportRequest) (*sportv1.Sport, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoSport(res), nil
}

func (h *SportHandler) ListSports(ctx context.Context, req *sportv1.ListSportsRequest) (*sportv1.ListSportsResponse, error) {
	var tierTagID *uuid.UUID
	if req.TierTagId != nil && *req.TierTagId != "" {
		id, err := uuid.Parse(*req.TierTagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier tag id"))
		}
		tierTagID = &id
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		id, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &id
	}

	dtoReq := dto.ListSportsRequest{
		TierTagID: tierTagID,
		StatusID:  statusID,
		Search:    req.Search,
		Page:      int(req.GetPage()),
		PageSize:  int(req.GetPageSize()),
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	sports := make([]*sportv1.Sport, len(res.Sports))
	for i, s := range res.Sports {
		sports[i] = mapper.ToProtoSport(s)
	}

	return &sportv1.ListSportsResponse{
		Sports:   sports,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *SportHandler) CreateSport(ctx context.Context, req *sportv1.CreateSportRequest) (*sportv1.Sport, error) {
	tierTagID, err := uuid.Parse(req.GetTierTagId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier tag id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	var iconAttachmentID *uuid.UUID
	if req.IconAttachmentId != "" {
		id, err := uuid.Parse(req.IconAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid icon attachment id"))
		}
		iconAttachmentID = &id
	}

	dtoReq := dto.CreateSportRequest{
		Name:             req.GetName(),
		Slug:             req.GetSlug(),
		Description:      req.GetDescription(),
		IconAttachmentID: iconAttachmentID,
		TierTagID:        tierTagID,
		StatusID:         statusID,
		CreatedByID:      createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoSport(res), nil
}

func (h *SportHandler) UpdateSport(ctx context.Context, req *sportv1.UpdateSportRequest) (*sportv1.Sport, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var iconAttachmentID *uuid.UUID
	if req.IconAttachmentId != nil && *req.IconAttachmentId != "" {
		iconID, err := uuid.Parse(*req.IconAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid icon attachment id"))
		}
		iconAttachmentID = &iconID
	}

	var tierTagID *uuid.UUID
	if req.TierTagId != nil && *req.TierTagId != "" {
		tagID, err := uuid.Parse(*req.TierTagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid tier tag id"))
		}
		tierTagID = &tagID
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		statID, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &statID
	}

	dtoReq := dto.UpdateSportRequest{
		ID:               id,
		Name:             name,
		Description:      description,
		IconAttachmentID: iconAttachmentID,
		TierTagID:        tierTagID,
		StatusID:         statusID,
		UpdatedByID:      updatedByID,
	}

	res, err := h.uc.Update(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoSport(res), nil
}

func (h *SportHandler) DeleteSport(ctx context.Context, req *sportv1.DeleteSportRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	if err := h.uc.Delete(ctx, id); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *SportHandler) UpdateSportConfig(ctx context.Context, req *sportv1.UpdateSportConfigRequest) (*sportv1.SportConfig, error) {
	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	participantTypeTagID, err := uuid.Parse(req.GetParticipantTypeTagId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid participant type tag id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	statsInput := make([]dto.SportStatConfigInput, len(req.GetStats()))
	for i, s := range req.GetStats() {
		tagID, err := uuid.Parse(s.GetStatTypeTagId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument(fmt.Sprintf("invalid stat tag id at index %d", i)))
		}
		statsInput[i] = dto.SportStatConfigInput{
			StatTypeTagID:     tagID,
			AggregationMethod: s.GetAggregationMethod(),
		}
	}

	var rulesURL *string
	if req.RulesUrl != nil {
		rulesURL = req.RulesUrl
	}

	var desc string
	if req.Description != nil {
		desc = *req.Description
	}

	dtoReq := dto.UpdateSportConfigRequest{
		SportID:              sportID,
		Stats:                statsInput,
		ParticipantTypeTagID: participantTypeTagID,
		MinRosterSize:        req.GetMinRosterSize(),
		MaxRosterSize:        req.GetMaxRosterSize(),
		TypicalRosterSize:    req.GetTypicalRosterSize(),
		RulesURL:             rulesURL,
		Description:          desc,
		UpdatedByID:          updatedByID,
	}

	res, err := h.uc.UpdateConfig(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoSportConfig(res), nil
}

func (h *SportHandler) GetSportConfig(ctx context.Context, req *sportv1.GetSportConfigRequest) (*sportv1.SportConfig, error) {
	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	res, err := h.uc.GetConfig(ctx, sportID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoSportConfig(res), nil
}
