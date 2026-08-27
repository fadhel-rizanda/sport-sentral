package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	venuev1 "microservice-golang/gen/venue/v1"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type VenueHandler struct {
	client *client.VenueClient
}

func NewVenueHandler(client *client.VenueClient) *VenueHandler {
	return &VenueHandler{client: client}
}

func (h *VenueHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	// Venue Routes
	venues := router.Group("/venues", auth)
	venues.Get("/", h.ListVenues)
	venues.Get("/:id", h.GetVenue)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminVenues := router.Group("/admin/venues", adminMiddlewares...)
	adminVenues.Post("/", h.CreateVenue)
	adminVenues.Put("/:id", h.UpdateVenue)
	adminVenues.Delete("/:id", h.DeleteVenue)

	// Court Routes
	courts := router.Group("/courts", auth)
	courts.Get("/", h.ListCourts)
	courts.Get("/:id", h.GetCourt)
	courts.Get("/:id/availability", h.CheckAvailability)
	courts.Get("/:id/slots", h.ListCourtSlots)
	courts.Post("/:id/bookings", h.CreateBooking)

	adminCourts := router.Group("/admin/courts", adminMiddlewares...)
	adminCourts.Post("/", h.CreateCourt)
	adminCourts.Put("/:id", h.UpdateCourt)
	adminCourts.Delete("/:id", h.DeleteCourt)

	// Admin Court Slots Management
	adminCourts.Post("/:id/slots", h.CreateCourtSlots)
	adminCourts.Delete("/slots/:slotId", h.DeleteCourtSlot)

	// Booking Routes
	bookings := router.Group("/bookings", auth)
	bookings.Get("/", h.ListBookings)
	bookings.Get("/:id", h.GetBooking)
	bookings.Put("/:id/status", h.UpdateBookingStatus)
}

// ─── VENUE HANDLERS ──────────────────────────────────────────────────────────

// CreateVenue godoc
// @Summary      Create new sports venue (Admin)
// @Description  Register a new sports venue or building complex.
// @Tags         Venues
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateVenueRequest true "New Venue Data"
// @Success      201  {object}  response.Response{data=dto.VenueResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/venues [post]
// @Security     BearerAuth
func (h *VenueHandler) CreateVenue(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.CreateVenueRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &venuev1.CreateVenueRequest{
		Name:            reqBody.Name,
		Description:     reqBody.Description,
		StreetAddress:   reqBody.StreetAddress,
		AdminDivisionId: reqBody.AdminDivisionID,
		OwnerId:         userID, // Set current user as owner
		StatusId:        reqBody.StatusID,
	}

	if reqBody.Notes != "" {
		req.Notes = &reqBody.Notes
	}
	if reqBody.Latitude != 0 {
		req.Latitude = &reqBody.Latitude
	}
	if reqBody.Longitude != 0 {
		req.Longitude = &reqBody.Longitude
	}

	resp, err := h.client.Venue.CreateVenue(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToVenueResponseFromProto(resp.Venue))
}

// GetVenue godoc
// @Summary      Get sports venue details
// @Description  Retrieve venue profile and address data by ID.
// @Tags         Venues
// @Produce      json
// @Param        id   path      string  true  "Venue ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.VenueResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /venues/{id} [get]
// @Security     BearerAuth
func (h *VenueHandler) GetVenue(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Venue.GetVenue(c.Context(), &venuev1.GetVenueRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToVenueResponseFromProto(resp.Venue))
}

