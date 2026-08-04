package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AthleteProfile struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AthleteID               uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_athlete_sport"`
	SportID                 uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_athlete_sport;index"`
	CurrentAcademyID        *uuid.UUID `gorm:"type:uuid"`
	CurrentAcademyName      string     `gorm:"type:varchar(255)"`
	Age                     int32
	LeaderboardRank         int32                              `gorm:"index"`
	LeaderboardScore        float64                            `gorm:"index"`
	HighlightStats          datatypes.JSONType[map[string]any] `gorm:"type:jsonb;default:'[]'"`
	HighestCompetitionLevel string                             `gorm:"type:varchar(50)"`
	TotalCompetitions       int32
	TotalAcademies          int32
	YearsActive             int32
	LastUpdatedAt           time.Time
	CreatedAt               time.Time `gorm:"not null"`
	UpdatedAt               time.Time `gorm:"not null"`

	// Relations
	Athlete *User `gorm:"foreignKey:AthleteID;references:ID;constraint:-;"`
}

func (AthleteProfile) TableName() string {
	return "athlete_profiles"
}
