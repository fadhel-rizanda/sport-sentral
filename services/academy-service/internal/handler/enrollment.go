package handler

import (
	"context"
	"time"

	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/mapper"
	"microservice-golang/services/academy-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type EnrollmentHandler struct {
	academyv1.UnimplementedEnrollmentServiceServer
	uc usecase.EnrollmentUseCase
}

func NewEnrollmentHandler(uc usecase.EnrollmentUseCase) *EnrollmentHandler {
	return &EnrollmentHandler{uc: uc}
}

func (h *EnrollmentHandler) RegisterGRPC(s *grpc.Server) {
	academyv1.RegisterEnrollmentServiceServer(s, h)
}

func (h *EnrollmentHandler) CreateEnrollment(ctx context.Context, req *academyv1.CreateEnrollmentRequest) (*academyv1.CreateEnrollmentResponse, error) {
	branchID, err := uuid.Parse(req.GetAcademyBranchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	statusID, err := uuid.Parse(req.GetStatusId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	createdByID, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created by id"))
	}

	var joinedAt time.Time
	if req.JoinedAt != nil {
		joinedAt = req.JoinedAt.AsTime()
	} else {
		joinedAt = time.Now()
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := req.ExpiresAt.AsTime()
		expiresAt = &t
	}

	dtoReq := dto.CreateEnrollmentRequest{
		AcademyBranchID: branchID,
		AthleteID:       athleteID,
		JoinedAt:        joinedAt,
		ExpiresAt:       expiresAt,
		StatusID:        statusID,
		CreatedByID:     createdByID,
	}

	res, err := h.uc.Create(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.CreateEnrollmentResponse{
		Enrollment: mapper.ToProtoEnrollment(res),
	}, nil
}

func (h *EnrollmentHandler) GetEnrollment(ctx context.Context, req *academyv1.GetEnrollmentRequest) (*academyv1.GetEnrollmentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid enrollment id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetEnrollmentResponse{
		Enrollment: mapper.ToProtoEnrollment(res),
	}, nil
}

func (h *EnrollmentHandler) GetEnrollmentByBranchAndAthlete(ctx context.Context, req *academyv1.GetEnrollmentByBranchAndAthleteRequest) (*academyv1.GetEnrollmentByBranchAndAthleteResponse, error) {
	branchID, err := uuid.Parse(req.GetBranchId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
	}

	res, err := h.uc.GetByBranchAndAthlete(ctx, branchID, athleteID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.GetEnrollmentByBranchAndAthleteResponse{
		Enrollment: mapper.ToProtoEnrollment(res),
	}, nil
}

func (h *EnrollmentHandler) ListEnrollments(ctx context.Context, req *academyv1.ListEnrollmentsRequest) (*academyv1.ListEnrollmentsResponse, error) {
	var branchID *uuid.UUID
	if req.BranchId != nil && *req.BranchId != "" {
		id, err := uuid.Parse(*req.BranchId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid branch id"))
		}
		branchID = &id
	}

	var athleteID *uuid.UUID
	if req.AthleteId != nil && *req.AthleteId != "" {
		id, err := uuid.Parse(*req.AthleteId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete id"))
		}
		athleteID = &id
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		id, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &id
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}

	dtoReq := dto.ListEnrollmentsRequest{
		BranchID:  branchID,
		AthleteID: athleteID,
		StatusID:  statusID,
		Search:    req.Search,
		Page:      page,
		PageSize:  pageSize,
	}

	res, err := h.uc.List(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	enrollments := make([]*academyv1.Enrollment, len(res.Enrollments))
	for i, e := range res.Enrollments {
		enrollments[i] = mapper.ToProtoEnrollment(e)
	}

	return &academyv1.ListEnrollmentsResponse{
		Enrollments: enrollments,
		Total:       res.Total,
		Page:        int32(res.Page),
		PageSize:    int32(res.PageSize),
	}, nil
}

func (h *EnrollmentHandler) UpdateEnrollment(ctx context.Context, req *academyv1.UpdateEnrollmentRequest) (*academyv1.UpdateEnrollmentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid enrollment id"))
	}

	updatedByID, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated by id"))
	}

	var leftAt *time.Time
	if req.LeftAt != nil {
		t := req.LeftAt.AsTime()
		leftAt = &t
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := req.ExpiresAt.AsTime()
		expiresAt = &t
	}

	var statusID *uuid.UUID
	if req.StatusId != nil && *req.StatusId != "" {
		id, err := uuid.Parse(*req.StatusId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
		}
		statusID = &id
	}

	dtoReq := dto.UpdateEnrollmentRequest{
		LeftAt:      leftAt,
		ExpiresAt:   expiresAt,
		StatusID:    statusID,
		UpdatedByID: updatedByID,
	}

	res, err := h.uc.Update(ctx, id, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.UpdateEnrollmentResponse{
		Enrollment: mapper.ToProtoEnrollment(res),
	}, nil
}

func (h *EnrollmentHandler) DeleteEnrollment(ctx context.Context, req *academyv1.DeleteEnrollmentRequest) (*academyv1.DeleteEnrollmentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid enrollment id"))
	}

	deletedByID, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by id"))
	}

	dtoReq := dto.DeleteEnrollmentRequest{
		ID:          id,
		DeletedByID: deletedByID,
	}

	err = h.uc.Delete(ctx, dtoReq)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &academyv1.DeleteEnrollmentResponse{
		Id: req.GetId(),
	}, nil
}
