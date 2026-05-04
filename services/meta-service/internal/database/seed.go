package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
)

func Seed(db *gorm.DB) error {
	return seedStatuses(db)
}

func seedStatuses(db *gorm.DB) error {
	statuses := []entity.Status{
		// user
		{Type: entity.StatusTypeUser, Name: entity.StatusActive},
		{Type: entity.StatusTypeUser, Name: entity.StatusPending},
		{Type: entity.StatusTypeUser, Name: entity.StatusBanned},

		// user_role
		{Type: entity.StatusTypeUserRole, Name: entity.StatusActive},
		{Type: entity.StatusTypeUserRole, Name: entity.StatusPending},

		// academy
		{Type: entity.StatusTypeAcademy, Name: entity.StatusActive},
		{Type: entity.StatusTypeAcademy, Name: entity.StatusPending},
		{Type: entity.StatusTypeAcademy, Name: entity.StatusSuspended},

		// academy_member
		{Type: entity.StatusTypeAcademyMember, Name: entity.StatusActive},
		{Type: entity.StatusTypeAcademyMember, Name: entity.StatusPending},
		{Type: entity.StatusTypeAcademyMember, Name: entity.StatusRejected},
		{Type: entity.StatusTypeAcademyMember, Name: entity.StatusInactive},

		// court
		{Type: entity.StatusTypeCourt, Name: entity.StatusActive},
		{Type: entity.StatusTypeCourt, Name: entity.StatusPending},
		{Type: entity.StatusTypeCourt, Name: entity.StatusSuspended},

		// court_booking
		{Type: entity.StatusTypeCourtBooking, Name: entity.StatusBooked},
		{Type: entity.StatusTypeCourtBooking, Name: entity.StatusCancelled},
		{Type: entity.StatusTypeCourtBooking, Name: entity.StatusCompleted},

		// competition
		{Type: entity.StatusTypeCompetition, Name: entity.StatusActive},
		{Type: entity.StatusTypeCompetition, Name: entity.StatusPending},
		{Type: entity.StatusTypeCompetition, Name: entity.StatusCompleted},
		{Type: entity.StatusTypeCompetition, Name: entity.StatusCancelled},

		// event
		{Type: entity.StatusTypeEvent, Name: entity.StatusActive},
		{Type: entity.StatusTypeEvent, Name: entity.StatusPending},
		{Type: entity.StatusTypeEvent, Name: entity.StatusCompleted},
		{Type: entity.StatusTypeEvent, Name: entity.StatusCancelled},

		// participant
		{Type: entity.StatusTypeParticipant, Name: entity.StatusActive},
		{Type: entity.StatusTypeParticipant, Name: entity.StatusPending},
		{Type: entity.StatusTypeParticipant, Name: entity.StatusRejected},
	}

	for _, s := range statuses {
		var existing entity.Status
		err := db.Where("type = ? AND name = ?", s.Type, s.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.ID = uuid.New()
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
