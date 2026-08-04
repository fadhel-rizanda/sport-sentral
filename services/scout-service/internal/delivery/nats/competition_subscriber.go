package nats

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/scout-service/internal/entity"
	"microservice-golang/services/scout-service/internal/repository"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"
)

type CompetitionSubscriber struct {
	repo        repository.ScoutRepository
	nats        *messaging.Client
	logger      *zap.Logger
	durableName string
}

func NewCompetitionSubscriber(
	repo repository.ScoutRepository,
	nats *messaging.Client,
	logger *zap.Logger,
	durableName string,
) *CompetitionSubscriber {
	return &CompetitionSubscriber{
		repo:        repo,
		nats:        nats,
		logger:      logger,
		durableName: durableName,
	}
}

func (s *CompetitionSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case events.SubjectAthleteStatsAggregated:
		var evt competitionv1.AthleteStatsAggregateEvent
		if err := proto.Unmarshal(data, &evt); err != nil {
			s.logger.Warn("invalid competition stat event protobuf – skipping",
				zap.String("subject", subject),
				zap.Error(err),
			)
			return nil
		}

		athleteID, err := uuid.Parse(evt.AthleteId)
		if err != nil {
			s.logger.Warn("invalid athlete_id in competition stat event", zap.String("athlete_id", evt.AthleteId))
			return nil
		}

		sportID, err := uuid.Parse(evt.SportId)
		if err != nil {
			s.logger.Warn("invalid sport_id in competition stat event", zap.String("sport_id", evt.SportId))
			return nil
		}

		profile, err := s.repo.GetAthleteProfile(ctx, athleteID, &sportID)
		if err != nil {
			s.logger.Error("failed to fetch athlete profile", zap.Error(err))
			return err
		}

		if profile == nil {
			profileID, _ := uuid.NewV7()
			profile = &entity.AthleteProfile{
				ID:                profileID,
				AthleteID:         athleteID,
				SportID:           sportID,
				TotalCompetitions: 1,
				CreatedAt:         time.Now(),
			}
		}

		// Update aggregated score and timestamp
		profile.LeaderboardScore = profile.LeaderboardScore + evt.AvgValue
		profile.LastUpdatedAt = time.Now()
		profile.UpdatedAt = time.Now()

		if err := s.repo.UpsertAthleteProfile(ctx, profile); err != nil {
			s.logger.Error("failed to upsert athlete profile on stat aggregate event", zap.Error(err))
			return err
		}

		s.logger.Info("athlete profile updated from competition event",
			zap.String("athlete_id", evt.AthleteId),
			zap.String("sport_id", evt.SportId),
		)
		return nil

	default:
		s.logger.Warn("unknown subject in competition stream – skipping", zap.String("subject", subject))
		return nil
	}
}

func (s *CompetitionSubscriber) Listen(ctx context.Context) error {
	return s.nats.Subscribe(
		ctx,
		events.CompetitionStreamName,
		s.durableName,
		[]string{
			events.SubjectAthleteStatsAggregated,
		},
		s.handleMessage,
	)
}
