package repository

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnrollmentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Enrollment, error)
	GetByBranchAndAthlete(ctx context.Context, branchID, athleteID uuid.UUID) (*entity.Enrollment, error)
	List(ctx context.Context, filters EnrollmentFilters, page, pageSize int) ([]*entity.Enrollment, int64, error)
	Create(ctx context.Context, enrollment *entity.Enrollment) error
	Update(ctx context.Context, enrollment *entity.Enrollment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EnrollmentFilters struct {
	BranchID  *uuid.UUID
	AthleteID *uuid.UUID
	StatusID  *uuid.UUID
	Search    *string
}

type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{
		db: db,
	}
}

func (r *enrollmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Enrollment, error) {
	var enrollment entity.Enrollment
	// Menambahkan WithContext di semua method
	err := r.db.WithContext(ctx).
		Preload("AcademyBranch.Holding").
		Preload("AcademyBranch.Sport").
		Preload("Athlete").
		Preload("ApprovedBy").
		Preload("CreatedBy").
		Preload("Status").
		First(&enrollment, "id = ?", id).Error

	return &enrollment, err
}

func (r *enrollmentRepository) GetByBranchAndAthlete(ctx context.Context, branchID, athleteID uuid.UUID) (*entity.Enrollment, error) {
	var enrollment entity.Enrollment
	err := r.db.WithContext(ctx).
		Preload("AcademyBranch.Holding").
		Preload("AcademyBranch.Sport").
		Preload("Athlete").
		Preload("ApprovedBy").
		Preload("CreatedBy").
		Preload("Status").
		First(&enrollment, "academy_branch_id = ? AND athlete_id = ?", branchID, athleteID).Error
	return &enrollment, err
}

func (r *enrollmentRepository) List(ctx context.Context, filters EnrollmentFilters, page, pageSize int) ([]*entity.Enrollment, int64, error) {
	var enrollments []*entity.Enrollment
	var count int64
	offset := pageSize * (page - 1)

	query := r.db.WithContext(ctx).Model(&entity.Enrollment{})

	if filters.BranchID != nil {
		query = query.Where("enrollments.academy_branch_id = ?", filters.BranchID)
	}
	if filters.AthleteID != nil {
		query = query.Where("enrollments.athlete_id = ?", filters.AthleteID)
	}
	if filters.StatusID != nil {
		query = query.Where("enrollments.status_id = ?", filters.StatusID)
	}

	if filters.Search != nil && *filters.Search != "" {
		if _, err := uuid.Parse(*filters.Search); err == nil {
			query = query.Where("enrollments.id = ?", *filters.Search)
		}
		//else {
		//	// Karena mencari nama atlet, kita lakukan JOIN ke tabel replikasi user
		//	query = query.
		//		Joins("JOIN replicated_users ON replicated_users.id = enrollments.athlete_id").
		//		Where("replicated_users.full_name ILIKE ? OR replicated_users.username ILIKE ?", "%"+*filters.Search+"%", "%"+*filters.Search+"%")
		//}
	}

	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("AcademyBranch.Holding").
		Preload("AcademyBranch.Sport").
		Preload("Athlete").
		Preload("Status").
		Order("enrollments.created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&enrollments).Error

	return enrollments, count, err
}

func (r *enrollmentRepository) Create(ctx context.Context, enrollment *entity.Enrollment) error {
	return r.db.WithContext(ctx).Create(enrollment).Error
}

func (r *enrollmentRepository) Update(ctx context.Context, enrollment *entity.Enrollment) error {
	return r.db.WithContext(ctx).Save(enrollment).Error
}

func (r *enrollmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Enrollment{}, "id = ?", id).Error
}
