package database

import (
	"errors"
	"fmt"
	"log"
	"microservice-golang/services/log-service/internal/entity"
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
		&entity.AuditLog{},
		&entity.ActivityLog{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_audit_created_at_desc
		ON audit_logs(created_at DESC);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_activity_created_at_desc
		ON activity_logs(created_at DESC);
	`).Error; err != nil {
		return err
	}

	return nil
}
