package mapper

import (
	"time"

	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToSportStatResponse(st *sportv1.SportStat) dto.SportStatResponse {
	if st == nil {
		return dto.SportStatResponse{}
	}
	return dto.SportStatResponse{
		ID:                st.Id,
		SportID:           st.SportId,
		StatTypeTagID:     st.StatTypeTagId,
		StatTypeTag:       ToTagSimpleResponse(st.StatTypeTag),
		AggregationMethod: st.AggregationMethod,
	}
}

func ToSportConfigResponse(c *sportv1.SportConfig) dto.SportConfigResponse {
	if c == nil {
		return dto.SportConfigResponse{}
	}

	stats := make([]dto.SportStatResponse, len(c.Stats))
	for i, st := range c.Stats {
		stats[i] = ToSportStatResponse(st)
	}

	return dto.SportConfigResponse{
		ID:                 c.Id,
		SportID:            c.SportId,
		Stats:              stats,
		ParticipantTypeTag: ToTagSimpleResponse(c.ParticipantTypeTag),
		MinRosterSize:      c.MinRosterSize,
		MaxRosterSize:      c.MaxRosterSize,
		TypicalRosterSize:  c.TypicalRosterSize,
		RulesURL:           c.RulesUrl,
		Description:        c.Description,
		CreatedAt:          c.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:          c.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToRegulatorStaffResponse(st *sportv1.RegulatorStaff) dto.RegulatorStaffResponse {
	if st == nil {
		return dto.RegulatorStaffResponse{}
	}
	return dto.RegulatorStaffResponse{
		ID:          st.Id,
		RegulatorID: st.RegulatorId,
		UserID:      st.UserId,
		User:        ToUserSimpleResponse(st.User),
		RoleTag:     ToTagSimpleResponse(st.RoleTag),
		JoinedAt:    st.JoinedAt.AsTime().UTC().Format(time.RFC3339),
		CreatedAt:   st.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   st.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToRegulatorResponse(r *sportv1.Regulator) dto.RegulatorResponse {
	if r == nil {
		return dto.RegulatorResponse{}
	}

	staffList := make([]dto.RegulatorStaffResponse, len(r.Staff))
	for i, st := range r.Staff {
		staffList[i] = ToRegulatorStaffResponse(st)
	}

	return dto.RegulatorResponse{
		ID:               r.Id,
		OrganizationName: r.OrganizationName,
		Code:             r.Code,
		LogoAttachmentID: r.LogoAttachmentId,
		ContactEmail:     r.ContactEmail,
		PhoneNumber:      r.PhoneNumber,
		WebsiteURL:       r.WebsiteUrl,
		Status:           ToStatusSimpleResponse(r.Status),
		Staff:            staffList,
		CreatedAt:        r.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:        r.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToSportResponse(s *sportv1.Sport) dto.SportResponse {
	if s == nil {
		return dto.SportResponse{}
	}

	var activeRegulator *dto.RegulatorResponse
	if s.ActiveRegulator != nil {
		reg := ToRegulatorResponse(s.ActiveRegulator)
		activeRegulator = &reg
	}

	var config *dto.SportConfigResponse
	if s.Config != nil {
		cfg := ToSportConfigResponse(s.Config)
		config = &cfg
	}

	return dto.SportResponse{
		ID:                s.Id,
		Name:              s.Name,
		Slug:              s.Slug,
		Description:       s.Description,
		IconAttachmentID:  s.IconAttachmentId,
		Status:            ToStatusSimpleResponse(s.Status),
		TierTag:           ToTagSimpleResponse(s.TierTag),
		ActiveRegulator:   activeRegulator,
		Config:            config,
		TotalCompetitions: s.TotalCompetitions,
		TotalAthletes:     s.TotalAthletes,
		TotalAcademies:    s.TotalAcademies,
		CreatedAt:         s.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:         s.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}
