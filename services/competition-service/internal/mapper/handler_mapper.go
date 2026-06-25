package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "microservice-golang/gen/common/v1"
	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/competition-service/internal/dto"
	"microservice-golang/shared/pkg/utils"
)

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

// ─── Roster Member ────────────────────────────────────────────────────────────

func ToProtoRosterMember(m dto.RosterMemberResponse) *competitionv1.RosterMember {
	var removedAt *timestamppb.Timestamp
	if m.RemovedAt != nil {
		removedAt = timestamppb.New(*m.RemovedAt)
	}

	var removedByID *string
	if m.RemovedByID != nil {
		str := m.RemovedByID.String()
		removedByID = &str
	}

	var removalReason *string
	if m.RemovalReason != nil {
		removalReason = m.RemovalReason
	}

	var jNum *int32
	if m.JerseyNumber != nil {
		val := int32(*m.JerseyNumber)
		jNum = &val
	}

	res := &competitionv1.RosterMember{
		Id:            m.ID.String(),
		RosterId:      m.RosterID.String(),
		AthleteId:     m.AthleteID.String(),
		JerseyNumber:  jNum,
		PositionId:    m.PositionID.String(),
		StatusId:      m.StatusID.String(),
		AddedAt:       timestamppb.New(m.AddedAt),
		AddedById:     m.AddedByID.String(),
		RemovedAt:     removedAt,
		RemovedById:   removedByID,
		RemovalReason: removalReason,
		Athlete:       ToProtoUserSimple(m.Athlete),
		Position:      ToProtoTagSimple(m.Position),
		Status:        ToProtoStatusSimple(m.Status),
		AddedBy:       ToProtoUserSimple(m.AddedBy),
	}

	if m.RemovedBy != nil {
		remBy := ToProtoUserSimple(*m.RemovedBy)
		res.RemovedBy = remBy
	}

	return res
}

// ─── Roster ───────────────────────────────────────────────────────────────────

func ToProtoRoster(r *dto.RosterResponse) *competitionv1.Roster {
	if r == nil {
		return nil
	}

	var branchSimple *commonv1.AcademyBranchSimple
	if r.AcademyBranch != nil {
		branchSimple = ToProtoAcademyBranchSimple(r.AcademyBranch)
	}

	members := make([]*competitionv1.RosterMember, len(r.Members))
	for i, m := range r.Members {
		members[i] = ToProtoRosterMember(m)
	}

	var deletedAt *timestamppb.Timestamp
	if r.DeletedAt != nil {
		deletedAt = timestamppb.New(*r.DeletedAt)
	}

	return &competitionv1.Roster{
		Id:              r.ID.String(),
		AcademyBranchId: r.AcademyBranchID.String(),
		CompetitionId:   utils.UUIDToStringPtr(r.CompetitionID),
		Name:            r.Name,
		TagId:           r.TagID.String(),
		StatusId:        r.StatusID.String(),
		MaxSize:         int32(r.MaxSize),
		CreatedAt:       timestamppb.New(r.CreatedAt),
		UpdatedAt:       timestamppb.New(r.UpdatedAt),
		DeletedAt:       deletedAt,
		MemberCount:     r.MemberCount,
		AcademyBranch:   branchSimple,
		Members:         members,
		Tag:             ToProtoTagSimple(r.Tag),
		Status:          ToProtoStatusSimple(r.Status),
	}
}

// ─── Competition ─────────────────────────────────────────────────────────────

