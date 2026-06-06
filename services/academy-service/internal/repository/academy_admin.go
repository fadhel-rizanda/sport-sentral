package repository

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademyAdminRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyAdmin, error)
	List(ctx context.Context, filters AdminFilters, page, pageSize int) ([]*entity.AcademyAdmin, int64, error)
	GetByUser(ctx context.Context, scope AdminScope, userID uuid.UUID) (*entity.AcademyAdmin, error)
	Create(ctx context.Context, admin *entity.AcademyAdmin) error
	Update(ctx context.Context, admin *entity.AcademyAdmin) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckUserIsAdmin(ctx context.Context, scope AdminScope, userID uuid.UUID) (bool, error)
}

type AdminScope struct {
	HoldingID *uuid.UUID
	BranchID  *uuid.UUID
}
type AdminFilters struct {
	AdminScope
	StatusID *uuid.UUID
	Search   *string
}

type academyAdminRepository struct {
	db *gorm.DB
}

func NewAcademyAdminRepository(db *gorm.DB) AcademyAdminRepository {
	return &academyAdminRepository{
		db: db,
	}
}

func (r *academyAdminRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AcademyAdmin, error) {
	var admin entity.AcademyAdmin
	err := r.db.WithContext(ctx).
		Preload("Academy").
		Preload("Branch").
		Preload("User").
		Preload("Role").
		Preload("Role.Permissions").
		Preload("ApprovedBy").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *academyAdminRepository) List(ctx context.Context, filters AdminFilters, page, pageSize int) ([]*entity.AcademyAdmin, int64, error) {
	var admins []*entity.AcademyAdmin
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&entity.AcademyAdmin{})

	if filters.HoldingID != nil {
		query = query.Where("academy_admins.academy_id = ? AND academy_admins.branch_id IS NULL", filters.HoldingID)
	}
	if filters.BranchID != nil {
		query = query.Where("academy_admins.branch_id = ?", filters.BranchID)
	}
	if filters.StatusID != nil {
		query = query.Where("academy_admins.role_id = ?", filters.StatusID)
	}

	if filters.Search != nil && *filters.Search != "" {
		if _, err := uuid.Parse(*filters.Search); err == nil {
			query = query.Where("academy_admins.id = ? OR academy_admins.user_id = ?", *filters.Search, *filters.Search)
		} else {
			query = query.
				Joins("JOIN replicated_users ON replicated_users.id = academy_admins.user_id").
				Where("replicated_users.full_name ILIKE ? OR replicated_users.username ILIKE ?", "%"+*filters.Search+"%", "%"+*filters.Search+"%")
		}
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Preload("Role").
		Offset(offset).
		Limit(pageSize).
		Order("academy_admins.approved_at DESC").
		Find(&admins).Error
	if err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

func (r *academyAdminRepository) GetByUser(ctx context.Context, scope AdminScope, userID uuid.UUID) (*entity.AcademyAdmin, error) {
	var admin entity.AcademyAdmin
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if scope.HoldingID != nil {
		query = query.Where("academy_id = ? AND branch_id IS NULL", scope.HoldingID)
	}
	if scope.BranchID != nil {
		query = query.Where("branch_id = ?", scope.BranchID)
	}

	err := query.
		Preload("Academy").
		Preload("Branch").
		Preload("User").
		Preload("Role").
		Preload("Role.Permissions").
		Preload("ApprovedBy").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *academyAdminRepository) Create(ctx context.Context, admin *entity.AcademyAdmin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *academyAdminRepository) Update(ctx context.Context, admin *entity.AcademyAdmin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *academyAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.AcademyAdmin{}, "id = ?", id).Error
}

func (r *academyAdminRepository) CheckUserIsAdmin(ctx context.Context, scope AdminScope, userID uuid.UUID) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL AND approved_at IS NOT NULL", userID)

	if scope.HoldingID != nil {
		query = query.Where("academy_id = ? AND branch_id IS NULL", scope.HoldingID)
	}
	if scope.BranchID != nil {
		query = query.Where("branch_id = ?", scope.BranchID)
	}

	err := query.Model(&entity.AcademyAdmin{}).Count(&count).Error
	return count > 0, err
}
