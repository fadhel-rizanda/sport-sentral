package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
)

type StatusRepository interface {
	GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error)
}

type statusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{db: db}
}

func (r *statusRepository) GetByTypeAndName(ctx context.Context, statusType, name string) (*entity.Status, error) {
	var status entity.Status
	err := r.db.WithContext(ctx).
		First(&status, "type = ? AND name = ?", statusType, name).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Status, error) {
	var status entity.Status
	err := r.db.WithContext(ctx).First(&status, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}
