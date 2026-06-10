package database

import (
	"errors"
	"fmt"
	"log"
	"microservice-golang/services/identity-service/internal/entity"
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
		&entity.User{},
		&entity.Permission{},
		&entity.Role{},
		&entity.UserRole{},
		&entity.Status{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_permissions_slug 
		ON permissions(slug) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name 
		ON roles(name) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_roles_slug 
		ON roles(slug) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_permission_ids 
		ON role_permissions(role_id, permission_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email 
		ON users(email) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username 
		ON users(username) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
