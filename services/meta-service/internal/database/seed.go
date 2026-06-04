package database

import (
	"context"
	"errors"
	"fmt"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/usecase"
	"microservice-golang/shared/pkg/constants"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	//return seedStatusesDB(db)
	return nil
}

func SeedStatuses(uc usecase.StatusUseCase) error {
	createdBy, err := uuid.Parse("389d7e0d-4bdd-4ecc-9d7b-88ad8a8055db")
	if err != nil {
		return err
	}

	data := []entity.Status{
		// user
		{Type: constants.StatusTypeUser, Name: constants.StatusActive, Slug: "user-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeUser, Name: constants.StatusPending, Slug: "user-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeUser, Name: constants.StatusBanned, Slug: "user-banned", CreatedByID: createdBy, UpdatedByID: createdBy},

		// user_role
		{Type: constants.StatusTypeUserRole, Name: constants.StatusActive, Slug: "user-role-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeUserRole, Name: constants.StatusPending, Slug: "user-role-pending", CreatedByID: createdBy, UpdatedByID: createdBy},

		// academy
		{Type: constants.StatusTypeAcademy, Name: constants.StatusActive, Slug: "academy-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusPending, Slug: "academy-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeAcademy, Name: constants.StatusSuspended, Slug: "academy-suspended", CreatedByID: createdBy, UpdatedByID: createdBy},

		// academy_member
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusActive, Slug: "academy-member-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusPending, Slug: "academy-member-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusRejected, Slug: "academy-member-rejected", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeAcademyMember, Name: constants.StatusInactive, Slug: "academy-member-inactive", CreatedByID: createdBy, UpdatedByID: createdBy},

		// court
		{Type: constants.StatusTypeCourt, Name: constants.StatusActive, Slug: "court-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCourt, Name: constants.StatusPending, Slug: "court-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCourt, Name: constants.StatusSuspended, Slug: "court-suspended", CreatedByID: createdBy, UpdatedByID: createdBy},

		// court_booking
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusBooked, Slug: "court-booking-booked", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCancelled, Slug: "court-booking-cancelled", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCourtBooking, Name: constants.StatusCompleted, Slug: "court-booking-completed", CreatedByID: createdBy, UpdatedByID: createdBy},

		// competition
		{Type: constants.StatusTypeCompetition, Name: constants.StatusActive, Slug: "competition-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusPending, Slug: "competition-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCompleted, Slug: "competition-completed", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeCompetition, Name: constants.StatusCancelled, Slug: "competition-cancelled", CreatedByID: createdBy, UpdatedByID: createdBy},

		// event
		{Type: constants.StatusTypeEvent, Name: constants.StatusActive, Slug: "event-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeEvent, Name: constants.StatusPending, Slug: "event-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCompleted, Slug: "event-completed", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeEvent, Name: constants.StatusCancelled, Slug: "event-cancelled", CreatedByID: createdBy, UpdatedByID: createdBy},

		// participant
		{Type: constants.StatusTypeParticipant, Name: constants.StatusActive, Slug: "participant-active", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusPending, Slug: "participant-pending", CreatedByID: createdBy, UpdatedByID: createdBy},
		{Type: constants.StatusTypeParticipant, Name: constants.StatusRejected, Slug: "participant-rejected", CreatedByID: createdBy, UpdatedByID: createdBy},
	}

	for _, s := range data {
		ctx := context.Background()
		_, err := uc.GetByTypeAndName(ctx, s.Type, s.Name)

		if err != nil {
			slug := fmt.Sprintf("%s-%s", s.Type, s.Name)

			_, err := uc.Create(ctx, dto.CreateStatusRequest{
				Type:        s.Type,
				Name:        s.Name,
				Slug:        slug,
				CreatedByID: uuid.MustParse("389d7e0d-4bdd-4ecc-9d7b-88ad8a8055db"),
			})
			if err != nil {
				return err
			}
		}
	}
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
			s.Slug = fmt.Sprintf("%s-%s", s.Type, s.Name)
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
