package database

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/usecase"
	"microservice-golang/shared/pkg/constants"
)

func SeedStatuses(uc usecase.StatusUseCase) error {
	data := []entity.Status{
		// user
		{Type: constants.StatusTypeUser, Name: constants.StatusActive},
		{Type: constants.StatusTypeUser, Name: constants.StatusPending},
		{Type: constants.StatusTypeUser, Name: constants.StatusBanned},

		// user_role
		{Type: constants.StatusTypeUserRole, Name: constants.StatusActive},
		{Type: constants.StatusTypeUserRole, Name: constants.StatusPending},

		// academy
		{Type: constants.StatusTypeAcademy, Name: constants.StatusActive},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusPending},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusSuspended},

		// academy_member
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusActive},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusPending},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusRejected},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusInactive},

		// court
		{Type: constants.StatusTypeCourt, Name: constants.StatusActive},
		{Type: constants.StatusTypeCourt, Name: constants.StatusPending},
		{Type: constants.StatusTypeCourt, Name: constants.StatusSuspended},

		// court_booking
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusBooked},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCancelled},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCompleted},

		// competition
		{Type: constants.StatusTypeCompetition, Name: constants.StatusActive},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusPending},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCompleted},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCancelled},

		// event
		{Type: constants.StatusTypeEvent, Name: constants.StatusActive},
		{Type: constants.StatusTypeEvent, Name: constants.StatusPending},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCompleted},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCancelled},

		// participant
		{Type: constants.StatusTypeParticipant, Name: constants.StatusActive},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusPending},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusRejected},
	}

	for _, s := range data {
		ctx := context.Background()
		_, err := uc.GetByTypeAndName(ctx, s.Type, s.Name)

		if err != nil {
			_, err := uc.Create(ctx, usecase.CreateStatusRequest{
				Type:        s.Type,
				Name:        s.Name,
				CreatedByID: uuid.MustParse("00000000-0000-0000-0000-000000000000"), // System ID
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Seed(db *gorm.DB) error {
	//return seedStatusesDB(db)
	return nil
}

func seedStatusesDB(db *gorm.DB) error {
	statuses := []entity.Status{
		// user
		{Type: constants.StatusTypeUser, Name: constants.StatusActive},
		{Type: constants.StatusTypeUser, Name: constants.StatusPending},
		{Type: constants.StatusTypeUser, Name: constants.StatusBanned},

		// user_role
		{Type: constants.StatusTypeUserRole, Name: constants.StatusActive},
		{Type: constants.StatusTypeUserRole, Name: constants.StatusPending},

		// academy
		{Type: constants.StatusTypeAcademy, Name: constants.StatusActive},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusPending},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusSuspended},

		// academy_member
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusActive},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusPending},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusRejected},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusInactive},

		// court
		{Type: constants.StatusTypeCourt, Name: constants.StatusActive},
		{Type: constants.StatusTypeCourt, Name: constants.StatusPending},
		{Type: constants.StatusTypeCourt, Name: constants.StatusSuspended},

		// court_booking
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusBooked},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCancelled},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCompleted},

		// competition
		{Type: constants.StatusTypeCompetition, Name: constants.StatusActive},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusPending},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCompleted},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCancelled},

		// event
		{Type: constants.StatusTypeEvent, Name: constants.StatusActive},
		{Type: constants.StatusTypeEvent, Name: constants.StatusPending},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCompleted},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCancelled},

		// participant
		{Type: constants.StatusTypeParticipant, Name: constants.StatusActive},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusPending},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusRejected},
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
