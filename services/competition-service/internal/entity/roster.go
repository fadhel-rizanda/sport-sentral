package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Roster struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	AcademyBranchID uuid.UUID `gorm:"type:uuid;not null;index"`
	CompetitionID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name            string    `gorm:"type:varchar(255);not null"`
	TagID           uuid.UUID `gorm:"type:uuid;index"`
	StatusID        uuid.UUID `gorm:"type:uuid;not null"`
	MaxSize         int       `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	MemberCount int64 `gorm:"->"`

	// Relations
	AcademyBranch *AcademyBranch `gorm:"foreignKey:AcademyBranchID;references:ID;constraint:-;"`
	Competition   *Competition   `gorm:"foreignKey:CompetitionID;references:ID;constraint:-;"`
	Members       []RosterMember `gorm:"foreignKey:RosterID;constraint:OnDelete:CASCADE"`
	Tag           *Tag           `gorm:"foreignKey:TagID;references:ID;constraint:-;"`
	Status        *Status        `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
}

type RosterMember struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	RosterID     uuid.UUID `gorm:"type:uuid;not null;index"`
	AthleteID    uuid.UUID `gorm:"type:uuid;not null;index"`
	JerseyNumber *int
	PositionID   uuid.UUID `gorm:"type:uuid;not null"`
	StatusID     uuid.UUID `gorm:"type:uuid;not null"`

	AddedAt   time.Time `gorm:"not null;default:NOW()"`
	AddedByID uuid.UUID `gorm:"type:uuid;not null"`

	RemovedAt     *time.Time
	RemovedByID   *uuid.UUID `gorm:"type:uuid;index"`
	RemovalReason *string    `gorm:"type:text"`

	// Relations
	Roster    *Roster `gorm:"foreignKey:RosterID"`
	Athlete   *User   `gorm:"foreignKey:AthleteID;references:ID;constraint:-;"`
	Position  *Tag    `gorm:"foreignKey:PositionID;references:ID;constraint:-;"`
	Status    *Status `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	AddedBy   *User   `gorm:"foreignKey:AddedByID;references:ID;constraint:-;"`
	RemovedBy *User   `gorm:"foreignKey:RemovedByID;references:ID;constraint:-;"`
}
