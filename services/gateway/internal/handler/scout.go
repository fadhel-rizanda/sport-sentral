package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type ScoutHandler struct {
	client *client.ScoutClient
}

func NewScoutHandler(client *client.ScoutClient) *ScoutHandler {
	return &ScoutHandler{client: client}
}

func (h *ScoutHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	scouts := router.Group("/scouts", auth)

	// Scout Profile
	scouts.Post("/profile", h.CreateScoutProfile)
	scouts.Get("/profile", h.GetScoutProfile)
	scouts.Put("/profile/:id", h.UpdateScoutProfile)

	// Watchlist
	scouts.Post("/watchlist", h.AddToWatchlist)
	scouts.Get("/watchlist", h.ListWatchlist)
	scouts.Put("/watchlist/:id", h.UpdateWatchlistEntry)
	scouts.Delete("/watchlist/:athlete_id", h.RemoveFromWatchlist)

	// Athlete Profiles & Leaderboard
	scouts.Get("/athletes", h.ListAthleteProfiles)
	scouts.Get("/athletes/:id", h.GetAthleteProfile)
	scouts.Get("/leaderboard", h.GetLeaderboard)

	// Activity Logs
	scouts.Post("/logs", h.LogScoutActivity)
	scouts.Get("/logs", h.ListScoutActivities)
}

func (h *ScoutHandler) CreateScoutProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.CreateScoutProfileRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &scoutv1.CreateScoutProfileRequest{
		UserId:             userID,
		OrganizationName:   reqBody.OrganizationName,
		Bio:                reqBody.Bio,
		AvatarAttachmentId: reqBody.AvatarAttachment,
	}

	resp, err := h.client.Scout.CreateScoutProfile(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToScoutProfileResponseFromProto(resp.Scout))
}

func (h *ScoutHandler) GetScoutProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	resp, err := h.client.Scout.GetScoutProfile(c.Context(), &scoutv1.GetScoutProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToScoutProfileResponseFromProto(resp.Scout))
}

func (h *ScoutHandler) UpdateScoutProfile(c *fiber.Ctx) error {
	scoutID := c.Params("id")

	var reqBody dto.UpdateScoutProfileRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &scoutv1.UpdateScoutProfileRequest{
		ScoutId:            scoutID,
		OrganizationName:   reqBody.OrganizationName,
		Bio:                reqBody.Bio,
		AvatarAttachmentId: reqBody.AvatarAttachment,
	}

	resp, err := h.client.Scout.UpdateScoutProfile(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToScoutProfileResponseFromProto(resp.Scout))
}

func (h *ScoutHandler) AddToWatchlist(c *fiber.Ctx) error {
	scoutID := c.Query("scout_id", "")
	if scoutID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id parameter is required")
	}

	var reqBody dto.AddToWatchlistRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &scoutv1.AddToWatchlistRequest{
		ScoutId:       scoutID,
		AthleteId:     reqBody.AthleteID,
		Notes:         reqBody.Notes,
		PriorityTagId: reqBody.PriorityTagID,
	}

	resp, err := h.client.Scout.AddToWatchlist(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToWatchlistEntryResponseFromProto(resp.Entry))
}

func (h *ScoutHandler) RemoveFromWatchlist(c *fiber.Ctx) error {
	scoutID := c.Query("scout_id", "")
	athleteID := c.Params("athlete_id")
	if scoutID == "" || athleteID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id and athlete_id parameters are required")
	}

	_, err := h.client.Scout.RemoveFromWatchlist(c.Context(), &scoutv1.RemoveFromWatchlistRequest{
		ScoutId:   scoutID,
		AthleteId: athleteID,
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ScoutHandler) ListWatchlist(c *fiber.Ctx) error {
	scoutID := c.Query("scout_id", "")
	if scoutID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id parameter is required")
	}

	page, pageSize := request.ParsePagination(c)

	resp, err := h.client.Scout.ListWatchlist(c.Context(), &scoutv1.ListWatchlistRequest{
		ScoutId: scoutID,
		Page:    int32(page),
		Limit:   int32(pageSize),
	})
	if err != nil {
		return err
	}

	entries := make([]*dto.WatchlistEntryResponse, len(resp.Entries))
	for i, entry := range resp.Entries {
		entries[i] = mapper.ToWatchlistEntryResponseFromProto(entry)
	}

	return response.OK(c, fiber.Map{
		"entries":     entries,
		"total_hits":  resp.TotalHits,
		"page":        resp.Page,
		"total_pages": resp.TotalPages,
	})
}

