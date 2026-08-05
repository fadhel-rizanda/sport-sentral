package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusPending = "PENDING"

	SeverityInfo     = "INFO"
	SeverityWarn     = "WARN"
	SeverityError    = "ERROR"
	SeverityCritical = "CRITICAL"

	ServiceNameNatsBus     = "nats-event-bus"
	ModuleEventSync        = "event-sync"
	EntityTypeEventPayload = "EventPayload"
)

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ServiceName  string     `gorm:"type:varchar(100);not null;index:idx_audit_service_action" json:"service_name"`
	Module       string     `gorm:"type:varchar(100);not null" json:"module"`
	Action       string     `gorm:"type:varchar(100);not null;index:idx_audit_service_action" json:"action"`
	EntityType   string     `gorm:"type:varchar(100);not null;index:idx_audit_entity" json:"entity_type"`
	EntityID     *string    `gorm:"type:varchar(255);index:idx_audit_entity" json:"entity_id,omitempty"`
	UserID       *uuid.UUID `gorm:"type:uuid;index:idx_audit_user" json:"user_id,omitempty"`
	UserEmail    *string    `gorm:"type:varchar(255)" json:"user_email,omitempty"`
	UserRole     *string    `gorm:"type:varchar(100)" json:"user_role,omitempty"`
	IPAddress    *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    *string    `gorm:"type:text" json:"user_agent,omitempty"`
	Status       string     `gorm:"type:varchar(50);not null;default:'SUCCESS';index:idx_audit_status" json:"status"`
	Severity     string     `gorm:"type:varchar(50);not null;default:'INFO';index:idx_audit_severity" json:"severity"`
	OldValue     *string    `gorm:"type:text" json:"old_value,omitempty"`
	NewValue     *string    `gorm:"type:text" json:"new_value,omitempty"`
	Metadata     *string    `gorm:"type:text" json:"metadata,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs   int64      `gorm:"type:bigint;default:0" json:"duration_ms"`
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime;index:idx_audit_created_at" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
