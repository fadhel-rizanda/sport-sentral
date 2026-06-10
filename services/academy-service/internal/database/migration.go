package database

import (
	"errors"
	"fmt"
	"log"
	"microservice-golang/services/academy-service/internal/entity"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

func RunExternalMigrations(dsn string) error {
	migrationPath := "db/migrations"

	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		log.Printf("Warning: Migration folder '%s' not found. Skipping external migrations.", migrationPath)
		return nil
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not run up migrations: %v", err)
	}

	log.Println("External SQL migrations ran successfully")
	return nil
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Sport{},
		&entity.Status{},
		&entity.Tag{},
		&entity.Role{},
		&entity.Permission{},
		&entity.Country{},
		&entity.AdministrativeDivision{},
	)
	if err != nil {
		return err
	}

	return db.AutoMigrate(
		&entity.AcademyHoldingAddress{},
		&entity.AcademyHolding{},
		&entity.AcademyBranch{},
		&entity.AcademyBranchAddress{},
		&entity.AcademyAdmin{},
		&entity.Enrollment{},
		&entity.Roster{},
		&entity.RosterMember{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_academy_holdings_name 
		ON academy_holdings(name) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_academy_athlete 
		ON enrollments(academy_branch_id, athlete_id) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_roster_athlete 
		ON roster_members(roster_id, athlete_id) 
		WHERE removed_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
