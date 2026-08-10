package mapper

import (
	"time"

	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToScoutProfileResponseFromProto(s *scoutv1.ScoutProfile) *dto.ScoutProfileResponse {
	if s == nil {
		return nil
	}
	var avatar *string
	if s.AvatarAttachmentId != nil {
		avatar = s.AvatarAttachmentId
	}
	return &dto.ScoutProfileResponse{
		ID:               s.Id,
		UserID:           s.UserId,
		OrganizationName: s.OrganizationName,
		Bio:              s.Bio,
		AvatarAttachment: avatar,
		CreatedAt:        s.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:        s.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToAthleteProfileResponseFromProto(ap *scoutv1.AthleteProfile) *dto.AthleteProfileResponse {
	if ap == nil {
		return nil
	}
	var currentAcademyID *string
	if ap.CurrentAcademyId != nil {
		currentAcademyID = ap.CurrentAcademyId
	}
	return &dto.AthleteProfileResponse{
		ID:                      ap.Id,
		AthleteID:               ap.AthleteId,
		SportID:                 ap.SportId,
		CurrentAcademyID:        currentAcademyID,
		CurrentAcademyName:      ap.CurrentAcademyName,
		Age:                     ap.Age,
		LeaderboardRank:         ap.LeaderboardRank,
		LeaderboardScore:        ap.LeaderboardScore,
		HighestCompetitionLevel: ap.HighestCompetitionLevel,
		TotalCompetitions:       ap.TotalCompetitions,
		TotalAcademies:          ap.TotalAcademies,
		YearsActive:             ap.YearsActive,
		CreatedAt:               ap.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:               ap.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}

func ToWatchlistEntryResponseFromProto(w *scoutv1.WatchlistEntry) *dto.WatchlistEntryResponse {
	if w == nil {
		return nil
	}
	var priorityTagID *string
	if w.PriorityTagId != nil {
		priorityTagID = w.PriorityTagId
	}
	return &dto.WatchlistEntryResponse{
		ID:             w.Id,
		ScoutID:        w.ScoutId,
		AthleteID:      w.AthleteId,
		Notes:          w.Notes,
		PriorityTagID:  priorityTagID,
		AddedAt:        w.AddedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:      w.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
		AthleteProfile: ToAthleteProfileResponseFromProto(w.AthleteProfile),
	}
}

func ToLeaderboardEntryResponseFromProto(l *scoutv1.LeaderboardEntry) *dto.LeaderboardEntryResponse {
	if l == nil {
		return nil
	}
	var periodTagID *string
	if l.PeriodTagId != nil {
		periodTagID = l.PeriodTagId
	}
	return &dto.LeaderboardEntryResponse{
		ID:             l.Id,
		SportID:        l.SportId,
		PeriodTagID:    periodTagID,
		Rank:           l.Rank,
		AthleteID:      l.AthleteId,
		Score:          l.Score,
		PreviousScore:  l.PreviousScore,
		CalculatedAt:   l.CalculatedAt.AsTime().UTC().Format(time.RFC3339),
		AthleteProfile: ToAthleteProfileResponseFromProto(l.AthleteProfile),
	}
}

func ToScoutActivityLogResponseFromProto(log *scoutv1.ScoutActivityLog) *dto.ScoutActivityLogResponse {
	if log == nil {
		return nil
	}
	var resourceID *string
	if log.ResourceId != nil {
		resourceID = log.ResourceId
	}
	var resourceType *string
	if log.ResourceType != nil {
		resourceType = log.ResourceType
	}
	return &dto.ScoutActivityLogResponse{
		ID:           log.Id,
		ScoutID:      log.ScoutId,
		Action:       log.Action,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CreatedAt:    log.CreatedAt.AsTime().UTC().Format(time.RFC3339),
	}
}
