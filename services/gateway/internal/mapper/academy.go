package mapper

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	academyv1 "microservice-golang/gen/academy/v1"
	commonv1 "microservice-golang/gen/common/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func FormatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

func FormatTimestampPtr(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	str := ts.AsTime().UTC().Format(time.RFC3339)
	return &str
}

func ToCountrySimpleResponse(c *commonv1.CountrySimple) *dto.CountrySimpleResponse {
	if c == nil {
		return nil
	}
	return &dto.CountrySimpleResponse{
		ID:           c.Id,
		Name:         c.Name,
		ISOAlpha2:    c.IsoAlpha_2,
		ISOAlpha3:    c.IsoAlpha_3,
		PhoneCode:    c.PhoneCode,
		CurrencyCode: c.CurrencyCode,
	}
}

func ToAdministrativeDivisionSimpleResponse(ad *commonv1.AdministrativeDivisionSimple) dto.AdministrativeDivisionSimpleResponse {
	if ad == nil {
		return dto.AdministrativeDivisionSimpleResponse{}
	}
	var parentID *string
	if ad.ParentId != "" {
		parentID = &ad.ParentId
	}
	return dto.AdministrativeDivisionSimpleResponse{
		ID:         ad.Id,
		CountryID:  ad.Country.GetId(),
		ParentID:   parentID,
		Name:       ad.Name,
		Level:      ad.Level,
		PostalCode: ad.PostalCode,
		Country:    ToCountrySimpleResponse(ad.Country),
	}
}

func ToAcademyHoldingAddressResponse(addr *academyv1.AcademyHoldingAddress) *dto.AcademyHoldingAddressResponse {
	if addr == nil {
		return nil
	}
	return &dto.AcademyHoldingAddressResponse{
		ID:                     addr.Id,
		StreetAddress:          addr.StreetAddress,
		Notes:                  addr.Notes,
		Latitude:               addr.Latitude,
		Longitude:              addr.Longitude,
		AdministrativeDivision: ToAdministrativeDivisionSimpleResponse(addr.AdministrativeDivision),
	}
}

func ToAcademyBranchAddressResponse(addr *academyv1.AcademyBranchAddress) dto.AcademyBranchAddressResponse {
	if addr == nil {
		return dto.AcademyBranchAddressResponse{}
	}
	return dto.AcademyBranchAddressResponse{
		ID:                     addr.Id,
		BranchID:               addr.BranchId,
		IsPrimary:              addr.IsPrimary,
		StreetAddress:          addr.StreetAddress,
		Notes:                  addr.Notes,
		Latitude:               addr.Latitude,
		Longitude:              addr.Longitude,
		AdministrativeDivision: ToAdministrativeDivisionSimpleResponse(addr.AdministrativeDivision),
	}
}

func ToSportSimpleResponse(s *commonv1.SportSimple) *dto.SportSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.SportSimpleResponse{
		ID:               s.Id,
		Name:             s.Name,
		Slug:             s.Slug,
		IconAttachmentID: s.IconAttachmentId,
		IsVerified:       s.IsVerified,
		RegulatorID:      s.RegulatorId,
		Tier:             s.Tier,
	}
}

func ToAcademyHoldingSimpleResponse(h *commonv1.AcademyHoldingSimple) *dto.AcademyHoldingSimpleResponse {
	if h == nil {
		return nil
	}
	return &dto.AcademyHoldingSimpleResponse{
		ID:                h.Id,
		Name:              h.Name,
		Description:       h.Description,
		Email:             h.Email,
		PhoneNumber:       h.PhoneNumber,
		ImageAttachmentID: h.ImageAttachmentId,
		BranchesCount:     h.BranchesCount,
		StatusID:          h.StatusId,
		CreatedAt:         FormatTimestamp(h.CreatedAt),
		UpdatedAt:         FormatTimestamp(h.UpdatedAt),
	}
}

func ToAcademyBranchSimpleResponse(b *commonv1.AcademyBranchSimple) *dto.AcademyBranchSimpleResponse {
	if b == nil {
		return nil
	}
	return &dto.AcademyBranchSimpleResponse{
		ID:          b.Id,
		HoldingID:   b.HoldingId,
		SportID:     b.SportId,
		SportName:   b.SportName,
		Name:        b.Name,
		Email:       b.Email,
		PhoneNumber: b.PhoneNumber,
		StatusID:    b.StatusId,
		MemberCount: b.MemberCount,
		CreatedAt:   FormatTimestamp(b.CreatedAt),
		UpdatedAt:   FormatTimestamp(b.UpdatedAt),
	}
}

