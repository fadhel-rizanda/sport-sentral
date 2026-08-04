package entity

import (
	"time"

	"github.com/google/uuid"
)

// DEPRECATED
type ScoutActivityLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ScoutID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Action       string     `gorm:"type:varchar(100);not null"`
	ResourceID   *uuid.UUID `gorm:"type:uuid"`
	ResourceType string     `gorm:"type:varchar(50)"`
	CreatedAt    time.Time  `gorm:"not null;index"`

	// Relations
	Scout *Scout `gorm:"foreignKey:ScoutID;references:ID;constraint:OnDelete:CASCADE"`
}

func (ScoutActivityLog) TableName() string {
	return "scout_activity_logs"
}
