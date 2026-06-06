package mapper

import (
	"microservice-golang/services/academy-service/internal/dto"
	"microservice-golang/services/academy-service/internal/entity"
	"time"
)

// ─── Replicated Helpers ───────────────────────────────────────────────────────

func ToUserSimpleResponse(u *entity.User) dto.UserSimpleResponse {
	if u == nil {
		return dto.UserSimpleResponse{}
	}
	return dto.UserSimpleResponse{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToStatusSimpleResponse(s *entity.Status) dto.StatusSimpleResponse {
	if s == nil {
		return dto.StatusSimpleResponse{}
	}
	return dto.StatusSimpleResponse{
		ID:   s.ID,
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToTagSimpleResponse(t *entity.Tag) dto.TagSimpleResponse {
	if t == nil {
		return dto.TagSimpleResponse{}
	}
	return dto.TagSimpleResponse{
		ID:   t.ID,
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToPermissionSimpleResponse(p entity.Permission) dto.PermissionSimpleResponse {
	return dto.PermissionSimpleResponse{
		ID:       p.ID,
		Resource: p.Resource,
		Action:   p.Action,
		Slug:     p.Slug,
	}
}

func ToRoleSimpleResponse(r *entity.Role) dto.RoleSimpleResponse {
	if r == nil {
		return dto.RoleSimpleResponse{}
	}
	permissions := make([]dto.PermissionSimpleResponse, len(r.Permissions))
	permissionsIDs := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		permissions[i] = ToPermissionSimpleResponse(p)
		permissionsIDs[i] = p.ID.String()
	}
	return dto.RoleSimpleResponse{
		ID:             r.ID,
		Name:           r.Name,
		Slug:           r.Slug,
		Permissions:    permissions,
		PermissionsIDs: permissionsIDs,
	}
}

func ToCountrySimpleResponse(c entity.Country) dto.CountrySimpleResponse {
	return dto.CountrySimpleResponse{
		ID:           c.ID,
		Name:         c.Name,
		ISOAlpha2:    c.ISOAlpha2,
		ISOAlpha3:    c.ISOAlpha3,
		PhoneCode:    c.PhoneCode,
		CurrencyCode: c.CurrencyCode,
	}
}

func ToAdministrativeDivisionSimpleResponse(ad *entity.AdministrativeDivision) dto.AdministrativeDivisionSimpleResponse {
	if ad == nil {
		return dto.AdministrativeDivisionSimpleResponse{}
	}
	res := dto.AdministrativeDivisionSimpleResponse{
		ID:         ad.ID,
		CountryID:  ad.CountryID,
		ParentID:   ad.ParentID,
		Name:       ad.Name,
		Level:      ad.Level,
		PostalCode: ad.PostalCode,
		Country:    ToCountrySimpleResponse(ad.Country),
	}
	if ad.Parent != nil {
		parent := ToAdministrativeDivisionSimpleResponse(ad.Parent)
		res.Parent = &parent
	}
	return res
}

func ToSportSimpleResponse(s *entity.Sport) *dto.SportSimpleResponse {
	if s == nil {
		return nil
	}
	return &dto.SportSimpleResponse{
		ID:               s.ID,
		Name:             s.Name,
		Slug:             s.Slug,
		IconAttachmentID: s.IconAttachmentID,
		IsVerified:       s.IsVerified,
		RegulatorID:      s.RegulatorID,
		Tier:             s.Tier,
	}
}

// ─── Address Helpers ──────────────────────────────────────────────────────────

func ToAcademyHoldingAddressResponse(addr *entity.AcademyHoldingAddress) *dto.AcademyHoldingAddressResponse {
	if addr == nil {
		return nil
	}
	res := &dto.AcademyHoldingAddressResponse{
		ID:                     addr.ID,
		StreetAddress:          addr.StreetAddress,
		Notes:                  addr.Notes,
		Latitude:               addr.Latitude,
		Longitude:              addr.Longitude,
		AdministrativeDivision: ToAdministrativeDivisionSimpleResponse(addr.AdministrativeDivision),
	}
	return res
}

func ToAcademyBranchAddressResponse(addr *entity.AcademyBranchAddress) dto.AcademyBranchAddressResponse {
	if addr == nil {
		return dto.AcademyBranchAddressResponse{}
	}
	res := dto.AcademyBranchAddressResponse{
		ID:                     addr.ID,
		BranchID:               addr.BranchID,
		IsPrimary:              addr.IsPrimary,
		StreetAddress:          addr.StreetAddress,
		Notes:                  addr.Notes,
		Latitude:               addr.Latitude,
		Longitude:              addr.Longitude,
		AdministrativeDivision: ToAdministrativeDivisionSimpleResponse(addr.AdministrativeDivision),
	}
	return res
}

// ─── Academy Holding Mapper ───────────────────────────────────────────────────

func ToAcademyHoldingResponse(holding *entity.AcademyHolding) *dto.AcademyHoldingResponse {
	if holding == nil {
		return nil
	}

	branches := make([]dto.AcademyBranchSimpleResponse, len(holding.Branches))
	for i, b := range holding.Branches {
		branches[i] = *ToAcademyBranchSimpleResponse(&b)
	}

	var statusSimple *dto.StatusSimpleResponse
	if holding.Status != nil {
		s := ToStatusSimpleResponse(holding.Status)
		statusSimple = &s
	}

	var imageAttachmentID *string
	if holding.ImageAttachmentID != nil {
		str := holding.ImageAttachmentID.String()
		imageAttachmentID = &str
	}

	res := &dto.AcademyHoldingResponse{
		ID:                holding.ID.String(),
		Name:              holding.Name,
		Description:       holding.Description,
		Email:             holding.Email,
		PhoneNumber:       holding.PhoneNumber,
		ImageAttachmentID: imageAttachmentID,
		Branches:          branches,
		Address:           ToAcademyHoldingAddressResponse(holding.Address),
		Status:            statusSimple,
		StatusID:          holding.StatusID.String(),
		CreatedAt:         holding.CreatedAt,
		UpdatedAt:         holding.UpdatedAt,
		CreatedBy:         ToUserSimpleResponse(holding.CreatedBy),
		UpdatedBy:         ToUserSimpleResponse(holding.UpdatedBy),
	}

	if holding.DeletedByID != nil && holding.DeletedBy != nil {
		deletedBy := ToUserSimpleResponse(holding.DeletedBy)
		res.DeletedBy = &deletedBy
	}

	return res
}

func ToAcademyHoldingSimpleResponse(holding *entity.AcademyHolding, branchesCount int64) *dto.AcademyHoldingSimpleResponse {
	if holding == nil {
		return nil
	}
	var imageAttachmentID *string
	if holding.ImageAttachmentID != nil {
		str := holding.ImageAttachmentID.String()
		imageAttachmentID = &str
	}
	return &dto.AcademyHoldingSimpleResponse{
		ID:                holding.ID.String(),
		Name:              holding.Name,
		Description:       holding.Description,
		Email:             holding.Email,
		PhoneNumber:       holding.PhoneNumber,
		ImageAttachmentID: imageAttachmentID,
		BranchesCount:     branchesCount,
		StatusID:          holding.StatusID.String(),
		CreatedAt:         holding.CreatedAt,
		UpdatedAt:         holding.UpdatedAt,
	}
}

// ─── Academy Branch Mapper ────────────────────────────────────────────────────

func ToAcademyBranchResponse(branch *entity.AcademyBranch) *dto.AcademyBranchResponse {
	if branch == nil {
		return nil
	}

	var sportName string
	if branch.Sport != nil {
		sportName = branch.Sport.Name
	}

	var statusSimple *dto.StatusSimpleResponse
	if branch.Status != nil {
		s := ToStatusSimpleResponse(branch.Status)
		statusSimple = &s
	}

	addresses := make([]dto.AcademyBranchAddressResponse, len(branch.Addresses))
	for i, addr := range branch.Addresses {
		addresses[i] = ToAcademyBranchAddressResponse(&addr)
	}

	var holdingSimple *dto.AcademyHoldingSimpleResponse
	if branch.Holding != nil {
		holdingSimple = &dto.AcademyHoldingSimpleResponse{
			ID:          branch.Holding.ID.String(),
			Name:        branch.Holding.Name,
			Description: branch.Holding.Description,
			Email:       branch.Holding.Email,
			PhoneNumber: branch.Holding.PhoneNumber,
			StatusID:    branch.Holding.StatusID.String(),
			CreatedAt:   branch.Holding.CreatedAt,
			UpdatedAt:   branch.Holding.UpdatedAt,
		}
	}

	res := &dto.AcademyBranchResponse{
		ID:          branch.ID.String(),
		HoldingID:   branch.HoldingID.String(),
		SportID:     branch.SportID.String(),
		SportName:   sportName,
		Name:        branch.Name,
		Email:       branch.Email,
		PhoneNumber: branch.PhoneNumber,
		StatusID:    branch.StatusID.String(),
		CreatedAt:   branch.CreatedAt,
		UpdatedAt:   branch.UpdatedAt,
		Holding:     holdingSimple,
		Sport:       ToSportSimpleResponse(branch.Sport),
		Status:      statusSimple,
		Addresses:   addresses,
		CreatedBy:   ToUserSimpleResponse(branch.CreatedBy),
		UpdatedBy:   ToUserSimpleResponse(branch.UpdatedBy),
	}

	if branch.DeletedByID != nil && branch.DeletedBy != nil {
		deletedBy := ToUserSimpleResponse(branch.DeletedBy)
		res.DeletedBy = &deletedBy
	}

	return res
}

func ToAcademyBranchSimpleResponse(branch *entity.AcademyBranch) *dto.AcademyBranchSimpleResponse {
	if branch == nil {
		return nil
	}
	var sportName string
	if branch.Sport != nil {
		sportName = branch.Sport.Name
	}
	return &dto.AcademyBranchSimpleResponse{
		ID:          branch.ID.String(),
		HoldingID:   branch.HoldingID.String(),
		SportID:     branch.SportID.String(),
		SportName:   sportName,
		Name:        branch.Name,
		Email:       branch.Email,
		PhoneNumber: branch.PhoneNumber,
		StatusID:    branch.StatusID.String(),
		CreatedAt:   branch.CreatedAt,
		UpdatedAt:   branch.UpdatedAt,
	}
}

// ─── Academy Admin Mapper ─────────────────────────────────────────────────────

func ToAcademyAdminResponse(admin *entity.AcademyAdmin) *dto.AcademyAdminResponse {
	if admin == nil {
		return nil
	}

	var academySimple *dto.AcademyHoldingSimpleResponse
	if admin.Academy != nil {
		academySimple = ToAcademyHoldingSimpleResponse(admin.Academy, 0)
	}

	var branchSimple *dto.AcademyBranchSimpleResponse
	if admin.Branch != nil {
		branchSimple = ToAcademyBranchSimpleResponse(admin.Branch)
	}

	var userSimple dto.UserSimpleResponse
	if admin.User != nil {
		userSimple = ToUserSimpleResponse(admin.User)
	}

	var roleSimple dto.RoleSimpleResponse
	if admin.Role != nil {
		roleSimple = ToRoleSimpleResponse(admin.Role)
	}

	var approvedBy *dto.UserSimpleResponse
	if admin.ApprovedBy != nil {
		apBy := ToUserSimpleResponse(admin.ApprovedBy)
		approvedBy = &apBy
	}

	var deletedAt *time.Time
	if admin.DeletedAt.Valid {
		deletedAt = &admin.DeletedAt.Time
	}

	res := &dto.AcademyAdminResponse{
		ID:           admin.ID,
		AcademyID:    admin.AcademyID,
		BranchID:     admin.BranchID,
		UserID:       admin.UserID,
		RoleID:       admin.RoleID,
		ApprovedAt:   admin.ApprovedAt,
		ApprovedByID: admin.ApprovedByID,
		CreatedAt:    admin.CreatedAt,
		UpdatedAt:    admin.UpdatedAt,
		DeletedAt:    deletedAt,
		Academy:      academySimple,
		Branch:       branchSimple,
		User:         userSimple,
		Role:         roleSimple,
		ApprovedBy:   approvedBy,
		CreatedBy:    ToUserSimpleResponse(admin.CreatedBy),
		UpdatedBy:    ToUserSimpleResponse(admin.UpdatedBy),
	}

	if admin.DeletedByID != nil && admin.DeletedBy != nil {
		deletedBy := ToUserSimpleResponse(admin.DeletedBy)
		res.DeletedBy = &deletedBy
	}

	return res
}

// ─── Enrollment Mapper ────────────────────────────────────────────────────────

func ToEnrollmentResponse(enrollment *entity.Enrollment) *dto.EnrollmentResponse {
	if enrollment == nil {
		return nil
	}

	var branchSimple *dto.AcademyBranchSimpleResponse
	if enrollment.AcademyBranch != nil {
		branchSimple = ToAcademyBranchSimpleResponse(enrollment.AcademyBranch)
	}

	var athleteSimple dto.UserSimpleResponse
	if enrollment.Athlete != nil {
		athleteSimple = ToUserSimpleResponse(enrollment.Athlete)
	}

	var approvedBy *dto.UserSimpleResponse
	if enrollment.ApprovedBy != nil {
		apBy := ToUserSimpleResponse(enrollment.ApprovedBy)
		approvedBy = &apBy
	}

	var statusSimple dto.StatusSimpleResponse
	if enrollment.Status != nil {
		statusSimple = ToStatusSimpleResponse(enrollment.Status)
	}

	var deletedAt *time.Time
	if enrollment.DeletedAt.Valid {
		deletedAt = &enrollment.DeletedAt.Time
	}

	res := &dto.EnrollmentResponse{
		ID:              enrollment.ID,
		AcademyBranchID: enrollment.AcademyBranchID,
		AthleteID:       enrollment.AthleteID,
		JoinedAt:        enrollment.JoinedAt,
		LeftAt:          enrollment.LeftAt,
		ExpiresAt:       enrollment.ExpiresAt,
		ApprovedAt:      enrollment.ApprovedAt,
		ApprovedByID:    enrollment.ApprovedByID,
		StatusID:        enrollment.StatusID,
		CreatedAt:       enrollment.CreatedAt,
		UpdatedAt:       enrollment.UpdatedAt,
		DeletedAt:       deletedAt,
		AcademyBranch:   branchSimple,
		Athlete:         athleteSimple,
		ApprovedBy:      approvedBy,
		Status:          statusSimple,
		CreatedBy:       ToUserSimpleResponse(enrollment.CreatedBy),
		UpdatedBy:       ToUserSimpleResponse(enrollment.UpdatedBy),
	}

	if enrollment.DeletedByID != nil && enrollment.DeletedBy != nil {
		deletedBy := ToUserSimpleResponse(enrollment.DeletedBy)
		res.DeletedBy = &deletedBy
	}

	return res
}

// ─── Roster Member Mapper ─────────────────────────────────────────────────────

func ToRosterMemberResponse(member *entity.RosterMember) dto.RosterMemberResponse {
	if member == nil {
		return dto.RosterMemberResponse{}
	}

	var athleteSimple dto.UserSimpleResponse
	if member.Athlete != nil {
		athleteSimple = ToUserSimpleResponse(member.Athlete)
	}

	var positionSimple dto.TagSimpleResponse
	if member.Position != nil {
		positionSimple = ToTagSimpleResponse(member.Position)
	}

	var statusSimple dto.StatusSimpleResponse
	if member.Status != nil {
		statusSimple = ToStatusSimpleResponse(member.Status)
	}

	var addedBy dto.UserSimpleResponse
	if member.AddedBy != nil {
		addedBy = ToUserSimpleResponse(member.AddedBy)
	}

	var removedBy *dto.UserSimpleResponse
	if member.RemovedBy != nil {
		remBy := ToUserSimpleResponse(member.RemovedBy)
		removedBy = &remBy
	}

	return dto.RosterMemberResponse{
		ID:            member.ID,
		RosterID:      member.RosterID,
		AthleteID:     member.AthleteID,
		JerseyNumber:  member.JerseyNumber,
		PositionID:    member.PositionID,
		StatusID:      member.StatusID,
		AddedAt:       member.AddedAt,
		AddedByID:     member.AddedByID,
		RemovedAt:     member.RemovedAt,
		RemovedByID:   member.RemovedByID,
		RemovalReason: member.RemovalReason,
		Athlete:       athleteSimple,
		Position:      positionSimple,
		Status:        statusSimple,
		AddedBy:       addedBy,
		RemovedBy:     removedBy,
	}
}

// ─── Roster Mapper ────────────────────────────────────────────────────────────

func ToRosterResponse(roster *entity.Roster) *dto.RosterResponse {
	if roster == nil {
		return nil
	}

	var branchSimple *dto.AcademyBranchSimpleResponse
	if roster.AcademyBranch != nil {
		branchSimple = ToAcademyBranchSimpleResponse(roster.AcademyBranch)
	}

	members := make([]dto.RosterMemberResponse, len(roster.Members))
	for i, m := range roster.Members {
		members[i] = ToRosterMemberResponse(&m)
	}

	var tagSimple dto.TagSimpleResponse
	if roster.Tag != nil {
		tagSimple = ToTagSimpleResponse(roster.Tag)
	}

	var statusSimple dto.StatusSimpleResponse
	if roster.Status != nil {
		statusSimple = ToStatusSimpleResponse(roster.Status)
	}

	var deletedAt *time.Time
	if roster.DeletedAt.Valid {
		deletedAt = &roster.DeletedAt.Time
	}

	return &dto.RosterResponse{
		ID:              roster.ID,
		AcademyBranchID: roster.AcademyBranchID,
		CompetitionID:   roster.CompetitionID,
		Name:            roster.Name,
		TagID:           roster.TagID,
		StatusID:        roster.StatusID,
		MaxSize:         roster.MaxSize,
		CreatedAt:       roster.CreatedAt,
		UpdatedAt:       roster.UpdatedAt,
		DeletedAt:       deletedAt,
		MemberCount:     roster.MemberCount,
		AcademyBranch:   branchSimple,
		Members:         members,
		Tag:             tagSimple,
		Status:          statusSimple,
	}
}
