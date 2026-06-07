package usecase

import (
	"context"

	"github.com/google/uuid"

	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
)

type RosterUseCase interface {
	Create(ctx context.Context, req dto.CreateRosterRequest) (*dto.RosterResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RosterResponse, error)
	List(ctx context.Context, req dto.ListRostersRequest) (*dto.ListRostersResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateRosterRequest) (*dto.RosterResponse, error)
	Delete(ctx context.Context, req dto.DeleteRosterRequest) error

	GetMembers(ctx context.Context, rosterID uuid.UUID) ([]dto.RosterMemberResponse, error)
	GetMemberByID(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) (*dto.RosterMemberResponse, error)
	AddMember(ctx context.Context, req dto.AddRosterMemberRequest) (*dto.RosterMemberResponse, error)
	RemoveMember(ctx context.Context, rosterID, memberID uuid.UUID, req dto.RemoveRosterMemberRequest) error
	DeleteMember(ctx context.Context, rosterID, memberID uuid.UUID) error
}

type rosterUseCase struct {
	repo repository.RosterRepository
}

func NewRosterUseCase(repo repository.RosterRepository) RosterUseCase {
	return &rosterUseCase{
		repo: repo,
	}
}

func (uc *rosterUseCase) Create(ctx context.Context, req dto.CreateRosterRequest) (*dto.RosterResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	roster := &entity.Roster{
		ID:              id,
		AcademyBranchID: req.AcademyBranchID,
		CompetitionID:   req.CompetitionID,
		Name:            req.Name,
		TagID:           req.TagID,
		StatusID:        req.StatusID,
		MaxSize:         req.MaxSize,
	}

	if err := uc.repo.Create(ctx, roster); err != nil {
		return nil, apperr.Internal(err)
	}

	resRoster, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRosterResponse(resRoster), nil
}

func (uc *rosterUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.RosterResponse, error) {
	roster, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("roster")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToRosterResponse(roster), nil
}

func (uc *rosterUseCase) List(ctx context.Context, req dto.ListRostersRequest) (*dto.ListRostersResponse, error) {
	filters := repository.RosterFilters{
		CompetitionID: req.CompetitionID,
		BranchID:      req.BranchID,
		StatusID:      req.StatusID,
		TagID:         req.TagID,
		Search:        req.Search,
	}

	rosters, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.RosterResponse, len(rosters))
	for i, r := range rosters {
		result[i] = mapper.ToRosterResponse(r)
	}

	return &dto.ListRostersResponse{
		Rosters:  result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (uc *rosterUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateRosterRequest) (*dto.RosterResponse, error) {
	roster, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("roster")
		}
		return nil, apperr.Internal(err)
	}

	updated := false
	if req.CompetitionID != nil {
		roster.CompetitionID = req.CompetitionID
		updated = true
	}
	if req.Name != nil {
		roster.Name = *req.Name
		updated = true
	}
	if req.TagID != nil {
		roster.TagID = *req.TagID
		updated = true
	}
	if req.StatusID != nil {
		roster.StatusID = *req.StatusID
		updated = true
	}
	if req.MaxSize != nil {
		roster.MaxSize = *req.MaxSize
		updated = true
	}

	if updated {
		if err := uc.repo.Update(ctx, roster); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	resRoster, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToRosterResponse(resRoster), nil
}

func (uc *rosterUseCase) Delete(ctx context.Context, req dto.DeleteRosterRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("roster")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *rosterUseCase) GetMembers(ctx context.Context, rosterID uuid.UUID) ([]dto.RosterMemberResponse, error) {
	_, err := uc.repo.GetByID(ctx, rosterID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("roster")
		}
		return nil, apperr.Internal(err)
	}

	members, err := uc.repo.GetMembers(ctx, rosterID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]dto.RosterMemberResponse, len(members))
	for i, m := range members {
		result[i] = mapper.ToRosterMemberResponse(m)
	}

	return result, nil
}

func (uc *rosterUseCase) GetMemberByID(ctx context.Context, rosterID uuid.UUID, memberID uuid.UUID) (*dto.RosterMemberResponse, error) {
	member, err := uc.repo.GetMemberByID(ctx, rosterID, memberID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("roster member")
		}
		return nil, apperr.Internal(err)
	}
	res := mapper.ToRosterMemberResponse(member)
	return &res, nil
}

func (uc *rosterUseCase) AddMember(ctx context.Context, req dto.AddRosterMemberRequest) (*dto.RosterMemberResponse, error) {
	roster, err := uc.repo.GetByID(ctx, req.RosterID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("roster")
		}
		return nil, apperr.Internal(err)
	}

	// Check max size
	activeMembers := 0
	for _, m := range roster.Members {
		if m.RemovedAt == nil {
			activeMembers++
		}
	}
	if activeMembers >= roster.MaxSize {
		return nil, apperr.Conflict("roster has reached its maximum size limit")
	}

	// Check if already in roster
	exists, err := uc.repo.CheckAthleteInRoster(ctx, req.RosterID, req.AthleteID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if exists {
		return nil, apperr.Conflict("athlete is already an active member of this roster")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	member := &entity.RosterMember{
		ID:           id,
		RosterID:     req.RosterID,
		AthleteID:    req.AthleteID,
		JerseyNumber: req.JerseyNumber,
		PositionID:   req.PositionID,
		StatusID:     req.StatusID,
		AddedByID:    req.AddedByID,
	}

	if err := uc.repo.AddMember(ctx, member); err != nil {
		if postgres.IsUniqueConstraint(err, "idx_roster_athlete") || postgres.IsUniqueConstraint(err, "uq_roster_athlete") {
			return nil, apperr.Conflict("athlete is already in this roster")
		}
		return nil, apperr.Internal(err)
	}

	resMember, err := uc.repo.GetMemberByID(ctx, req.RosterID, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	res := mapper.ToRosterMemberResponse(resMember)
	return &res, nil
}

func (uc *rosterUseCase) RemoveMember(ctx context.Context, rosterID, memberID uuid.UUID, req dto.RemoveRosterMemberRequest) error {
	_, err := uc.repo.GetMemberByID(ctx, rosterID, memberID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("roster member")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.RemoveMember(ctx, rosterID, memberID, req.RemovedByID, req.RemovalReason); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (uc *rosterUseCase) DeleteMember(ctx context.Context, rosterID, memberID uuid.UUID) error {
	_, err := uc.repo.GetMemberByID(ctx, rosterID, memberID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("roster member")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.DeleteMember(ctx, rosterID, memberID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}
