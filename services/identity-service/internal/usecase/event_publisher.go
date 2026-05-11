package usecase

import (
	"context"
	userv1 "microservice-golang/gen/user/v1"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, evt *userv1.UserEvent) error
	PublishUserUpdated(ctx context.Context, evt *userv1.UserEvent) error
	PublishUserDeleted(ctx context.Context, evt *userv1.UserEvent) error
}
