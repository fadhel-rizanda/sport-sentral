package handler

import (
	"context"
	"math"

	"github.com/google/uuid"
	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/services/scout-service/internal/mapper"
	"microservice-golang/services/scout-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ScoutHandler struct {
	scoutv1.UnimplementedScoutServiceServer
	uc usecase.ScoutUsecase
}

func NewScoutHandler(uc usecase.ScoutUsecase) *ScoutHandler {
	return &ScoutHandler{uc: uc}
}

func (h *ScoutHandler) CreateScoutProfile(ctx context.Context, req *scoutv1.CreateScoutProfileRequest) (*scoutv1.ScoutProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user_id format"))
	}

	var avatarID *uuid.UUID
	if req.AvatarAttachmentId != nil && *req.AvatarAttachmentId != "" {
		id, err := uuid.Parse(*req.AvatarAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid avatar_attachment_id format"))
		}
		avatarID = &id
	}

	scout, err := h.uc.CreateScoutProfile(ctx, userID, req.GetOrganizationName(), req.GetBio(), avatarID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.ScoutProfileResponse{Scout: mapper.ScoutToProto(scout)}, nil
}

func (h *ScoutHandler) GetScoutProfile(ctx context.Context, req *scoutv1.GetScoutProfileRequest) (*scoutv1.ScoutProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid user_id format"))
	}

	scout, err := h.uc.GetScoutProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.ScoutProfileResponse{Scout: mapper.ScoutToProto(scout)}, nil
}

func (h *ScoutHandler) UpdateScoutProfile(ctx context.Context, req *scoutv1.UpdateScoutProfileRequest) (*scoutv1.ScoutProfileResponse, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	var avatarID *uuid.UUID
	if req.AvatarAttachmentId != nil && *req.AvatarAttachmentId != "" {
		id, err := uuid.Parse(*req.AvatarAttachmentId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid avatar_attachment_id format"))
		}
		avatarID = &id
	}

	scout, err := h.uc.UpdateScoutProfile(ctx, scoutID, req.OrganizationName, req.Bio, avatarID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.ScoutProfileResponse{Scout: mapper.ScoutToProto(scout)}, nil
}

// TODO: Tier & Membership management is handled via RBAC permissions (identity-service).
// Future Payment Service (Phase 4/5) will trigger RBAC role/permission assignments upon subscription events.

func (h *ScoutHandler) AddToWatchlist(ctx context.Context, req *scoutv1.AddToWatchlistRequest) (*scoutv1.WatchlistEntryResponse, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete_id format"))
	}

	var tagID *uuid.UUID
	if req.PriorityTagId != nil && *req.PriorityTagId != "" {
		id, err := uuid.Parse(*req.PriorityTagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid priority_tag_id format"))
		}
		tagID = &id
	}

	entry, err := h.uc.AddToWatchlist(ctx, scoutID, athleteID, req.GetNotes(), tagID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.WatchlistEntryResponse{Entry: mapper.WatchlistEntryToProto(entry)}, nil
}

func (h *ScoutHandler) RemoveFromWatchlist(ctx context.Context, req *scoutv1.RemoveFromWatchlistRequest) (*emptypb.Empty, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete_id format"))
	}

	if err := h.uc.RemoveFromWatchlist(ctx, scoutID, athleteID); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *ScoutHandler) ListWatchlist(ctx context.Context, req *scoutv1.ListWatchlistRequest) (*scoutv1.ListWatchlistResponse, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	entries, totalHits, err := h.uc.ListWatchlist(ctx, scoutID, page, limit)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	items := make([]*scoutv1.WatchlistEntry, len(entries))
	for i, entry := range entries {
		items[i] = mapper.WatchlistEntryToProto(&entry)
	}

	totalPages := int32(math.Ceil(float64(totalHits) / float64(limit)))

	return &scoutv1.ListWatchlistResponse{
		Entries:    items,
		TotalHits:  totalHits,
		Page:       int32(page),
		TotalPages: totalPages,
	}, nil
}

func (h *ScoutHandler) UpdateWatchlistEntry(ctx context.Context, req *scoutv1.UpdateWatchlistEntryRequest) (*scoutv1.WatchlistEntryResponse, error) {
	entryID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid entry id format"))
	}

	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	var tagID *uuid.UUID
	if req.PriorityTagId != nil && *req.PriorityTagId != "" {
		id, err := uuid.Parse(*req.PriorityTagId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid priority_tag_id format"))
		}
		tagID = &id
	}

	entry, err := h.uc.UpdateWatchlistEntry(ctx, entryID, scoutID, req.Notes, tagID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.WatchlistEntryResponse{Entry: mapper.WatchlistEntryToProto(entry)}, nil
}

