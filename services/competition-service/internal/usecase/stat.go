package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/entity"
	"microservice-golang/services/competition-service/internal/mapper"
	"microservice-golang/services/competition-service/internal/repository"
	replicatedRepo "microservice-golang/services/competition-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type StatUseCase interface {
	RecordMatchStat(ctx context.Context, adminID, matchID uuid.UUID, req dto.RecordStatRequest) (*dto.MatchStatResponse, error)
	UpdateMatchStat(ctx context.Context, adminID, statID uuid.UUID, value float64) error
	DeleteMatchStat(ctx context.Context, adminID, statID uuid.UUID) error
	GetMatchStats(ctx context.Context, matchID uuid.UUID) ([]*dto.MatchStatResponse, error)
	GetAthleteAggregate(ctx context.Context, athleteID, competitionID uuid.UUID) ([]*dto.AggregateResponse, error)
	RecalculateAggregate(ctx context.Context, adminID, athleteID, competitionID uuid.UUID) error // triggers NATS publish
}

type statUseCase struct {
	permissionRepo   replicatedRepo.PermissionRepository
	statRepo         repository.MatchStatRepository
	matchRepo        repository.MatchRepository
	branchRepo       repository.CompetitionBranchRepository
	competitionRepo  repository.CompetitionRepository
	sportRepo        replicatedRepo.SportRepository
	academyAdminRepo replicatedRepo.AcademyAdminRepository
}

func NewStatUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	statRepo repository.MatchStatRepository,
	matchRepo repository.MatchRepository,
	branchRepo repository.CompetitionBranchRepository,
	competitionRepo repository.CompetitionRepository,
	sportRepo replicatedRepo.SportRepository,
	academyAdminRepo replicatedRepo.AcademyAdminRepository,
) StatUseCase {
	return &statUseCase{
		permissionRepo:   permissionRepo,
		statRepo:         statRepo,
		matchRepo:        matchRepo,
		branchRepo:       branchRepo,
		competitionRepo:  competitionRepo,
		sportRepo:        sportRepo,
		academyAdminRepo: academyAdminRepo,
	}
}

func (uc *statUseCase) RecordMatchStat(ctx context.Context, adminID, matchID uuid.UUID, req dto.RecordStatRequest) (*dto.MatchStatResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return nil, err
	}
	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("match")
		}
		return nil, apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.competitionRepo, uc.academyAdminRepo); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	stat := &entity.MatchStat{
		ID:           id,
		MatchID:      matchID,
		AthleteID:    req.AthleteID,
		StatTypeID:   req.StatTypeID,
		Value:        req.Value,
		RecordedByID: adminID,
	}

	if err := uc.statRepo.Create(ctx, stat); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_match_athlete_stat") {
			return nil, apperr.Conflict("this statistic has already been recorded for this athlete in this match")
		}
		return nil, apperr.Internal(err)
	}

	// Recalculate aggregate for athlete in the competition
	if err := uc.RecalculateAggregate(ctx, adminID, req.AthleteID, match.Branch.CompetitionID); err != nil {
		return nil, err
	}

	res, err := uc.statRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToMatchStatResponse(res), nil
}

func (uc *statUseCase) UpdateMatchStat(ctx context.Context, adminID, statID uuid.UUID, value float64) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	stat, err := uc.statRepo.GetByID(ctx, statID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match stat")
		}
		return apperr.Internal(err)
	}

	// Fetch match to get competition ID
	match, err := uc.matchRepo.GetByID(ctx, stat.MatchID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.competitionRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	stat.Value = value
	if err := uc.statRepo.Update(ctx, stat); err != nil {
		return apperr.Internal(err)
	}

	_ = uc.RecalculateAggregate(ctx, adminID, stat.AthleteID, match.Branch.CompetitionID)

	return nil
}

func (uc *statUseCase) DeleteMatchStat(ctx context.Context, adminID, statID uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	stat, err := uc.statRepo.GetByID(ctx, statID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match stat")
		}
		return apperr.Internal(err)
	}

	// Fetch match to get competition ID
	match, err := uc.matchRepo.GetByID(ctx, stat.MatchID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.competitionRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	if err := uc.statRepo.Delete(ctx, statID); err != nil {
		return apperr.Internal(err)
	}

	_ = uc.RecalculateAggregate(ctx, adminID, stat.AthleteID, match.Branch.CompetitionID)

	return nil
}

