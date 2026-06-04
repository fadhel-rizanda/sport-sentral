package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Status struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Type      string         `gorm:"not null"`
	Name      string         `gorm:"not null"`
	Slug      string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
