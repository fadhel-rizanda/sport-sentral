package mapper

import (
	academyv1 "microservice-golang/gen/academy/v1"
	commonv1 "microservice-golang/gen/common/v1"
	"microservice-golang/services/academy-service/internal/dto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"microservice-golang/shared/pkg/utils"
)

// ─── Shared Helper Converters ──────────────────────────────────────────────────

func ToProtoUserSimple(u dto.UserSimpleResponse) *commonv1.UserSimple {
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToProtoStatusSimple(s dto.StatusSimpleResponse) *commonv1.StatusSimple {
	return &commonv1.StatusSimple{
		Id:   s.ID.String(),
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToProtoTagSimple(t dto.TagSimpleResponse) *commonv1.TagSimple {
	return &commonv1.TagSimple{
		Id:   t.ID.String(),
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToProtoPermissionSimple(p dto.PermissionSimpleResponse) *commonv1.PermissionSimple {
	return &commonv1.PermissionSimple{
		Id:       p.ID.String(),
		Resource: p.Resource,
		Action:   p.Action,
		Slug:     p.Slug,
	}
}

func ToProtoRoleSimple(r dto.RoleSimpleResponse) *commonv1.RoleSimple {
	permissions := make([]*commonv1.PermissionSimple, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = ToProtoPermissionSimple(p)
	}
	return &commonv1.RoleSimple{
		Id:          r.ID.String(),
		Name:        r.Name,
		Slug:        r.Slug,
		Permissions: permissions,
	}
}

func ToProtoSportSimple(s *dto.SportSimpleResponse) *commonv1.SportSimple {
	if s == nil {
		return nil
	}
	var iconID *string
	if s.IconAttachmentID != nil {
		str := s.IconAttachmentID.String()
		iconID = &str
	}
	var regulatorID *string
	if s.RegulatorID != nil {
		str := s.RegulatorID.String()
		regulatorID = &str
	}
	return &commonv1.SportSimple{
		Id:               s.ID.String(),
		Name:             s.Name,
		Slug:             s.Slug,
		IconAttachmentId: iconID,
		IsVerified:       s.IsVerified,
		RegulatorId:      regulatorID,
		Tier:             s.Tier,
	}
}

func ToProtoAdministrativeDivisionSimple(ad dto.AdministrativeDivisionSimpleResponse) *commonv1.AdministrativeDivisionSimple {
	var countrySimple *commonv1.CountrySimple
	if ad.Country.ID != uuid.Nil {
		countrySimple = &commonv1.CountrySimple{
			Id:           ad.Country.ID.String(),
			Name:         ad.Country.Name,
			IsoAlpha_2:   ad.Country.ISOAlpha2,
			IsoAlpha_3:   ad.Country.ISOAlpha3,
			PhoneCode:    ad.Country.PhoneCode,
			CurrencyCode: ad.Country.CurrencyCode,
		}
	}
	var parentID string
	if ad.ParentID != nil {
		parentID = ad.ParentID.String()
	}
	return &commonv1.AdministrativeDivisionSimple{
		Id:         ad.ID.String(),
		Name:       ad.Name,
		Level:      ad.Level,
		PostalCode: ad.PostalCode,
		Country:    countrySimple,
		ParentId:   parentID,
	}
}

// ─── Address Helpers ──────────────────────────────────────────────────────────

func ToProtoAcademyHoldingAddress(addr *dto.AcademyHoldingAddressResponse) *academyv1.AcademyHoldingAddress {
	if addr == nil {
		return nil
	}
	var notes *string
	if addr.Notes != nil {
		notes = addr.Notes
	}
	var lat *float64
	if addr.Latitude != nil {
		lat = addr.Latitude
	}
	var lon *float64
	if addr.Longitude != nil {
		lon = addr.Longitude
	}
	return &academyv1.AcademyHoldingAddress{
		Id:                     addr.ID.String(),
		StreetAddress:          addr.StreetAddress,
		Notes:                  notes,
		Latitude:               lat,
		Longitude:              lon,
		AdministrativeDivision: ToProtoAdministrativeDivisionSimple(addr.AdministrativeDivision),
	}
}

func ToProtoAcademyBranchAddress(addr dto.AcademyBranchAddressResponse) *academyv1.AcademyBranchAddress {
	var notes *string
	if addr.Notes != nil {
		notes = addr.Notes
	}
	var lat *float64
	if addr.Latitude != nil {
		lat = addr.Latitude
	}
	var lon *float64
	if addr.Longitude != nil {
		lon = addr.Longitude
	}
	return &academyv1.AcademyBranchAddress{
		Id:                     addr.ID.String(),
		BranchId:               addr.BranchID.String(),
		IsPrimary:              addr.IsPrimary,
		StreetAddress:          addr.StreetAddress,
		Notes:                  notes,
		Latitude:               lat,
		Longitude:              lon,
		AdministrativeDivision: ToProtoAdministrativeDivisionSimple(addr.AdministrativeDivision),
	}
}

// ─── Academy Holding ──────────────────────────────────────────────────────────

func ToProtoAcademyHolding(h *dto.AcademyHoldingResponse) *academyv1.AcademyHolding {
	if h == nil {
		return nil
	}

	branches := make([]*commonv1.AcademyBranchSimple, len(h.Branches))
	for i, b := range h.Branches {
		branches[i] = ToProtoAcademyBranchSimple(&b)
	}

	var statusSimple *commonv1.StatusSimple
	if h.Status != nil {
		statusSimple = ToProtoStatusSimple(*h.Status)
	}

	res := &academyv1.AcademyHolding{
		Id:                h.ID,
		Name:              h.Name,
		Description:       h.Description,
		Email:             h.Email,
		PhoneNumber:       h.PhoneNumber,
		ImageAttachmentId: h.ImageAttachmentID,
		Branches:          branches,
		Address:           ToProtoAcademyHoldingAddress(h.Address),
		Status:            statusSimple,
		StatusId:          h.StatusID,
		CreatedAt:         timestamppb.New(h.CreatedAt),
		UpdatedAt:         timestamppb.New(h.UpdatedAt),
		CreatedBy:         ToProtoUserSimple(h.CreatedBy),
		UpdatedBy:         ToProtoUserSimple(h.UpdatedBy),
	}

	if h.DeletedBy != nil {
		deletedBy := ToProtoUserSimple(*h.DeletedBy)
		res.DeletedBy = deletedBy
	}

	return res
}

func ToProtoAcademyHoldingSimple(h *dto.AcademyHoldingSimpleResponse) *commonv1.AcademyHoldingSimple {
	if h == nil {
		return nil
	}
	return &commonv1.AcademyHoldingSimple{
		Id:                h.ID,
		Name:              h.Name,
		Description:       h.Description,
		Email:             h.Email,
		PhoneNumber:       h.PhoneNumber,
		ImageAttachmentId: h.ImageAttachmentID,
		BranchesCount:     h.BranchesCount,
		StatusId:          h.StatusID,
		CreatedAt:         timestamppb.New(h.CreatedAt),
		UpdatedAt:         timestamppb.New(h.UpdatedAt),
	}
}

// ─── Academy Branch ───────────────────────────────────────────────────────────

func ToProtoAcademyBranch(b *dto.AcademyBranchResponse) *academyv1.AcademyBranch {
	if b == nil {
		return nil
	}

	var holdingSimple *commonv1.AcademyHoldingSimple
	if b.Holding != nil {
		holdingSimple = ToProtoAcademyHoldingSimple(b.Holding)
	}

	var statusSimple *commonv1.StatusSimple
	if b.Status != nil {
		statusSimple = ToProtoStatusSimple(*b.Status)
	}

	addresses := make([]*academyv1.AcademyBranchAddress, len(b.Addresses))
	for i, addr := range b.Addresses {
		addresses[i] = ToProtoAcademyBranchAddress(addr)
	}

	res := &academyv1.AcademyBranch{
		Id:          b.ID,
		HoldingId:   b.HoldingID,
		SportId:     b.SportID,
		SportName:   b.SportName,
		Name:        b.Name,
		Email:       b.Email,
		PhoneNumber: b.PhoneNumber,
		StatusId:    b.StatusID,
		MemberCount: b.MemberCount,
		Holding:     holdingSimple,
		Sport:       ToProtoSportSimple(b.Sport),
		Status:      statusSimple,
		Addresses:   addresses,
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
		CreatedBy:   ToProtoUserSimple(b.CreatedBy),
		UpdatedBy:   ToProtoUserSimple(b.UpdatedBy),
	}

	if b.DeletedBy != nil {
		deletedBy := ToProtoUserSimple(*b.DeletedBy)
		res.DeletedBy = deletedBy
	}

	return res
}

func ToProtoAcademyBranchSimple(b *dto.AcademyBranchSimpleResponse) *commonv1.AcademyBranchSimple {
	if b == nil {
		return nil
	}
	return &commonv1.AcademyBranchSimple{
		Id:          b.ID,
		HoldingId:   b.HoldingID,
		SportId:     b.SportID,
		SportName:   b.SportName,
		Name:        b.Name,
		Email:       b.Email,
		PhoneNumber: b.PhoneNumber,
		StatusId:    b.StatusID,
		MemberCount: b.MemberCount,
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
	}
}

// ─── Academy Admin ────────────────────────────────────────────────────────────

func ToProtoAcademyAdmin(a *dto.AcademyAdminResponse) *academyv1.AcademyAdmin {
	if a == nil {
		return nil
	}

	var academySimple *commonv1.AcademyHoldingSimple
	if a.Academy != nil {
		academySimple = ToProtoAcademyHoldingSimple(a.Academy)
	}

	var branchSimple *commonv1.AcademyBranchSimple
	if a.Branch != nil {
		branchSimple = ToProtoAcademyBranchSimple(a.Branch)
	}

	var approvedAt *timestamppb.Timestamp
	if a.ApprovedAt != nil {
		approvedAt = timestamppb.New(*a.ApprovedAt)
	}

	var approvedByID *string
	if a.ApprovedByID != nil {
		str := a.ApprovedByID.String()
		approvedByID = &str
	}

	var deletedAt *timestamppb.Timestamp
	if a.DeletedAt != nil {
		deletedAt = timestamppb.New(*a.DeletedAt)
	}

	res := &academyv1.AcademyAdmin{
		Id:           a.ID.String(),
		AcademyId:    a.AcademyID.String(),
		BranchId:     utils.UUIDPtrToStringPtr(a.BranchID),
		UserId:       a.UserID.String(),
		RoleId:       a.RoleID.String(),
		ApprovedAt:   approvedAt,
		ApprovedById: approvedByID,
		CreatedAt:    timestamppb.New(a.CreatedAt),
		UpdatedAt:    timestamppb.New(a.UpdatedAt),
		DeletedAt:    deletedAt,
		Academy:      academySimple,
		Branch:       branchSimple,
		User:         ToProtoUserSimple(a.User),
		Role:         ToProtoRoleSimple(a.Role),
		CreatedBy:    ToProtoUserSimple(a.CreatedBy),
		UpdatedBy:    ToProtoUserSimple(a.UpdatedBy),
	}

	if a.ApprovedBy != nil {
		apBy := ToProtoUserSimple(*a.ApprovedBy)
		res.ApprovedBy = apBy
	}
	if a.DeletedBy != nil {
		deletedBy := ToProtoUserSimple(*a.DeletedBy)
		res.DeletedBy = deletedBy
	}

	return res
}

// ─── Enrollment ───────────────────────────────────────────────────────────────

func ToProtoEnrollment(e *dto.EnrollmentResponse) *academyv1.Enrollment {
	if e == nil {
		return nil
	}

	var branchSimple *commonv1.AcademyBranchSimple
	if e.AcademyBranch != nil {
		branchSimple = ToProtoAcademyBranchSimple(e.AcademyBranch)
	}

	var leftAt *timestamppb.Timestamp
	if e.LeftAt != nil {
		leftAt = timestamppb.New(*e.LeftAt)
	}

	var expiresAt *timestamppb.Timestamp
	if e.ExpiresAt != nil {
		expiresAt = timestamppb.New(*e.ExpiresAt)
	}

	var approvedAt *timestamppb.Timestamp
	if e.ApprovedAt != nil {
		approvedAt = timestamppb.New(*e.ApprovedAt)
	}

	var approvedByID *string
	if e.ApprovedByID != nil {
		str := e.ApprovedByID.String()
		approvedByID = &str
	}

	var deletedAt *timestamppb.Timestamp
	if e.DeletedAt != nil {
		deletedAt = timestamppb.New(*e.DeletedAt)
	}

	res := &academyv1.Enrollment{
		Id:              e.ID.String(),
		AcademyBranchId: e.AcademyBranchID.String(),
		AthleteId:       e.AthleteID.String(),
		JoinedAt:        timestamppb.New(e.JoinedAt),
		LeftAt:          leftAt,
		ExpiresAt:       expiresAt,
		ApprovedAt:      approvedAt,
		ApprovedById:    approvedByID,
		StatusId:        e.StatusID.String(),
		CreatedAt:       timestamppb.New(e.CreatedAt),
		UpdatedAt:       timestamppb.New(e.UpdatedAt),
		DeletedAt:       deletedAt,
		AcademyBranch:   branchSimple,
		Athlete:         ToProtoUserSimple(e.Athlete),
		Status:          ToProtoStatusSimple(e.Status),
		CreatedBy:       ToProtoUserSimple(e.CreatedBy),
		UpdatedBy:       ToProtoUserSimple(e.UpdatedBy),
	}

	if e.ApprovedBy != nil {
		apBy := ToProtoUserSimple(*e.ApprovedBy)
		res.ApprovedBy = apBy
	}
	if e.DeletedBy != nil {
		deletedBy := ToProtoUserSimple(*e.DeletedBy)
		res.DeletedBy = deletedBy
	}

	return res
}

// ─── Utils ────────────────────────────────────────────────────────────────────
