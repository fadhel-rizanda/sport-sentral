package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Country struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"type:varchar(255);not null"`
	ISOAlpha2    string    `gorm:"type:varchar(255);not null"`
	ISOAlpha3    string    `gorm:"type:varchar(255);not null"`
	PhoneCode    string    `gorm:"type:varchar(255);not null"`
	CurrencyCode string    `gorm:"type:varchar(255);not null"`

	CreatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	CreatedBy *User `gorm:"foreignkey:CreatedByID"`
	UpdatedBy *User `gorm:"foreignkey:UpdatedByID"`
	DeletedBy *User `gorm:"foreignkey:DeletedByID"`
}