func ToProtoCompetition(c *dto.CompetitionResponse) *competitionv1.Competition {
	if c == nil {
		return nil
	}

	var sport *commonv1.SportSimple
	if c.Sport != nil {
		sport = &commonv1.SportSimple{
			Id:               c.Sport.ID.String(),
			Name:             c.Sport.Name,
			Slug:             c.Sport.Slug,
			IconAttachmentId: utils.UUIDPtrToStringPtr(c.Sport.IconAttachmentID),
			IsVerified:       c.Sport.IsVerified,
			RegulatorId:      utils.UUIDPtrToStringPtr(c.Sport.RegulatorID),
			Tier:             c.Sport.Tier,
		}
	}

	var hostAcademy *commonv1.AcademyBranchSimple
	if c.HostAcademy != nil {
		hostAcademy = ToProtoAcademyBranchSimple(c.HostAcademy)
	}

	var tierTag *commonv1.TagSimple
	if c.TierTag != nil {
		tierTag = ToProtoTagSimple(*c.TierTag)
	}

	var status *commonv1.StatusSimple
	if c.Status != nil {
		status = ToProtoStatusSimple(*c.Status)
	}

	var createdBy *commonv1.UserSimple
	if c.CreatedBy != nil {
		createdBy = ToProtoUserSimple(*c.CreatedBy)
	}

	var updatedBy *commonv1.UserSimple
	if c.UpdatedBy != nil {
		updatedBy = ToProtoUserSimple(*c.UpdatedBy)
	}

	var deletedByID *string
	if c.DeletedByID != nil {
		str := c.DeletedByID.String()
		deletedByID = &str
	}

	var deletedAt *timestamppb.Timestamp
	if c.DeletedAt != nil {
		deletedAt = timestamppb.New(*c.DeletedAt)
	}

	var endDate *timestamppb.Timestamp
	if c.EndDate != nil {
		endDate = timestamppb.New(*c.EndDate)
	}

	branches := make([]*competitionv1.Branch, len(c.Branches))
	for i, b := range c.Branches {
		branches[i] = ToProtoBranch(&b)
	}

	return &competitionv1.Competition{
		Id:                  c.ID.String(),
		SportId:             c.SportID.String(),
		HostAcademyBranchId: utils.UUIDPtrToString(c.HostAcademyBranchID),
		Name:                c.Name,
		Description:         c.Description,
		TierId:              c.TierID.String(),
		StartDate:           timestamppb.New(c.StartDate),
		EndDate:             endDate,
		StatusId:            c.StatusID.String(),
		CreatedById:         c.CreatedByID.String(),
		UpdatedById:         c.UpdatedByID.String(),
		DeletedById:         deletedByID,
		CreatedAt:           timestamppb.New(c.CreatedAt),
		UpdatedAt:           timestamppb.New(c.UpdatedAt),
		DeletedAt:           deletedAt,
		Sport:               sport,
		HostAcademy:         hostAcademy,
		TierTag:             tierTag,
		Status:              status,
		CreatedBy:           createdBy,
		UpdatedBy:           updatedBy,
		Branches:            branches,
	}
}

// ─── Competition Branch ──────────────────────────────────────────────────────

func ToProtoBranch(b *dto.BranchResponse) *competitionv1.Branch {
	if b == nil {
		return nil
	}

	var parentBranch *competitionv1.Branch
	if b.ParentBranch != nil {
		parentBranch = ToProtoBranch(b.ParentBranch)
	}

	var status *commonv1.StatusSimple
	if b.Status != nil {
		status = ToProtoStatusSimple(*b.Status)
	}

	var createdBy *commonv1.UserSimple
	if b.CreatedBy != nil {
		createdBy = ToProtoUserSimple(*b.CreatedBy)
	}

	var updatedBy *commonv1.UserSimple
	if b.UpdatedBy != nil {
		updatedBy = ToProtoUserSimple(*b.UpdatedBy)
	}

	var parentBranchID *string
	if b.ParentBranchID != nil {
		str := b.ParentBranchID.String()
		parentBranchID = &str
	}

	var deletedByID *string
	if b.DeletedByID != nil {
		str := b.DeletedByID.String()
		deletedByID = &str
	}

	var deletedAt *timestamppb.Timestamp
	if b.DeletedAt != nil {
		deletedAt = timestamppb.New(*b.DeletedAt)
	}

	childBranches := make([]*competitionv1.Branch, len(b.ChildBranches))
	for i, cb := range b.ChildBranches {
		childBranches[i] = ToProtoBranch(&cb)
	}

	matches := make([]*competitionv1.Match, len(b.Matches))
	for i, m := range b.Matches {
		matches[i] = ToProtoMatch(&m)
	}

	return &competitionv1.Branch{
		Id:             b.ID.String(),
		CompetitionId:  b.CompetitionID.String(),
		ParentBranchId: parentBranchID,
		Name:           b.Name,
		Description:    b.Description,
		StatusId:       b.StatusID.String(),
		CreatedById:    b.CreatedByID.String(),
		UpdatedById:    b.UpdatedByID.String(),
		DeletedById:    deletedByID,
		CreatedAt:      timestamppb.New(b.CreatedAt),
		UpdatedAt:      timestamppb.New(b.UpdatedAt),
		DeletedAt:      deletedAt,
		ParentBranch:   parentBranch,
		Status:         status,
		ChildBranches:  childBranches,
		Matches:        matches,
		CreatedBy:      createdBy,
		UpdatedBy:      updatedBy,
	}
}

// ─── Match ───────────────────────────────────────────────────────────────────

