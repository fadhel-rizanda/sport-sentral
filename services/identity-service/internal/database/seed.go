package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/shared/pkg/constants"
)

func Seed(db *gorm.DB) error {
	//if err := seedStatuses(db); err != nil {
	//	return err
	//}
	return seedRoles(db)
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{Name: constants.RoleAthlete, Description: "Default role. Can join academy, competitions, book courts."},
		{Name: constants.RoleScout, Description: "Talent finder. Freemium access to athlete profiles and leaderboard."},
		{Name: constants.RoleCourtOwner, Description: "Manages courts. Requires admin verification."},
		{Name: constants.RoleAcademyAdmin, Description: "Manages academies. Requires admin verification."},
		{Name: constants.RoleRegulator, Description: "Official sport body. Assigned by platform admin only."},
		{Name: constants.RolePlatformAdmin, Description: "Internal platform administrator."},
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
