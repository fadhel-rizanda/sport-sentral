package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email     string         `gorm:"index"`
	Username  string         `gorm:"index"`
	FullName  string         `gorm:"default:''"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "replicated_users"
}

type Status struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Type      string         `gorm:"type:varchar(50)"`
	Name      string         `gorm:"type:varchar(100)"`
	Slug      string         `gorm:"type:varchar(100)"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Status) TableName() string {
	return "replicated_statuses"
}

type Tag struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Type      string         `gorm:"type:varchar(50)"`
	Name      string         `gorm:"type:varchar(100)"`
	Slug      string         `gorm:"type:varchar(100)"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Tag) TableName() string {
	return "replicated_tags"
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