func ToProtoMatch(m *dto.MatchResponse) *competitionv1.Match {
	if m == nil {
		return nil
	}

	var startedAt *timestamppb.Timestamp
	if m.StartedAt != nil {
		startedAt = timestamppb.New(*m.StartedAt)
	}

	var endedAt *timestamppb.Timestamp
	if m.EndedAt != nil {
		endedAt = timestamppb.New(*m.EndedAt)
	}

	var deletedAt *timestamppb.Timestamp
	if m.DeletedAt != nil {
		deletedAt = timestamppb.New(*m.DeletedAt)
	}

	participants := make([]*competitionv1.MatchParticipant, len(m.Participants))
	for i, p := range m.Participants {
		participants[i] = ToProtoMatchParticipant(p)
	}

	return &competitionv1.Match{
		Id:           m.ID.String(),
		BranchId:     m.BranchID.String(),
		ScheduledAt:  timestamppb.New(m.ScheduledAt),
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		Status:       m.Status,
		Location:     m.Location,
		Referee:      m.Referee,
		Notes:        m.Notes,
		CreatedAt:    timestamppb.New(m.CreatedAt),
		UpdatedAt:    timestamppb.New(m.UpdatedAt),
		DeletedAt:    deletedAt,
		Participants: participants,
	}
}

func ToProtoMatchParticipant(p dto.MatchParticipantResponse) *competitionv1.MatchParticipant {
	var roster *competitionv1.Roster
	if p.Roster != nil {
		roster = ToProtoRoster(p.Roster)
	}

	var formatTag *commonv1.TagSimple
	if p.FormatTag != nil {
		formatTag = ToProtoTagSimple(*p.FormatTag)
	}

	var resultTag *commonv1.TagSimple
	if p.ResultTag != nil {
		resultTag = ToProtoTagSimple(*p.ResultTag)
	}

	return &competitionv1.MatchParticipant{
		Id:          p.ID.String(),
		MatchId:     p.MatchID.String(),
		RosterId:    p.RosterID.String(),
		FormatTagId: p.FormatTagID.String(),
		ResultTagId: p.ResultTagID.String(),
		Score:       p.Score,
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
		Roster:      roster,
		FormatTag:   formatTag,
		ResultTag:   resultTag,
	}
}

// ─── Stat ─────────────────────────────────────────────────────────────────────

func ToProtoMatchStat(s *dto.MatchStatResponse) *competitionv1.MatchStat {
	if s == nil {
		return nil
	}

	var athlete *commonv1.UserSimple
	if s.Athlete != nil {
		athlete = ToProtoUserSimple(*s.Athlete)
	}

	var statType *commonv1.TagSimple
	if s.StatType != nil {
		statType = ToProtoTagSimple(*s.StatType)
	}

	var recordedBy *commonv1.UserSimple
	if s.RecordedBy != nil {
		recordedBy = ToProtoUserSimple(*s.RecordedBy)
	}

	return &competitionv1.MatchStat{
		Id:           s.ID.String(),
		MatchId:      s.MatchID.String(),
		AthleteId:    s.AthleteID.String(),
		StatTypeId:   s.StatTypeID.String(),
		Value:        s.Value,
		RecordedById: s.RecordedByID.String(),
		CreatedAt:    timestamppb.New(s.CreatedAt),
		UpdatedAt:    timestamppb.New(s.UpdatedAt),
		Athlete:      athlete,
		StatType:     statType,
		RecordedBy:   recordedBy,
	}
}

func ToProtoAthleteStatsAggregate(a *dto.AggregateResponse) *competitionv1.AthleteStatsAggregate {
	if a == nil {
		return nil
	}

	var athlete *commonv1.UserSimple
	if a.Athlete != nil {
		athlete = ToProtoUserSimple(*a.Athlete)
	}

	var statType *commonv1.TagSimple
	if a.StatType != nil {
		statType = ToProtoTagSimple(*a.StatType)
	}

	return &competitionv1.AthleteStatsAggregate{
		Id:            a.ID.String(),
		CompetitionId: a.CompetitionID.String(),
		BranchId:      a.BranchID.String(),
		AthleteId:     a.AthleteID.String(),
		StatTypeId:    a.StatTypeID.String(),
		TotalMatches:  a.TotalMatches,
		AvgValue:      a.AvgValue,
		MaxValue:      a.MaxValue,
		MinValue:      a.MinValue,
		LastUpdatedAt: timestamppb.New(a.LastUpdatedAt),
		Athlete:       athlete,
		StatType:      statType,
		SumValue:      a.SumValue,
	}
}

// ─── Utils ────────────────────────────────────────────────────────────────────
