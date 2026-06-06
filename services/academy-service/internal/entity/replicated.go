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
}

func (User) TableName() string {
	return "replicated_users"
}

type Sport struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name             string    `gorm:"type:varchar(100);not null"`
	Slug             string    `gorm:"type:varchar(100);not null"`
	IconAttachmentID *uuid.UUID

	IsVerified  bool       `gorm:"default:false;index"`
	RegulatorID *uuid.UUID `gorm:"type:uuid;index"`
	Tier        string     `gorm:"type:varchar(20)"`
}

func (Sport) TableName() string {
	return "replicated_sports"
}

type Status struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Type      string         `gorm:"not null"`
	Name      string         `gorm:"not null"`
	Slug      string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Status) TableName() string {
	return "replicated_statuses"
}

type Tag struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Type      string         `gorm:"not null"`
	Name      string         `gorm:"not null"`
	Slug      string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Tag) TableName() string {
	return "replicated_tags"
}

type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"uniqueIndex;not null"`
	Description string         `gorm:"not null;default:''"`
	Slug        string         `gorm:"not null;unique"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string {
	return "replicated_roles"
}

type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Resource    string         `gorm:"not null"`
	Action      string         `gorm:"not null"`
	Description string         `gorm:"not null;default:''"`
	Slug        string         `gorm:"not null;unique"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (Permission) TableName() string {
	return "replicated_permissions"
}

type Country struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name         string         `gorm:"type:varchar(255);not null"`
	ISOAlpha2    string         `gorm:"type:varchar(255);not null"`
	ISOAlpha3    string         `gorm:"type:varchar(255);not null"`
	PhoneCode    string         `gorm:"type:varchar(255);not null"`
	CurrencyCode string         `gorm:"type:varchar(255);not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (Country) TableName() string {
	return "replicated_countries"
}

type AdministrativeDivision struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CountryID  uuid.UUID      `gorm:"type:uuid;not null"`
	ParentID   *uuid.UUID     `gorm:"type:uuid"`
	Name       string         `gorm:"type:varchar(255);not null"`
	Level      string         `gorm:"type:varchar(255);not null"`
	PostalCode string         `gorm:"type:varchar(255);not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	Parent  *AdministrativeDivision `gorm:"foreignkey:ParentID"`
	Country Country                 `gorm:"foreignkey:CountryID"`
}

func (AdministrativeDivision) TableName() string {
	return "replicated_administrative_divisions"
}
