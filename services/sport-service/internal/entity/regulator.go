package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Regulator struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrganizationName string     `gorm:"type:varchar(100);not null"`
	Code             string     `gorm:"type:varchar(50);not null"`
	LogoAttachmentID *uuid.UUID `gorm:"type:uuid"`
	ContactEmail     string     `gorm:"type:varchar(255);not null"`
	PhoneNumber      *string    `gorm:"type:varchar(50)"`
	WebsiteURL       *string    `gorm:"type:text"`
	StatusID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	AddressID        uuid.UUID  `gorm:"type:uuid;not null;index"`

	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Status    *Status           `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	Staff     []RegulatorStaff  `gorm:"foreignKey:RegulatorID;constraint:OnDelete:CASCADE"`
	Address   *RegulatorAddress `gorm:"foreignKey:AddressID;constraint:OnDelete:CASCADE"`
	CreatedBy *User             `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy *User             `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy *User             `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}

type RegulatorAddress struct {
	ID                       uuid.UUID `gorm:"type:uuid;primaryKey"`
	StreetAddress            string    `gorm:"not null"`
	Notes                    *string   `gorm:"type:text"`
	Latitude                 *float64  `gorm:"type:decimal(10,8)"`
	Longitude                *float64  `gorm:"type:decimal(11,8)"`
	AdministrativeDivisionID uuid.UUID `gorm:"type:uuid;not null"`

	//	Relations
	AdministrativeDivision *AdministrativeDivision `gorm:"foreignKey:AdministrativeDivisionID;references:ID;constraint:-;"`
}

type RegulatorStaff struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	RegulatorID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	RoleTagID   uuid.UUID `gorm:"type:uuid;not null;index"`
	JoinedAt    time.Time `gorm:"not null"`

	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`

	// Relations
	Regulator *Regulator `gorm:"foreignKey:RegulatorID;references:ID;constraint:-;"`
	User      *User      `gorm:"foreignKey:UserID;references:ID;constraint:-;"`
	RoleTag   *Tag       `gorm:"foreignKey:RoleTagID;references:ID;constraint:-;"`
	CreatedBy *User      `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy *User      `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy *User      `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}
