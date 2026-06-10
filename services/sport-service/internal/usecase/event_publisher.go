package usecase

import (
	"context"
	sportv1 "microservice-golang/gen/sport/v1"
)

type SportEventPublisher interface {
	PublishSportCreated(ctx context.Context, evt *sportv1.SportEvent) error
	PublishSportUpdated(ctx context.Context, evt *sportv1.SportEvent) error
	PublishSportDeleted(ctx context.Context, evt *sportv1.SportEvent) error
}
