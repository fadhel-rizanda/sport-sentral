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
}

func (User) TableName() string {
	return "replicated_users"
}

type Sport struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name             string    `gorm:"type:varchar(100)"`
	Slug             string    `gorm:"type:varchar(100)"`
	IconAttachmentID *uuid.UUID

	IsVerified  bool       `gorm:"default:false;index"`
	RegulatorID *uuid.UUID `gorm:"type:uuid;index"`
	Tier        string     `gorm:"type:varchar(20)"`

	Stats []SportStat `gorm:"foreignKey:SportID;constraint:OnDelete:CASCADE"`
}

func (Sport) TableName() string {
	return "replicated_sports"
}

type SportStat struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	SportID           uuid.UUID `gorm:"type:uuid;not null;index"`
	StatTypeTagID     uuid.UUID `gorm:"type:uuid;not null;index"`
	AggregationMethod string    `gorm:"type:varchar(20);not null;default:'AVG'"`
}

func (SportStat) TableName() string {
	return "replicated_sport_stats"
}

type Status struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Type      string
	Name      string
	Slug      string
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Status) TableName() string {
	return "replicated_statuses"
}

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Type      string
	Name      string
	Slug      string
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Tag) TableName() string {
	return "replicated_tags"
}

type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"index"`
	Description string         `gorm:"default:''"`
	Slug        string         `gorm:"index"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Permissions []Permission `gorm:"many2many:role_permissions;constraint:-;"`
}

func (Role) TableName() string {
	return "replicated_roles"
}

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Resource    string
	Action      string
	Description string         `gorm:"default:''"`
	Slug        string         `gorm:"index"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (Permission) TableName() string {
	return "replicated_permissions"
}

type Country struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name         string         `gorm:"type:varchar(255)"`
	ISOAlpha2    string         `gorm:"type:varchar(255)"`
	ISOAlpha3    string         `gorm:"type:varchar(255)"`
	PhoneCode    string         `gorm:"type:varchar(255)"`
	CurrencyCode string         `gorm:"type:varchar(255)"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (Country) TableName() string {
	return "replicated_countries"
}

type AdministrativeDivision struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CountryID  uuid.UUID      `gorm:"type:uuid"`
	ParentID   *uuid.UUID     `gorm:"type:uuid"`
	Name       string         `gorm:"type:varchar(255)"`
	Level      string         `gorm:"type:varchar(255)"`
	PostalCode string         `gorm:"type:varchar(255)"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	Parent  *AdministrativeDivision `gorm:"foreignKey:ParentID;constraint:-;"`
	Country Country                 `gorm:"foreignKey:CountryID;constraint:-;"`
}

func (AdministrativeDivision) TableName() string {
	return "replicated_administrative_divisions"
}
