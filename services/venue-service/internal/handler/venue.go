package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	courtv1 "microservice-golang/gen/venue/v1"
	"microservice-golang/services/venue-service/internal/dto"
	"microservice-golang/services/venue-service/internal/mapper"
	"microservice-golang/services/venue-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type VenueHandler struct {
	courtv1.UnimplementedVenueServiceServer
	venueUC   usecase.VenueUseCase
	courtUC   usecase.CourtUseCase
	bookingUC usecase.BookingUseCase
	slotUC    usecase.CourtSlotUseCase
}

func NewVenueHandler(
	venueUC usecase.VenueUseCase,
	courtUC usecase.CourtUseCase,
	bookingUC usecase.BookingUseCase,
	slotUC usecase.CourtSlotUseCase,
) *VenueHandler {
	return &VenueHandler{
		venueUC:   venueUC,
		courtUC:   courtUC,
		bookingUC: bookingUC,
		slotUC:    slotUC,
	}
}

func (h *VenueHandler) RegisterGRPC(s *grpc.Server) {
	courtv1.RegisterVenueServiceServer(s, h)
}

// Venue Management

func (h *VenueHandler) CreateVenue(ctx context.Context, req *courtv1.CreateVenueRequest) (*courtv1.CreateVenueResponse, error) {
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid owner id"))
	}

	adminDivisionID, err := uuid.Parse(req.GetAdminDivisionId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid admin division id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	var notes *string
	if req.Notes != nil {
		notes = req.Notes
	}

	var lat *float64
	if req.Latitude != nil {
		lat = req.Latitude
	}

	var lng *float64
	if req.Longitude != nil {
		lng = req.Longitude
	}

	dtoReq := dto.CreateVenueRequest{
		Name:            req.GetName(),
		Description:     req.GetDescription(),
		StreetAddress:   req.GetStreetAddress(),
		Notes:           notes,
		Latitude:        lat,
		Longitude:       lng,
		AdminDivisionID: adminDivisionID,
		OwnerID:         ownerID,
		StatusID:        statusID,
	}

	res, err := h.venueUC.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.CreateVenueResponse{
		Venue: mapper.ToProtoVenueFromDTO(res),
	}, nil
}

func (h *VenueHandler) GetVenue(ctx context.Context, req *courtv1.GetVenueRequest) (*courtv1.GetVenueResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid venue id"))
	}

	res, err := h.venueUC.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.GetVenueResponse{
		Venue: mapper.ToProtoVenueFromDTO(res),
	}, nil
}

