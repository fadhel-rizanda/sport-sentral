package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdministrativeDivision struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CountryID  uuid.UUID  `gorm:"type:uuid;not null"`
	ParentID   *uuid.UUID `gorm:"type:uuid"`
	Name       string     `gorm:"type:varchar(255);not null"`
	Level      string     `gorm:"type:varchar(255);not null"`
	PostalCode string     `gorm:"type:varchar(255);not null"`

	CreatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Parent    *AdministrativeDivision `gorm:"foreignkey:ParentID"`
	Country   *Country                `gorm:"foreignkey:CountryID"`
	CreatedBy *User                   `gorm:"foreignkey:CreatedByID"`
	UpdatedBy *User                   `gorm:"foreignkey:UpdatedByID"`
	DeletedBy *User                   `gorm:"foreignkey:DeletedByID"`
}