func (h *ScoutHandler) UpdateWatchlistEntry(c *fiber.Ctx) error {
	entryID := c.Params("id")
	scoutID := c.Query("scout_id", "")
	if scoutID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id parameter is required")
	}

	var reqBody dto.UpdateWatchlistEntryRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &scoutv1.UpdateWatchlistEntryRequest{
		Id:            entryID,
		ScoutId:       scoutID,
		Notes:         reqBody.Notes,
		PriorityTagId: reqBody.PriorityTagID,
	}

	resp, err := h.client.Scout.UpdateWatchlistEntry(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToWatchlistEntryResponseFromProto(resp.Entry))
}

func (h *ScoutHandler) ListAthleteProfiles(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)

	req := &scoutv1.ListAthleteProfilesRequest{
		Page:  int32(page),
		Limit: int32(pageSize),
	}

	if sportID := c.Query("sport_id", ""); sportID != "" {
		req.SportId = &sportID
	}
	if minAgeStr := c.Query("min_age", ""); minAgeStr != "" {
		if minAge, err := strconv.Atoi(minAgeStr); err == nil {
			v := int32(minAge)
			req.MinAge = &v
		}
	}
	if maxAgeStr := c.Query("max_age", ""); maxAgeStr != "" {
		if maxAge, err := strconv.Atoi(maxAgeStr); err == nil {
			v := int32(maxAge)
			req.MaxAge = &v
		}
	}
	if level := c.Query("level", ""); level != "" {
		req.HighestCompetitionLevel = &level
	}

	resp, err := h.client.Scout.ListAthleteProfiles(c.Context(), req)
	if err != nil {
		return err
	}

	profiles := make([]*dto.AthleteProfileResponse, len(resp.Profiles))
	for i, p := range resp.Profiles {
		profiles[i] = mapper.ToAthleteProfileResponseFromProto(p)
	}

	return response.OK(c, fiber.Map{
		"profiles":    profiles,
		"total_hits":  resp.TotalHits,
		"page":        resp.Page,
		"total_pages": resp.TotalPages,
	})
}

func (h *ScoutHandler) GetAthleteProfile(c *fiber.Ctx) error {
	athleteID := c.Params("id")

	req := &scoutv1.GetAthleteProfileRequest{
		AthleteId: athleteID,
	}
	if sportID := c.Query("sport_id", ""); sportID != "" {
		req.SportId = &sportID
	}

	resp, err := h.client.Scout.GetAthleteProfile(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAthleteProfileResponseFromProto(resp.Profile))
}

func (h *ScoutHandler) GetLeaderboard(c *fiber.Ctx) error {
	sportID := c.Query("sport_id", "")
	if sportID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sport_id query parameter is required")
	}

	page, pageSize := request.ParsePagination(c)

	req := &scoutv1.GetLeaderboardRequest{
		SportId: sportID,
		Page:    int32(page),
		Limit:   int32(pageSize),
	}

	if periodTagID := c.Query("period_tag_id", ""); periodTagID != "" {
		req.PeriodTagId = &periodTagID
	}

	resp, err := h.client.Scout.GetLeaderboard(c.Context(), req)
	if err != nil {
		return err
	}

	entries := make([]*dto.LeaderboardEntryResponse, len(resp.Entries))
	for i, entry := range resp.Entries {
		entries[i] = mapper.ToLeaderboardEntryResponseFromProto(entry)
	}

	return response.OK(c, fiber.Map{
		"entries":     entries,
		"total_hits":  resp.TotalHits,
		"page":        resp.Page,
		"total_pages": resp.TotalPages,
	})
}

func (h *ScoutHandler) LogScoutActivity(c *fiber.Ctx) error {
	scoutID := c.Query("scout_id", "")
	if scoutID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id parameter is required")
	}

	var reqBody dto.LogScoutActivityRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	req := &scoutv1.LogScoutActivityRequest{
		ScoutId:      scoutID,
		Action:       reqBody.Action,
		ResourceId:   reqBody.ResourceID,
		ResourceType: reqBody.ResourceType,
	}

	_, err := h.client.Scout.LogScoutActivity(c.Context(), req)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ScoutHandler) ListScoutActivities(c *fiber.Ctx) error {
	scoutID := c.Query("scout_id", "")
	if scoutID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scout_id parameter is required")
	}

	page, pageSize := request.ParsePagination(c)

	resp, err := h.client.Scout.ListScoutActivities(c.Context(), &scoutv1.ListScoutActivitiesRequest{
		ScoutId: scoutID,
		Page:    int32(page),
		Limit:   int32(pageSize),
	})
	if err != nil {
		return err
	}

	logs := make([]*dto.ScoutActivityLogResponse, len(resp.Logs))
	for i, log := range resp.Logs {
		logs[i] = mapper.ToScoutActivityLogResponseFromProto(log)
	}

	return response.OK(c, fiber.Map{
		"logs":        logs,
		"total_hits":  resp.TotalHits,
		"page":        resp.Page,
		"total_pages": resp.TotalPages,
	})
}
