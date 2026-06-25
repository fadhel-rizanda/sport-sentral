package usecase

import (
	"context"
	academyv1 "microservice-golang/gen/academy/v1"
)

type AcademyHoldingEventPublisher interface {
	PublishHoldingCreated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
	PublishHoldingUpdated(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
	PublishHoldingDeleted(ctx context.Context, evt *academyv1.AcademyHoldingEvent) error
}

type AcademyBranchEventPublisher interface {
	PublishBranchCreated(ctx context.Context, evt *academyv1.AcademyBranchEvent) error
	PublishBranchUpdated(ctx context.Context, evt *academyv1.AcademyBranchEvent) error
	PublishBranchDeleted(ctx context.Context, evt *academyv1.AcademyBranchEvent) error
}

type AcademyAdminEventPublisher interface {
	PublishAdminAssigned(ctx context.Context, evt *academyv1.AcademyAdminEvent) error
	PublishAdminRevoked(ctx context.Context, evt *academyv1.AcademyAdminEvent) error
}
