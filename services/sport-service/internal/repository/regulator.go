package repository

import (
	"context"
	"microservice-golang/services/sport-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegulatorRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Regulator, error)
	Create(ctx context.Context, regulator *entity.Regulator) error
	Update(ctx context.Context, regulator *entity.Regulator) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Staff Management
	GetStaffByID(ctx context.Context, id uuid.UUID) (*entity.RegulatorStaff, error)
	AddStaff(ctx context.Context, staff *entity.RegulatorStaff) error
	RemoveStaff(ctx context.Context, staffID uuid.UUID) error
}

type regulatorRepository struct {
	db *gorm.DB
}

func NewRegulatorRepository(db *gorm.DB) RegulatorRepository {
	return &regulatorRepository{
		db: db,
	}
}

func (r *regulatorRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Regulator, error) {
	var regulator entity.Regulator
	err := r.db.WithContext(ctx).
		Preload("Status").
		Preload("Address.AdministrativeDivision.Country").
		Preload("Staff.User").
		Preload("Staff.RoleTag").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&regulator, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &regulator, nil
}

func (r *regulatorRepository) Create(ctx context.Context, regulator *entity.Regulator) error {
	return r.db.WithContext(ctx).Create(regulator).Error
}

func (r *regulatorRepository) Update(ctx context.Context, regulator *entity.Regulator) error {
	return r.db.WithContext(ctx).Save(regulator).Error
}

func (r *regulatorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Regulator{}, "id = ?", id).Error
}

func (r *regulatorRepository) GetStaffByID(ctx context.Context, id uuid.UUID) (*entity.RegulatorStaff, error) {
	var staff entity.RegulatorStaff
	err := r.db.WithContext(ctx).
		Preload("Regulator").
		Preload("User").
		Preload("RoleTag").
		First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *regulatorRepository) AddStaff(ctx context.Context, staff *entity.RegulatorStaff) error {
	return r.db.WithContext(ctx).Create(staff).Error
}

func (r *regulatorRepository) RemoveStaff(ctx context.Context, staffID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.RegulatorStaff{}, "id = ?", staffID).Error
}
