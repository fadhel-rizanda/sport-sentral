package nats

import (
	"context"

	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"

	"go.uber.org/zap"
)

type CompetitionEventPublisher interface {
	PublishAthleteStatsAggregated(ctx context.Context, evt *competitionv1.AthleteStatsAggregateEvent) error
}

type competitionEventPublisher struct {
	nats   *messaging.Client
	logger *zap.Logger
}

func NewCompetitionEventPublisher(nats *messaging.Client, logger *zap.Logger) CompetitionEventPublisher {
	return &competitionEventPublisher{
		nats:   nats,
		logger: logger,
	}
}

func (p *competitionEventPublisher) PublishAthleteStatsAggregated(ctx context.Context, evt *competitionv1.AthleteStatsAggregateEvent) error {
	if err := p.nats.Publish(ctx, events.SubjectAthleteStatsAggregated, evt); err != nil {
		p.logger.Error("failed to publish athlete stats aggregate event",
			zap.String("event_id", evt.EventId),
			zap.String("athlete_id", evt.AthleteId),
			zap.Error(err),
		)
		return err
	}
	p.logger.Debug("athlete stats aggregate event published", zap.String("event_id", evt.EventId))
	return nil
}
