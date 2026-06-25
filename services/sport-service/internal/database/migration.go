package database

import (
	"log"
	"microservice-golang/services/sport-service/internal/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Status{},
		&entity.Tag{},
	)
	if err != nil {
		return err
	}

	log.Println("Replicated tables migrated successfully")

	err = db.AutoMigrate(
		&entity.RegulatorAddress{},
		&entity.Regulator{},
		&entity.RegulatorStaff{},
		&entity.Sport{},
		&entity.SportConfig{},
		&entity.SportStat{},
	)
	if err != nil {
		return err
	}

	log.Println("Core sport service tables migrated successfully")
	return nil
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sports_status_tier 
		ON sports(status_id, tier_tag_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_sports_name 
		ON sports(name) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_sports_slug 
		ON sports(slug) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_sport_configs_sport_id 
		ON sport_configs(sport_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uni_regulators_code 
		ON regulators(code) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
