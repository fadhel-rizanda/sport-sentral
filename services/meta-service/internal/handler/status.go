package handler

import (
	"context"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/mapper"
	"microservice-golang/services/meta-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type StatusHandler struct {
	metav1.UnimplementedStatusServiceServer
	uc usecase.StatusUseCase
}

func NewStatusHandler(uc usecase.StatusUseCase) *StatusHandler {
	return &StatusHandler{uc: uc}
}

func (h *StatusHandler) RegisterGRPC(s *grpc.Server) {
	metav1.RegisterStatusServiceServer(s, h)
}

func (h *StatusHandler) GetStatus(ctx context.Context, req *metav1.GetStatusRequest) (*metav1.GetStatusResponse, error) {
	res, err := h.uc.GetByTypeAndName(ctx, req.GetType(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &metav1.GetStatusResponse{
		Status: mapper.ToProtoStatus(res),
	}, nil
}

func (h *StatusHandler) GetStatusByID(ctx context.Context, req *metav1.GetStatusByIDRequest) (*metav1.GetStatusByIDResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.GetStatusByIDResponse{
		Status: mapper.ToProtoStatus(res),
	}, nil
}

func (h *StatusHandler) ListStatuses(ctx context.Context, req *metav1.ListStatusesRequest) (*metav1.ListStatusesResponse, error) {
	res, err := h.uc.List(ctx, dto.ListStatusesRequest{
		Type:     req.Type,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	statuses := make([]*metav1.Status, 0, res.Total)
	for _, status := range res.Statuses {
		statuses = append(statuses, mapper.ToProtoStatus(status))
	}
	return &metav1.ListStatusesResponse{
		Statuses: statuses,
		Total:    res.Total,
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
	}, nil
}

func (h *StatusHandler) CreateStatus(ctx context.Context, req *metav1.CreateStatusRequest) (*metav1.CreateStatusResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}
	res, err := h.uc.Create(ctx, dto.CreateStatusRequest{
		Type:        req.GetType(),
		Name:        req.GetName(),
		Slug:        req.GetSlug(),
		CreatedByID: createdBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.CreateStatusResponse{
		Status: mapper.ToProtoStatus(res),
	}, nil
}

func (h *StatusHandler) UpdateStatus(ctx context.Context, req *metav1.UpdateStatusRequest) (*metav1.UpdateStatusResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}
	updatedBy, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated_by id"))
	}

	res, err := h.uc.Update(ctx, id, dto.UpdateStatusRequest{
		Type:        req.Type,
		Name:        req.Name,
		Slug:        req.Slug,
		UpdatedByID: updatedBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.UpdateStatusResponse{
		Status: mapper.ToProtoStatus(res),
	}, nil
}

func (h *StatusHandler) DeleteStatus(ctx context.Context, req *metav1.DeleteStatusRequest) (*metav1.DeleteStatusResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}
	deletedBy, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted by"))
	}

	if req.GetIsPermanent() {
		err = h.uc.HardDelete(ctx, dto.DeleteStatusRequest{
			ID: id,
		})
	} else {
		err = h.uc.SoftDelete(ctx, dto.DeleteStatusRequest{
			ID:          id,
			DeletedByID: deletedBy,
		})
	}
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.DeleteStatusResponse{}, nil
}
