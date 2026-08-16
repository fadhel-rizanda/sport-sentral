package database

import (
	"microservice-golang/services/attachment-service/internal/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Attachment{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_attachments_status
		ON attachments(status);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_attachments_file_category
		ON attachments(file_category);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by_user_id
		ON attachments(uploaded_by_user_id);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_attachments_created_at_desc
		ON attachments(created_at DESC);
	`).Error; err != nil {
		return err
	}

	return nil
}
