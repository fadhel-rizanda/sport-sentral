package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/shared/pkg/constants"
)

func Seed(db *gorm.DB) error {
	if err := seedPermissions(db); err != nil {
		return err
	}
	if err := seedRoles(db); err != nil {
		return err
	}
	return seedRolePermissions(db)
}

func seedPermissions(db *gorm.DB) error {
	createdBy, err := uuid.Parse("389d7e0d-4bdd-4ecc-9d7b-88ad8a8055db")
	if err != nil {
		return err
	}

	permissions := []entity.Permission{
		// User Management
		{Resource: "user", Action: "read", Slug: "user.read", Description: "View user profiles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "user", Action: "create", Slug: "user.create", Description: "Create new users", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "user", Action: "update", Slug: "user.update", Description: "Update user information", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "user", Action: "delete", Slug: "user.delete", Description: "Delete users", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "user", Action: "verify", Slug: "user.verify", Description: "Verify user accounts", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Role Management
		{Resource: "role", Action: "read", Slug: "role.read", Description: "View roles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "role", Action: "create", Slug: "role.create", Description: "Create new roles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "role", Action: "update", Slug: "role.update", Description: "Update roles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "role", Action: "delete", Slug: "role.delete", Description: "Delete roles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "role", Action: "assign", Slug: "role.assign", Description: "Assign roles to users", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Permission Management
		{Resource: "permission", Action: "read", Slug: "permission.read", Description: "View permissions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "permission", Action: "create", Slug: "permission.create", Description: "Create permissions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "permission", Action: "update", Slug: "permission.update", Description: "Update permissions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "permission", Action: "delete", Slug: "permission.delete", Description: "Delete permissions", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Athlete Profiles
		{Resource: "athlete", Action: "read", Slug: "athlete.read", Description: "View athlete profiles", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "athlete", Action: "read_premium", Slug: "athlete.read_premium", Description: "View premium athlete data", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "athlete", Action: "update", Slug: "athlete.update", Description: "Update athlete profiles", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Academy Management
		{Resource: "academy", Action: "read", Slug: "academy.read", Description: "View academies", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "academy", Action: "create", Slug: "academy.create", Description: "Create academies", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "academy", Action: "update", Slug: "academy.update", Description: "Update academies", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "academy", Action: "delete", Slug: "academy.delete", Description: "Delete academies", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "academy", Action: "manage", Slug: "academy.manage", Description: "Full academy management", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Court Management
		{Resource: "court", Action: "read", Slug: "court.read", Description: "View courts", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "court", Action: "create", Slug: "court.create", Description: "Create courts", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "court", Action: "update", Slug: "court.update", Description: "Update courts", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "court", Action: "delete", Slug: "court.delete", Description: "Delete courts", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "court", Action: "manage", Slug: "court.manage", Description: "Full court management", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Booking Management
		{Resource: "booking", Action: "read", Slug: "booking.read", Description: "View bookings", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "booking", Action: "create", Slug: "booking.create", Description: "Create bookings", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "booking", Action: "update", Slug: "booking.update", Description: "Update bookings", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "booking", Action: "delete", Slug: "booking.delete", Description: "Cancel bookings", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Competition Management
		{Resource: "competition", Action: "read", Slug: "competition.read", Description: "View competitions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "competition", Action: "create", Slug: "competition.create", Description: "Create competitions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "competition", Action: "update", Slug: "competition.update", Description: "Update competitions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "competition", Action: "delete", Slug: "competition.delete", Description: "Delete competitions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "competition", Action: "join", Slug: "competition.join", Description: "Join competitions", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "competition", Action: "regulate", Slug: "competition.regulate", Description: "Regulate competitions", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Leaderboard
		{Resource: "leaderboard", Action: "read", Slug: "leaderboard.read", Description: "View leaderboard", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "leaderboard", Action: "read_premium", Slug: "leaderboard.read_premium", Description: "View premium leaderboard data", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Verification & Approval
		{Resource: "verification", Action: "approve", Slug: "verification.approve", Description: "Approve verification requests", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "verification", Action: "reject", Slug: "verification.reject", Description: "Reject verification requests", CreatedByID: createdBy, UpdatedByID: createdBy},

		// Platform Administration
		{Resource: "platform", Action: "admin", Slug: "platform.admin", Description: "Platform administration access", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Resource: "platform", Action: "analytics", Slug: "platform.analytics", Description: "View platform analytics", CreatedByID: createdBy, UpdatedByID: createdBy},
	}

	for _, p := range permissions {
		var existing entity.Permission
		err := db.Where("slug = ?", p.Slug).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p.ID = uuid.New()
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	createdBy, err := uuid.Parse("389d7e0d-4bdd-4ecc-9d7b-88ad8a8055db")
	if err != nil {
		return err
	}
	roles := []entity.Role{
		{
			Name:        constants.RoleAthlete,
			Slug:        "athlete",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Default role. Can join academy, competitions, book courts.",
		},
		{
			Name:        constants.RoleScout,
			Slug:        "scout",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Talent finder. Freemium access to athlete profiles and leaderboard.",
		},
		{
			Name:        constants.RoleCourtOwner,
			Slug:        "court-owner",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Manages courts. Requires admin verification.",
		},
		{
			Name:        constants.RoleAcademyAdmin,
			Slug:        "academy-admin",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Manages academies. Requires admin verification.",
		},
		{
			Name:        constants.RoleRegulator,
			Slug:        "regulator",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Official sport body. Assigned by platform admin only.",
		},
		{
			Name:        constants.RolePlatformAdmin,
			Slug:        "platform-admin",
			CreatedByID: createdBy,
			UpdatedByID: createdBy,
			Description: "Internal platform administrator.",
		},
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

func seedRolePermissions(db *gorm.DB) error {
	rolePermissionMap := map[string][]string{
		constants.RoleAthlete: {
			"user.read", "user.update",
			"athlete.read", "athlete.update",
			"academy.read",
			"court.read",
			"booking.read", "booking.create", "booking.update", "booking.delete",
			"competition.read", "competition.join",
			"leaderboard.read",
		},
		constants.RoleScout: {
			"user.read",
			"athlete.read", "athlete.read_premium",
			"leaderboard.read", "leaderboard.read_premium",
		},
		constants.RoleCourtOwner: {
			"user.read", "user.update",
			"court.read", "court.create", "court.update", "court.delete", "court.manage",
			"booking.read", "booking.update",
		},
		constants.RoleAcademyAdmin: {
			"user.read", "user.update",
			"academy.read", "academy.create", "academy.update", "academy.delete", "academy.manage",
			"athlete.read",
			"competition.read", "competition.create",
		},
		constants.RoleRegulator: {
			"user.read",
			"athlete.read", "athlete.read_premium",
			"competition.read", "competition.regulate",
			"leaderboard.read", "leaderboard.read_premium",
			"verification.approve", "verification.reject",
		},
		constants.RolePlatformAdmin: {
			// Full access - all permissions
			"user.read", "user.create", "user.update", "user.delete", "user.verify",
			"role.read", "role.create", "role.update", "role.delete", "role.assign",
			"permission.read", "permission.create", "permission.update", "permission.delete",
			"athlete.read", "athlete.read_premium", "athlete.update",
			"academy.read", "academy.create", "academy.update", "academy.delete", "academy.manage",
			"court.read", "court.create", "court.update", "court.delete", "court.manage",
			"booking.read", "booking.create", "booking.update", "booking.delete",
			"competition.read", "competition.create", "competition.update", "competition.delete", "competition.regulate",
			"leaderboard.read", "leaderboard.read_premium",
			"verification.approve", "verification.reject",
			"platform.admin", "platform.analytics",
		},
	}

	for roleName, permSlugs := range rolePermissionMap {
		var role entity.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			continue
		}

		for _, permSlug := range permSlugs {
			var permission entity.Permission
			if err := db.Where("slug = ?", permSlug).First(&permission).Error; err != nil {
				continue
			}

			var count int64
			db.Table("role_permissions").
				Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).
				Count(&count)

			if count == 0 {
				if err := db.Model(&role).Association("Permissions").Append(&permission); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
