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

func (Status) TableName() string {
	return "replicated_statuses"
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
