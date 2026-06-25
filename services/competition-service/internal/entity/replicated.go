package entity

import (
	"time"

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

type AcademyHolding struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name              string     `gorm:"type:varchar(255);not null"`
	Description       string     `gorm:"type:text"`
	Email             string     `gorm:"type:varchar(255);not null"`
	PhoneNumber       string     `gorm:"type:varchar(255);not null"`
	ImageAttachmentID *uuid.UUID `gorm:"type:uuid"`
	StatusID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`

	// Relations
	Status   *Status         `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	Branches []AcademyBranch `gorm:"foreignKey:HoldingID;constraint:OnDelete:CASCADE;"`
}

func (AcademyHolding) TableName() string {
	return "replicated_academy_holding"
}

type AcademyBranch struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	HoldingID uuid.UUID `gorm:"type:uuid;not null;index"`
	SportID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:varchar(255);not null"`
	StatusID  uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Holding *AcademyHolding `gorm:"foreignKey:HoldingID;references:ID;constraint:-;"`
	Sport   *Sport          `gorm:"foreignKey:SportID;references:ID;constraint:-;"`
	Status  *Status         `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
}

func (AcademyBranch) TableName() string {
	return "replicated_academy_branches"
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

type AcademyAdmin struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	AcademyID uuid.UUID      `gorm:"type:uuid;not null;index:idx_academy_user_role"`
	BranchID  *uuid.UUID     `gorm:"type:uuid;index"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_academy_user_role"`
	RoleID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_academy_user_role"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Academy *AcademyHolding `gorm:"foreignKey:AcademyID;constraint:OnDelete:CASCADE"`
	Branch  *AcademyBranch  `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE"`
	User    *User           `gorm:"foreignKey:UserID;references:ID;constraint:-;"`
	Role    *Role           `gorm:"foreignKey:RoleID;references:ID;constraint:-;"`
}

func (AcademyAdmin) TableName() string {
	return "replicated_academy_admins"
}
