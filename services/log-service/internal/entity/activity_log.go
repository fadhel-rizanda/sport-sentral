package entity

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_activity_user" json:"user_id"`
	Action       string    `gorm:"type:varchar(100);not null;index:idx_activity_action" json:"action"`
	ResourceType string    `gorm:"type:varchar(100);not null;index:idx_activity_resource" json:"resource_type"`
	ResourceID   *string   `gorm:"type:varchar(255);index:idx_activity_resource" json:"resource_id,omitempty"`
	Description  string    `gorm:"type:text;not null" json:"description"`
	Metadata     *string   `gorm:"type:text" json:"metadata,omitempty"`
	IPAddress    *string   `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    *string   `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime;index:idx_activity_created_at" json:"created_at"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}
