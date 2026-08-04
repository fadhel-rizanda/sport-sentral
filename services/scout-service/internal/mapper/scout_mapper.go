package mapper

import (
	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/services/scout-service/internal/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ScoutToProto(s *entity.Scout) *scoutv1.ScoutProfile {
	if s == nil {
		return nil
	}

	var avatarID *string
	if s.AvatarAttachmentID != nil {
		str := s.AvatarAttachmentID.String()
		avatarID = &str
	}

	return &scoutv1.ScoutProfile{
		Id:                 s.ID.String(),
		UserId:             s.UserID.String(),
		OrganizationName:   s.OrganizationName,
		Bio:                s.Bio,
		AvatarAttachmentId: avatarID,
		CreatedAt:          timestamppb.New(s.CreatedAt),
		UpdatedAt:          timestamppb.New(s.UpdatedAt),
	}
}

func WatchlistEntryToProto(e *entity.WatchlistEntry) *scoutv1.WatchlistEntry {
	if e == nil {
		return nil
	}

	var priorityTagID *string
	if e.PriorityTagID != (entity.WatchlistEntry{}).PriorityTagID && e.PriorityTagID.String() != "00000000-0000-0000-0000-000000000000" {
		str := e.PriorityTagID.String()
		priorityTagID = &str
	}

	return &scoutv1.WatchlistEntry{
		Id:             e.ID.String(),
		ScoutId:        e.ScoutID.String(),
		AthleteId:      e.AthleteID.String(),
		Notes:          e.Notes,
		PriorityTagId:  priorityTagID,
		AddedAt:        timestamppb.New(e.AddedAt),
		UpdatedAt:      timestamppb.New(e.UpdatedAt),
		AthleteProfile: AthleteProfileToProto(e.AthleteProfile),
	}
}

func AthleteProfileToProto(ap *entity.AthleteProfile) *scoutv1.AthleteProfile {
	if ap == nil {
		return nil
	}

	var academyID *string
	if ap.CurrentAcademyID != nil {
		str := ap.CurrentAcademyID.String()
		academyID = &str
	}

	return &scoutv1.AthleteProfile{
		Id:                      ap.ID.String(),
		AthleteId:               ap.AthleteID.String(),
		SportId:                 ap.SportID.String(),
		CurrentAcademyId:        academyID,
		CurrentAcademyName:      ap.CurrentAcademyName,
		Age:                     ap.Age,
		LeaderboardRank:         ap.LeaderboardRank,
		LeaderboardScore:        ap.LeaderboardScore,
		HighestCompetitionLevel: ap.HighestCompetitionLevel,
		TotalCompetitions:       ap.TotalCompetitions,
		TotalAcademies:          ap.TotalAcademies,
		YearsActive:             ap.YearsActive,
		CreatedAt:               timestamppb.New(ap.CreatedAt),
		UpdatedAt:               timestamppb.New(ap.UpdatedAt),
	}
}

func LeaderboardEntryToProto(le *entity.LeaderboardEntry) *scoutv1.LeaderboardEntry {
	if le == nil {
		return nil
	}

	var periodTagID *string
	if le.PeriodTagID != nil {
		str := le.PeriodTagID.String()
		periodTagID = &str
	}

	return &scoutv1.LeaderboardEntry{
		Id:             le.ID.String(),
		SportId:        le.SportID.String(),
		PeriodTagId:    periodTagID,
		Rank:           le.Rank,
		AthleteId:      le.AthleteID.String(),
		Score:          le.Score,
		PreviousScore:  le.PreviousScore,
		CalculatedAt:   timestamppb.New(le.CalculatedAt),
		AthleteProfile: AthleteProfileToProto(le.AthleteProfile),
	}
}

func ActivityLogToProto(l *entity.ScoutActivityLog) *scoutv1.ScoutActivityLog {
	if l == nil {
		return nil
	}

	var resourceID *string
	if l.ResourceID != nil {
		str := l.ResourceID.String()
		resourceID = &str
	}

	var resourceType *string
	if l.ResourceType != "" {
		resType := l.ResourceType
		resourceType = &resType
	}

	return &scoutv1.ScoutActivityLog{
		Id:           l.ID.String(),
		ScoutId:      l.ScoutID.String(),
		Action:       l.Action,
		ResourceId:   resourceID,
		ResourceType: resourceType,
		CreatedAt:    timestamppb.New(l.CreatedAt),
	}
}
