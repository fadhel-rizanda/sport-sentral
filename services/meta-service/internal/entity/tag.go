package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Type string    `gorm:"not null"`
	Name string    `gorm:"not null;unique"`
	Slug string    `gorm:"not null;unique"`

	CreatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID `gorm:"type:uuid;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	CreatedBy *UserCache `gorm:"-"`
	UpdatedBy *UserCache `gorm:"-"`
	DeletedBy *UserCache `gorm:"-"`
}

// ─── Tag Types ────────────────────────────────────────────────────────────────

const (
	TagTypeSportCategory   = "sport_category"
	TagTypePlayerPosition  = "player_position"
	TagTypeCourtFacility   = "court_facility"
	TagTypeAcademyFocus    = "academy_focus"
	TagTypeCompetitionType = "competition_type"
)
