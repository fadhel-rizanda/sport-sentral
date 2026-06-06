package mapper

import (
	commonv1 "microservice-golang/gen/common/v1"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/dto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoUserSimple(u dto.UserSimpleResponse) *commonv1.UserSimple {
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToProtoStatus(s *dto.StatusResponse) *metav1.Status {
	res := &metav1.Status{
		Id:        s.ID.String(),
		Type:      s.Type,
		Name:      s.Name,
		Slug:      s.Slug,
		CreatedBy: ToProtoUserSimple(s.CreatedBy),
		UpdatedBy: ToProtoUserSimple(s.UpdatedBy),
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}

	if s.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*s.DeletedAt)
		if s.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*s.DeletedBy)
		}
	}

	return res
}

func ToProtoTag(t *dto.TagResponse) *metav1.Tag {
	res := &metav1.Tag{
		Id:        t.ID.String(),
		Type:      t.Type,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedBy: ToProtoUserSimple(t.CreatedBy),
		UpdatedBy: ToProtoUserSimple(t.UpdatedBy),
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}

	if t.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*t.DeletedAt)
		if t.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*t.DeletedBy)
		}
	}

	return res
}

func ToProtoCountrySimple(c dto.CountrySimpleResponse) *commonv1.CountrySimple {
	return &commonv1.CountrySimple{
		Id:           c.ID.String(),
		Name:         c.Name,
		IsoAlpha_2:   c.ISOAlpha2,
		IsoAlpha_3:   c.ISOAlpha3,
		PhoneCode:    c.PhoneCode,
		CurrencyCode: c.CurrencyCode,
	}
}

func ToProtoCountry(c *dto.CountryResponse) *metav1.Country {
	if c == nil {
		return nil
	}
	res := &metav1.Country{
		Id:           c.ID.String(),
		Name:         c.Name,
		IsoAlpha_2:   c.ISOAlpha2,
		IsoAlpha_3:   c.ISOAlpha3,
		PhoneCode:    c.PhoneCode,
		CurrencyCode: c.CurrencyCode,
		CreatedBy:    ToProtoUserSimple(c.CreatedBy),
		UpdatedBy:    ToProtoUserSimple(c.UpdatedBy),
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}

	if c.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*c.DeletedAt)
		if c.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*c.DeletedBy)
		}
	}

	return res
}

func ToProtoAdministrativeDivisionSimple(ad *dto.AdministrativeDivisionSimpleResponse) *commonv1.AdministrativeDivisionSimple {
	if ad == nil {
		return nil
	}
	res := &commonv1.AdministrativeDivisionSimple{
		Id:         ad.ID.String(),
		Name:       ad.Name,
		Level:      ad.Level,
		PostalCode: ad.PostalCode,
		Country:    ToProtoCountrySimple(ad.Country),
	}
	if ad.ParentID != nil {
		res.ParentId = ad.ParentID.String()
	}
	return res
}

func ToProtoAdministrativeDivision(ad *dto.AdministrativeDivisionResponse) *metav1.AdministrativeDivision {
	if ad == nil {
		return nil
	}
	res := &metav1.AdministrativeDivision{
		Id:         ad.ID.String(),
		Name:       ad.Name,
		Level:      ad.Level,
		PostalCode: ad.PostalCode,
		CreatedBy:  ToProtoUserSimple(ad.CreatedBy),
		UpdatedBy:  ToProtoUserSimple(ad.UpdatedBy),
		CreatedAt:  timestamppb.New(ad.CreatedAt),
		UpdatedAt:  timestamppb.New(ad.UpdatedAt),
		Country:    ToProtoCountrySimple(ad.Country),
	}

	if ad.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*ad.DeletedAt)
		if ad.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*ad.DeletedBy)
		}
	}

	if ad.Parent != nil {
		res.Parent = ToProtoAdministrativeDivisionSimple(ad.Parent)
	}

	return res
}
