package mapper

import (
	"microservice-golang/services/meta-service/internal/dto"
	"microservice-golang/services/meta-service/internal/entity"
)

func ToStatusResponse(status *entity.Status) *dto.StatusResponse {
	res := &dto.StatusResponse{
		ID:        status.ID,
		Type:      status.Type,
		Name:      status.Name,
		Slug:      status.Slug,
		CreatedAt: status.CreatedAt,
		UpdatedAt: status.UpdatedAt,
	}

	if status.CreatedBy != nil {
		res.CreatedBy = dto.UserSimpleResponse{
			ID:       status.CreatedByID,
			Email:    status.CreatedBy.Email,
			Username: status.CreatedBy.Username,
			FullName: status.CreatedBy.FullName,
		}
	}

	if status.UpdatedBy != nil {
		res.UpdatedBy = dto.UserSimpleResponse{
			ID:       status.UpdatedByID,
			Email:    status.UpdatedBy.Email,
			Username: status.UpdatedBy.Username,
			FullName: status.UpdatedBy.FullName,
		}
	}

	if status.DeletedAt.Valid {
		res.DeletedAt = &status.DeletedAt.Time
		if status.DeletedByID != nil && status.DeletedBy != nil {
			res.DeletedBy = &dto.UserSimpleResponse{
				ID:       *status.DeletedByID,
				Email:    status.DeletedBy.Email,
				Username: status.DeletedBy.Username,
				FullName: status.DeletedBy.FullName,
			}
		}
	}

	return res
}

func ToTagResponse(tag *entity.Tag) *dto.TagResponse {
	res := &dto.TagResponse{
		ID:        tag.ID,
		Type:      tag.Type,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}

	if tag.CreatedBy != nil {
		res.CreatedBy = dto.UserSimpleResponse{
			ID:       tag.CreatedByID,
			Email:    tag.CreatedBy.Email,
			Username: tag.CreatedBy.Username,
			FullName: tag.CreatedBy.FullName,
		}
	}

	if tag.UpdatedBy != nil {
		res.UpdatedBy = dto.UserSimpleResponse{
			ID:       tag.UpdatedByID,
			Email:    tag.UpdatedBy.Email,
			Username: tag.UpdatedBy.Username,
			FullName: tag.UpdatedBy.FullName,
		}
	}

	if tag.DeletedAt.Valid {
		res.DeletedAt = &tag.DeletedAt.Time
		if tag.DeletedByID != nil && tag.DeletedBy != nil {
			res.DeletedBy = &dto.UserSimpleResponse{
				ID:       *tag.DeletedByID,
				Email:    tag.DeletedBy.Email,
				Username: tag.DeletedBy.Username,
				FullName: tag.DeletedBy.FullName,
			}
		}
	}

	return res
}

func ToCountryResponse(country *entity.Country) *dto.CountryResponse {
	if country == nil {
		return nil
	}
	res := &dto.CountryResponse{
		ID:           country.ID,
		Name:         country.Name,
		ISOAlpha2:    country.ISOAlpha2,
		ISOAlpha3:    country.ISOAlpha3,
		PhoneCode:    country.PhoneCode,
		CurrencyCode: country.CurrencyCode,
		CreatedAt:    country.CreatedAt,
		UpdatedAt:    country.UpdatedAt,
	}

	if country.CreatedBy != nil {
		res.CreatedBy = dto.UserSimpleResponse{
			ID:       country.CreatedByID,
			Email:    country.CreatedBy.Email,
			Username: country.CreatedBy.Username,
			FullName: country.CreatedBy.FullName,
		}
	}

	if country.UpdatedBy != nil {
		res.UpdatedBy = dto.UserSimpleResponse{
			ID:       country.UpdatedByID,
			Email:    country.UpdatedBy.Email,
			Username: country.UpdatedBy.Username,
			FullName: country.UpdatedBy.FullName,
		}
	}

	if country.DeletedAt.Valid {
		res.DeletedAt = &country.DeletedAt.Time
		if country.DeletedByID != nil && country.DeletedBy != nil {
			res.DeletedBy = &dto.UserSimpleResponse{
				ID:       *country.DeletedByID,
				Email:    country.DeletedBy.Email,
				Username: country.DeletedBy.Username,
				FullName: country.DeletedBy.FullName,
			}
		}
	}

	return res
}

func ToCountrySimpleResponse(country *entity.Country) dto.CountrySimpleResponse {
	if country == nil {
		return dto.CountrySimpleResponse{}
	}
	return dto.CountrySimpleResponse{
		ID:           country.ID,
		Name:         country.Name,
		ISOAlpha2:    country.ISOAlpha2,
		ISOAlpha3:    country.ISOAlpha3,
		PhoneCode:    country.PhoneCode,
		CurrencyCode: country.CurrencyCode,
	}
}

func ToAdministrativeDivisionResponse(division *entity.AdministrativeDivision) *dto.AdministrativeDivisionResponse {
	if division == nil {
		return nil
	}
	res := &dto.AdministrativeDivisionResponse{
		ID:         division.ID,
		CountryID:  division.CountryID,
		ParentID:   division.ParentID,
		Name:       division.Name,
		Level:      division.Level,
		PostalCode: division.PostalCode,
		CreatedAt:  division.CreatedAt,
		UpdatedAt:  division.UpdatedAt,
	}

	if division.CreatedBy != nil {
		res.CreatedBy = dto.UserSimpleResponse{
			ID:       division.CreatedByID,
			Email:    division.CreatedBy.Email,
			Username: division.CreatedBy.Username,
			FullName: division.CreatedBy.FullName,
		}
	}

	if division.UpdatedBy != nil {
		res.UpdatedBy = dto.UserSimpleResponse{
			ID:       division.UpdatedByID,
			Email:    division.UpdatedBy.Email,
			Username: division.UpdatedBy.Username,
			FullName: division.UpdatedBy.FullName,
		}
	}

	if division.DeletedAt.Valid {
		res.DeletedAt = &division.DeletedAt.Time
		if division.DeletedByID != nil && division.DeletedBy != nil {
			res.DeletedBy = &dto.UserSimpleResponse{
				ID:       *division.DeletedByID,
				Email:    division.DeletedBy.Email,
				Username: division.DeletedBy.Username,
				FullName: division.DeletedBy.FullName,
			}
		}
	}

	if division.Country != nil {
		res.Country = ToCountrySimpleResponse(division.Country)
	}

	if division.Parent != nil {
		res.Parent = ToAdministrativeDivisionSimpleResponse(division.Parent)
	}

	return res
}

func ToAdministrativeDivisionSimpleResponse(division *entity.AdministrativeDivision) *dto.AdministrativeDivisionSimpleResponse {
	if division == nil {
		return nil
	}
	res := &dto.AdministrativeDivisionSimpleResponse{
		ID:         division.ID,
		Name:       division.Name,
		Level:      division.Level,
		PostalCode: division.PostalCode,
		ParentID:   division.ParentID,
	}

	if division.Country != nil {
		res.Country = ToCountrySimpleResponse(division.Country)
	}

	return res
}
