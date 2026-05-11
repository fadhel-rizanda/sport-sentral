package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"microservice-golang/services/identity-service/internal/entity"
)

type userWithStatus struct {
	entity.User
	StatusCacheID   uuid.UUID `gorm:"column:status_cache_id"`
	StatusCacheType string    `gorm:"column:status_cache_type"`
	StatusCacheName string    `gorm:"column:status_cache_name"`
}

type UserRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	AssignRoles(ctx context.Context, uid uuid.UUID, rids []uuid.UUID) error
	RemoveRoles(ctx context.Context, uid uuid.UUID, rids []uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []userWithStatus
	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("LEFT JOIN status_caches ur_sc ON ur_sc.id = user_roles.status_id").
				Select("user_roles.*, ur_sc.id AS status_id, ur_sc.type AS status_type, ur_sc.name AS status_name")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name").
		Offset(offset).Limit(pageSize).
		Find(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, len(rows))
	for i := range rows {
		rows[i].User.Status = entity.StatusCache{
			ID:   rows[i].StatusCacheID,
			Type: rows[i].StatusCacheType,
			Name: rows[i].StatusCacheName,
		}
		users[i] = &rows[i].User
	}

	return users, total, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var row userWithStatus

	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("LEFT JOIN status_caches ur_sc ON ur_sc.id = user_roles.status_id").
				Select("user_roles.*, ur_sc.id AS status_id, ur_sc.type AS status_type, ur_sc.name AS status_name")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name").
		First(&row, "users.id = ?", id).Error

	if err != nil {
		return nil, err
	}

	row.User.Status = entity.StatusCache{
		ID:   row.StatusCacheID,
		Type: row.StatusCacheType,
		Name: row.StatusCacheName,
	}

	return &row.User, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var row userWithStatus

	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("LEFT JOIN status_caches ur_sc ON ur_sc.id = user_roles.status_id").
				Select("user_roles.*, ur_sc.id AS status_id, ur_sc.type AS status_type, ur_sc.name AS status_name")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name").
		First(&row, "users.email = ?", email).Error

	if err != nil {
		return nil, err
	}

	row.User.Status = entity.StatusCache{
		ID:   row.StatusCacheID,
		Type: row.StatusCacheType,
		Name: row.StatusCacheName,
	}

	return &row.User, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}
	return r.db.WithContext(ctx).Model(user).Association("Roles").Append(roles)
}

func (r *userRepository) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}
	return r.db.WithContext(ctx).Model(user).Association("Roles").Delete(roles)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func toRoleRefs(ids []uuid.UUID) []*entity.Role {
	roles := make([]*entity.Role, len(ids))
	for i, id := range ids {
		roles[i] = &entity.Role{ID: id}
	}
	return roles
}
