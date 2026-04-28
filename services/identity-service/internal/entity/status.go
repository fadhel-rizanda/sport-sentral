package entity

import "github.com/google/uuid"

type Status struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Type string    `gorm:"not null"`
	Name string    `gorm:"not null"`
	// unique index on (type, name) di migration
}

const (
	StatusTypeUser     = "user"
	StatusTypeUserRole = "user_role"
)

const (
	UserStatusActive  = "active"
	UserStatusPending = "pending"
	UserStatusBanned  = "banned"
)

const (
	UserRoleStatusActive  = "active"
	UserRoleStatusPending = "pending"
)