func (h *VenueHandler) ListVenues(ctx context.Context, req *courtv1.ListVenuesRequest) (*courtv1.ListVenuesResponse, error) {
	var ownerID *uuid.UUID
	if req.OwnerId != nil && *req.OwnerId != "" {
		id, err := uuid.Parse(*req.OwnerId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid owner id"))
		}
		ownerID = &id
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

	dtoReq := dto.ListVenuesRequest{
		OwnerID:  ownerID,
		StatusID: statusID,
		Search:   req.Search,
		Page:     page,
		PageSize: pageSize,
	}

	res, err := h.venueUC.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	venues := make([]*courtv1.Venue, len(res.Venues))
	for i, v := range res.Venues {
		venues[i] = mapper.ToProtoVenueFromDTO(v)
	}

	return &courtv1.ListVenuesResponse{
		Venues:   venues,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *VenueHandler) UpdateVenue(ctx context.Context, req *courtv1.UpdateVenueRequest) (*courtv1.UpdateVenueResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid venue id"))
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var streetAddress *string
	if req.StreetAddress != nil {
		streetAddress = req.StreetAddress
	}

	var notes *string
	if req.Notes != nil {
		notes = req.Notes
	}

	var lat *float64
	if req.Latitude != nil {
		lat = req.Latitude
	}

	var lng *float64
	if req.Longitude != nil {
		lng = req.Longitude
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		parsed, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &parsed
	}

	dtoReq := dto.UpdateVenueRequest{
		Name:          name,
		Description:   description,
		StreetAddress: streetAddress,
		Notes:         notes,
		Latitude:      lat,
		Longitude:     lng,
		StatusID:      statusID,
	}

	res, err := h.venueUC.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.UpdateVenueResponse{
		Venue: mapper.ToProtoVenueFromDTO(res),
	}, nil
}

func (h *VenueHandler) DeleteVenue(ctx context.Context, req *courtv1.DeleteVenueRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid venue id"))
	}

	err = h.venueUC.Delete(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

// Court Management

func (h *VenueHandler) CreateCourt(ctx context.Context, req *courtv1.CreateCourtRequest) (*courtv1.CreateCourtResponse, error) {
	venueID, err := uuid.Parse(req.GetVenueId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid venue id"))
	}

	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	var imageAttachment *uuid.UUID
	if req.ImageAttachmentId != nil && *req.ImageAttachmentId != "" {
		id, err := uuid.Parse(*req.ImageAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid image attachment id"))
		}
		imageAttachment = &id
	}

	dtoReq := dto.CreateCourtRequest{
		VenueID:         venueID,
		Name:            req.GetName(),
		Description:     req.GetDescription(),
		SportID:         sportID,
		PricePerHour:    req.GetPricePerHour(),
		ImageAttachment: imageAttachment,
		StatusID:        statusID,
	}

	res, err := h.courtUC.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.CreateCourtResponse{
		Court: mapper.ToProtoCourtFromDTO(res),
	}, nil
}

func (h *VenueHandler) GetCourt(ctx context.Context, req *courtv1.GetCourtRequest) (*courtv1.GetCourtResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	res, err := h.courtUC.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.GetCourtResponse{
		Court: mapper.ToProtoCourtFromDTO(res),
	}, nil
}

func (h *VenueHandler) ListCourts(ctx context.Context, req *courtv1.ListCourtsRequest) (*courtv1.ListCourtsResponse, error) {
	var venueID *uuid.UUID
	if req.VenueId != nil && *req.VenueId != "" {
		id, err := uuid.Parse(*req.VenueId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid venue id"))
		}
		venueID = &id
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

	dtoReq := dto.ListCourtsRequest{
		VenueID:  venueID,
		SportID:  sportID,
		StatusID: statusID,
		Search:   req.Search,
		Page:     page,
		PageSize: pageSize,
	}

	res, err := h.courtUC.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	courts := make([]*courtv1.Court, len(res.Courts))
	for i, c := range res.Courts {
		courts[i] = mapper.ToProtoCourtFromDTO(c)
	}

	return &courtv1.ListCourtsResponse{
		Courts:   courts,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *VenueHandler) UpdateCourt(ctx context.Context, req *courtv1.UpdateCourtRequest) (*courtv1.UpdateCourtResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		parsed, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport id"))
		}
		sportID = &parsed
	}

	var pricePerHour *int64
	if req.PricePerHour != nil {
		pricePerHour = req.PricePerHour
	}

	var imageAttachment *uuid.UUID
	if req.ImageAttachmentId != nil && *req.ImageAttachmentId != "" {
		parsed, err := uuid.Parse(*req.ImageAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid image attachment id"))
		}
		imageAttachment = &parsed
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		parsed, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &parsed
	}

	dtoReq := dto.UpdateCourtRequest{
		Name:            name,
		Description:     description,
		SportID:         sportID,
		PricePerHour:    pricePerHour,
		ImageAttachment: imageAttachment,
		StatusID:        statusID,
	}

	res, err := h.courtUC.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.UpdateCourtResponse{
		Court: mapper.ToProtoCourtFromDTO(res),
	}, nil
}

func (h *VenueHandler) DeleteCourt(ctx context.Context, req *courtv1.DeleteCourtRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	err = h.courtUC.Delete(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

// Slot Management

func (h *VenueHandler) CreateCourtSlots(ctx context.Context, req *courtv1.CreateCourtSlotsRequest) (*courtv1.CreateCourtSlotsResponse, error) {
	courtID, err := uuid.Parse(req.GetCourtId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	slotsInput := make([]dto.SlotInput, len(req.Slots))
	for i, s := range req.Slots {
		if s.StartTime == nil || s.EndTime == nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("start time and end time are required for all slots"))
		}
		var price *int64
		if s.Price != nil {
			price = s.Price
		}
		slotsInput[i] = dto.SlotInput{
			StartTime: s.StartTime.AsTime(),
			EndTime:   s.EndTime.AsTime(),
			Price:     price,
		}
	}

	dtoReq := dto.CreateCourtSlotsRequest{
		CourtID:  courtID,
		Slots:    slotsInput,
		StatusID: statusID,
	}

	res, err := h.slotUC.CreateBulk(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	protoSlots := make([]*courtv1.CourtSlot, len(res))
	for i, s := range res {
		protoSlots[i] = mapper.ToProtoCourtSlotFromDTO(s)
	}

	return &courtv1.CreateCourtSlotsResponse{
		Slots: protoSlots,
	}, nil
}

func (h *VenueHandler) ListCourtSlots(ctx context.Context, req *courtv1.ListCourtSlotsRequest) (*courtv1.ListCourtSlotsResponse, error) {
	courtID, err := uuid.Parse(req.GetCourtId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		id, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &id
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

	dtoReq := dto.ListCourtSlotsRequest{
		CourtID:   courtID,
		StatusID:  statusID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	res, err := h.slotUC.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	protoSlots := make([]*courtv1.CourtSlot, len(res))
	for i, s := range res {
		protoSlots[i] = mapper.ToProtoCourtSlotFromDTO(s)
	}

	return &courtv1.ListCourtSlotsResponse{
		Slots: protoSlots,
	}, nil
}

func (h *VenueHandler) DeleteCourtSlot(ctx context.Context, req *courtv1.DeleteCourtSlotRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid slot id"))
	}

	err = h.slotUC.Delete(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

// Booking Management

func (h *VenueHandler) CreateBooking(ctx context.Context, req *courtv1.CreateBookingRequest) (*courtv1.CreateBookingResponse, error) {
	courtID, err := uuid.Parse(req.GetCourtId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	slotIDs := make([]uuid.UUID, len(req.GetSlotIds()))
	for i, sIDStr := range req.GetSlotIds() {
		id, err := uuid.Parse(sIDStr)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument(interface{}("invalid slot id: " + sIDStr).(string)))
		}
		slotIDs[i] = id
	}

	dtoReq := dto.CreateBookingRequest{
		CourtID:  courtID,
		UserID:   userID,
		SlotIDs:  slotIDs,
		StatusID: statusID,
	}

	res, err := h.bookingUC.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.CreateBookingResponse{
		Booking: mapper.ToProtoBookingFromDTO(res),
	}, nil
}

func (h *VenueHandler) GetBooking(ctx context.Context, req *courtv1.GetBookingRequest) (*courtv1.GetBookingResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid booking id"))
	}

	res, err := h.bookingUC.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.GetBookingResponse{
		Booking: mapper.ToProtoBookingFromDTO(res),
	}, nil
}

func (h *VenueHandler) ListBookings(ctx context.Context, req *courtv1.ListBookingsRequest) (*courtv1.ListBookingsResponse, error) {
	var courtID *uuid.UUID
	if req.CourtId != nil && *req.CourtId != "" {
		id, err := uuid.Parse(*req.CourtId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid court id"))
		}
		courtID = &id
	}

	var userID *uuid.UUID
	if req.UserId != nil && *req.UserId != "" {
		id, err := uuid.Parse(*req.UserId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user id"))
		}
		userID = &id
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

	dtoReq := dto.ListBookingsRequest{
		CourtID:  courtID,
		UserID:   userID,
		StatusID: statusID,
		Page:     page,
		PageSize: pageSize,
	}

	res, err := h.bookingUC.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	bookings := make([]*courtv1.Booking, len(res.Bookings))
	for i, b := range res.Bookings {
		bookings[i] = mapper.ToProtoBookingFromDTO(b)
	}

	return &courtv1.ListBookingsResponse{
		Bookings: bookings,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *VenueHandler) UpdateBookingStatus(ctx context.Context, req *courtv1.UpdateBookingStatusRequest) (*courtv1.UpdateBookingStatusResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid booking id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	var paymentStatus *string
	if req.PaymentStatus != nil {
		paymentStatus = req.PaymentStatus
	}

	dtoReq := dto.UpdateBookingStatusRequest{
		StatusID:      statusID,
		PaymentStatus: paymentStatus,
	}

	res, err := h.bookingUC.UpdateStatus(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &courtv1.UpdateBookingStatusResponse{
		Booking: mapper.ToProtoBookingFromDTO(res),
	}, nil
}
