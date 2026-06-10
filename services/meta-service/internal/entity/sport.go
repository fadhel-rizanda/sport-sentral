package entity

import "github.com/google/uuid"

// TODO: PHASE 2 PINDAHIN KE REGULATOR SERVICE
type Sport struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name             string    `gorm:"type:varchar(100);not null"`
	Slug             string    `gorm:"type:varchar(100);not null"`
	IconAttachmentID *uuid.UUID

	IsVerified  bool       `gorm:"default:false;index"`
	RegulatorID *uuid.UUID `gorm:"type:uuid;index"`
	Tier        string     `gorm:"type:varchar(20)"`
}
