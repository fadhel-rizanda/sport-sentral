package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// TODO kurang audit kaya tag/status
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null;default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Permissions []*Permission `gorm:"many2many:role_permissions;"`
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

func NewRole(name, description string) *Role {
	return &Role{
		Name:        name,
		Description: description,
	}
}

const (
	RoleAthlete       = "athlete"
	RoleScout         = "scout"
	RoleCourtOwner    = "court_owner"
	RoleAcademyAdmin  = "academy_admin"
	RoleRegulator     = "regulator"
	RolePlatformAdmin = "platform_admin"
)

var RolesInstantActive = map[string]bool{
	RoleAthlete: true,
	RoleScout:   true,
}

var RolesPendingApproval = map[string]bool{
	RoleCourtOwner:   true,
	RoleAcademyAdmin: true,
}

var RolesAdminAssignOnly = map[string]bool{
	RoleRegulator:     true,
	RolePlatformAdmin: true,
}

var RolesExpandableFrom = map[string][]string{
	RoleAthlete: {RoleScout, RoleCourtOwner, RoleAcademyAdmin},
}
