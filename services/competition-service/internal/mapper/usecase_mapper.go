package mapper

import (
	"time"

	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/services/competition-service/internal/entity"
)

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

func ToAcademyBranchSimpleResponse(branch *entity.AcademyBranch) *dto.AcademyBranchSimpleResponse {
	if branch == nil {
		return nil
	}
	var sportName string
	if branch.Sport != nil {
		sportName = branch.Sport.Name
	}
	var email, phone string
	if branch.Holding != nil {
		email = branch.Holding.Email
		phone = branch.Holding.PhoneNumber
	}
	return &dto.AcademyBranchSimpleResponse{
		ID:          branch.ID.String(),
		HoldingID:   branch.HoldingID.String(),
		SportID:     branch.SportID.String(),
		SportName:   sportName,
		Name:        branch.Name,
		Email:       email,
		PhoneNumber: phone,
		StatusID:    branch.StatusID.String(),
		CreatedAt:   branch.CreatedAt,
		UpdatedAt:   branch.UpdatedAt,
	}
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

// ─── Sport Mapper ────────────────────────────────────────────────────────────

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

// ─── Competition Mapper ───────────────────────────────────────────────────────

func ToCompetitionResponse(c *entity.Competition) *dto.CompetitionResponse {
	if c == nil {
		return nil
	}

	var sportSimple *dto.SportSimpleResponse
	if c.Sport != nil {
		sportSimple = ToSportSimpleResponse(c.Sport)
	}

	var hostAcademySimple *dto.AcademyBranchSimpleResponse
	if c.HostAcademy != nil {
		hostAcademySimple = ToAcademyBranchSimpleResponse(c.HostAcademy)
	}

	var tierTagSimple *dto.TagSimpleResponse
	if c.TierTag != nil {
		tag := ToTagSimpleResponse(c.TierTag)
		tierTagSimple = &tag
	}

	var statusSimple *dto.StatusSimpleResponse
	if c.Status != nil {
		status := ToStatusSimpleResponse(c.Status)
		statusSimple = &status
	}

	var createdBySimple *dto.UserSimpleResponse
	if c.CreatedBy != nil {
		user := ToUserSimpleResponse(c.CreatedBy)
		createdBySimple = &user
	}

	var updatedBySimple *dto.UserSimpleResponse
	if c.UpdatedBy != nil {
		user := ToUserSimpleResponse(c.UpdatedBy)
		updatedBySimple = &user
	}

	branches := make([]dto.BranchResponse, len(c.Branches))
	for i, b := range c.Branches {
		branches[i] = *ToBranchResponse(&b)
	}

	var deletedAt *time.Time
	if c.DeletedAt.Valid {
		deletedAt = &c.DeletedAt.Time
	}

	return &dto.CompetitionResponse{
		ID:                  c.ID,
		SportID:             c.SportID,
		HostAcademyBranchID: c.HostAcademyBranchID,
		Name:                c.Name,
		Description:         c.Description,
		TierID:              c.TierID,
		StartDate:           c.StartDate,
		EndDate:             c.EndDate,
		StatusID:            c.StatusID,
		CreatedByID:         c.CreatedByID,
		UpdatedByID:         c.UpdatedByID,
		DeletedByID:         c.DeletedByID,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		DeletedAt:           deletedAt,
		Sport:               sportSimple,
		HostAcademy:         hostAcademySimple,
		TierTag:             tierTagSimple,
		Status:              statusSimple,
		CreatedBy:           createdBySimple,
		UpdatedBy:           updatedBySimple,
		Branches:            branches,
	}
}

// ─── Competition Branch Mapper ────────────────────────────────────────────────

func ToBranchResponse(b *entity.CompetitionBranch) *dto.BranchResponse {
	if b == nil {
		return nil
	}

	var parentBranchSimple *dto.BranchResponse
	if b.ParentBranch != nil {
		parentBranchSimple = ToBranchResponse(b.ParentBranch)
	}

	var statusSimple *dto.StatusSimpleResponse
	if b.Status != nil {
		status := ToStatusSimpleResponse(b.Status)
		statusSimple = &status
	}

	childBranches := make([]dto.BranchResponse, len(b.ChildBranches))
	for i, cb := range b.ChildBranches {
		childBranches[i] = *ToBranchResponse(&cb)
	}

	matches := make([]dto.MatchResponse, len(b.Matches))
	for i, m := range b.Matches {
		matches[i] = *ToMatchResponse(&m)
	}

	var createdBySimple *dto.UserSimpleResponse
	if b.CreatedBy != nil {
		user := ToUserSimpleResponse(b.CreatedBy)
		createdBySimple = &user
	}

	var updatedBySimple *dto.UserSimpleResponse
	if b.UpdatedBy != nil {
		user := ToUserSimpleResponse(b.UpdatedBy)
		updatedBySimple = &user
	}

	var deletedAt *time.Time
	if b.DeletedAt.Valid {
		deletedAt = &b.DeletedAt.Time
	}

	return &dto.BranchResponse{
		ID:             b.ID,
		CompetitionID:  b.CompetitionID,
		ParentBranchID: b.ParentBranchID,
		Name:           b.Name,
		Description:    b.Description,
		StatusID:       b.StatusID,
		CreatedByID:    b.CreatedByID,
		UpdatedByID:    b.UpdatedByID,
		DeletedByID:    b.DeletedByID,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		DeletedAt:      deletedAt,
		ParentBranch:   parentBranchSimple,
		Status:         statusSimple,
		ChildBranches:  childBranches,
		Matches:        matches,
		CreatedBy:      createdBySimple,
		UpdatedBy:      updatedBySimple,
	}
}

// ─── Match Mapper ────────────────────────────────────────────────────────────

func ToMatchResponse(m *entity.Match) *dto.MatchResponse {
	if m == nil {
		return nil
	}

	var branchSimple *dto.BranchResponse
	if m.Branch != nil {
		branchSimple = ToBranchResponse(m.Branch)
	}

	participants := make([]dto.MatchParticipantResponse, len(m.Participants))
	for i, p := range m.Participants {
		participants[i] = ToMatchParticipantResponse(p)
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &dto.MatchResponse{
		ID:           m.ID,
		BranchID:     m.BranchID,
		ScheduledAt:  m.ScheduledAt,
		StartedAt:    m.StartedAt,
		EndedAt:      m.EndedAt,
		Status:       m.Status,
		Location:     m.Location,
		Referee:      m.Referee,
		Notes:        m.Notes,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    deletedAt,
		Branch:       branchSimple,
		Participants: participants,
	}
}

func ToMatchParticipantResponse(p entity.MatchParticipant) dto.MatchParticipantResponse {
	var rosterSimple *dto.RosterResponse
	if p.Roster != nil {
		rosterSimple = ToRosterResponse(p.Roster)
	}

	var formatTagSimple *dto.TagSimpleResponse
	if p.FormatTag != nil {
		tag := ToTagSimpleResponse(p.FormatTag)
		formatTagSimple = &tag
	}

	var resultTagSimple *dto.TagSimpleResponse
	if p.ResultTag != nil {
		tag := ToTagSimpleResponse(p.ResultTag)
		resultTagSimple = &tag
	}

	return dto.MatchParticipantResponse{
		ID:          p.ID,
		MatchID:     p.MatchID,
		RosterID:    p.RosterID,
		FormatTagID: p.FormatTagID,
		ResultTagID: p.ResultTagID,
		Score:       p.Score,
		UpdatedAt:   p.UpdatedAt,
		Roster:      rosterSimple,
		FormatTag:   formatTagSimple,
		ResultTag:   resultTagSimple,
	}
}

// ─── Stat Mapper ─────────────────────────────────────────────────────────────

func ToMatchStatResponse(s *entity.MatchStat) *dto.MatchStatResponse {
	if s == nil {
		return nil
	}

	var athleteSimple *dto.UserSimpleResponse
	if s.Athlete != nil {
		user := ToUserSimpleResponse(s.Athlete)
		athleteSimple = &user
	}

	var tagSimple *dto.TagSimpleResponse
	if s.StatType != nil {
		tag := ToTagSimpleResponse(s.StatType)
		tagSimple = &tag
	}

	var recordedBySimple *dto.UserSimpleResponse
	if s.RecordedBy != nil {
		user := ToUserSimpleResponse(s.RecordedBy)
		recordedBySimple = &user
	}

	return &dto.MatchStatResponse{
		ID:           s.ID,
		MatchID:      s.MatchID,
		AthleteID:    s.AthleteID,
		StatTypeID:   s.StatTypeID,
		Value:        s.Value,
		RecordedByID: s.RecordedByID,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
		Athlete:      athleteSimple,
		StatType:     tagSimple,
		RecordedBy:   recordedBySimple,
	}
}

func ToAggregateResponse(a *entity.AthleteStatsAggregate) *dto.AggregateResponse {
	if a == nil {
		return nil
	}

	var athleteSimple *dto.UserSimpleResponse
	if a.Athlete != nil {
		user := ToUserSimpleResponse(a.Athlete)
		athleteSimple = &user
	}

	var tagSimple *dto.TagSimpleResponse
	if a.StatType != nil {
		tag := ToTagSimpleResponse(a.StatType)
		tagSimple = &tag
	}

	return &dto.AggregateResponse{
		ID:            a.ID,
		CompetitionID: a.CompetitionID,
		BranchID:      a.BranchID,
		AthleteID:     a.AthleteID,
		StatTypeID:    a.StatTypeID,
		TotalMatches:  a.TotalMatches,
		AvgValue:      a.AvgValue,
		MaxValue:      a.MaxValue,
		MinValue:      a.MinValue,
		LastUpdatedAt: a.LastUpdatedAt,
		Athlete:       athleteSimple,
		StatType:      tagSimple,
		SumValue:      a.SumValue,
	}
}
