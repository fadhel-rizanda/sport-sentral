package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Enrollment struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	AcademyBranchID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_academy_athlete"`
	AthleteID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_academy_athlete"`

	JoinedAt     time.Time `gorm:"not null"`
	LeftAt       *time.Time
	ExpiresAt    *time.Time
	ApprovedAt   *time.Time
	ApprovedByID *uuid.UUID `gorm:"type:uuid;index"`

	StatusID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	AcademyBranch *AcademyBranch `gorm:"foreignKey:AcademyBranchID;constraint:OnDelete:CASCADE"`
	Athlete       *User          `gorm:"foreignKey:AthleteID;references:ID;constraint:-;"`
	ApprovedBy    *User          `gorm:"foreignKey:ApprovedByID;references:ID;constraint:-;"`
	Status        *Status        `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	CreatedBy     *User          `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy     *User          `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy     *User          `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}
