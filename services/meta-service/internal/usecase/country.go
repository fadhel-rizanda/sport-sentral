package usecase

import (
	"context"
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/mapper"
	"time"

	"gorm.io/gorm"

	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/entity"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/shared/infrastructure/postgres"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CountryUseCase interface {
	Create(ctx context.Context, req dto.CreateCountryRequest) (*dto.CountryResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CountryResponse, error)
	List(ctx context.Context, req dto.ListCountriesRequest) (*dto.ListCountriesResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateCountryRequest) (*dto.CountryResponse, error)
	SoftDelete(ctx context.Context, req dto.DeleteCountryRequest) error
	HardDelete(ctx context.Context, req dto.DeleteCountryRequest) error
}

type countryUseCase struct {
	repo      repository.CountryRepository
	publisher CountryEventPublisher
}

func NewCountryUseCase(repo repository.CountryRepository, publisher CountryEventPublisher) CountryUseCase {
	return &countryUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *countryUseCase) Create(ctx context.Context, req dto.CreateCountryRequest) (*dto.CountryResponse, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	country := &entity.Country{
		ID:           id,
		Name:         req.Name,
		ISOAlpha2:    req.ISOAlpha2,
		ISOAlpha3:    req.ISOAlpha3,
		PhoneCode:    req.PhoneCode,
		CurrencyCode: req.CurrencyCode,
		CreatedByID:  req.CreatedByID,
		UpdatedByID:  req.CreatedByID,
	}

	if err := uc.repo.Create(ctx, country); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_countries_iso_alpha2") || postgres.IsUniqueConstraint(err, "uni_countries_iso_alpha3") {
			return nil, apperr.Conflict("iso_alpha")
		}
		return nil, apperr.Internal(err)
	}

	// Fetch country with preloaded user relations for the response mapping
	countryWithRelations, err := uc.repo.GetByID(ctx, id)
	if err == nil {
		country = countryWithRelations
	}

	evt := uc.buildEvent(
		metav1.CountryEventType_COUNTRY_EVENT_TYPE_CREATED,
		country,
	)
	if err := uc.publisher.PublishCountryCreated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCountryResponse(country), nil
}

func (uc *countryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.CountryResponse, error) {
	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("country")
		}
		return nil, apperr.Internal(err)
	}
	return mapper.ToCountryResponse(c), nil
}

func (uc *countryUseCase) List(ctx context.Context, req dto.ListCountriesRequest) (*dto.ListCountriesResponse, error) {
	filters := repository.CountryFilters{
		Name:      req.Name,
		ISOAlpha2: req.ISOAlpha2,
		ISOAlpha3: req.ISOAlpha3,
	}

	countries, total, err := uc.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	result := make([]*dto.CountryResponse, len(countries))
	for i, c := range countries {
		result[i] = mapper.ToCountryResponse(c)
	}

	return &dto.ListCountriesResponse{
		Countries: result,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

func (uc *countryUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCountryRequest) (*dto.CountryResponse, error) {
	country, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, apperr.NotFound("country")
		}
		return nil, apperr.Internal(err)
	}

	hasChanges := false
	if req.Name != nil {
		country.Name = *req.Name
		hasChanges = true
	}
	if req.ISOAlpha2 != nil {
		country.ISOAlpha2 = *req.ISOAlpha2
		hasChanges = true
	}
	if req.ISOAlpha3 != nil {
		country.ISOAlpha3 = *req.ISOAlpha3
		hasChanges = true
	}
	if req.PhoneCode != nil {
		country.PhoneCode = *req.PhoneCode
		hasChanges = true
	}
	if req.CurrencyCode != nil {
		country.CurrencyCode = *req.CurrencyCode
		hasChanges = true
	}

	if hasChanges {
		country.UpdatedByID = req.UpdatedByID
	}

	if err := uc.repo.Update(ctx, country); err != nil {
		if postgres.IsUniqueConstraint(err, "uni_countries_iso_alpha2") || postgres.IsUniqueConstraint(err, "uni_countries_iso_alpha3") {
			return nil, apperr.Conflict("iso_alpha")
		}
		return nil, apperr.Internal(err)
	}

	// Refetch to get populated relations
	countryWithRelations, err := uc.repo.GetByID(ctx, id)
	if err == nil {
		country = countryWithRelations
	}

	evt := uc.buildEvent(
		metav1.CountryEventType_COUNTRY_EVENT_TYPE_UPDATED,
		country,
	)
	if err := uc.publisher.PublishCountryUpdated(ctx, evt); err != nil {
		return nil, apperr.Internal(err)
	}

	return mapper.ToCountryResponse(country), nil
}

func (uc *countryUseCase) SoftDelete(ctx context.Context, req dto.DeleteCountryRequest) error {
	country, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("country")
		}
		return apperr.Internal(err)
	}

	country.DeletedByID = &req.DeletedByID
	country.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	if err := uc.repo.Update(ctx, country); err != nil {
		return apperr.Internal(err)
	}

	evt := uc.buildEvent(
		metav1.CountryEventType_COUNTRY_EVENT_TYPE_DELETED,
		country,
	)
	if err := uc.publisher.PublishCountryDeleted(ctx, evt); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *countryUseCase) HardDelete(ctx context.Context, req dto.DeleteCountryRequest) error {
	_, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return apperr.NotFound("country")
		}
		return apperr.Internal(err)
	}

	if err := uc.repo.Delete(ctx, req.ID); err != nil {
		return apperr.Internal(err)
	}

	return nil
}

func (uc *countryUseCase) buildEvent(
	eventType metav1.CountryEventType,
	c *entity.Country,
) *metav1.CountryEvent {
	evtID, _ := uuid.NewV7()

	evt := &metav1.CountryEvent{
		EventId:             evtID.String(),
		EventType:           eventType,
		OccurredAt:          timestamppb.Now(),
		CountryId:           c.ID.String(),
		CountryName:         c.Name,
		CountryIsoAlpha_2:   c.ISOAlpha2,
		CountryIsoAlpha_3:   c.ISOAlpha3,
		CountryPhoneCode:    c.PhoneCode,
		CountryCurrencyCode: c.CurrencyCode,
	}
	if c.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(c.DeletedAt.Time)
	}

	return evt
}