func (h *ScoutHandler) ListAthleteProfiles(ctx context.Context, req *scoutv1.ListAthleteProfilesRequest) (*scoutv1.ListAthleteProfilesResponse, error) {
	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		id, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport_id format"))
		}
		sportID = &id
	}

	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	profiles, totalHits, err := h.uc.ListAthleteProfiles(ctx, sportID, req.MinAge, req.MaxAge, req.HighestCompetitionLevel, page, limit)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	items := make([]*scoutv1.AthleteProfile, len(profiles))
	for i, p := range profiles {
		items[i] = mapper.AthleteProfileToProto(&p)
	}

	totalPages := int32(math.Ceil(float64(totalHits) / float64(limit)))

	return &scoutv1.ListAthleteProfilesResponse{
		Profiles:   items,
		TotalHits:  totalHits,
		Page:       int32(page),
		TotalPages: totalPages,
	}, nil
}

func (h *ScoutHandler) GetAthleteProfile(ctx context.Context, req *scoutv1.GetAthleteProfileRequest) (*scoutv1.AthleteProfileResponse, error) {
	athleteID, err := uuid.Parse(req.GetAthleteId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid athlete_id format"))
	}

	var sportID *uuid.UUID
	if req.SportId != nil && *req.SportId != "" {
		id, err := uuid.Parse(*req.SportId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport_id format"))
		}
		sportID = &id
	}

	profile, err := h.uc.GetAthleteProfile(ctx, athleteID, sportID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &scoutv1.AthleteProfileResponse{Profile: mapper.AthleteProfileToProto(profile)}, nil
}

func (h *ScoutHandler) GetLeaderboard(ctx context.Context, req *scoutv1.GetLeaderboardRequest) (*scoutv1.GetLeaderboardResponse, error) {
	sportID, err := uuid.Parse(req.GetSportId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid sport_id format"))
	}

	var periodTagID *uuid.UUID
	if req.PeriodTagId != nil && *req.PeriodTagId != "" {
		id, err := uuid.Parse(req.GetPeriodTagId())
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid period_tag_id format"))
		}
		periodTagID = &id
	}

	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	entries, totalHits, err := h.uc.GetLeaderboard(ctx, sportID, periodTagID, page, limit)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	items := make([]*scoutv1.LeaderboardEntry, len(entries))
	for i, entry := range entries {
		items[i] = mapper.LeaderboardEntryToProto(&entry)
	}

	totalPages := int32(math.Ceil(float64(totalHits) / float64(limit)))

	return &scoutv1.GetLeaderboardResponse{
		Entries:    items,
		TotalHits:  totalHits,
		Page:       int32(page),
		TotalPages: totalPages,
	}, nil
}

func (h *ScoutHandler) LogScoutActivity(ctx context.Context, req *scoutv1.LogScoutActivityRequest) (*emptypb.Empty, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	var resourceID *uuid.UUID
	if req.ResourceId != nil && *req.ResourceId != "" {
		id, err := uuid.Parse(*req.ResourceId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid resource_id format"))
		}
		resourceID = &id
	}

	resType := ""
	if req.ResourceType != nil {
		resType = *req.ResourceType
	}

	if err := h.uc.LogScoutActivity(ctx, scoutID, req.GetAction(), resourceID, resType); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *ScoutHandler) ListScoutActivities(ctx context.Context, req *scoutv1.ListScoutActivitiesRequest) (*scoutv1.ListScoutActivitiesResponse, error) {
	scoutID, err := uuid.Parse(req.GetScoutId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid scout_id format"))
	}

	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	logs, totalHits, err := h.uc.ListActivityLogs(ctx, scoutID, page, limit)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	items := make([]*scoutv1.ScoutActivityLog, len(logs))
	for i, l := range logs {
		items[i] = mapper.ActivityLogToProto(&l)
	}

	totalPages := int32(math.Ceil(float64(totalHits) / float64(limit)))

	return &scoutv1.ListScoutActivitiesResponse{
		Logs:       items,
		TotalHits:  totalHits,
		Page:       int32(page),
		TotalPages: totalPages,
	}, nil
}
