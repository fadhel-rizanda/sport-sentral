package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email        string         `gorm:"index;not null"`
	Username     string         `gorm:"index;not null"`
	FullName     string         `gorm:"not null;default:''"`
	ActiveRoleID uuid.UUID      `gorm:"type:uuid;not null"`
	StatusID     uuid.UUID      `gorm:"type:uuid;not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Status       Status         `gorm:"-"`
	StatusName   string         `gorm:"->"`
}

func (User) TableName() string {
	return "replicated_users"
}
