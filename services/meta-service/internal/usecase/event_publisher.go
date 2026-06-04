package usecase

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
)

type StatusEventPublisher interface {
	PublishStatusCreated(ctx context.Context, evt *metav1.StatusEvent) error
	PublishStatusUpdated(ctx context.Context, evt *metav1.StatusEvent) error
	PublishStatusDeleted(ctx context.Context, evt *metav1.StatusEvent) error
}

type TagEventPublisher interface {
	PublishTagCreated(ctx context.Context, evt *metav1.TagEvent) error
	PublishTagUpdated(ctx context.Context, evt *metav1.TagEvent) error
	PublishTagDeleted(ctx context.Context, evt *metav1.TagEvent) error
}