func ToAcademyHoldingResponse(h *academyv1.AcademyHolding) dto.AcademyHoldingResponse {
	if h == nil {
		return dto.AcademyHoldingResponse{}
	}

	branches := make([]dto.AcademyBranchSimpleResponse, len(h.Branches))
	for i, b := range h.Branches {
		branches[i] = *ToAcademyBranchSimpleResponse(b)
	}

	var statusSimple *dto.StatusSimpleResponse
	if h.Status != nil {
		st := ToStatusSimpleResponse(h.Status)
		statusSimple = &st
	}

	var deletedBy *dto.UserSimpleResponse
	if h.DeletedBy != nil {
		db := ToUserSimpleResponse(h.DeletedBy)
		deletedBy = &db
	}

	return dto.AcademyHoldingResponse{
		ID:                h.Id,
		Name:              h.Name,
		Description:       h.Description,
		Email:             h.Email,
		PhoneNumber:       h.PhoneNumber,
		ImageAttachmentID: h.ImageAttachmentId,
		Branches:          branches,
		Address:           ToAcademyHoldingAddressResponse(h.Address),
		Status:            statusSimple,
		StatusID:          h.StatusId,
		CreatedAt:         FormatTimestamp(h.CreatedAt),
		UpdatedAt:         FormatTimestamp(h.UpdatedAt),
		CreatedBy:         ToUserSimpleResponse(h.CreatedBy),
		UpdatedBy:         ToUserSimpleResponse(h.UpdatedBy),
		DeletedBy:         deletedBy,
	}
}

func ToAcademyBranchResponse(b *academyv1.AcademyBranch) dto.AcademyBranchResponse {
	if b == nil {
		return dto.AcademyBranchResponse{}
	}

	addresses := make([]dto.AcademyBranchAddressResponse, len(b.Addresses))
	for i, addr := range b.Addresses {
		addresses[i] = ToAcademyBranchAddressResponse(addr)
	}

	var statusSimple *dto.StatusSimpleResponse
	if b.Status != nil {
		st := ToStatusSimpleResponse(b.Status)
		statusSimple = &st
	}

	var deletedBy *dto.UserSimpleResponse
	if b.DeletedBy != nil {
		db := ToUserSimpleResponse(b.DeletedBy)
		deletedBy = &db
	}

	return dto.AcademyBranchResponse{
		ID:          b.Id,
		HoldingID:   b.HoldingId,
		SportID:     b.SportId,
		SportName:   b.SportName,
		Name:        b.Name,
		Email:       b.Email,
		PhoneNumber: b.PhoneNumber,
		StatusID:    b.StatusId,
		MemberCount: b.MemberCount,
		CreatedAt:   FormatTimestamp(b.CreatedAt),
		UpdatedAt:   FormatTimestamp(b.UpdatedAt),
		Holding:     ToAcademyHoldingSimpleResponse(b.Holding),
		Sport:       ToSportSimpleResponse(b.Sport),
		Status:      statusSimple,
		Addresses:   addresses,
		CreatedBy:   ToUserSimpleResponse(b.CreatedBy),
		UpdatedBy:   ToUserSimpleResponse(b.UpdatedBy),
		DeletedBy:   deletedBy,
	}
}

func ToAcademyAdminResponse(a *academyv1.AcademyAdmin) dto.AcademyAdminResponse {
	if a == nil {
		return dto.AcademyAdminResponse{}
	}

	var approvedBy *dto.UserSimpleResponse
	if a.ApprovedBy != nil {
		ap := ToUserSimpleResponse(a.ApprovedBy)
		approvedBy = &ap
	}

	var deletedBy *dto.UserSimpleResponse
	if a.DeletedBy != nil {
		db := ToUserSimpleResponse(a.DeletedBy)
		deletedBy = &db
	}

	return dto.AcademyAdminResponse{
		ID:           a.Id,
		AcademyID:    a.AcademyId,
		BranchID:     a.BranchId,
		UserID:       a.UserId,
		RoleID:       a.RoleId,
		ApprovedAt:   FormatTimestampPtr(a.ApprovedAt),
		ApprovedByID: a.ApprovedById,
		CreatedAt:    FormatTimestamp(a.CreatedAt),
		UpdatedAt:    FormatTimestamp(a.UpdatedAt),
		DeletedAt:    FormatTimestampPtr(a.DeletedAt),
		Academy:      ToAcademyHoldingSimpleResponse(a.Academy),
		Branch:       ToAcademyBranchSimpleResponse(a.Branch),
		User:         ToUserSimpleResponse(a.User),
		Role:         ToRoleSimpleResponse(a.Role),
		ApprovedBy:   approvedBy,
		CreatedBy:    ToUserSimpleResponse(a.CreatedBy),
		UpdatedBy:    ToUserSimpleResponse(a.UpdatedBy),
		DeletedBy:    deletedBy,
	}
}

