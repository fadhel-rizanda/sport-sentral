package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "microservice-golang/gen/common/v1"
	courtv1 "microservice-golang/gen/venue/v1"
	"microservice-golang/services/venue-service/internal/dto"
	"microservice-golang/services/venue-service/internal/entity"
)

// ToSportSimpleResponse maps replicated Sport entity to dto
func ToSportSimpleResponse(s *entity.Sport) *dto.SportSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.SportSimpleResponse{
		ID:   s.ID.String(),
		Name: s.Name,
		Slug: s.Slug,
	}
}

// ToUserSimpleResponse maps replicated User entity to dto
func ToUserSimpleResponse(u *entity.User) *dto.UserSimpleResponse {
	if u == nil {
		return nil
	}
	return &dto.UserSimpleResponse{
		ID:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

// ToStatusSimpleResponse maps replicated Status entity to dto
func ToStatusSimpleResponse(s *entity.Status) *dto.StatusSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.StatusSimpleResponse{
		ID:   s.ID.String(),
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

// ToVenueResponse maps Venue entity to DTO
func ToVenueResponse(v *entity.Venue) *dto.VenueResponse {
	if v == nil {
		return nil
	}
	var adminDiv string
	if v.AdministrativeDivision != nil {
		adminDiv = v.AdministrativeDivision.Name
	}
	return &dto.VenueResponse{
		ID:                     v.ID.String(),
		Name:                   v.Name,
		Description:            v.Description,
		StreetAddress:          v.StreetAddress,
		Notes:                  v.Notes,
		Latitude:               v.Latitude,
		Longitude:              v.Longitude,
		AdministrativeDivision: adminDiv,
		OwnerID:                v.OwnerID.String(),
		Owner:                  ToUserSimpleResponse(v.Owner),
		Status:                 ToStatusSimpleResponse(v.Status),
		StatusID:               v.StatusID.String(),
		CreatedAt:              v.CreatedAt,
		UpdatedAt:              v.UpdatedAt,
	}
}

// ToCourtResponse maps Court entity to CourtResponse dto
func ToCourtResponse(c *entity.Court) *dto.CourtResponse {
	if c == nil {
		return nil
	}
	var imageStr string
	if c.ImageAttachmentID != nil {
		imageStr = c.ImageAttachmentID.String()
	}
	return &dto.CourtResponse{
		ID:                c.ID.String(),
		VenueID:           c.VenueID.String(),
		Venue:             ToVenueResponse(c.Venue),
		Name:              c.Name,
		Description:       c.Description,
		SportID:           c.SportID.String(),
		Sport:             ToSportSimpleResponse(c.Sport),
		PricePerHour:      c.PricePerHour,
		ImageAttachmentID: &imageStr,
		Status:            ToStatusSimpleResponse(c.Status),
		StatusID:          c.StatusID.String(),
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// ToCourtSlotResponse maps CourtSlot entity to dto
func ToCourtSlotResponse(s *entity.CourtSlot) *dto.CourtSlotResponse {
	if s == nil {
		return nil
	}
	var bookingStr string
	if s.BookingID != nil {
		bookingStr = s.BookingID.String()
	}
	return &dto.CourtSlotResponse{
		ID:        s.ID.String(),
		CourtID:   s.CourtID.String(),
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Price:     s.Price,
		StatusID:  s.StatusID.String(),
		Status:    ToStatusSimpleResponse(s.Status),
		BookingID: &bookingStr,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// ToBookingResponse maps Booking entity to BookingResponse dto
func ToBookingResponse(b *entity.Booking) *dto.BookingResponse {
	if b == nil {
		return nil
	}
	slots := make([]*dto.CourtSlotResponse, len(b.Slots))
	for i, s := range b.Slots {
		slots[i] = ToCourtSlotResponse(&s)
	}
	return &dto.BookingResponse{
		ID:            b.ID.String(),
		CourtID:       b.CourtID.String(),
		Court:         ToCourtResponse(b.Court),
		UserID:        b.UserID.String(),
		User:          ToUserSimpleResponse(b.User),
		Slots:         slots,
		TotalPrice:    b.TotalPrice,
		StatusID:      b.StatusID.String(),
		Status:        ToStatusSimpleResponse(b.Status),
		PaymentStatus: b.PaymentStatus,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

// ToProtoVenueFromDTO maps Venue dto to protobuf
func ToProtoVenueFromDTO(v *dto.VenueResponse) *courtv1.Venue {
	if v == nil {
		return nil
	}
	res := &courtv1.Venue{
		Id:            v.ID,
		Name:          v.Name,
		Description:   v.Description,
		StreetAddress: v.StreetAddress,
		OwnerId:       v.OwnerID,
		StatusId:      v.StatusID,
		CreatedAt:     timestamppb.New(v.CreatedAt),
		UpdatedAt:     timestamppb.New(v.UpdatedAt),
	}
	if v.Notes != nil {
		res.Notes = v.Notes
	}
	if v.Latitude != nil {
		res.Latitude = v.Latitude
	}
	if v.Longitude != nil {
		res.Longitude = v.Longitude
	}
	if v.AdministrativeDivision != "" {
		res.AdministrativeDivision = &commonv1.AdministrativeDivisionSimple{
			Name: v.AdministrativeDivision,
		}
	}
	if v.Owner != nil {
		res.Owner = &commonv1.UserSimple{
			Id:       v.Owner.ID,
			Email:    v.Owner.Email,
			Username: v.Owner.Username,
			FullName: v.Owner.FullName,
		}
	}
	if v.Status != nil {
		res.Status = &commonv1.StatusSimple{
			Id:   v.Status.ID,
			Type: v.Status.Type,
			Name: v.Status.Name,
			Slug: v.Status.Slug,
		}
	}
	return res
}

// ToProtoCourtFromDTO maps CourtResponse dto to protobuf
func ToProtoCourtFromDTO(c *dto.CourtResponse) *courtv1.Court {
	if c == nil {
		return nil
	}
	res := &courtv1.Court{
		Id:           c.ID,
		VenueId:      c.VenueID,
		Name:         c.Name,
		Description:  c.Description,
		SportId:      c.SportID,
		PricePerHour: c.PricePerHour,
		StatusId:     c.StatusID,
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}
	if c.ImageAttachmentID != nil {
		res.ImageAttachmentId = c.ImageAttachmentID
	}
	if c.Venue != nil {
		res.Venue = ToProtoVenueFromDTO(c.Venue)
	}
	if c.Sport != nil {
		res.Sport = &commonv1.SportSimple{
			Id:   c.Sport.ID,
			Name: c.Sport.Name,
			Slug: c.Sport.Slug,
		}
	}
	if c.Status != nil {
		res.Status = &commonv1.StatusSimple{
			Id:   c.Status.ID,
			Type: c.Status.Type,
			Name: c.Status.Name,
			Slug: c.Status.Slug,
		}
	}
	return res
}

// ToProtoCourtSlotFromDTO maps CourtSlotResponse dto to protobuf
func ToProtoCourtSlotFromDTO(s *dto.CourtSlotResponse) *courtv1.CourtSlot {
	if s == nil {
		return nil
	}
	res := &courtv1.CourtSlot{
		Id:        s.ID,
		CourtId:   s.CourtID,
		StartTime: timestamppb.New(s.StartTime),
		EndTime:   timestamppb.New(s.EndTime),
		Price:     s.Price,
		StatusId:  s.StatusID,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}
	if s.Status != nil {
		res.Status = &commonv1.StatusSimple{
			Id:   s.Status.ID,
			Type: s.Status.Type,
			Name: s.Status.Name,
			Slug: s.Status.Slug,
		}
	}
	if s.BookingID != nil {
		res.BookingId = s.BookingID
	}
	return res
}

// ToProtoBookingFromDTO maps BookingResponse dto to protobuf
func ToProtoBookingFromDTO(b *dto.BookingResponse) *courtv1.Booking {
	if b == nil {
		return nil
	}
	slots := make([]*courtv1.CourtSlot, len(b.Slots))
	for i, s := range b.Slots {
		slots[i] = ToProtoCourtSlotFromDTO(s)
	}
	res := &courtv1.Booking{
		Id:            b.ID,
		CourtId:       b.CourtID,
		UserId:        b.UserID,
		Slots:         slots,
		TotalPrice:    b.TotalPrice,
		StatusId:      b.StatusID,
		PaymentStatus: b.PaymentStatus,
		CreatedAt:     timestamppb.New(b.CreatedAt),
		UpdatedAt:     timestamppb.New(b.UpdatedAt),
	}
	if b.Court != nil {
		res.Court = ToProtoCourtFromDTO(b.Court)
	}
	if b.User != nil {
		res.User = &commonv1.UserSimple{
			Id:       b.User.ID,
			Email:    b.User.Email,
			Username: b.User.Username,
			FullName: b.User.FullName,
		}
	}
	if b.Status != nil {
		res.Status = &commonv1.StatusSimple{
			Id:   b.Status.ID,
			Type: b.Status.Type,
			Name: b.Status.Name,
			Slug: b.Status.Slug,
		}
	}
	return res
}
