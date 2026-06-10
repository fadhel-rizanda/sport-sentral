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
	)
	if err != nil {
		return err
	}

	log.Println("Core sport service tables migrated successfully")
	return nil
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sport_status_tier 
		ON replicated_sports(status_id, tier_tag_id) WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sport_config_sport 
		ON sport_configs(sport_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_regulator_status 
		ON regulators(status_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sport_reg_active 
		ON sport_regulator_assignments(sport_id, is_active) 
		WHERE is_active = true
	`).Error; err != nil {
		return err
	}

	return nil
}
