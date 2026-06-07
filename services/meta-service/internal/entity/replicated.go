package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email        string         `gorm:"index"`
	Username     string         `gorm:"index"`
	FullName     string         `gorm:"default:''"`
	ActiveRoleID uuid.UUID      `gorm:"type:uuid"`
	StatusID     uuid.UUID      `gorm:"type:uuid"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Status       Status         `gorm:"-"`
	StatusName   string         `gorm:"->"`
}

func (User) TableName() string {
	return "replicated_users"
}
