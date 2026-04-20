package usecase

import (
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/shared/pkg/jwt"
)

type LoginResponse struct {
	UserID   string
	Email    string
	Username string
	Roles    []*rbacv1.Role
	Tokens   *jwt.TokenPair
}
