package dto

import (
	"time"

	"github.com/google/uuid"
)

type SportSimpleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UserSimpleResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type StatusSimpleResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type VenueResponse struct {
	ID                     string                `json:"id"`
	Name                   string                `json:"name"`
	Description            string                `json:"description"`
	StreetAddress          string                `json:"street_address"`
	Notes                  *string               `json:"notes,omitempty"`
	Latitude               *float64              `json:"latitude,omitempty"`
	Longitude              *float64              `json:"longitude,omitempty"`
	AdministrativeDivision string                `json:"administrative_division"`
	OwnerID                string                `json:"owner_id"`
	Owner                  *UserSimpleResponse   `json:"owner,omitempty"`
	Status                 *StatusSimpleResponse `json:"status,omitempty"`
	StatusID               string                `json:"status_id"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

type CreateVenueRequest struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	StreetAddress   string    `json:"street_address"`
	Notes           *string   `json:"notes,omitempty"`
	Latitude        *float64  `json:"latitude,omitempty"`
	Longitude       *float64  `json:"longitude,omitempty"`
	AdminDivisionID uuid.UUID `json:"admin_division_id"`
	OwnerID         uuid.UUID `json:"owner_id"`
	StatusID        uuid.UUID `json:"status_id"`
}

type UpdateVenueRequest struct {
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	StreetAddress *string    `json:"street_address,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	Latitude      *float64   `json:"latitude,omitempty"`
	Longitude     *float64   `json:"longitude,omitempty"`
	StatusID      *uuid.UUID `json:"status_id,omitempty"`
}

type ListVenuesRequest struct {
	OwnerID  *uuid.UUID `json:"owner_id,omitempty"`
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	Search   *string    `json:"search,omitempty"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type ListVenuesResponse struct {
	Venues   []*VenueResponse `json:"venues"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type CourtResponse struct {
	ID                string                `json:"id"`
	VenueID           string                `json:"venue_id"`
	Venue             *VenueResponse        `json:"venue,omitempty"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	SportID           string                `json:"sport_id"`
	Sport             *SportSimpleResponse  `json:"sport,omitempty"`
	PricePerHour      int64                 `json:"price_per_hour"` // in cents
	ImageAttachmentID *string               `json:"image_attachment_id,omitempty"`
	Status            *StatusSimpleResponse `json:"status,omitempty"`
	StatusID          string                `json:"status_id"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type CreateCourtRequest struct {
	VenueID         uuid.UUID  `json:"venue_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	SportID         uuid.UUID  `json:"sport_id"`
	PricePerHour    int64      `json:"price_per_hour"` // in cents
	ImageAttachment *uuid.UUID `json:"image_attachment_id,omitempty"`
	StatusID        uuid.UUID  `json:"status_id"`
}

type ListCourtsRequest struct {
	VenueID  *uuid.UUID `json:"venue_id,omitempty"`
	SportID  *uuid.UUID `json:"sport_id,omitempty"`
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	Search   *string    `json:"search,omitempty"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type ListCourtsResponse struct {
	Courts   []*CourtResponse `json:"courts"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type UpdateCourtRequest struct {
	Name            *string    `json:"name,omitempty"`
	Description     *string    `json:"description,omitempty"`
	SportID         *uuid.UUID `json:"sport_id,omitempty"`
	PricePerHour    *int64     `json:"price_per_hour,omitempty"`
	ImageAttachment *uuid.UUID `json:"image_attachment_id,omitempty"`
	StatusID        *uuid.UUID `json:"status_id,omitempty"`
}

type CourtSlotResponse struct {
	ID        string                `json:"id"`
	CourtID   string                `json:"court_id"`
	StartTime time.Time             `json:"start_time"`
	EndTime   time.Time             `json:"end_time"`
	Price     int64                 `json:"price"` // in cents
	StatusID  string                `json:"status_id"`
	Status    *StatusSimpleResponse `json:"status,omitempty"`
	BookingID *string               `json:"booking_id,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type SlotInput struct {
	StartTime time.Time
	EndTime   time.Time
	Price     *int64 // in cents
}

type CreateCourtSlotsRequest struct {
	CourtID  uuid.UUID
	Slots    []SlotInput
	StatusID uuid.UUID
}

type ListCourtSlotsRequest struct {
	CourtID   uuid.UUID
	StatusID  *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}

type BookingResponse struct {
	ID            string                `json:"id"`
	CourtID       string                `json:"court_id"`
	Court         *CourtResponse        `json:"court,omitempty"`
	UserID        string                `json:"user_id"`
	User          *UserSimpleResponse   `json:"user,omitempty"`
	Slots         []*CourtSlotResponse  `json:"slots"`
	TotalPrice    int64                 `json:"total_price"` // in cents
	StatusID      string                `json:"status_id"`
	Status        *StatusSimpleResponse `json:"status,omitempty"`
	PaymentStatus string                `json:"payment_status"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type CreateBookingRequest struct {
	CourtID  uuid.UUID   `json:"court_id"`
	UserID   uuid.UUID   `json:"user_id"`
	SlotIDs  []uuid.UUID `json:"slot_ids"`
	StatusID uuid.UUID   `json:"status_id"`
}

type ListBookingsRequest struct {
	CourtID  *uuid.UUID `json:"court_id,omitempty"`
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type ListBookingsResponse struct {
	Bookings []*BookingResponse `json:"bookings"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type UpdateBookingStatusRequest struct {
	StatusID      uuid.UUID `json:"status_id"`
	PaymentStatus *string   `json:"payment_status,omitempty"`
}
