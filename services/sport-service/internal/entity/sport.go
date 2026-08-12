package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Sport struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name             string     `gorm:"type:varchar(100);not null"`
	Slug             string     `gorm:"type:varchar(100);not null"`
	Description      string     `gorm:"type:text"`
	IconAttachmentID *uuid.UUID `gorm:"type:uuid"`
	StatusID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	TierTagID        uuid.UUID  `gorm:"type:uuid;not null;index"`

	RegulatorID                 *uuid.UUID `gorm:"type:uuid;index"`
	RequiresApprovalForOfficial bool       `gorm:"default:false;not null"`
	RequiresApprovalForRegional bool       `gorm:"default:false;not null"`

	TotalCompetitions int32 `gorm:"default:0"`
	TotalAthletes     int32 `gorm:"default:0"`
	TotalAcademies    int32 `gorm:"default:0"`

	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Status          *Status      `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	TierTag         *Tag         `gorm:"foreignKey:TierTagID;references:ID;constraint:-;"`
	Config          *SportConfig `gorm:"foreignKey:SportID;references:ID;constraint:OnDelete:CASCADE"`
	ActiveRegulator *Regulator   `gorm:"foreignKey:RegulatorID;references:ID;constraint:OnDelete:SET NULL"`
	CreatedBy       *User        `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy       *User        `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy       *User        `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}

type SportConfig struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey"`
	SportID              uuid.UUID `gorm:"type:uuid;not null"`
	ParticipantTypeTagID uuid.UUID `gorm:"type:uuid;not null;index"`
	MinRosterSize        int32     `gorm:"not null"`
	MaxRosterSize        int32     `gorm:"not null"`
	TypicalRosterSize    int32     `gorm:"not null"`
	RulesURL             *string   `gorm:"type:text"`
	Description          string    `gorm:"type:text"`

	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Sport              *Sport      `gorm:"foreignKey:SportID;references:ID;constraint:-;"`
	ParticipantTypeTag *Tag        `gorm:"foreignKey:ParticipantTypeTagID;references:ID;constraint:-;"`
	Stats              []SportStat `gorm:"foreignKey:SportID;references:SportID;constraint:-;"`
}

type SportStat struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey"`
	SportID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	StatTypeTagID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	AggregationMethod string         `gorm:"type:varchar(20);not null;default:'AVG'"` // AVG, SUM, MAX, MIN
	CreatedAt         time.Time      `gorm:"not null"`
	UpdatedAt         time.Time      `gorm:"not null"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`

	// Relations
	Sport       *Sport `gorm:"foreignKey:SportID;references:ID;constraint:-;"`
	StatTypeTag *Tag   `gorm:"foreignKey:StatTypeTagID;references:ID;constraint:-;"`
}

func (SportStat) TableName() string {
	return "sport_stats"
}
