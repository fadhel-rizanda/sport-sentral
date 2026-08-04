package entity

import (
	"time"

	"github.com/google/uuid"
)

type WatchlistEntry struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	ScoutID       uuid.UUID `gorm:"type:uuid;not null;index:idx_scout_athlete"`
	AthleteID     uuid.UUID `gorm:"type:uuid;not null;index:idx_scout_athlete"`
	Notes         string    `gorm:"type:text"`
	PriorityTagID uuid.UUID `gorm:"type:uuid"`
	AddedAt       time.Time `gorm:"not null;index"`
	UpdatedAt     time.Time `gorm:"not null"`

	// Relations
	Scout          *Scout          `gorm:"foreignKey:ScoutID;references:ID;constraint:OnDelete:CASCADE"`
	AthleteProfile *AthleteProfile `gorm:"foreignKey:AthleteID;references:AthleteID;constraint:-;"`
}

func (WatchlistEntry) TableName() string {
	return "scout_watchlists"
}
