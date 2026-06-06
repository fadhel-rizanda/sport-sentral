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

type CountryHandler struct {
	metav1.UnimplementedCountryServiceServer
	uc usecase.CountryUseCase
}

func NewCountryHandler(uc usecase.CountryUseCase) *CountryHandler {
	return &CountryHandler{uc: uc}
}

func (h *CountryHandler) RegisterGRPC(s *grpc.Server) {
	metav1.RegisterCountryServiceServer(s, h)
}

func (h *CountryHandler) GetCountryByID(ctx context.Context, req *metav1.GetCountryByIDRequest) (*metav1.GetCountryByIDResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.GetCountryByIDResponse{
		Country: mapper.ToProtoCountry(res),
	}, nil
}

func (h *CountryHandler) ListCountries(ctx context.Context, req *metav1.ListCountriesRequest) (*metav1.ListCountriesResponse, error) {
	res, err := h.uc.List(ctx, dto.ListCountriesRequest{
		Name:      req.Name,
		ISOAlpha2: req.IsoAlpha_2,
		ISOAlpha3: req.IsoAlpha_3,
		Page:      1,
		PageSize:  100,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	countries := make([]*metav1.Country, 0, len(res.Countries))
	for _, c := range res.Countries {
		countries = append(countries, mapper.ToProtoCountry(c))
	}

	return &metav1.ListCountriesResponse{
		Countries: countries,
		Total:     res.Total,
		Page:      int32(res.Page),
		PageSize:  int32(res.PageSize),
	}, nil
}

func (h *CountryHandler) CreateCountry(ctx context.Context, req *metav1.CreateCountryRequest) (*metav1.CreateCountryResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}

	res, err := h.uc.Create(ctx, dto.CreateCountryRequest{
		Name:         req.GetName(),
		ISOAlpha2:    req.GetIsoAlpha_2(),
		ISOAlpha3:    req.GetIsoAlpha_3(),
		PhoneCode:    req.GetPhoneCode(),
		CurrencyCode: req.GetCurrencyCode(),
		CreatedByID:  createdBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.CreateCountryResponse{
		Country: mapper.ToProtoCountry(res),
	}, nil
}

func (h *CountryHandler) UpdateCountry(ctx context.Context, req *metav1.UpdateCountryRequest) (*metav1.UpdateCountryResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
	}
	updatedBy, err := uuid.Parse(req.GetUpdatedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid updated_by id"))
	}

	res, err := h.uc.Update(ctx, id, dto.UpdateCountryRequest{
		Name:         req.Name,
		ISOAlpha2:    req.IsoAlpha_2,
		ISOAlpha3:    req.IsoAlpha_3,
		PhoneCode:    req.PhoneCode,
		CurrencyCode: req.CurrencyCode,
		UpdatedByID:  updatedBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.UpdateCountryResponse{
		Country: mapper.ToProtoCountry(res),
	}, nil
}

func (h *CountryHandler) DeleteCountry(ctx context.Context, req *metav1.DeleteCountryRequest) (*metav1.DeleteCountryResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid country id"))
	}
	deletedBy, err := uuid.Parse(req.GetDeletedById())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid deleted_by id"))
	}

	if req.GetIsPermanent() {
		err = h.uc.HardDelete(ctx, dto.DeleteCountryRequest{
			ID: id,
		})
	} else {
		err = h.uc.SoftDelete(ctx, dto.DeleteCountryRequest{
			ID:          id,
			DeletedByID: deletedBy,
		})
	}
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.DeleteCountryResponse{
		Success: true,
	}, nil
}
