package repository

import (
	"context"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/shared/infrastructure/postgres"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userWithStatus struct {
	entity.User
	StatusID   uuid.UUID `gorm:"column:status_id"`
	StatusType string    `gorm:"column:status_type"`
	StatusName string    `gorm:"column:status_name"`
	StatusSlug string    `gorm:"column:status_slug"`
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

	if err := postgres.GetTx(ctx, r.db).WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []userWithStatus
	err := postgres.GetTx(ctx, r.db).WithContext(ctx).
		Table("users").
		Preload("UserRoles", "is_active = ?", true).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN replicated_statuses sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_id, sc.type AS status_type, sc.name AS status_name, sc.slug AS status_slug").
		Offset(offset).Limit(pageSize).
		Find(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, len(rows))
	for i := range rows {
		rows[i].User.Status = entity.Status{
			ID:   rows[i].StatusID,
			Type: rows[i].StatusType,
			Name: rows[i].StatusName,
			Slug: rows[i].StatusSlug,
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
			var replicatedStatuses []entity.Status
			if err := postgres.GetTx(ctx, r.db).WithContext(ctx).
				Table("replicated_statuses").
				Where("id IN ?", statusIDs).
				Find(&replicatedStatuses).Error; err != nil {
				return nil, 0, err
			}

			statusMap := make(map[uuid.UUID]entity.Status)
			for _, s := range replicatedStatuses {
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

	err := postgres.GetTx(ctx, r.db).WithContext(ctx).
		Table("users").
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("user_roles.created_at DESC")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN replicated_statuses sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_id, sc.type AS status_type, sc.name AS status_name, sc.slug AS status_slug").
		First(&row, "users.id = ?", id).Error

	if err != nil {
		return nil, err
	}

	// Only set status if the join returned a valid record
	if row.StatusID != uuid.Nil {
		row.User.Status = entity.Status{
			ID:   row.StatusID,
			Type: row.StatusType,
			Name: row.StatusName,
			Slug: row.StatusSlug,
		}
	}

	if len(row.User.UserRoles) > 0 {
		statusIDs := make([]uuid.UUID, len(row.User.UserRoles))
		roleIDs := make([]uuid.UUID, len(row.User.UserRoles))
		for i, ur := range row.User.UserRoles {
			statusIDs[i] = ur.StatusID
			roleIDs[i] = ur.RoleID
		}

		var replicatedStatuses []entity.Status
		if err := postgres.GetTx(ctx, r.db).WithContext(ctx).
			Table("replicated_statuses").
			Where("id IN ?", statusIDs).
			Find(&replicatedStatuses).Error; err != nil {
			return nil, err
		}

		statusMap := make(map[uuid.UUID]entity.Status)
		for _, s := range replicatedStatuses {
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

	err := postgres.GetTx(ctx, r.db).WithContext(ctx).
		Table("users").
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("user_roles.created_at DESC")
		}).
		Preload("UserRoles.Role").
		Joins("LEFT JOIN replicated_statuses sc ON sc.id = users.status_id").
		Select("users.*, sc.id AS status_id, sc.type AS status_type, sc.name AS status_name, sc.slug AS status_slug").
		First(&row, "users.email = ?", email).Error

	if err != nil {
		return nil, err
	}

	// Only set status if the join returned a valid record
	if row.StatusID != uuid.Nil {
		row.User.Status = entity.Status{
			ID:   row.StatusID,
			Type: row.StatusType,
			Name: row.StatusName,
			Slug: row.StatusSlug,
		}
	}

	if len(row.User.UserRoles) > 0 {
		statusIDs := make([]uuid.UUID, len(row.User.UserRoles))
		roleIDs := make([]uuid.UUID, len(row.User.UserRoles))
		for i, ur := range row.User.UserRoles {
			statusIDs[i] = ur.StatusID
			roleIDs[i] = ur.RoleID
		}

		var replicatedStatuses []entity.Status
		if err := postgres.GetTx(ctx, r.db).WithContext(ctx).
			Table("replicated_statuses").
			Where("id IN ?", statusIDs).
			Find(&replicatedStatuses).Error; err != nil {
			return nil, err
		}

		statusMap := make(map[uuid.UUID]entity.Status)
		for _, s := range replicatedStatuses {
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
	return postgres.GetTx(ctx, r.db).WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return postgres.GetTx(ctx, r.db).WithContext(ctx).Save(user).Error
}

func (r *userRepository) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}
	return postgres.GetTx(ctx, r.db).WithContext(ctx).Model(user).Association("Roles").Append(roles)
}

func (r *userRepository) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}
	return postgres.GetTx(ctx, r.db).WithContext(ctx).Model(user).Association("Roles").Delete(roles)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func toRoleRefs(ids []uuid.UUID) []*entity.Role {
	roles := make([]*entity.Role, len(ids))
	for i, id := range ids {
		roles[i] = &entity.Role{ID: id}
	}
	return roles
}
