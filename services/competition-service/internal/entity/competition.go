package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Competition struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`

	StartDate time.Time
	EndDate   *time.Time
	StatusID  uuid.UUID `gorm:"type:uuid;not null;index"`

	SportID             uuid.UUID  `gorm:"type:uuid;not null;index"`
	TierID              uuid.UUID  `gorm:"type:uuid;not null;index"`
	HostAcademyBranchID *uuid.UUID `gorm:"type:uuid;index"`

	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Sport       *Sport              `gorm:"foreignKey:SportID;references:ID;constraint:-;"`
	HostAcademy *AcademyBranch      `gorm:"foreignKey:HostAcademyBranchID;references:ID;constraint:-;"`
	TierTag     *Tag                `gorm:"foreignKey:TierID;references:ID;constraint:-;"`
	Status      *Status             `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	CreatedBy   *User               `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy   *User               `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy   *User               `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
	Branches    []CompetitionBranch `gorm:"foreignKey:CompetitionID;constraint:OnDelete:CASCADE"`
}

type CompetitionBranch struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompetitionID  uuid.UUID  `gorm:"type:uuid;not null;index:idx_comp_branch"`
	ParentBranchID *uuid.UUID `gorm:"type:uuid;index"`

	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	StatusID    uuid.UUID `gorm:"type:uuid;not null;index"`

	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Competition   *Competition        `gorm:"foreignKey:CompetitionID;constraint:OnDelete:CASCADE"`
	ParentBranch  *CompetitionBranch  `gorm:"foreignKey:ParentBranchID;references:ID;constraint:OnDelete:CASCADE"`
	Status        *Status             `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	ChildBranches []CompetitionBranch `gorm:"foreignKey:ParentBranchID;constraint:OnDelete:CASCADE"`
	Matches       []Match             `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE"`
	CreatedBy     *User               `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy     *User               `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy     *User               `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}

type CompetitionAdmin struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CompetitionID uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	RoleID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Competition *Competition `gorm:"foreignKey:CompetitionID;constraint:OnDelete:CASCADE"`
	User        *User        `gorm:"foreignKey:UserID;references:ID;constraint:-;"`
	Role        *Role        `gorm:"foreignKey:RoleID;references:ID;constraint:-;"`
}

func (CompetitionAdmin) TableName() string {
	return "competition_admins"
}
