package mapper

import (
	"microservice-golang/services/sport-service/internal/dto"
	"microservice-golang/services/sport-service/internal/entity"
)

func ToUserSimpleResponse(u *entity.User) *dto.UserSimpleResponse {
	if u == nil {
		return nil
	}
	return &dto.UserSimpleResponse{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToStatusSimpleResponse(s *entity.Status) *dto.StatusSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.StatusSimpleResponse{
		ID:   s.ID,
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToTagSimpleResponse(t *entity.Tag) *dto.TagSimpleResponse {
	if t == nil {
		return nil
	}
	return &dto.TagSimpleResponse{
		ID:   t.ID,
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToRegulatorSimpleResponse(r *entity.Regulator) *dto.RegulatorSimpleResponse {
	if r == nil {
		return nil
	}
	return &dto.RegulatorSimpleResponse{
		ID:               r.ID,
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentID: r.LogoAttachmentID,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      r.PhoneNumber,
		WebsiteURL:       r.WebsiteURL,
		Status:           ToStatusSimpleResponse(r.Status),
	}
}

func ToSportResponse(s *entity.Sport) *dto.SportResponse {
	if s == nil {
		return nil
	}
	return &dto.SportResponse{
		ID:                          s.ID,
		Name:                        s.Name,
		Slug:                        s.Slug,
		Description:                 s.Description,
		IconAttachmentID:            s.IconAttachmentID,
		Status:                      ToStatusSimpleResponse(s.Status),
		TierTag:                     ToTagSimpleResponse(s.TierTag),
		ActiveRegulator:             ToRegulatorSimpleResponse(s.ActiveRegulator),
		Config:                      ToSportConfigResponse(s.Config),
		RequiresApprovalForOfficial: s.RequiresApprovalForOfficial,
		RequiresApprovalForRegional: s.RequiresApprovalForRegional,
		TotalCompetitions:           s.TotalCompetitions,
		TotalAthletes:               s.TotalAthletes,
		TotalAcademies:              s.TotalAcademies,
		CreatedAt:                   s.CreatedAt,
		UpdatedAt:                   s.UpdatedAt,
		CreatedByID:                 s.CreatedByID,
		UpdatedByID:                 s.UpdatedByID,
		DeletedByID:                 s.DeletedByID,
	}
}

func ToSportConfigResponse(c *entity.SportConfig) *dto.SportConfigResponse {
	if c == nil {
		return nil
	}
	statTags := make([]dto.TagSimpleResponse, len(c.StatTags))
	for i, t := range c.StatTags {
		statTags[i] = *ToTagSimpleResponse(&t)
	}
	return &dto.SportConfigResponse{
		ID:                 c.ID,
		SportID:            c.SportID,
		StatTags:           statTags,
		ParticipantTypeTag: ToTagSimpleResponse(c.ParticipantTypeTag),
		MinRosterSize:      c.MinRosterSize,
		MaxRosterSize:      c.MaxRosterSize,
		TypicalRosterSize:  c.TypicalRosterSize,
		RulesURL:           c.RulesURL,
		Description:        c.Description,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

func ToRegulatorResponse(r *entity.Regulator) *dto.RegulatorResponse {
	if r == nil {
		return nil
	}
	staff := make([]*dto.RegulatorStaffResponse, len(r.Staff))
	for i, s := range r.Staff {
		staff[i] = ToRegulatorStaffResponse(&s)
	}
	return &dto.RegulatorResponse{
		ID:               r.ID,
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentID: r.LogoAttachmentID,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      r.PhoneNumber,
		WebsiteURL:       r.WebsiteURL,
		Status:           ToStatusSimpleResponse(r.Status),
		Staff:            staff,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		CreatedByID:      r.CreatedByID,
		UpdatedByID:      r.UpdatedByID,
		DeletedByID:      r.DeletedByID,
	}
}

func ToRegulatorStaffResponse(s *entity.RegulatorStaff) *dto.RegulatorStaffResponse {
	if s == nil {
		return nil
	}
	return &dto.RegulatorStaffResponse{
		ID:          s.ID,
		RegulatorID: s.RegulatorID,
		UserID:      s.UserID,
		User:        ToUserSimpleResponse(s.User),
		RoleTag:     ToTagSimpleResponse(s.RoleTag),
		JoinedAt:    s.JoinedAt,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		CreatedByID: s.CreatedByID,
		UpdatedByID: s.UpdatedByID,
		DeletedByID: s.DeletedByID,
	}
}