func ToEnrollmentResponse(e *academyv1.Enrollment) dto.EnrollmentResponse {
	if e == nil {
		return dto.EnrollmentResponse{}
	}

	var approvedBy *dto.UserSimpleResponse
	if e.ApprovedBy != nil {
		ap := ToUserSimpleResponse(e.ApprovedBy)
		approvedBy = &ap
	}

	var deletedBy *dto.UserSimpleResponse
	if e.DeletedBy != nil {
		db := ToUserSimpleResponse(e.DeletedBy)
		deletedBy = &db
	}

	return dto.EnrollmentResponse{
		ID:              e.Id,
		AcademyBranchID: e.AcademyBranchId,
		AthleteID:       e.AthleteId,
		JoinedAt:        FormatTimestamp(e.JoinedAt),
		LeftAt:          FormatTimestampPtr(e.LeftAt),
		ExpiresAt:       FormatTimestampPtr(e.ExpiresAt),
		ApprovedAt:      FormatTimestampPtr(e.ApprovedAt),
		ApprovedByID:    e.ApprovedById,
		StatusID:        e.StatusId,
		CreatedAt:       FormatTimestamp(e.CreatedAt),
		UpdatedAt:       FormatTimestamp(e.UpdatedAt),
		DeletedAt:       FormatTimestampPtr(e.DeletedAt),
		AcademyBranch:   ToAcademyBranchSimpleResponse(e.AcademyBranch),
		Athlete:         ToUserSimpleResponse(e.Athlete),
		ApprovedBy:      approvedBy,
		Status:          ToStatusSimpleResponse(e.Status),
		CreatedBy:       ToUserSimpleResponse(e.CreatedBy),
		UpdatedBy:       ToUserSimpleResponse(e.UpdatedBy),
		DeletedBy:       deletedBy,
	}
}

func ToRosterMemberResponse(m *academyv1.RosterMember) dto.RosterMemberResponse {
	if m == nil {
		return dto.RosterMemberResponse{}
	}

	var removedBy *dto.UserSimpleResponse
	if m.RemovedBy != nil {
		rb := ToUserSimpleResponse(m.RemovedBy)
		removedBy = &rb
	}

	return dto.RosterMemberResponse{
		ID:            m.Id,
		RosterID:      m.RosterId,
		AthleteID:     m.AthleteId,
		JerseyNumber:  m.JerseyNumber,
		PositionID:    m.PositionId,
		StatusID:      m.StatusId,
		AddedAt:       FormatTimestamp(m.AddedAt),
		AddedByID:     m.AddedById,
		RemovedAt:     FormatTimestampPtr(m.RemovedAt),
		RemovedByID:   m.RemovedById,
		RemovalReason: m.RemovalReason,
		Athlete:       ToUserSimpleResponse(m.Athlete),
		Position:      ToTagSimpleResponse(m.Position),
		Status:        ToStatusSimpleResponse(m.Status),
		AddedBy:       ToUserSimpleResponse(m.AddedBy),
		RemovedBy:     removedBy,
	}
}

func ToRosterResponse(r *academyv1.Roster) dto.RosterResponse {
	if r == nil {
		return dto.RosterResponse{}
	}

	members := make([]dto.RosterMemberResponse, len(r.Members))
	for i, m := range r.Members {
		members[i] = ToRosterMemberResponse(m)
	}

	return dto.RosterResponse{
		ID:              r.Id,
		AcademyBranchID: r.AcademyBranchId,
		CompetitionID:   r.CompetitionId,
		Name:            r.Name,
		TagID:           r.TagId,
		StatusID:        r.StatusId,
		MaxSize:         r.MaxSize,
		CreatedAt:       FormatTimestamp(r.CreatedAt),
		UpdatedAt:       FormatTimestamp(r.UpdatedAt),
		DeletedAt:       FormatTimestampPtr(r.DeletedAt),
		MemberCount:     r.MemberCount,
		AcademyBranch:   ToAcademyBranchSimpleResponse(r.AcademyBranch),
		Members:         members,
		Tag:             ToTagSimpleResponse(r.Tag),
		Status:          ToStatusSimpleResponse(r.Status),
	}
}
