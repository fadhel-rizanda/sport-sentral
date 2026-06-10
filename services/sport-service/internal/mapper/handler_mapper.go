package mapper

import (
	commonv1 "microservice-golang/gen/common/v1"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/dto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoUserSimple(u *dto.UserSimpleResponse) *commonv1.UserSimple {
	if u == nil {
		return nil
	}
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToProtoStatusSimple(s *dto.StatusSimpleResponse) *commonv1.StatusSimple {
	if s == nil {
		return nil
	}
	return &commonv1.StatusSimple{
		Id:   s.ID.String(),
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToProtoTagSimple(t *dto.TagSimpleResponse) *commonv1.TagSimple {
	if t == nil {
		return nil
	}
	return &commonv1.TagSimple{
		Id:   t.ID.String(),
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToProtoRegulatorSimple(r *dto.RegulatorSimpleResponse) *sportv1.Regulator {
	if r == nil {
		return nil
	}
	var logoID string
	if r.LogoAttachmentID != nil {
		logoID = r.LogoAttachmentID.String()
	}
	var phone string
	if r.PhoneNumber != nil {
		phone = *r.PhoneNumber
	}
	var web string
	if r.WebsiteURL != nil {
		web = *r.WebsiteURL
	}
	return &sportv1.Regulator{
		Id:               r.ID.String(),
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentId: logoID,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      phone,
		WebsiteUrl:       web,
		Status:           ToProtoStatusSimple(r.Status),
	}
}

func ToProtoSport(s *dto.SportResponse) *sportv1.Sport {
	if s == nil {
		return nil
	}
	var iconID string
	if s.IconAttachmentID != nil {
		iconID = s.IconAttachmentID.String()
	}
	return &sportv1.Sport{
		Id:                s.ID.String(),
		Name:              s.Name,
		Slug:              s.Slug,
		Description:       s.Description,
		IconAttachmentId:  iconID,
		Status:            ToProtoStatusSimple(s.Status),
		TierTag:           ToProtoTagSimple(s.TierTag),
		ActiveRegulator:   ToProtoRegulator(ToRegulatorFromSimple(s.ActiveRegulator)),
		Config:            ToProtoSportConfig(s.Config),
		TotalCompetitions: s.TotalCompetitions,
		TotalAthletes:     s.TotalAthletes,
		TotalAcademies:    s.TotalAcademies,
		CreatedAt:         timestamppb.New(s.CreatedAt),
		UpdatedAt:         timestamppb.New(s.UpdatedAt),
	}
}

func ToProtoSportConfig(c *dto.SportConfigResponse) *sportv1.SportConfig {
	if c == nil {
		return nil
	}
	statTags := make([]*commonv1.TagSimple, len(c.StatTags))
	for i, t := range c.StatTags {
		statTags[i] = ToProtoTagSimple(&t)
	}
	var rules string
	if c.RulesURL != nil {
		rules = *c.RulesURL
	}
	return &sportv1.SportConfig{
		Id:                 c.ID.String(),
		SportId:            c.SportID.String(),
		StatTags:           statTags,
		ParticipantTypeTag: ToProtoTagSimple(c.ParticipantTypeTag),
		MinRosterSize:      c.MinRosterSize,
		MaxRosterSize:      c.MaxRosterSize,
		TypicalRosterSize:  c.TypicalRosterSize,
		RulesUrl:           rules,
		Description:        c.Description,
		CreatedAt:          timestamppb.New(c.CreatedAt),
		UpdatedAt:          timestamppb.New(c.UpdatedAt),
	}
}

func ToProtoRegulator(r *dto.RegulatorResponse) *sportv1.Regulator {
	if r == nil {
		return nil
	}
	var logoID string
	if r.LogoAttachmentID != nil {
		logoID = r.LogoAttachmentID.String()
	}
	var phone string
	if r.PhoneNumber != nil {
		phone = *r.PhoneNumber
	}
	var web string
	if r.WebsiteURL != nil {
		web = *r.WebsiteURL
	}
	staff := make([]*sportv1.RegulatorStaff, len(r.Staff))
	for i, s := range r.Staff {
		staff[i] = ToProtoRegulatorStaff(s)
	}
	return &sportv1.Regulator{
		Id:               r.ID.String(),
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentId: logoID,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      phone,
		WebsiteUrl:       web,
		Status:           ToProtoStatusSimple(r.Status),
		Staff:            staff,
		CreatedAt:        timestamppb.New(r.CreatedAt),
		UpdatedAt:        timestamppb.New(r.UpdatedAt),
	}
}

func ToProtoRegulatorStaff(s *dto.RegulatorStaffResponse) *sportv1.RegulatorStaff {
	if s == nil {
		return nil
	}
	return &sportv1.RegulatorStaff{
		Id:          s.ID.String(),
		RegulatorId: s.RegulatorID.String(),
		UserId:      s.UserID.String(),
		User:        ToProtoUserSimple(s.User),
		RoleTag:     ToProtoTagSimple(s.RoleTag),
		JoinedAt:    timestamppb.New(s.JoinedAt),
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}

func ToRegulatorFromSimple(r *dto.RegulatorSimpleResponse) *dto.RegulatorResponse {
	if r == nil {
		return nil
	}
	return &dto.RegulatorResponse{
		ID:               r.ID,
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentID: r.LogoAttachmentID,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      r.PhoneNumber,
		WebsiteURL:       r.WebsiteURL,
		Status:           r.Status,
	}
}
