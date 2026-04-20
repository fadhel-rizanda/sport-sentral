package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	pgerr "microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"

	"microservice-golang/services/user-service/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)

	AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error

	GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmailWithRoles(ctx context.Context, email string) (*entity.User, error)
	GetRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
}

type gormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepo{db: db}
}

func (r *gormUserRepo) Create(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if pgerr.IsUniqueConstraint(result.Error, "idx_users_email") {
			return apperr.Conflict("email already exists")
		}
		if pgerr.IsUniqueConstraint(result.Error, "idx_users_username") {
			return apperr.Conflict("username already exists")
		}
		return apperr.Internal(result.Error)
	}
	return nil
}

func (r *gormUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal(result.Error)
	}
	return &user, nil
}

func (r *gormUserRepo) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&user, "id = ?", id)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal(result.Error)
	}
	return &user, nil
}

func (r *gormUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal(result.Error)
	}
	return &user, nil
}

func (r *gormUserRepo) GetByEmailWithRoles(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		Where("email = ?", email).
		First(&user)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal(result.Error)
	}
	return &user, nil
}

func (r *gormUserRepo) Update(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).
		Model(user).
		Updates(map[string]any{
			"full_name": user.FullName,
			"username":  user.Username,
		})

	if result.Error != nil {
		if pgerr.IsUniqueConstraint(result.Error, "idx_users_username") {
			return apperr.Conflict("username already exists")
		}
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("user not found")
	}
	return nil
}

func (r *gormUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity.User{})

	if result.Error != nil {
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("user not found")
	}
	return nil
}

func (r *gormUserRepo) List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&users)

	if result.Error != nil {
		return nil, 0, apperr.Internal(result.Error)
	}
	return users, total, nil
}

func (r *gormUserRepo) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}

	if err := r.db.WithContext(ctx).Model(user).Association("Roles").Append(roles); err != nil {
		if pgerr.IsUniqueViolation(err) {
			return apperr.Conflict("user already has one or more roles")
		}
		if pgerr.IsForeignKeyViolation(err) {
			return apperr.NotFound("one or more roles not found")
		}
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormUserRepo) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}

	if err := r.db.WithContext(ctx).Model(user).Association("Roles").Delete(roles); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormUserRepo) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	roles := toRoleRefs(roleIDs)
	user := &entity.User{ID: userID}

	if err := r.db.WithContext(ctx).Model(user).Association("Roles").Replace(roles); err != nil {
		if pgerr.IsUniqueViolation(err) {
			return apperr.Conflict("user already has one or more roles")
		}
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormUserRepo) GetRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&user, "id = ?", userID)

	if result.Error != nil {
		if pgerr.IsNotFound(result.Error) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal(result.Error)
	}
	return user.Roles, nil
}

func toRoleRefs(ids []uuid.UUID) []*entity.Role {
	roles := make([]*entity.Role, len(ids))
	for i, id := range ids {
		roles[i] = &entity.Role{ID: id}
	}
	return roles
}
