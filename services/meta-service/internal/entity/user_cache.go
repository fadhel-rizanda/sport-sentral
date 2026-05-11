package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserCache struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email          string         `gorm:"index;not null"`
	Username       string         `gorm:"index;not null"`
	FullName       string         `gorm:"not null;default:''"`
	ActiveRoleID   uuid.UUID      `gorm:"type:uuid;not null"`
	ActiveRoleName string         `gorm:"not null;default:''"`
	StatusID       uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Status         Status         `gorm:"-"`
	StatusName     string         `gorm:"->"`
}
