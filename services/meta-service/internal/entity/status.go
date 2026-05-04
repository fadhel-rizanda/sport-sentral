package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Status struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
	Type      string     `gorm:"not null"`
	Name      string     `gorm:"not null"`
	CreatedBy uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedBy uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedBy *uuid.UUID `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

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
	// shared
	StatusActive    = "active"
	StatusPending   = "pending"
	StatusSuspended = "suspended"
	StatusRejected  = "rejected"
	StatusBanned    = "banned"
	StatusInactive  = "inactive"

	// court booking specific
	StatusBooked    = "booked"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)
