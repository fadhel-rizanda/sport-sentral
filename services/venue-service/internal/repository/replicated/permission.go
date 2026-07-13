package replicated

import (
	"context"
	"microservice-golang/services/venue-service/internal/entity"
	"microservice-golang/shared/infrastructure/postgres"
	sharedgrpc "microservice-golang/shared/pkg/grpc"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PermissionRepository interface {
	Upsert(ctx context.Context, p entity.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error)
	Validate(ctx context.Context, permissionSlug string) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

func (r *permissionRepository) Upsert(ctx context.Context, p entity.Permission) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"resource", "action", "description", "slug", "deleted_at"}),
		}).
		Create(&p).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, "id = ?", id).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	var p entity.Permission
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	return &p, err
}

func (r *permissionRepository) Validate(ctx context.Context, permissionSlug string) error {
	return sharedgrpc.ValidatePermission(ctx, postgres.GetTx(ctx, r.db), permissionSlug)
}
