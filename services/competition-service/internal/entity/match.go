package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Match struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	BranchID uuid.UUID `gorm:"type:uuid;not null;index"`

	ScheduledAt time.Time `gorm:"index"`
	StartedAt   *time.Time
	EndedAt     *time.Time

	Status   string `gorm:"type:varchar(50);not null;default:'SCHEDULED'"`
	Location string `gorm:"type:varchar(255)"`
	Referee  string `gorm:"type:varchar(255)"`
	Notes    string `gorm:"type:text"`

	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Branch       *CompetitionBranch `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;"`
	Stats        []MatchStat        `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;"`
	Participants []MatchParticipant `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;"`
}

type MatchParticipant struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	MatchID     uuid.UUID `gorm:"type:uuid;not null;index"`
	RosterID    uuid.UUID `gorm:"type:uuid;not null;index"`
	FormatTagID uuid.UUID `gorm:"type:uuid;not null"`
	ResultTagID uuid.UUID `gorm:"type:uuid;not null"`
	Score       int32     `gorm:"default:0"`
	UpdatedAt   time.Time

	// Relations
	Match     *Match  `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;"`
	Roster    *Roster `gorm:"foreignKey:RosterID;references:ID;constraint:-;"`
	FormatTag *Tag    `gorm:"foreignKey:FormatTagID;references:ID;constraint:-;"`
	ResultTag *Tag    `gorm:"foreignKey:ResultTagID;references:ID;constraint:-;"`
}

type MatchParticipantScoreLog struct {
	ID                 uuid.UUID
	MatchParticipantID uuid.UUID
	PreviousScore      int32
	NewScore           int32
	Action             string
	UpdatedByID        uuid.UUID
	UpdatedAt          time.Time
}
