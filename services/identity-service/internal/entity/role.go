package entity

import (
	"microservice-golang/shared/pkg/constants"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex;not null"`
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

	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

func (r *Role) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		r.ID = id
	}
	return nil
}

var RolesInstantActive = map[string]bool{
	constants.RoleAthlete: true,
	constants.RoleScout:   true,
}

var RolesPendingApproval = map[string]bool{
	constants.RoleCourtOwner:   true,
	constants.RoleAcademyAdmin: true,
	constants.RoleOrganizer:    true,
}

var RolesAdminAssignOnly = map[string]bool{
	constants.RoleRegulator:     true,
	constants.RolePlatformAdmin: true,
}

var RolesExpandableFrom = map[string][]string{
	constants.RoleAthlete: {constants.RoleScout, constants.RoleCourtOwner, constants.RoleAcademyAdmin, constants.RoleOrganizer},
}
