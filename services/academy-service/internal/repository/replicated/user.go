package replicated

import (
	"context"
	"microservice-golang/services/academy-service/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Upsert(ctx context.Context, user entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Upsert(ctx context.Context, user entity.User) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"email",
				"username",
				"full_name",
				"active_role_id",
				"status_id",
				"deleted_at",
			}),
		}).
		Create(&user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, err
}
