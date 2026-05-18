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
	StatusCacheSlug string    `gorm:"column:status_cache_slug"`
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
		Table("users").
		Preload("UserRoles", "is_active = ?", true).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name, sc.slug AS status_cache_slug").
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
			Slug: rows[i].StatusCacheSlug,
		}
		users[i] = &rows[i].User
	}

	if len(users) > 0 {
		statusIDs := make([]uuid.UUID, 0)
		for _, user := range users {
			for _, ur := range user.UserRoles {
				statusIDs = append(statusIDs, ur.StatusID)
			}
		}

		if len(statusIDs) > 0 {
			var statuses []entity.StatusCache
			if err := r.db.WithContext(ctx).
				Table("status_caches").
				Where("id IN ?", statusIDs).
				Find(&statuses).Error; err != nil {
				return nil, 0, err
			}

			statusMap := make(map[uuid.UUID]entity.StatusCache)
			for _, s := range statuses {
				statusMap[s.ID] = s
			}

			for _, user := range users {
				for i := range user.UserRoles {
					if status, ok := statusMap[user.UserRoles[i].StatusID]; ok {
						user.UserRoles[i].Status = status
					}
				}
			}
		}
	}

	return users, total, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var row userWithStatus

	err := r.db.WithContext(ctx).
		Table("users").
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("user_roles.created_at DESC")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name, sc.slug AS status_cache_slug").
		First(&row, "users.id = ?", id).Error

	if err != nil {
		return nil, err
	}

	// Only set status if the join returned a valid record
	if row.StatusCacheID != uuid.Nil {
		row.User.Status = entity.StatusCache{
			ID:   row.StatusCacheID,
			Type: row.StatusCacheType,
			Name: row.StatusCacheName,
			Slug: row.StatusCacheSlug,
		}
	}

	if len(row.User.UserRoles) > 0 {
		statusIDs := make([]uuid.UUID, len(row.User.UserRoles))
		roleIDs := make([]uuid.UUID, len(row.User.UserRoles))
		for i, ur := range row.User.UserRoles {
			statusIDs[i] = ur.StatusID
			roleIDs[i] = ur.RoleID
		}

		var statuses []entity.StatusCache
		if err := r.db.WithContext(ctx).
			Table("status_caches").
			Where("id IN ?", statusIDs).
			Find(&statuses).Error; err != nil {
			return nil, err
		}

		statusMap := make(map[uuid.UUID]entity.StatusCache)
		for _, s := range statuses {
			statusMap[s.ID] = s
		}

		var roles []entity.Role
		if err := r.db.WithContext(ctx).
			Preload("Permissions").
			Where("id IN ?", roleIDs).
			Find(&roles).Error; err != nil {
			return nil, err
		}

		roleMap := make(map[uuid.UUID]entity.Role)
		for _, role := range roles {
			roleMap[role.ID] = role
		}

		for i := range row.User.UserRoles {
			if status, ok := statusMap[row.User.UserRoles[i].StatusID]; ok {
				row.User.UserRoles[i].Status = status
			}
			if role, ok := roleMap[row.User.UserRoles[i].RoleID]; ok {
				row.User.UserRoles[i].Role = role
			}
		}
	}

	return &row.User, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var row userWithStatus

	err := r.db.WithContext(ctx).
		Table("users").
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("user_roles.created_at DESC")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN status_caches sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_cache_id, sc.type AS status_cache_type, sc.name AS status_cache_name, sc.slug AS status_cache_slug").
		First(&row, "users.email = ?", email).Error

	if err != nil {
		return nil, err
	}

	// Only set status if the join returned a valid record
	if row.StatusCacheID != uuid.Nil {
		row.User.Status = entity.StatusCache{
			ID:   row.StatusCacheID,
			Type: row.StatusCacheType,
			Name: row.StatusCacheName,
			Slug: row.StatusCacheSlug,
		}
	}

	if len(row.User.UserRoles) > 0 {
		statusIDs := make([]uuid.UUID, len(row.User.UserRoles))
		roleIDs := make([]uuid.UUID, len(row.User.UserRoles))
		for i, ur := range row.User.UserRoles {
			statusIDs[i] = ur.StatusID
			roleIDs[i] = ur.RoleID
		}

		var statuses []entity.StatusCache
		if err := r.db.WithContext(ctx).
			Table("status_caches").
			Where("id IN ?", statusIDs).
			Find(&statuses).Error; err != nil {
			return nil, err
		}

		statusMap := make(map[uuid.UUID]entity.StatusCache)
		for _, s := range statuses {
			statusMap[s.ID] = s
		}

		var roles []entity.Role
		if err := r.db.WithContext(ctx).
			Preload("Permissions").
			Where("id IN ?", roleIDs).
			Find(&roles).Error; err != nil {
			return nil, err
		}

		roleMap := make(map[uuid.UUID]entity.Role)
		for _, role := range roles {
			roleMap[role.ID] = role
		}

		for i := range row.User.UserRoles {
			if status, ok := statusMap[row.User.UserRoles[i].StatusID]; ok {
				row.User.UserRoles[i].Status = status
			}
			if role, ok := roleMap[row.User.UserRoles[i].RoleID]; ok {
				row.User.UserRoles[i].Role = role
			}
		}
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
