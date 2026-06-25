package usecase

import (
	"context"

	"github.com/google/uuid"

	"microservice-golang/services/competition-service/internal/repository"
	replicatedRepo "microservice-golang/services/competition-service/internal/repository/replicated"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"
	sharedgrpc "microservice-golang/shared/pkg/grpc"
)

func validateCompetitionAdmin(
	ctx context.Context,
	adminID uuid.UUID,
	competitionID uuid.UUID,
	compRepo repository.CompetitionRepository,
	academyAdminRepo replicatedRepo.AcademyAdminRepository,
) error {
	// 1. Get user's active role from context
	activeRole, err := sharedgrpc.ExtractActiveRole(ctx)
	if err != nil {
		return err
	}

	// Platform admin bypasses check
	if activeRole == "platform_admin" {
		return nil
	}

	// 2. Fetch the competition
	comp, err := compRepo.GetByID(ctx, competitionID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("competition")
		}
		return apperr.Internal(err)
	}

	// 3. If creator, allow
	if comp.CreatedByID == adminID {
		return nil
	}

	// 4. Check if competition admin exists
	_, err = compRepo.GetAdmin(ctx, competitionID, adminID)
	if err == nil {
		return nil
	}

	// 5. If competition has a host academy branch, check if the user is an academy admin for it
	if comp.HostAcademyBranchID != nil && *comp.HostAcademyBranchID != uuid.Nil {
		academyAdmin, err := academyAdminRepo.GetByUserID(ctx, adminID)
		if err == nil {
			if academyAdmin.BranchID != nil && *academyAdmin.BranchID == *comp.HostAcademyBranchID {
				return nil
			}
			if comp.HostAcademy != nil && academyAdmin.AcademyID == comp.HostAcademy.HoldingID {
				return nil
			}
		}
	}

	return apperr.Forbidden("insufficient permissions to manage this competition")
}
