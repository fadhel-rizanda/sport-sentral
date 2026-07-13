package mapper

import (
	"time"

	commonv1 "microservice-golang/gen/common/v1"
	venuev1 "microservice-golang/gen/venue/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToSportSimpleResponseFromCommon(s *commonv1.SportSimple) *dto.SportSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.SportSimpleResponse{
		ID:   s.Id,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToVenueResponseFromProto(v *venuev1.Venue) *dto.VenueResponse {
	if v == nil {
		return nil
	}
	var notesVal string
	if v.Notes != nil {
		notesVal = *v.Notes
	}
	var latVal float64
	if v.Latitude != nil {
		latVal = *v.Latitude
	}
	var lngVal float64
	if v.Longitude != nil {
		lngVal = *v.Longitude
	}
	ownerSimple := ToUserSimpleResponse(v.Owner)
	statusSimple := ToStatusSimpleResponse(v.Status)
	return &dto.VenueResponse{
		ID:            v.Id,
		Name:          v.Name,
		Description:   v.Description,
		StreetAddress: v.StreetAddress,
		Notes:         notesVal,
		Latitude:      latVal,
		Longitude:     lngVal,
		OwnerID:       v.OwnerId,
		Owner:         &ownerSimple,
		Status:        &statusSimple,
		StatusID:      v.StatusId,
		CreatedAt:     v.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:     v.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToCourtResponseFromProto(c *venuev1.Court) *dto.CourtResponse {
	if c == nil {
		return nil
	}
	var imageStr string
	if c.ImageAttachmentId != nil {
		imageStr = *c.ImageAttachmentId
	}
	statusSimple := ToStatusSimpleResponse(c.Status)
	return &dto.CourtResponse{
		ID:                c.Id,
		VenueID:           c.VenueId,
		Venue:             ToVenueResponseFromProto(c.Venue),
		Name:              c.Name,
		Description:       c.Description,
		SportID:           c.SportId,
		Sport:             ToSportSimpleResponseFromCommon(c.Sport),
		PricePerHour:      c.PricePerHour,
		ImageAttachmentID: imageStr,
		Status:            &statusSimple,
		StatusID:          c.StatusId,
		CreatedAt:         c.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:         c.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToCourtSlotResponseFromProto(s *venuev1.CourtSlot) *dto.CourtSlotResponse {
	if s == nil {
		return nil
	}
	var bookingStr string
	if s.BookingId != nil {
		bookingStr = *s.BookingId
	}
	statusSimple := ToStatusSimpleResponse(s.Status)
	return &dto.CourtSlotResponse{
		ID:        s.Id,
		CourtID:   s.CourtId,
		StartTime: s.StartTime.AsTime().UTC().Format(time.RFC3339),
		EndTime:   s.EndTime.AsTime().UTC().Format(time.RFC3339),
		Price:     s.Price,
		StatusID:  s.StatusId,
		Status:    &statusSimple,
		BookingID: bookingStr,
		CreatedAt: s.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToBookingResponseFromProto(b *venuev1.Booking) *dto.BookingResponse {
	if b == nil {
		return nil
	}
	userSimple := ToUserSimpleResponse(b.User)
	statusSimple := ToStatusSimpleResponse(b.Status)
	slots := make([]*dto.CourtSlotResponse, len(b.Slots))
	for i, s := range b.Slots {
		slots[i] = ToCourtSlotResponseFromProto(s)
	}
	return &dto.BookingResponse{
		ID:            b.Id,
		CourtID:       b.CourtId,
		Court:         ToCourtResponseFromProto(b.Court),
		UserID:        b.UserId,
		User:          &userSimple,
		Slots:         slots,
		TotalPrice:    b.TotalPrice,
		StatusID:      b.StatusId,
		Status:        &statusSimple,
		PaymentStatus: b.PaymentStatus,
		CreatedAt:     b.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:     b.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}
