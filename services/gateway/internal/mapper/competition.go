package mapper

import (
	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToCompetitionResponse(c *competitionv1.Competition) dto.CompetitionResponse {
	if c == nil {
		return dto.CompetitionResponse{}
	}

	var sport *dto.SportSimpleResponse
	if c.Sport != nil {
		sport = ToSportSimpleResponse(c.Sport)
	}

	var hostAcademy *dto.AcademyBranchSimpleResponse
	if c.HostAcademy != nil {
		hostAcademy = ToAcademyBranchSimpleResponse(c.HostAcademy)
	}

	var tierTag *dto.TagSimpleResponse
	if c.TierTag != nil {
		tag := ToTagSimpleResponse(c.TierTag)
		tierTag = &tag
	}

	var status *dto.StatusSimpleResponse
	if c.Status != nil {
		stat := ToStatusSimpleResponse(c.Status)
		status = &stat
	}

	var createdBy *dto.UserSimpleResponse
	if c.CreatedBy != nil {
		user := ToUserSimpleResponse(c.CreatedBy)
		createdBy = &user
	}

	var updatedBy *dto.UserSimpleResponse
	if c.UpdatedBy != nil {
		user := ToUserSimpleResponse(c.UpdatedBy)
		updatedBy = &user
	}

	branches := make([]dto.CompetitionBranchResponse, len(c.Branches))
	for i, b := range c.Branches {
		branches[i] = ToCompetitionBranchResponse(b)
	}

	return dto.CompetitionResponse{
		ID:                  c.Id,
		SportID:             c.SportId,
		HostAcademyBranchID: c.HostAcademyBranchId,
		Name:                c.Name,
		Description:         c.Description,
		TierID:              c.TierId,
		StartDate:           FormatTimestamp(c.StartDate),
		EndDate:             FormatTimestampPtr(c.EndDate),
		StatusID:            c.StatusId,
		CreatedByID:         c.CreatedById,
		UpdatedByID:         c.UpdatedById,
		DeletedByID:         c.DeletedById,
		CreatedAt:           FormatTimestamp(c.CreatedAt),
		UpdatedAt:           FormatTimestamp(c.UpdatedAt),
		DeletedAt:           FormatTimestampPtr(c.DeletedAt),
		Sport:               sport,
		HostAcademy:         hostAcademy,
		TierTag:             tierTag,
		Status:              status,
		CreatedBy:           createdBy,
		UpdatedBy:           updatedBy,
		Branches:            branches,
	}
}

func ToCompetitionBranchResponse(b *competitionv1.Branch) dto.CompetitionBranchResponse {
	if b == nil {
		return dto.CompetitionBranchResponse{}
	}

	var parentBranch *dto.CompetitionBranchResponse
	if b.ParentBranch != nil {
		pb := ToCompetitionBranchResponse(b.ParentBranch)
		parentBranch = &pb
	}

	var status *dto.StatusSimpleResponse
	if b.Status != nil {
		stat := ToStatusSimpleResponse(b.Status)
		status = &stat
	}

	var createdBy *dto.UserSimpleResponse
	if b.CreatedBy != nil {
		user := ToUserSimpleResponse(b.CreatedBy)
		createdBy = &user
	}

	var updatedBy *dto.UserSimpleResponse
	if b.UpdatedBy != nil {
		user := ToUserSimpleResponse(b.UpdatedBy)
		updatedBy = &user
	}

	childBranches := make([]dto.CompetitionBranchResponse, len(b.ChildBranches))
	for i, cb := range b.ChildBranches {
		childBranches[i] = ToCompetitionBranchResponse(cb)
	}

	matches := make([]dto.MatchResponse, len(b.Matches))
	for i, m := range b.Matches {
		matches[i] = ToMatchResponse(m)
	}

	return dto.CompetitionBranchResponse{
		ID:             b.Id,
		CompetitionID:  b.CompetitionId,
		ParentBranchID: b.ParentBranchId,
		Name:           b.Name,
		Description:    b.Description,
		StatusID:       b.StatusId,
		CreatedByID:    b.CreatedById,
		UpdatedByID:    b.UpdatedById,
		DeletedByID:    b.DeletedById,
		CreatedAt:      FormatTimestamp(b.CreatedAt),
		UpdatedAt:      FormatTimestamp(b.UpdatedAt),
		DeletedAt:      FormatTimestampPtr(b.DeletedAt),
		ParentBranch:   parentBranch,
		Status:         status,
		ChildBranches:  childBranches,
		Matches:        matches,
		CreatedBy:      createdBy,
		UpdatedBy:      updatedBy,
	}
}

func ToMatchParticipantResponse(p *competitionv1.MatchParticipant) dto.MatchParticipantResponse {
	if p == nil {
		return dto.MatchParticipantResponse{}
	}

	var roster *dto.RosterResponse
	if p.Roster != nil {
		ros := ToRosterResponse(p.Roster)
		roster = &ros
	}

	var formatTag *dto.TagSimpleResponse
	if p.FormatTag != nil {
		tag := ToTagSimpleResponse(p.FormatTag)
		formatTag = &tag
	}

	var resultTag *dto.TagSimpleResponse
	if p.ResultTag != nil {
		tag := ToTagSimpleResponse(p.ResultTag)
		resultTag = &tag
	}

	return dto.MatchParticipantResponse{
		ID:          p.Id,
		MatchID:     p.MatchId,
		RosterID:    p.RosterId,
		FormatTagID: p.FormatTagId,
		ResultTagID: p.ResultTagId,
		Score:       p.Score,
		UpdatedAt:   FormatTimestamp(p.UpdatedAt),
		Roster:      roster,
		FormatTag:   formatTag,
		ResultTag:   resultTag,
	}
}

func ToMatchResponse(m *competitionv1.Match) dto.MatchResponse {
	if m == nil {
		return dto.MatchResponse{}
	}

	participants := make([]dto.MatchParticipantResponse, len(m.Participants))
	for i, p := range m.Participants {
		participants[i] = ToMatchParticipantResponse(p)
	}

	return dto.MatchResponse{
		ID:           m.Id,
		BranchID:     m.BranchId,
		ScheduledAt:  FormatTimestamp(m.ScheduledAt),
		StartedAt:    FormatTimestampPtr(m.StartedAt),
		EndedAt:      FormatTimestampPtr(m.EndedAt),
		Status:       m.Status,
		Location:     m.Location,
		Referee:      m.Referee,
		Notes:        m.Notes,
		CreatedAt:    FormatTimestamp(m.CreatedAt),
		UpdatedAt:    FormatTimestamp(m.UpdatedAt),
		DeletedAt:    FormatTimestampPtr(m.DeletedAt),
		Participants: participants,
	}
}

func ToMatchStatResponse(st *competitionv1.MatchStat) dto.MatchStatResponse {
	if st == nil {
		return dto.MatchStatResponse{}
	}

	var athlete *dto.UserSimpleResponse
	if st.Athlete != nil {
		user := ToUserSimpleResponse(st.Athlete)
		athlete = &user
	}

	var statType *dto.TagSimpleResponse
	if st.StatType != nil {
		tag := ToTagSimpleResponse(st.StatType)
		statType = &tag
	}

	var recordedBy *dto.UserSimpleResponse
	if st.RecordedBy != nil {
		user := ToUserSimpleResponse(st.RecordedBy)
		recordedBy = &user
	}

	return dto.MatchStatResponse{
		ID:           st.Id,
		MatchID:      st.MatchId,
		AthleteID:    st.AthleteId,
		StatTypeID:   st.StatTypeId,
		Value:        st.Value,
		RecordedByID: st.RecordedById,
		CreatedAt:    FormatTimestamp(st.CreatedAt),
		UpdatedAt:    FormatTimestamp(st.UpdatedAt),
		Athlete:      athlete,
		StatType:     statType,
		RecordedBy:   recordedBy,
	}
}

func ToAthleteStatsAggregateResponse(a *competitionv1.AthleteStatsAggregate) dto.AthleteStatsAggregateResponse {
	if a == nil {
		return dto.AthleteStatsAggregateResponse{}
	}

	var athlete *dto.UserSimpleResponse
	if a.Athlete != nil {
		user := ToUserSimpleResponse(a.Athlete)
		athlete = &user
	}

	var statType *dto.TagSimpleResponse
	if a.StatType != nil {
		tag := ToTagSimpleResponse(a.StatType)
		statType = &tag
	}

	return dto.AthleteStatsAggregateResponse{
		ID:            a.Id,
		CompetitionID: a.CompetitionId,
		BranchID:      a.BranchId,
		AthleteID:     a.AthleteId,
		StatTypeID:    a.StatTypeId,
		TotalMatches:  a.TotalMatches,
		SumValue:      a.SumValue,
		AvgValue:      a.AvgValue,
		MaxValue:      a.MaxValue,
		MinValue:      a.MinValue,
		LastUpdatedAt: FormatTimestamp(a.LastUpdatedAt),
		Athlete:       athlete,
		StatType:      statType,
	}
}
