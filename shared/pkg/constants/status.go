package constants

// ─── Status Types ─────────────────────────────────────────────────────────────

const (
	StatusTypeUser          = "user"
	StatusTypeUserRole      = "user_role"
	StatusTypeAcademy       = "academy"
	StatusTypeAcademyMember = "academy_member"
	StatusTypeCourt         = "court"
	StatusTypeCourtBooking  = "court_booking"
	StatusTypeCompetition   = "competition"
	StatusTypeEvent         = "event"
	StatusTypeParticipant   = "participant"
)

// ─── Status Values ────────────────────────────────────────────────────────────

const (
	StatusActive    = "active"
	StatusPending   = "pending"
	StatusSuspended = "suspended"
	StatusRejected  = "rejected"
	StatusBanned    = "banned"
	StatusInactive  = "inactive"

	StatusBooked    = "booked"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)
