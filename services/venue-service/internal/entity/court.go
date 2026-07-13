package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Venue struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name                     string         `gorm:"type:varchar(255);not null"`
	Description              string         `gorm:"type:text"`
	StreetAddress            string         `gorm:"type:text;not null"`
	Notes                    *string        `gorm:"type:text"`
	Latitude                 *float64       `gorm:"type:decimal(10,8)"`
	Longitude                *float64       `gorm:"type:decimal(11,8)"`
	AdministrativeDivisionID uuid.UUID      `gorm:"type:uuid;not null;index"`
	OwnerID                  uuid.UUID      `gorm:"type:uuid;not null;index"`
	StatusID                 uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt                time.Time      `gorm:"not null"`
	UpdatedAt                time.Time      `gorm:"not null"`
	DeletedAt                gorm.DeletedAt `gorm:"index"`

	// Relations
	AdministrativeDivision *AdministrativeDivision `gorm:"foreignKey:AdministrativeDivisionID;references:ID;constraint:-;"`
	Owner                  *User                   `gorm:"foreignKey:OwnerID;references:ID;constraint:-;"`
	Status                 *Status                 `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	Courts                 []Court                 `gorm:"foreignKey:VenueID;references:ID"`
}

func (Venue) TableName() string {
	return "venues"
}

// BeforeDelete hook to cascade soft-delete to child courts
func (v *Venue) BeforeDelete(tx *gorm.DB) (err error) {
	var courts []Court
	if err := tx.Where("venue_id = ? AND deleted_at IS NULL", v.ID).Find(&courts).Error; err != nil {
		return err
	}
	for _, court := range courts {
		if err := tx.Delete(&court).Error; err != nil {
			return err
		}
	}
	return nil
}

type Court struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey"`
	VenueID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name              string         `gorm:"type:varchar(255);not null"`
	Description       string         `gorm:"type:text"`
	SportID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	PricePerHour      int64          `gorm:"type:bigint;not null"` // in cents
	ImageAttachmentID *uuid.UUID     `gorm:"type:uuid"`
	StatusID          uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt         time.Time      `gorm:"not null"`
	UpdatedAt         time.Time      `gorm:"not null"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`

	// Relations
	Venue  *Venue  `gorm:"foreignKey:VenueID;references:ID;constraint:OnDelete:CASCADE"`
	Sport  *Sport  `gorm:"foreignKey:SportID;references:ID;constraint:-;"`
	Status *Status `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
}

func (Court) TableName() string {
	return "courts"
}

func (c *Court) BeforeDelete(tx *gorm.DB) (err error) {
	now := time.Now()
	if err := tx.Model(&CourtSlot{}).Where("court_id = ? AND deleted_at IS NULL", c.ID).Update("deleted_at", now).Error; err != nil {
		return err
	}
	var bookings []Booking
	if err := tx.Where("court_id = ? AND deleted_at IS NULL", c.ID).Find(&bookings).Error; err != nil {
		return err
	}
	for _, booking := range bookings {
		if err := tx.Delete(&booking).Error; err != nil {
			return err
		}
	}
	return nil
}

type CourtSlot struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CourtID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	StartTime time.Time      `gorm:"not null"`
	EndTime   time.Time      `gorm:"not null;check:end_time > start_time"`
	Price     int64          `gorm:"type:bigint;not null"` // in cents
	StatusID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	BookingID *uuid.UUID     `gorm:"type:uuid;index"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Court  *Court  `gorm:"foreignKey:CourtID;references:ID;constraint:-;"`
	Status *Status `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
}

func (CourtSlot) TableName() string {
	return "court_slots"
}

type Booking struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CourtID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	TotalPrice    int64          `gorm:"type:bigint;not null"` // in cents
	StatusID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	PaymentStatus string         `gorm:"type:varchar(50);not null;default:'UNPAID'"` // UNPAID, PAID
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Court  *Court      `gorm:"foreignKey:CourtID;references:ID;constraint:-;"`
	User   *User       `gorm:"foreignKey:UserID;references:ID;constraint:-;"`
	Status *Status     `gorm:"foreignKey:StatusID;references:ID;constraint:-;"`
	Slots  []CourtSlot `gorm:"foreignKey:BookingID;references:ID"`
}

func (Booking) TableName() string {
	return "court_bookings"
}

func (b *Booking) BeforeDelete(tx *gorm.DB) (err error) {
	var status Status
	if err := tx.First(&status, "type = ? AND slug = ?", "booking_slot", "available").Error; err != nil {
		if err := tx.First(&status, "type = ? AND slug = ?", "slot", "available").Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&CourtSlot{}).
		Where("booking_id = ? AND deleted_at IS NULL", b.ID).
		Updates(map[string]interface{}{
			"booking_id": nil,
			"status_id":  status.ID,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}
	return nil
}
