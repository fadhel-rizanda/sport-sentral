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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// kalau mau custom nama tabel, bisa pake method TableName() seperti ini, secara default gorm akan buat tabel dengan nama "permissions" (plural dari struct name)
//func (Permission) TableName() string { return "permissions" }

func (p *Permission) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func NewPermission(resource, action, description string) *Permission {
	return &Permission{
		Resource:    resource,
		Action:      action,
		Description: description,
	}
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}
