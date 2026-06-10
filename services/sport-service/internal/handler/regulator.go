package handler

import (
	"context"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/dto"
	"microservice-golang/services/sport-service/internal/mapper"
	"microservice-golang/services/sport-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RegulatorHandler struct {
	sportv1.UnimplementedRegulatorServiceServer
	uc usecase.RegulatorUseCase
}

func NewRegulatorHandler(uc usecase.RegulatorUseCase) *RegulatorHandler {
	return &RegulatorHandler{uc: uc}
}

func (h *RegulatorHandler) RegisterGRPC(s *grpc.Server) {
	sportv1.RegisterRegulatorServiceServer(s, h)
}

func (h *RegulatorHandler) CreateRegulator(ctx context.Context, req *sportv1.CreateRegulatorRequest) (*sportv1.Regulator, error) {
	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	adminDivisionID, err := uuid.Parse(req.GetAdministrativeDivisionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid administrative division id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	var logoAttachmentID *uuid.UUID
	if req.LogoAttachmentId != "" {
		id, err := uuid.Parse(req.LogoAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid logo attachment id"))
		}
		logoAttachmentID = &id
	}

	var phone *string
	if req.PhoneNumber != nil {
		phone = req.PhoneNumber
	}

	var website *string
	if req.WebsiteUrl != nil {
		website = req.WebsiteUrl
	}

	var addressNotes *string
	if req.AddressNotes != nil {
		addressNotes = req.AddressNotes
	}

	var lat *float64
	if req.Latitude != nil {
		lat = req.Latitude
	}

	var lng *float64
	if req.Longitude != nil {
		lng = req.Longitude
	}

	dtoReq := dto.CreateRegulatorRequest{
		OrganizationName: req.GetOrganizationName(),
		Code:             req.GetCode(),
		LogoAttachmentID: logoAttachmentID,
		ContactEmail:     req.GetContactEmail(),
		PhoneNumber:      phone,
		WebsiteURL:       website,
		StatusID:         statusID,
		StreetAddress:    req.GetStreetAddress(),
		Notes:            addressNotes,
		Latitude:         lat,
		Longitude:        lng,
		AdminDivisionID:  adminDivisionID,
		CreatedByID:      createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoRegulator(res), nil
}

func (h *RegulatorHandler) GetRegulator(ctx context.Context, req *sportv1.GetRegulatorRequest) (*sportv1.Regulator, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid regulator id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoRegulator(res), nil
}

func (h *RegulatorHandler) AddRegulatorStaff(ctx context.Context, req *sportv1.AddStaffRequest) (*sportv1.RegulatorStaff, error) {
	regulatorID, err := uuid.Parse(req.GetRegulatorId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid regulator id"))
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	roleTagID, err := uuid.Parse(req.GetRoleTagId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid role tag id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	dtoReq := dto.AddStaffRequest{
		RegulatorID: regulatorID,
		UserID:      userID,
		RoleTagID:   roleTagID,
		CreatedByID: createdByID,
	}

	res, err := h.uc.AddStaff(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return mapper.ToProtoRegulatorStaff(res), nil
}

func (h *RegulatorHandler) RemoveRegulatorStaff(ctx context.Context, req *sportv1.RemoveStaffRequest) (*emptypb.Empty, error) {
	staffID, err := uuid.Parse(req.GetStaffId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid staff id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.RemoveStaffRequest{
		StaffID:     staffID,
		DeletedByID: deletedByID,
	}

	if err := h.uc.RemoveStaff(ctx, dtoReq); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *RegulatorHandler) AssignSportToRegulator(ctx context.Context, req *sportv1.AssignSportToRegulatorRequest) (*emptypb.Empty, error) {
	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	regulatorID, err := uuid.Parse(req.GetRegulatorId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid regulator id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	dtoReq := dto.AssignSportToRegulatorRequest{
		SportID:                     sportID,
		RegulatorID:                 regulatorID,
		RequiresApprovalForOfficial: req.GetRequiresApprovalForOfficial(),
		RequiresApprovalForRegional: req.GetRequiresApprovalForRegional(),
		UpdatedByID:                 updatedByID,
	}

	if err := h.uc.AssignSport(ctx, dtoReq); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}
