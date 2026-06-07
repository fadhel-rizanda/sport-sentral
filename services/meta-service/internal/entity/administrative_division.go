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

	Parent *AdministrativeDivision `gorm:"foreignKey:ParentID;references:ID;constraint:-;"`

	Country   *Country `gorm:"foreignKey:CountryID;references:ID;constraint:-;"`
	CreatedBy *User    `gorm:"foreignKey:CreatedByID;references:ID;constraint:-;"`
	UpdatedBy *User    `gorm:"foreignKey:UpdatedByID;references:ID;constraint:-;"`
	DeletedBy *User    `gorm:"foreignKey:DeletedByID;references:ID;constraint:-;"`
}
