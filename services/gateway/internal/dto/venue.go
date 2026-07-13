package dto

type VenueResponse struct {
	ID                     string                                `json:"id"`
	Name                   string                                `json:"name"`
	Description            string                                `json:"description"`
	StreetAddress          string                                `json:"street_address"`
	Notes                  string                                `json:"notes,omitempty"`
	Latitude               float64                               `json:"latitude,omitempty"`
	Longitude              float64                               `json:"longitude,omitempty"`
	AdministrativeDivision *AdministrativeDivisionSimpleResponse `json:"administrative_division,omitempty"`
	OwnerID                string                                `json:"owner_id"`
	Owner                  *UserSimpleResponse                   `json:"owner,omitempty"`
	Status                 *StatusSimpleResponse                 `json:"status,omitempty"`
	StatusID               string                                `json:"status_id"`
	CreatedAt              string                                `json:"created_at"`
	UpdatedAt              string                                `json:"updated_at"`
}

type CreateVenueRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	StreetAddress   string  `json:"street_address"`
	Notes           string  `json:"notes,omitempty"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	AdminDivisionID string  `json:"admin_division_id"`
	StatusID        string  `json:"status_id"`
}

type UpdateVenueRequest struct {
	Name          *string  `json:"name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	StreetAddress *string  `json:"street_address,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	StatusID      *string  `json:"status_id,omitempty"`
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
	ImageAttachmentID string                `json:"image_attachment_id,omitempty"`
	Status            *StatusSimpleResponse `json:"status,omitempty"`
	StatusID          string                `json:"status_id"`
	CreatedAt         string                `json:"created_at"`
	UpdatedAt         string                `json:"updated_at"`
}

type CreateCourtRequest struct {
	VenueID         string `json:"venue_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	SportID         string `json:"sport_id"`
	PricePerHour    int64  `json:"price_per_hour"` // in cents
	ImageAttachment string `json:"image_attachment_id,omitempty"`
	StatusID        string `json:"status_id"`
}

type UpdateCourtRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	SportID         *string `json:"sport_id,omitempty"`
	PricePerHour    *int64  `json:"price_per_hour,omitempty"`
	ImageAttachment *string `json:"image_attachment_id,omitempty"`
	StatusID        *string `json:"status_id,omitempty"`
}

type CourtSlotResponse struct {
	ID        string                `json:"id"`
	CourtID   string                `json:"court_id"`
	StartTime string                `json:"start_time"`
	EndTime   string                `json:"end_time"`
	Price     int64                 `json:"price"` // in cents
	StatusID  string                `json:"status_id"`
	Status    *StatusSimpleResponse `json:"status,omitempty"`
	BookingID string                `json:"booking_id,omitempty"`
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
}

type SlotInput struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Price     *int64 `json:"price,omitempty"` // in cents
}

type CreateCourtSlotsRequest struct {
	Slots    []SlotInput `json:"slots"`
	StatusID string      `json:"status_id"`
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
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
}

type CreateBookingRequest struct {
	CourtID  string   `json:"court_id"`
	SlotIDs  []string `json:"slot_ids"`
	StatusID string   `json:"status_id"`
}

type UpdateBookingStatusRequest struct {
	StatusID      string  `json:"status_id"`
	PaymentStatus *string `json:"payment_status,omitempty"`
}
