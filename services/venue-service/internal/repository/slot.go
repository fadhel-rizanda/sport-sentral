package repository

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/shared/infrastructure/postgres"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CourtSlotRepository interface {
	CreateBulk(ctx context.Context, slots []*entity.CourtSlot) error
	List(ctx context.Context, courtID uuid.UUID, statusID *uuid.UUID, startDate, endDate *time.Time) ([]*entity.CourtSlot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CourtSlot, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CourtSlot, error)
	GetByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*entity.CourtSlot, error)
	Update(ctx context.Context, slot *entity.CourtSlot) error
	UpdateStatusBulk(ctx context.Context, ids []uuid.UUID, statusID uuid.UUID, bookingID *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type courtSlotRepository struct {
	db *gorm.DB
}

func NewCourtSlotRepository(db *gorm.DB) CourtSlotRepository {
	return &courtSlotRepository{
		db: db,
	}
}

func (r *courtSlotRepository) CreateBulk(ctx context.Context, slots []*entity.CourtSlot) error {
	return postgres.GetTx(ctx, r.db).Create(slots).Error
}

func (r *courtSlotRepository) List(ctx context.Context, courtID uuid.UUID, statusID *uuid.UUID, startDate, endDate *time.Time) ([]*entity.CourtSlot, error) {
	var slots []*entity.CourtSlot
	query := postgres.GetTx(ctx, r.db).Where("court_id = ?", courtID)

	if statusID != nil {
		query = query.Where("status_id = ?", *statusID)
	}
	if startDate != nil {
		query = query.Where("start_time >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("end_time <= ?", *endDate)
	}

	err := query.Preload("Status").Order("start_time asc").Find(&slots).Error
	return slots, err
}

func (r *courtSlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CourtSlot, error) {
	var slot entity.CourtSlot
	err := postgres.GetTx(ctx, r.db).Preload("Status").First(&slot, "id = ?", id).Error
	return &slot, err
}

func (r *courtSlotRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CourtSlot, error) {
	var slots []*entity.CourtSlot
	err := postgres.GetTx(ctx, r.db).Preload("Status").Where("id IN ?", ids).Find(&slots).Error
	return slots, err
}

func (r *courtSlotRepository) GetByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*entity.CourtSlot, error) {
	var slots []*entity.CourtSlot
	err := postgres.GetTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Status").
		Where("id IN ?", ids).
		Find(&slots).Error
	return slots, err
}

func (r *courtSlotRepository) Update(ctx context.Context, slot *entity.CourtSlot) error {
	return postgres.GetTx(ctx, r.db).Save(slot).Error
}

func (r *courtSlotRepository) UpdateStatusBulk(ctx context.Context, ids []uuid.UUID, statusID uuid.UUID, bookingID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status_id":  statusID,
		"updated_at": time.Now(),
		"booking_id": bookingID,
	}

	query := postgres.GetTx(ctx, r.db).Model(&entity.CourtSlot{}).Where("id IN ?", ids)
	if bookingID != nil {
		query = query.Where("booking_id IS NULL")
	}

	res := query.Updates(updates)
	if res.Error != nil {
		return res.Error
	}

	if bookingID != nil && res.RowsAffected < int64(len(ids)) {
		return gorm.ErrDuplicatedKey
	}

	return nil
}

func (r *courtSlotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&entity.CourtSlot{}, "id = ?", id).Error
}
