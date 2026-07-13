package events

const (
	CourtStreamName = "COURT_EVENTS"
)

const (
	SubjectCourtCreated = "court.court.created"
	SubjectCourtUpdated = "court.court.updated"
	SubjectCourtDeleted = "court.court.deleted"

	SubjectBookingCreated   = "court.booking.created"
	SubjectBookingConfirmed = "court.booking.confirmed"
	SubjectBookingCancelled = "court.booking.cancelled"
)
