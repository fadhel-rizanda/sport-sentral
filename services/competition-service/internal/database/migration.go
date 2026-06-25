package database

import (
	"log"
	"microservice-golang/services/competition-service/internal/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// Replicated tables
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Sport{},
		&entity.SportStat{},
		&entity.Status{},
		&entity.Tag{},
		&entity.AcademyHolding{},
		&entity.AcademyBranch{},
		&entity.Role{},
		&entity.Permission{},
		&entity.AcademyAdmin{},
	)
	if err != nil {
		return err
	}

	log.Println("Replicated tables migrated successfully")

	// Core tables
	err = db.AutoMigrate(
		&entity.Competition{},
		&entity.CompetitionBranch{},
		&entity.CompetitionAdmin{},
		&entity.Match{},
		&entity.MatchParticipant{},
		&entity.MatchStat{},
		&entity.AthleteStatsAggregate{},
		&entity.Roster{},
		&entity.RosterMember{},
	)
	if err != nil {
		return err
	}

	log.Println("Core competition tables migrated successfully")
	return nil
}
func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_roster_athlete 
		ON roster_members(roster_id, athlete_id) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_match_athlete_stat 
		ON match_stats(match_id, athlete_id, stat_type_id) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_comp_athlete_stat 
		ON athlete_stats_aggregates(competition_id, branch_id, athlete_id, stat_type_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_match_roster 
		ON match_participants(match_id, roster_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_competition_name 
		ON competitions(name) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_name 
		ON competition_branches(competition_id, name) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}
