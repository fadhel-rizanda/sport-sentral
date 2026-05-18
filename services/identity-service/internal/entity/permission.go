package entity

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Resource    string    `gorm:"not null"`
	Action      string    `gorm:"not null"`
	Description string    `gorm:"not null;default:''"`
	Slug        string    `gorm:"not null;unique"`

	CreatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	UpdatedByID uuid.UUID  `gorm:"type:uuid;not null"`
	DeletedByID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	CreatedBy User  `gorm:"foreignKey:CreatedByID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	UpdatedBy User  `gorm:"foreignKey:UpdatedByID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	DeletedBy *User `gorm:"foreignKey:DeletedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// kalau mau custom nama tabel, bisa pake method TableName() seperti ini, secara default gorm akan buat tabel dengan nama "permissions" (plural dari struct name)
//func (Permission) TableName() string { return "permissions" }

func (p *Permission) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return nil
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}
