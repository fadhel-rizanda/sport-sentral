package repository

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourtRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Court, error)
	List(ctx context.Context, filters CourtFilters, page, pageSize int) ([]*entity.Court, int64, error)
	Create(ctx context.Context, court *entity.Court) error
	Update(ctx context.Context, court *entity.Court) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CourtFilters struct {
	VenueID  *uuid.UUID
	SportID  *uuid.UUID
	OwnerID  *uuid.UUID
	StatusID *uuid.UUID
	Search   *string
}

type courtRepository struct {
	db *gorm.DB
}

func NewCourtRepository(db *gorm.DB) CourtRepository {
	return &courtRepository{
		db: db,
	}
}

func (r *courtRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Court, error) {
	var court entity.Court
	err := r.db.WithContext(ctx).
		Preload("Sport").
		Preload("Status").
		Preload("Venue.AdministrativeDivision.Country").
		Preload("Venue.Owner").
		First(&court, "id = ? AND deleted_at IS NULL", id).Error
	return &court, err
}

func (r *courtRepository) List(ctx context.Context, filters CourtFilters, page, pageSize int) ([]*entity.Court, int64, error) {
	var courts []*entity.Court
	var count int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Court{}).Where("deleted_at IS NULL")

	if filters.VenueID != nil {
		query = query.Where("venue_id = ?", filters.VenueID)
	}
	if filters.SportID != nil {
		query = query.Where("sport_id = ?", filters.SportID)
	}
	if filters.OwnerID != nil {
		// Filter through Venue owner
		query = query.Joins("JOIN venues ON venues.id = courts.venue_id").Where("venues.owner_id = ?", filters.OwnerID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where("name ILIKE ?", "%"+*filters.Search+"%")
	}

	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Sport").
		Preload("Status").
		Preload("Venue.AdministrativeDivision.Country").
		Preload("Venue.Owner").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&courts).Error

	return courts, count, err
}

func (r *courtRepository) Create(ctx context.Context, court *entity.Court) error {
	return r.db.WithContext(ctx).Create(court).Error
}

func (r *courtRepository) Update(ctx context.Context, court *entity.Court) error {
	return r.db.WithContext(ctx).Save(court).Error
}

func (r *courtRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var court entity.Court
		if err := tx.First(&court, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
			return err
		}
		return tx.Delete(&court).Error
	})
}
