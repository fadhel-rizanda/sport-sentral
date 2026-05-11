package usecase

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
)

type EventPublisher interface {
	PublishStatusCreated(ctx context.Context, evt *metav1.StatusEvent) error
	PublishStatusUpdated(ctx context.Context, evt *metav1.StatusEvent) error
	PublishStatusDeleted(ctx context.Context, evt *metav1.StatusEvent) error
}
