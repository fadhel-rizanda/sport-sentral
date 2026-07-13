package repository

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VenueFilters struct {
	OwnerID  *uuid.UUID
	StatusID *uuid.UUID
	Search   *string
}

type VenueRepository interface {
	Create(ctx context.Context, venue *entity.Venue) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Venue, error)
	List(ctx context.Context, filters VenueFilters, page, pageSize int) ([]*entity.Venue, int64, error)
	Update(ctx context.Context, venue *entity.Venue) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type venueRepository struct {
	db *gorm.DB
}

func NewVenueRepository(db *gorm.DB) VenueRepository {
	return &venueRepository{db: db}
}

func (r *venueRepository) Create(ctx context.Context, venue *entity.Venue) error {
	return r.db.WithContext(ctx).Create(venue).Error
}

func (r *venueRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Venue, error) {
	var venue entity.Venue
	err := r.db.WithContext(ctx).
		Preload("AdministrativeDivision.Country").
		Preload("Owner").
		Preload("Status").
		First(&venue, "id = ? AND deleted_at IS NULL", id).Error
	return &venue, err
}

func (r *venueRepository) List(ctx context.Context, filters VenueFilters, page, pageSize int) ([]*entity.Venue, int64, error) {
	var venues []*entity.Venue
	var count int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Venue{}).Where("deleted_at IS NULL")

	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", filters.OwnerID)
	}
	if filters.StatusID != nil {
		query = query.Where("status_id = ?", filters.StatusID)
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+*filters.Search+"%", "%"+*filters.Search+"%")
	}

	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("AdministrativeDivision.Country").
		Preload("Owner").
		Preload("Status").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&venues).Error

	return venues, count, err
}

func (r *venueRepository) Update(ctx context.Context, venue *entity.Venue) error {
	return r.db.WithContext(ctx).Save(venue).Error
}

func (r *venueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var venue entity.Venue
		if err := tx.First(&venue, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
			return err
		}
		return tx.Delete(&venue).Error
	})
}
