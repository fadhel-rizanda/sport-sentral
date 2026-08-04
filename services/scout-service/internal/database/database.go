package database

import (
	"fmt"
	"log/slog"

	"microservice-golang/services/scout-service/internal/config"
	"microservice-golang/services/scout-service/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.GormDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect db failed: %w", err)
	}

	slog.Info("Connected to PostgreSQL for scout-service", "database", cfg.Database.Name)

	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Scout{},
		&entity.AthleteProfile{},
		&entity.WatchlistEntry{},
		&entity.LeaderboardEntry{},
		&entity.ScoutActivityLog{},
	)
}
