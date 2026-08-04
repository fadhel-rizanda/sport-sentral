package entity

import (
	"time"

	"github.com/google/uuid"
)

type LeaderboardEntry struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	SportID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_sport_period"`
	PeriodTagID   *uuid.UUID `gorm:"type:uuid;index:idx_sport_period"`
	Rank          int32      `gorm:"not null"`
	AthleteID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Score         float64    `gorm:"not null"`
	PreviousScore float64
	CalculatedAt  time.Time `gorm:"not null;index"`
	ExpiresAt     time.Time `gorm:"index"`

	// Relations
	AthleteProfile *AthleteProfile `gorm:"foreignKey:AthleteID;references:AthleteID;constraint:-;"`
}

func (LeaderboardEntry) TableName() string {
	return "leaderboard_entries"
}
