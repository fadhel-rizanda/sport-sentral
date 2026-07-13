package database

import (
	"errors"
	"fmt"
	"log"
	"microservice-golang/services/venue-service/internal/entity"
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

	log.Println("Replicated tables migrated successfully")

	err = db.AutoMigrate(
		&entity.Venue{},
		&entity.Court{},
		&entity.Booking{},
		&entity.CourtSlot{},
	)
	if err != nil {
		return err
	}

	log.Println("Core court service tables migrated successfully")
	return nil
}

func CreateIndexes(db *gorm.DB) error {
	// Courts Indexes
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_courts_sport_status 
		ON courts(sport_id, status_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	// Unique name per venue
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_courts_name_venue 
		ON courts(name, venue_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	// Court Slots Indexes to prevent overlaps
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_slots_court_times 
		ON court_slots(court_id, start_time, end_time) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	// Index for "show me available slots in this range"
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_slots_court_start 
		ON court_slots(court_id, start_time) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	// Bookings Indexes for "my bookings" screens
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_bookings_user_created 
		ON court_bookings(user_id, created_at DESC) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_bookings_user_status 
		ON court_bookings(user_id, status_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
