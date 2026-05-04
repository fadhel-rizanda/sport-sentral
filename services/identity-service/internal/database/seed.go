package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
)

func Seed(db *gorm.DB) error {
	//if err := seedStatuses(db); err != nil {
	//	return err
	//}
	return seedRoles(db)
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{Name: entity.RoleAthlete, Description: "Default role. Can join academy, competitions, book courts."},
		{Name: entity.RoleScout, Description: "Talent finder. Freemium access to athlete profiles and leaderboard."},
		{Name: entity.RoleCourtOwner, Description: "Manages courts. Requires admin verification."},
		{Name: entity.RoleAcademyAdmin, Description: "Manages academies. Requires admin verification."},
		{Name: entity.RoleRegulator, Description: "Official sport body. Assigned by platform admin only."},
		{Name: entity.RolePlatformAdmin, Description: "Internal platform administrator."},
	}

	for _, r := range roles {
		var existing entity.Role
		err := db.Where("name = ?", r.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.ID = uuid.New()
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
