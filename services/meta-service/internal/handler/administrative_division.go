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

type AdministrativeDivisionHandler struct {
	metav1.UnimplementedAdministrativeDivisionServiceServer
	uc usecase.AdministrativeDivisionUseCase
}

func NewAdministrativeDivisionHandler(uc usecase.AdministrativeDivisionUseCase) *AdministrativeDivisionHandler {
	return &AdministrativeDivisionHandler{uc: uc}
}

func (h *AdministrativeDivisionHandler) RegisterGRPC(s *grpc.Server) {
	metav1.RegisterAdministrativeDivisionServiceServer(s, h)
}

func (h *AdministrativeDivisionHandler) GetAdministrativeDivision(ctx context.Context, req *metav1.GetAdministrativeDivisionRequest) (*metav1.GetAdministrativeDivisionResponse, error) {
	countryID, err := uuid.Parse(req.GetCountryId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
	}

	var parentIDPtr *uuid.UUID
	if req.ParentId != nil {
		parentID, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent id"))
		}
		parentIDPtr = &parentID
	}

	name := req.GetName()
	level := req.GetLevel()
	postalCode := req.GetPostalCode()

	res, err := h.uc.List(ctx, dto.ListAdministrativeDivisionsRequest{
		Name:       &name,
		Level:      &level,
		PostalCode: &postalCode,
		CountryID:  &countryID,
		ParentID:   parentIDPtr,
		Page:       1,
		PageSize:   1,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	if len(res.AdministrativeDivisions) == 0 {
		return nil, apperr.ToGRPC(apperr.NotFound("administrative division"))
	}

	return &metav1.GetAdministrativeDivisionResponse{
		AdministrativeDivision: mapper.ToProtoAdministrativeDivision(res.AdministrativeDivisions[0]),
	}, nil
}

func (h *AdministrativeDivisionHandler) GetAdministrativeDivisionByID(ctx context.Context, req *metav1.GetAdministrativeDivisionByIDRequest) (*metav1.GetAdministrativeDivisionResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid administrative division id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.GetAdministrativeDivisionResponse{
		AdministrativeDivision: mapper.ToProtoAdministrativeDivision(res),
	}, nil
}

func (h *AdministrativeDivisionHandler) ListAdministrativeDivisions(ctx context.Context, req *metav1.ListAdministrativeDivisionsRequest) (*metav1.ListAdministrativeDivisionsResponse, error) {
	var countryIDPtr *uuid.UUID
	if req.CountryId != nil {
		countryID, err := uuid.Parse(req.GetCountryId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
		}
		countryIDPtr = &countryID
	}

	var parentIDPtr *uuid.UUID
	if req.ParentId != nil {
		parentID, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent id"))
		}
		parentIDPtr = &parentID
	}

	res, err := h.uc.List(ctx, dto.ListAdministrativeDivisionsRequest{
		Name:       req.Name,
		Level:      req.Level,
		PostalCode: req.PostalCode,
		CountryID:  countryIDPtr,
		ParentID:   parentIDPtr,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	divisions := make([]*metav1.AdministrativeDivision, 0, len(res.AdministrativeDivisions))
	for _, ad := range res.AdministrativeDivisions {
		divisions = append(divisions, mapper.ToProtoAdministrativeDivision(ad))
	}

	return &metav1.ListAdministrativeDivisionsResponse{
		AdministrativeDivisions: divisions,
	}, nil
}

func (h *AdministrativeDivisionHandler) CreateAdministrativeDivision(ctx context.Context, req *metav1.CreateAdministrativeDivisionRequest) (*metav1.CreateAdministrativeDivisionResponse, error) {
	countryID, err := uuid.Parse(req.GetCountryId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
	}
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}

	var parentIDPtr *uuid.UUID
	if req.ParentId != nil {
		parentID, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent id"))
		}
		parentIDPtr = &parentID
	}

	res, err := h.uc.Create(ctx, dto.CreateAdministrativeDivisionRequest{
		Name:        req.GetName(),
		Level:       req.GetLevel(),
		PostalCode:  req.GetPostalCode(),
		CountryID:   countryID,
		ParentID:    parentIDPtr,
		CreatedByID: createdBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.CreateAdministrativeDivisionResponse{
		AdministrativeDivision: mapper.ToProtoAdministrativeDivision(res),
	}, nil
}

func (h *AdministrativeDivisionHandler) UpdateAdministrativeDivision(ctx context.Context, req *metav1.UpdateAdministrativeDivisionRequest) (*metav1.UpdateAdministrativeDivisionResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid administrative division id"))
	}
	updatedBy, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated_by id"))
	}

	var countryIDPtr *uuid.UUID
	if req.CountryId != nil {
		countryID, err := uuid.Parse(req.GetCountryId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
		}
		countryIDPtr = &countryID
	}

	var parentIDPtr *uuid.UUID
	if req.ParentId != nil {
		parentID, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid parent id"))
		}
		parentIDPtr = &parentID
	}

	res, err := h.uc.Update(ctx, id, dto.UpdateAdministrativeDivisionRequest{
		Name:        req.Name,
		Level:       req.Level,
		PostalCode:  req.PostalCode,
		CountryID:   countryIDPtr,
		ParentID:    parentIDPtr,
		UpdatedByID: updatedBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.UpdateAdministrativeDivisionResponse{
		AdministrativeDivision: mapper.ToProtoAdministrativeDivision(res),
	}, nil
}

func (h *AdministrativeDivisionHandler) DeleteAdministrativeDivision(ctx context.Context, req *metav1.DeleteAdministrativeDivisionRequest) (*metav1.DeleteAdministrativeDivisionResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid administrative division id"))
	}
	deletedBy, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted_by id"))
	}

	if req.GetIsPermanent() {
		err = h.uc.HardDelete(ctx, dto.DeleteAdministrativeDivisionRequest{
			ID: id,
		})
	} else {
		err = h.uc.SoftDelete(ctx, dto.DeleteAdministrativeDivisionRequest{
			ID:          id,
			DeletedByID: deletedBy,
		})
	}
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.DeleteAdministrativeDivisionResponse{
		Id: id.String(),
	}, nil
}
