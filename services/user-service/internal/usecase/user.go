package usecase

import (
	"context"

	"github.com/google/uuid"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/repository"
	apperr "microservice-golang/shared/pkg/errors"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, email, username, fullName, password string) (*entity.User, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByEmailInternal(ctx context.Context, email string) (*entity.User, error) // return full entity termasuk HashedPassword
	UpdateUser(ctx context.Context, id, fullName, username string) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)
}

type userUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{repo: repo}
}

func (uc *userUseCase) CreateUser(ctx context.Context, email, username, fullName, password string) (*entity.User, error) {
	user, err := entity.NewUser(email, username, fullName, password)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid user id")
	}

	return uc.repo.GetByID(ctx, uid)
}

func (uc *userUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return uc.repo.GetByEmail(ctx, email)
}

func (uc *userUseCase) GetUserByEmailInternal(ctx context.Context, email string) (*entity.User, error) {
	return uc.repo.GetByEmail(ctx, email)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, id, fullName, username string) (*entity.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid user id")
	}

	user, err := uc.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	user.Update(fullName, username)

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperr.InvalidArgument("invalid user id")
	}

	return uc.repo.Delete(ctx, uid)
}

func (uc *userUseCase) ListUsers(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.List(ctx, page, pageSize)
}