func (uc *statUseCase) GetMatchStats(ctx context.Context, matchID uuid.UUID) ([]*dto.MatchStatResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	stats, _, err := uc.statRepo.List(ctx, repository.MatchStatFilters{MatchID: &matchID}, 1, 1000)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	res := make([]*dto.MatchStatResponse, len(stats))
	for i, s := range stats {
		res[i] = mapper.ToMatchStatResponse(s)
	}
	return res, nil
}

func (uc *statUseCase) GetAthleteAggregate(ctx context.Context, athleteID, competitionID uuid.UUID) ([]*dto.AggregateResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	aggregates, _, err := uc.statRepo.ListAggregates(ctx, repository.AggregateFilters{
		CompetitionID: &competitionID,
		AthleteID:     &athleteID,
	}, 1, 1000)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	res := make([]*dto.AggregateResponse, len(aggregates))
	for i, a := range aggregates {
		res[i] = mapper.ToAggregateResponse(a)
	}
	return res, nil
}

func (uc *statUseCase) RecalculateAggregate(ctx context.Context, adminID, athleteID, competitionID uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	if err := validateCompetitionAdmin(ctx, adminID, competitionID, uc.competitionRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	// 1. Fetch competition to get SportID
	competition, err := uc.competitionRepo.GetByID(ctx, competitionID)
	if err != nil {
		return apperr.Internal(err)
	}

	// 2. Fetch the replicated sport and its stats configuration
	sport, err := uc.sportRepo.GetByID(ctx, competition.SportID)
	if err != nil {
		return apperr.Internal(err)
	}

	// 3. Map configured stats by their StatTypeTagID for quick lookup
	statConfigMap := make(map[uuid.UUID]entity.SportStat)
	for _, s := range sport.Stats {
		statConfigMap[s.StatTypeTagID] = s
	}

	// 4. Fetch all branches of the competition
	branches, _, err := uc.branchRepo.List(ctx, repository.CompetitionBranchFilters{CompetitionID: &competitionID}, 1, 1000)
	if err != nil {
		return apperr.Internal(err)
	}

	branchMap := make(map[uuid.UUID]bool)
	for _, b := range branches {
		branchMap[b.ID] = true
	}

	// 5. Fetch all stats for the athlete
	stats, _, err := uc.statRepo.List(ctx, repository.MatchStatFilters{AthleteID: &athleteID}, 1, 10000)
	if err != nil {
		return apperr.Internal(err)
	}

	// Group by (BranchID, StatTypeID)
	type groupKey struct {
		BranchID   uuid.UUID
		StatTypeID uuid.UUID
	}
	groups := make(map[groupKey][]float64)

	for _, s := range stats {
		if s.Match == nil {
			continue
		}
		// Only calculate if the match's branch is in our competition AND the stat type is configured for the sport
		if branchMap[s.Match.BranchID] {
			if _, configured := statConfigMap[s.StatTypeID]; configured {
				key := groupKey{
					BranchID:   s.Match.BranchID,
					StatTypeID: s.StatTypeID,
				}
				groups[key] = append(groups[key], s.Value)
			}
		}
	}

	// For each group, calculate aggregates and save
	for key, values := range groups {
		if len(values) == 0 {
			continue
		}

		totalMatches := int32(len(values))
		var sum float64
		maxValue := values[0]
		minValue := values[0]

		for _, val := range values {
			sum += val
			if val > maxValue {
				maxValue = val
			}
			if val < minValue {
				minValue = val
			}
		}
		avgValue := sum / float64(totalMatches)

		// Get existing aggregate or create new
		aggregate, err := uc.statRepo.GetAggregate(ctx, competitionID, key.BranchID, athleteID, key.StatTypeID)
		if err != nil {
			if postgres.IsNotFound(err) {
				aggID, err := uuid.NewV7()
				if err != nil {
					return apperr.Internal(err)
				}
				aggregate = &entity.AthleteStatsAggregate{
					ID:            aggID,
					CompetitionID: competitionID,
					BranchID:      key.BranchID,
					AthleteID:     athleteID,
					StatTypeID:    key.StatTypeID,
				}
			} else {
				return apperr.Internal(err)
			}
		}

		aggregate.TotalMatches = totalMatches
		aggregate.AvgValue = avgValue
		aggregate.SumValue = sum
		aggregate.MaxValue = maxValue
		aggregate.MinValue = minValue
		aggregate.LastUpdatedAt = time.Now()

		if err := uc.statRepo.SaveAggregate(ctx, aggregate); err != nil {
			return apperr.Internal(err)
		}
	}

	return nil
}
