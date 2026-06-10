package database

import (
	"errors"
	"fmt"
	"log"
	"microservice-golang/services/meta-service/internal/entity"
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
	return db.AutoMigrate(
		&entity.Status{},
		&entity.Tag{},
		&entity.Country{},
		&entity.AdministrativeDivision{},
		&entity.User{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_countries_iso_alpha2 
		ON countries(iso_alpha2) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_countries_iso_alpha3 
		ON countries(iso_alpha3) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_statuses_slug 
		ON statuses(slug) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_tags_name 
		ON tags(name) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_tags_slug 
		ON tags(slug) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
