package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchStat struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	MatchID    uuid.UUID `gorm:"type:uuid;not null;index"`
	AthleteID  uuid.UUID `gorm:"type:uuid;not null;index"`
	StatTypeID uuid.UUID `gorm:"type:uuid;not null;index"`

	Value float64 `gorm:"not null"`

	RecordedByID uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relations
	Match      *Match `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE"`
	Athlete    *User  `gorm:"foreignKey:AthleteID;references:ID;constraint:-;"`
	StatType   *Tag   `gorm:"foreignKey:StatTypeID;references:ID;constraint:-;"`
	RecordedBy *User  `gorm:"foreignKey:RecordedByID;references:ID;constraint:-;"`
}

type AthleteStatsAggregate struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompetitionID uuid.UUID `gorm:"type:uuid;not null;index"`
	BranchID      uuid.UUID `gorm:"type:uuid;not null;index"`
	AthleteID     uuid.UUID `gorm:"type:uuid;not null;index"`
	StatTypeID    uuid.UUID `gorm:"type:uuid;not null;index"`

	TotalMatches int32   `gorm:"default:0"`
	AvgValue     float64 `gorm:"default:0"`
	SumValue     float64 `gorm:"default:0"`
	MaxValue     float64 `gorm:"default:0"`
	MinValue     float64 `gorm:"default:0"`

	LastUpdatedAt time.Time `gorm:"not null;index"`

	// Relations
	Competition *Competition   `gorm:"foreignKey:CompetitionID;constraint:OnDelete:CASCADE"`
	Branch      *AcademyBranch `gorm:"foreignKey:BranchID;references:ID;constraint:-;"`
	Athlete     *User          `gorm:"foreignKey:AthleteID;references:ID;constraint:-;"`
	StatType    *Tag           `gorm:"foreignKey:StatTypeID;references:ID;constraint:-;"`
}
