package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademyHolding struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name              string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description       string     `gorm:"type:text"`
	Email             string     `gorm:"type:varchar(255);not null"`
	PhoneNumber       string     `gorm:"type:varchar(255);not null"`
	ImageAttachmentID *uuid.UUID `gorm:"type:uuid"`
	StatusID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	AddressID         uuid.UUID  `gorm:"type:uuid;not null;index"`

	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Branches  []AcademyBranch        `gorm:"foreignKey:HoldingID;constraint:OnDelete:CASCADE"`
	Address   *AcademyHoldingAddress `gorm:"foreignKey:AddressID;constraint:OnDelete:CASCADE"`
	Status    *Status                `gorm:"foreignKey:StatusID;references:ID"`
	CreatedBy *User                  `gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy *User                  `gorm:"foreignKey:UpdatedByID;references:ID"`
	DeletedBy *User                  `gorm:"foreignKey:DeletedByID;references:ID"`
}

type AcademyBranch struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	HoldingID uuid.UUID `gorm:"type:uuid;not null;index"`
	SportID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:varchar(255);not null"`

	Email       string `gorm:"type:varchar(255);not null"`
	PhoneNumber string `gorm:"type:varchar(255);not null"`

	StatusID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Holding   *AcademyHolding `gorm:"foreignKey:HoldingID;constraint:OnDelete:CASCADE"`
	Sport     *Sport          `gorm:"foreignKey:SportID;references:ID"`
	Status    *Status         `gorm:"foreignKey:StatusID;references:ID"`
	CreatedBy *User           `gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy *User           `gorm:"foreignKey:UpdatedByID;references:ID"`
	DeletedBy *User           `gorm:"foreignKey:DeletedByID;references:ID"`

	Enrollments []Enrollment           `gorm:"foreignKey:AcademyBranchID;constraint:OnDelete:CASCADE"`
	Rosters     []Roster               `gorm:"foreignKey:AcademyBranchID;constraint:OnDelete:CASCADE"`
	Addresses   []AcademyBranchAddress `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE"`
}

// TODO assign or revoke not yet configured
type AcademyAdmin struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AcademyID uuid.UUID  `gorm:"type:uuid;not null;index:idx_academy_user_role"`
	BranchID  *uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_academy_user_role"`

	RoleID       uuid.UUID `gorm:"type:uuid;not null;index:idx_academy_user_role"`
	ApprovedAt   *time.Time
	ApprovedByID *uuid.UUID `gorm:"type:uuid"`

	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Academy    *AcademyHolding `gorm:"foreignKey:AcademyID;constraint:OnDelete:CASCADE"`
	Branch     *AcademyBranch  `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE"`
	User       *User           `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Role       *Role           `gorm:"foreignKey:RoleID;references:ID"`
	ApprovedBy *User           `gorm:"foreignKey:ApprovedByID;references:ID"`
	CreatedBy  *User           `gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy  *User           `gorm:"foreignKey:UpdatedByID;references:ID"`
	DeletedBy  *User           `gorm:"foreignKey:DeletedByID;references:ID"`
}

type AcademyHoldingAddress struct {
	ID                       uuid.UUID `gorm:"type:uuid;primaryKey"`
	StreetAddress            string    `gorm:"not null"`
	Notes                    *string   `gorm:"type:text"`
	Latitude                 *float64  `gorm:"type:decimal(10,8)"`
	Longitude                *float64  `gorm:"type:decimal(11,8)"`
	AdministrativeDivisionID uuid.UUID `gorm:"type:uuid;not null"`

	//	Relations
	AdministrativeDivision *AdministrativeDivision `gorm:"foreignKey:ID;constraint:OnDelete:CASCADE"`
}

type AcademyBranchAddress struct {
	ID                       uuid.UUID `gorm:"type:uuid;primaryKey"`
	BranchID                 uuid.UUID `gorm:"type:uuid;not null;index"`
	IsPrimary                bool      `gorm:"default:false"`
	StreetAddress            string    `gorm:"not null"`
	Notes                    *string   `gorm:"type:text"`
	Latitude                 *float64  `gorm:"type:decimal(10,8)"`
	Longitude                *float64  `gorm:"type:decimal(11,8)"`
	AdministrativeDivisionID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Relations
	Branch                 *AcademyBranch          `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE"`
	AdministrativeDivision *AdministrativeDivision `gorm:"foreignKey:AdministrativeDivisionID;constraint:OnDelete:CASCADE"`
}
