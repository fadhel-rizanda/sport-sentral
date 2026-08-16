package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusPending = "PENDING"
	StatusActive  = "ACTIVE"
	StatusFailed  = "FAILED"
	StatusDeleted = "DELETED"
)

const (
	CategoryGeneral     = "GENERAL"
	CategoryImage       = "IMAGE"
	CategoryAvatar      = "AVATAR"
	CategoryLogo        = "LOGO"
	CategoryDocument    = "DOCUMENT"
	CategoryCertificate = "CERTIFICATE"
	CategoryVideo       = "VIDEO"
)

var ValidCategories = map[string]bool{
	CategoryGeneral:     true,
	CategoryImage:       true,
	CategoryAvatar:      true,
	CategoryLogo:        true,
	CategoryDocument:    true,
	CategoryCertificate: true,
	CategoryVideo:       true,
}

func IsValidCategory(category string) bool {
	return ValidCategories[category]
}

type Attachment struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Filename         string         `gorm:"type:varchar(255);not null"`
	FilePath         string         `gorm:"type:varchar(512);not null;index"`
	BucketName       string         `gorm:"type:varchar(100);not null"`
	FileSize         int64          `gorm:"not null"`
	MimeType         string         `gorm:"type:varchar(100);not null"`
	FileCategory     string         `gorm:"type:varchar(50);not null;index"`
	UploadedByUserID *uuid.UUID     `gorm:"type:uuid;index"`
	Status           string         `gorm:"type:varchar(20);not null;default:'PENDING';index"`
	URL              string         `gorm:"-"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (Attachment) TableName() string {
	return "attachments"
}
