package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	IsActive  bool      `gorm:"not null;default:false"`
	StatusID  uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Status StatusCache `gorm:"-"`
	Role   Role        `gorm:"foreignKey:RoleID;references:ID"`
}
