package usecase

import (
	"context"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
)

type UserEventPublisher interface {
	PublishUserCreated(ctx context.Context, evt *userv1.UserEvent) error
	PublishUserUpdated(ctx context.Context, evt *userv1.UserEvent) error
	PublishUserDeleted(ctx context.Context, evt *userv1.UserEvent) error
}

type RoleEventPublisher interface {
	PublishRoleCreated(ctx context.Context, evt *rbacv1.RoleEvent) error
	PublishRoleUpdated(ctx context.Context, evt *rbacv1.RoleEvent) error
	PublishRoleDeleted(ctx context.Context, evt *rbacv1.RoleEvent) error
}

type PermissionEventPublisher interface {
	PublishPermissionCreated(ctx context.Context, evt *rbacv1.PermissionEvent) error
	PublishPermissionUpdated(ctx context.Context, evt *rbacv1.PermissionEvent) error
	PublishPermissionDeleted(ctx context.Context, evt *rbacv1.PermissionEvent) error
}
