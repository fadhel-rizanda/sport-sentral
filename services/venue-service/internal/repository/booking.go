package repository

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/shared/infrastructure/postgres"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	List(ctx context.Context, filters BookingFilters, page, pageSize int) ([]*entity.Booking, int64, error)
	Create(ctx context.Context, booking *entity.Booking) error
	Update(ctx context.Context, booking *entity.Booking) error
}

type BookingFilters struct {
	CourtID  *uuid.UUID
	UserID   *uuid.UUID
	StatusID *uuid.UUID
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{
		db: db,
	}
}

func (r *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error) {
	var booking entity.Booking
	err := postgres.GetTx(ctx, r.db).
		Preload("Court.Sport").
		Preload("Court.Owner").
		Preload("Court.Venue.AdministrativeDivision.Country").
		Preload("User").
		Preload("Status").
		Preload("Slots.Status").
		First(&booking, "id = ? AND deleted_at IS NULL", id).Error
	return &booking, err
}

func (r *bookingRepository) List(ctx context.Context, filters BookingFilters, page, pageSize int) ([]*entity.Booking, int64, error) {
	var bookings []*entity.Booking
	var count int64
	offset := pageSize * (page - 1)

	query := postgres.GetTx(ctx, r.db).Model(&entity.Booking{}).Where("deleted_at IS NULL")

	if filters.CourtID != nil {
		query = query.Where("court_id = ?", filters.CourtID)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}

	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Court.Sport").
		Preload("Court.Owner").
		Preload("Court.Venue.AdministrativeDivision.Country").
		Preload("User").
		Preload("Status").
		Preload("Slots.Status").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&bookings).Error

	return bookings, count, err
}

func (r *bookingRepository) Create(ctx context.Context, booking *entity.Booking) error {
	return postgres.GetTx(ctx, r.db).Create(booking).Error
}

func (r *bookingRepository) Update(ctx context.Context, booking *entity.Booking) error {
	return postgres.GetTx(ctx, r.db).Save(booking).Error
}
