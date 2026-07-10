package usecase

import (
	"context"
	"fmt"
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

type MatchUseCase interface {
	Create(ctx context.Context, adminID, branchID uuid.UUID, req dto.CreateMatchRequest) (*dto.MatchResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.MatchResponse, error)
	ListByBranch(ctx context.Context, branchID uuid.UUID, page, pageSize int) ([]*dto.MatchResponse, int64, error)
	UpdateStatus(ctx context.Context, adminID, id uuid.UUID, status string) error
	UpdateScore(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateScoreRequest) error
	Delete(ctx context.Context, adminID, id uuid.UUID) error
}

type matchUseCase struct {
	permissionRepo   replicatedRepo.PermissionRepository
	repo             repository.MatchRepository
	branchRepo       repository.CompetitionBranchRepository
	compRepo         repository.CompetitionRepository
	academyAdminRepo replicatedRepo.AcademyAdminRepository
}

func NewMatchUseCase(
	permissionRepo replicatedRepo.PermissionRepository,
	repo repository.MatchRepository,
	branchRepo repository.CompetitionBranchRepository,
	compRepo repository.CompetitionRepository,
	academyAdminRepo replicatedRepo.AcademyAdminRepository,
) MatchUseCase {
	return &matchUseCase{
		permissionRepo:   permissionRepo,
		repo:             repo,
		branchRepo:       branchRepo,
		compRepo:         compRepo,
		academyAdminRepo: academyAdminRepo,
	}
}

func (uc *matchUseCase) Create(ctx context.Context, adminID, branchID uuid.UUID, req dto.CreateMatchRequest) (*dto.MatchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return nil, err
	}
	branch, err := uc.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("competition branch")
		}
		return nil, apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	participants := make([]entity.MatchParticipant, len(req.Participants))
	for i, p := range req.Participants {
		pID, err := uuid.NewV7()
		if err != nil {
			return nil, apperr.Internal(err)
		}
		participants[i] = entity.MatchParticipant{
			ID:          pID,
			MatchID:     id,
			RosterID:    p.RosterID,
			FormatTagID: p.FormatTagID,
			ResultTagID: p.ResultTagID,
			Score:       p.Score,
			UpdatedAt:   time.Now(),
		}
	}

	match := &entity.Match{
		ID:           id,
		BranchID:     branchID,
		ScheduledAt:  req.ScheduledAt,
		Status:       "SCHEDULED",
		Location:     req.Location,
		Referee:      req.Referee,
		Notes:        req.Notes,
		Participants: participants,
	}

	if err := uc.repo.Create(ctx, match); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_match_roster") {
			return nil, apperr.Conflict("a roster cannot be added to the same match multiple times")
		}
		return nil, apperr.Internal(err)
	}

	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToMatchResponse(res), nil
}

func (uc *matchUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.MatchResponse, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, err
	}
	res, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("match")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToMatchResponse(res), nil
}

func (uc *matchUseCase) ListByBranch(ctx context.Context, branchID uuid.UUID, page, pageSize int) ([]*dto.MatchResponse, int64, error) {
	if err := uc.permissionRepo.Validate(ctx, "competition.read"); err != nil {
		return nil, 0, err
	}
	filters := repository.MatchFilters{
		BranchID: &branchID,
	}

	list, total, err := uc.repo.List(ctx, filters, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}

	res := make([]*dto.MatchResponse, len(list))
	for i, m := range list {
		res[i] = mapper.ToMatchResponse(m)
	}
	return res, total, nil
}

func (uc *matchUseCase) UpdateStatus(ctx context.Context, adminID, id uuid.UUID, status string) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	match, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	match.Status = status
	now := time.Now()
	if status == "STARTED" || status == "LIVE" {
		match.StartedAt = &now
	} else if status == "FINISHED" || status == "COMPLETED" {
		match.EndedAt = &now
	}

	if err := uc.repo.Update(ctx, match); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *matchUseCase) UpdateScore(ctx context.Context, adminID, id uuid.UUID, req dto.UpdateScoreRequest) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	match, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	// TODO need a transaction
	for _, p := range req.Participants {
		participant, err := uc.repo.GetParticipantByID(ctx, id, p.ParticipantID)
		if err != nil {
			if postgres.IsNotFound(err) {
				return apperr.NotFound(fmt.Sprintf("match participant %s", p.ParticipantID))
			}
			return apperr.Internal(err)
		}

		if participant.Score != p.Score {
			logID, err := uuid.NewV7()
			if err != nil {
				return apperr.Internal(err)
			}

			scoreLog := &entity.MatchParticipantScoreLog{
				ID:                 logID,
				MatchParticipantID: p.ParticipantID,
				PreviousScore:      participant.Score,
				NewScore:           p.Score,
				Action:             "UPDATE",
				UpdatedByID:        adminID,
				UpdatedAt:          time.Now(),
			}

			participant.Score = p.Score
			participant.UpdatedAt = time.Now()

			if err := uc.repo.UpdateParticipant(ctx, participant); err != nil {
				return apperr.Internal(err)
			}

			if err := uc.repo.CreateScoreLog(ctx, scoreLog); err != nil {
				return apperr.Internal(err)
			}
		}
	}

	return nil
}

func (uc *matchUseCase) Delete(ctx context.Context, adminID, id uuid.UUID) error {
	if err := uc.permissionRepo.Validate(ctx, "competition.regulate"); err != nil {
		return err
	}
	match, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("match")
		}
		return apperr.Internal(err)
	}
	if err := validateCompetitionAdmin(ctx, adminID, match.Branch.CompetitionID, uc.compRepo, uc.academyAdminRepo); err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
