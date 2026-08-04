package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scout struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"`
	OrganizationName   string         `gorm:"type:varchar(255)"`
	Bio                string         `gorm:"type:text"`
	AvatarAttachmentID *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt          time.Time      `gorm:"not null"`
	UpdatedAt          time.Time      `gorm:"not null"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`

	// Relations
	User         *User              `gorm:"foreignKey:UserID;references:ID;constraint:-;"`
	Watchlists   []WatchlistEntry   `gorm:"foreignKey:ScoutID;constraint:OnDelete:CASCADE"`
	ActivityLogs []ScoutActivityLog `gorm:"foreignKey:ScoutID;constraint:OnDelete:CASCADE"`
}

func (Scout) TableName() string {
	return "scouts"
}