// ListVenues godoc
// @Summary      Get list of sports venues
// @Description  Retrieve list of all venues with filters for owner, status, or search query.
// @Tags         Venues
// @Produce      json
// @Param        owner_id  query string false "Filter Owner ID (UUID)"
// @Param        status_id query string false "Filter Status ID (UUID)"
// @Param        search    query string false "Search venue name"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /venues [get]
// @Security     BearerAuth
func (h *VenueHandler) ListVenues(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	ownerID := c.Query("owner_id", "")
	statusID := c.Query("status_id", "")
	search := c.Query("search", "")

	req := &venuev1.ListVenuesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if ownerID != "" {
		req.OwnerId = &ownerID
	}
	if statusID != "" {
		req.StatusId = &statusID
	}
	if search != "" {
		req.Search = &search
	}

	resp, err := h.client.Venue.ListVenues(c.Context(), req)
	if err != nil {
		return err
	}

	venues := make([]*dto.VenueResponse, len(resp.Venues))
	for i, venue := range resp.Venues {
		venues[i] = mapper.ToVenueResponseFromProto(venue)
	}

	return response.OK(c, fiber.Map{
		"venues":    venues,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

// UpdateVenue godoc
// @Summary      Update sports venue (Admin)
// @Description  Update venue name, address, or status.
// @Tags         Venues
// @Accept       json
// @Produce      json
// @Param        id      path string                true "Venue ID (UUID)"
// @Param        request body dto.UpdateVenueRequest true "Venue Update Data"
// @Success      200  {object}  response.Response{data=dto.VenueResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/venues/{id} [put]
// @Security     BearerAuth
func (h *VenueHandler) UpdateVenue(c *fiber.Ctx) error {
	id := c.Params("id")

	var reqBody dto.UpdateVenueRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &venuev1.UpdateVenueRequest{
		Id: id,
	}

	if reqBody.Name != nil {
		req.Name = reqBody.Name
	}
	if reqBody.Description != nil {
		req.Description = reqBody.Description
	}
	if reqBody.StreetAddress != nil {
		req.StreetAddress = reqBody.StreetAddress
	}
	if reqBody.Notes != nil {
		req.Notes = reqBody.Notes
	}
	if reqBody.Latitude != nil {
		req.Latitude = reqBody.Latitude
	}
	if reqBody.Longitude != nil {
		req.Longitude = reqBody.Longitude
	}
	if reqBody.StatusID != nil {
		req.StatusId = reqBody.StatusID
	}

	resp, err := h.client.Venue.UpdateVenue(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToVenueResponseFromProto(resp.Venue))
}

// DeleteVenue godoc
// @Summary      Delete sports venue (Admin)
// @Description  Delete venue from the system.
// @Tags         Venues
// @Produce      json
// @Param        id   path      string  true  "Venue ID (UUID)"
// @Success      204
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/venues/{id} [delete]
// @Security     BearerAuth
func (h *VenueHandler) DeleteVenue(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Venue.DeleteVenue(c.Context(), &venuev1.DeleteVenueRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ─── COURT HANDLERS ──────────────────────────────────────────────────────────

// CreateCourt godoc
// @Summary      Create new court (Admin)
// @Description  Register a new court (e.g. Court A, Futsal 1) under a venue.
// @Tags         Courts
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCourtRequest true "New Court Data"
// @Success      201  {object}  response.Response{data=dto.CourtResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/courts [post]
// @Security     BearerAuth
func (h *VenueHandler) CreateCourt(c *fiber.Ctx) error {
	var reqBody dto.CreateCourtRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &venuev1.CreateCourtRequest{
		VenueId:      reqBody.VenueID,
		Name:         reqBody.Name,
		Description:  reqBody.Description,
		SportId:      reqBody.SportID,
		PricePerHour: reqBody.PricePerHour,
		StatusId:     reqBody.StatusID,
	}

	if reqBody.ImageAttachment != "" {
		req.ImageAttachmentId = &reqBody.ImageAttachment
	}

	resp, err := h.client.Venue.CreateCourt(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToCourtResponseFromProto(resp.Court))
}

// GetCourt godoc
// @Summary      Get court details
// @Description  Retrieve court specifications and hourly rate.
// @Tags         Courts
// @Produce      json
// @Param        id   path      string  true  "Court ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.CourtResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /courts/{id} [get]
// @Security     BearerAuth
func (h *VenueHandler) GetCourt(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Venue.GetCourt(c.Context(), &venuev1.GetCourtRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCourtResponseFromProto(resp.Court))
}

// ListCourts godoc
// @Summary      Get list of courts
// @Description  Retrieve list of all courts with filters for venue, sport, or status.
// @Tags         Courts
// @Produce      json
// @Param        venue_id  query string false "Filter Venue ID (UUID)"
// @Param        sport_id  query string false "Filter Sport ID (UUID)"
// @Param        status_id query string false "Filter Status ID (UUID)"
// @Param        search    query string false "Search court name"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /courts [get]
// @Security     BearerAuth
func (h *VenueHandler) ListCourts(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	venueID := c.Query("venue_id", "")
	sportID := c.Query("sport_id", "")
	statusID := c.Query("status_id", "")
	search := c.Query("search", "")

	req := &venuev1.ListCourtsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if venueID != "" {
		req.VenueId = &venueID
	}
	if sportID != "" {
		req.SportId = &sportID
	}
	if statusID != "" {
		req.StatusId = &statusID
	}
	if search != "" {
		req.Search = &search
	}

	resp, err := h.client.Venue.ListCourts(c.Context(), req)
	if err != nil {
		return err
	}

	courts := make([]*dto.CourtResponse, len(resp.Courts))
	for i, court := range resp.Courts {
		courts[i] = mapper.ToCourtResponseFromProto(court)
	}

	return response.OK(c, fiber.Map{
		"courts":    courts,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

// UpdateCourt godoc
// @Summary      Update court (Admin)
// @Description  Update court name, description, hourly rate, or photo.
// @Tags         Courts
// @Accept       json
// @Produce      json
// @Param        id      path string                true "Court ID (UUID)"
// @Param        request body dto.UpdateCourtRequest true "Court Update Data"
// @Success      200  {object}  response.Response{data=dto.CourtResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/courts/{id} [put]
// @Security     BearerAuth
func (h *VenueHandler) UpdateCourt(c *fiber.Ctx) error {
	id := c.Params("id")

	var reqBody dto.UpdateCourtRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &venuev1.UpdateCourtRequest{
		Id: id,
	}

	if reqBody.Name != nil {
		req.Name = reqBody.Name
	}
	if reqBody.Description != nil {
		req.Description = reqBody.Description
	}
	if reqBody.SportID != nil {
		req.SportId = reqBody.SportID
	}
	if reqBody.PricePerHour != nil {
		req.PricePerHour = reqBody.PricePerHour
	}
	if reqBody.ImageAttachment != nil {
		req.ImageAttachmentId = reqBody.ImageAttachment
	}
	if reqBody.StatusID != nil {
		req.StatusId = reqBody.StatusID
	}

	resp, err := h.client.Venue.UpdateCourt(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCourtResponseFromProto(resp.Court))
}

// DeleteCourt godoc
// @Summary      Delete court (Admin)
// @Description  Delete court from a venue.
// @Tags         Courts
// @Produce      json
// @Param        id   path      string  true  "Court ID (UUID)"
// @Success      204
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/courts/{id} [delete]
// @Security     BearerAuth
func (h *VenueHandler) DeleteCourt(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Venue.DeleteCourt(c.Context(), &venuev1.DeleteCourtRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ─── SLOT HANDLERS ───────────────────────────────────────────────────────────

// CreateCourtSlots godoc
// @Summary      Create court schedule slots (Admin)
// @Description  Open operational time slots (e.g. 08:00 - 09:00) with rental pricing.
// @Tags         Courts
// @Accept       json
// @Produce      json
// @Param        id      path string                    true "Court ID (UUID)"
// @Param        request body dto.CreateCourtSlotsRequest true "New Slot Schedule Data"
// @Success      201  {object}  response.Response{data=[]dto.CourtSlotResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/courts/{id}/slots [post]
// @Security     BearerAuth
func (h *VenueHandler) CreateCourtSlots(c *fiber.Ctx) error {
	courtID := c.Params("id")

	var reqBody dto.CreateCourtSlotsRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	protoSlotsInput := make([]*venuev1.SlotInput, len(reqBody.Slots))
	for i, s := range reqBody.Slots {
		startTime, err := time.Parse(time.RFC3339, s.StartTime)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid start_time format (use RFC3339)")
		}
		endTime, err := time.Parse(time.RFC3339, s.EndTime)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid end_time format (use RFC3339)")
		}

		protoSlotsInput[i] = &venuev1.SlotInput{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
			Price:     s.Price,
		}
	}

	resp, err := h.client.Venue.CreateCourtSlots(c.Context(), &venuev1.CreateCourtSlotsRequest{
		CourtId:  courtID,
		Slots:    protoSlotsInput,
		StatusId: reqBody.StatusID,
	})
	if err != nil {
		return err
	}

	slots := make([]*dto.CourtSlotResponse, len(resp.Slots))
	for i, s := range resp.Slots {
		slots[i] = mapper.ToCourtSlotResponseFromProto(s)
	}

	return response.Created(c, slots)
}

// ListCourtSlots godoc
// @Summary      Get list of court schedule slots
// @Description  Retrieve court slot availability within a specified date range.
// @Tags         Courts
// @Produce      json
// @Param        id         path  string true  "Court ID (UUID)"
// @Param        status_id  query string false "Filter Status ID (UUID)"
// @Param        start_date query string false "Start Date (RFC3339)"
// @Param        end_date   query string false "End Date (RFC3339)"
// @Success      200  {object}  response.Response{data=[]dto.CourtSlotResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /courts/{id}/slots [get]
// @Security     BearerAuth
func (h *VenueHandler) ListCourtSlots(c *fiber.Ctx) error {
	courtID := c.Params("id")
	statusID := c.Query("status_id", "")
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	req := &venuev1.ListCourtSlotsRequest{
		CourtId: courtID,
	}

	if statusID != "" {
		req.StatusId = &statusID
	}

	if startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid start_date format (use RFC3339)")
		}
		req.StartDate = timestamppb.New(startDate)
	}

	if endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid end_date format (use RFC3339)")
		}
		req.EndDate = timestamppb.New(endDate)
	}

	resp, err := h.client.Venue.ListCourtSlots(c.Context(), req)
	if err != nil {
		return err
	}

	slots := make([]*dto.CourtSlotResponse, len(resp.Slots))
	for i, s := range resp.Slots {
		slots[i] = mapper.ToCourtSlotResponseFromProto(s)
	}

	return response.OK(c, slots)
}

// DeleteCourtSlot godoc
// @Summary      Delete court schedule slot (Admin)
// @Description  Delete a court rental time slot.
// @Tags         Courts
// @Produce      json
// @Param        slotId path string true "Slot ID (UUID)"
// @Success      204
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/courts/slots/{slotId} [delete]
// @Security     BearerAuth
func (h *VenueHandler) DeleteCourtSlot(c *fiber.Ctx) error {
	slotID := c.Params("slotId")

	_, err := h.client.Venue.DeleteCourtSlot(c.Context(), &venuev1.DeleteCourtSlotRequest{
		Id: slotID,
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ─── BOOKING HANDLERS ─────────────────────────────────────────────────────────

// CreateBooking godoc
// @Summary      Create court booking
// @Description  Book one or more available time slots for a specific court.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Param        id      path string                 true "Court ID (UUID)"
// @Param        request body dto.CreateBookingRequest true "Booking Reservation Data"
// @Success      201  {object}  response.Response{data=dto.BookingResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /courts/{id}/bookings [post]
// @Security     BearerAuth
func (h *VenueHandler) CreateBooking(c *fiber.Ctx) error {
	courtID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.CreateBookingRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	resp, err := h.client.Venue.CreateBooking(c.Context(), &venuev1.CreateBookingRequest{
		CourtId:  courtID,
		UserId:   userID,
		SlotIds:  reqBody.SlotIDs,
		StatusId: reqBody.StatusID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToBookingResponseFromProto(resp.Booking))
}

// GetBooking godoc
// @Summary      Get booking details
// @Description  Retrieve detailed booking data including reserved slot details.
// @Tags         Bookings
// @Produce      json
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.BookingResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /bookings/{id} [get]
// @Security     BearerAuth
func (h *VenueHandler) GetBooking(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Venue.GetBooking(c.Context(), &venuev1.GetBookingRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToBookingResponseFromProto(resp.Booking))
}

// ListBookings godoc
// @Summary      Get list of booking transactions
// @Description  Retrieve list of all court bookings with filters for user, court, or status.
// @Tags         Bookings
// @Produce      json
// @Param        court_id  query string false "Filter Court ID (UUID)"
// @Param        user_id   query string false "Filter User ID (UUID)"
// @Param        status_id query string false "Filter Status ID (UUID)"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /bookings [get]
// @Security     BearerAuth
func (h *VenueHandler) ListBookings(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	courtID := c.Query("court_id", "")
	userID := c.Query("user_id", "")
	statusID := c.Query("status_id", "")

	req := &venuev1.ListBookingsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if courtID != "" {
		req.CourtId = &courtID
	}
	if userID != "" {
		req.UserId = &userID
	}
	if statusID != "" {
		req.StatusId = &statusID
	}

	resp, err := h.client.Venue.ListBookings(c.Context(), req)
	if err != nil {
		return err
	}

	bookings := make([]*dto.BookingResponse, len(resp.Bookings))
	for i, b := range resp.Bookings {
		bookings[i] = mapper.ToBookingResponseFromProto(b)
	}

	return response.OK(c, fiber.Map{
		"bookings":  bookings,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

// UpdateBookingStatus godoc
// @Summary      Update booking status
// @Description  Update booking verification status or payment status (Paid/Cancelled).
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Booking ID (UUID)"
// @Param        request body dto.UpdateBookingStatusRequest true "New Booking Status Data"
// @Success      200  {object}  response.Response{data=dto.BookingResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /bookings/{id}/status [put]
// @Security     BearerAuth
func (h *VenueHandler) UpdateBookingStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var reqBody dto.UpdateBookingStatusRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &venuev1.UpdateBookingStatusRequest{
		Id:       id,
		StatusId: reqBody.StatusID,
	}

	if reqBody.PaymentStatus != nil {
		req.PaymentStatus = reqBody.PaymentStatus
	}

	resp, err := h.client.Venue.UpdateBookingStatus(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToBookingResponseFromProto(resp.Booking))
}

// CheckAvailability godoc
// @Summary      Check availability for multiple court slots
// @Description  Check if a set of court time slots are simultaneously available.
// @Tags         Courts
// @Produce      json
// @Param        id       path  string true "Court ID (UUID)"
// @Param        slot_ids query string true "Comma-separated slot IDs (slot1,slot2)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /courts/{id}/availability [get]
// @Security     BearerAuth
func (h *VenueHandler) CheckAvailability(c *fiber.Ctx) error {
	courtID := c.Params("id")
	slotIDs := c.Query("slot_ids", "")

	if slotIDs == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slot_ids comma-separated query parameter is required")
	}

	resp, err := h.client.Venue.ListCourtSlots(c.Context(), &venuev1.ListCourtSlotsRequest{
		CourtId: courtID,
	})
	if err != nil {
		return err
	}

	slotMap := make(map[string]string)
	for _, s := range resp.Slots {
		if s.Status != nil {
			slotMap[s.Id] = s.Status.Slug
		}
	}

	var requestedIDs []string
	var item string
	for _, char := range slotIDs {
		if char == ',' {
			if item != "" {
				requestedIDs = append(requestedIDs, item)
				item = ""
			}
		} else {
			item += string(char)
		}
	}
	if item != "" {
		requestedIDs = append(requestedIDs, item)
	}

	if len(requestedIDs) == 0 {
		return response.OK(c, fiber.Map{"is_available": false})
	}

	isAvailable := true
	for _, id := range requestedIDs {
		status, exists := slotMap[id]
		if !exists || status != "available" {
			isAvailable = false
			break
		}
	}

	return response.OK(c, fiber.Map{
		"is_available": isAvailable,
	})
}
