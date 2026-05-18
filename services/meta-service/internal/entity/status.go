package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Status struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Type string    `gorm:"not null"`
	Name string    `gorm:"not null"`
	Slug string    `gorm:"not null;unique"`

	CreatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	CreatedBy *UserCache `gorm:"-"`
	UpdatedBy *UserCache `gorm:"-"`
	DeletedBy *UserCache `gorm:"-"`
}
