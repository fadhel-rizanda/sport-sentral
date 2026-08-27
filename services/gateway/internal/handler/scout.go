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

// CreateScoutProfile godoc
// @Summary      Create talent scout profile
// @Description  Create a new scout profile linked to the logged-in user account.
// @Tags         Scouts
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateScoutProfileRequest true "Scout Profile Data"
// @Success      201  {object}  response.Response{data=dto.ScoutProfileResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/profile [post]
// @Security     BearerAuth
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

// GetScoutProfile godoc
// @Summary      Get own talent scout profile
// @Description  Retrieve scout profile data for the logged-in account.
// @Tags         Scouts
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.ScoutProfileResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /scouts/profile [get]
// @Security     BearerAuth
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

// UpdateScoutProfile godoc
// @Summary      Update talent scout profile
// @Description  Update scout organization, bio, or avatar information.
// @Tags         Scouts
// @Accept       json
// @Produce      json
// @Param        id      path string                           true "Scout ID (UUID)"
// @Param        request body dto.UpdateScoutProfileRequest    true "Scout Profile Update Data"
// @Success      200  {object}  response.Response{data=dto.ScoutProfileResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/profile/{id} [put]
// @Security     BearerAuth
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

// AddToWatchlist godoc
// @Summary      Add athlete to watchlist
// @Description  Add an athlete to talent scout watchlist with notes and priority.
// @Tags         Scouts
// @Accept       json
// @Produce      json
// @Param        scout_id query string                      true "Scout ID (UUID)"
// @Param        request  body  dto.AddToWatchlistRequest   true "Athlete Watchlist Data"
// @Success      201  {object}  response.Response{data=dto.WatchlistEntryResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/watchlist [post]
// @Security     BearerAuth
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

// RemoveFromWatchlist godoc
// @Summary      Remove athlete from watchlist
// @Description  Remove an athlete from the scout watchlist.
// @Tags         Scouts
// @Produce      json
// @Param        athlete_id path  string true "Athlete ID (UUID)"
// @Param        scout_id   query string true "Scout ID (UUID)"
// @Success      204
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/watchlist/{athlete_id} [delete]
// @Security     BearerAuth
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

// ListWatchlist godoc
// @Summary      Get athlete watchlist
// @Description  Retrieve list of athletes monitored by scout with pagination.
// @Tags         Scouts
// @Produce      json
// @Param        scout_id  query string true  "Scout ID (UUID)"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/watchlist [get]
// @Security     BearerAuth
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

// UpdateWatchlistEntry godoc
// @Summary      Update watchlist entry notes
// @Description  Update notes or priority for an athlete in the watchlist.
// @Tags         Scouts
// @Accept       json
// @Produce      json
// @Param        id       path  string                              true "Watchlist Entry ID (UUID)"
// @Param        scout_id query string                              true "Scout ID (UUID)"
// @Param        request  body  dto.UpdateWatchlistEntryRequest     true "Watchlist Update Data"
// @Success      200  {object}  response.Response{data=dto.WatchlistEntryResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/watchlist/{id} [put]
// @Security     BearerAuth
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

// ListAthleteProfiles godoc
// @Summary      Search and filter athlete profiles
// @Description  Search athlete profile directory filtered by sport, age range, and highest competition level.
// @Tags         Scouts
// @Produce      json
// @Param        sport_id  query string false "Filter Sport ID (UUID)"
// @Param        min_age   query int    false "Minimum age"
// @Param        max_age   query int    false "Maximum age"
// @Param        level     query string false "Highest competition level"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/athletes [get]
// @Security     BearerAuth
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

// GetAthleteProfile godoc
// @Summary      Get athlete performance and statistics profile
// @Description  Retrieve career statistics, leaderboard rankings, competition history, and academy background for an athlete.
// @Tags         Scouts
// @Produce      json
// @Param        id       path  string  true  "Athlete User ID (UUID)"
// @Param        sport_id query string  false "Sport ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.AthleteProfileResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /scouts/athletes/{id} [get]
// @Security     BearerAuth
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

// GetLeaderboard godoc
// @Summary      Get athlete leaderboard rankings
// @Description  Retrieve top athlete rankings calculated from official tournament score weighting per sport.
// @Tags         Scouts
// @Produce      json
// @Param        sport_id      query string true  "Sport ID (UUID)"
// @Param        period_tag_id query string false "Period Tag ID (UUID)"
// @Param        page          query int    false "Page number (default 1)"
// @Param        page_size     query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/leaderboard [get]
// @Security     BearerAuth
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

// LogScoutActivity godoc
// @Summary      Log talent scout activity
// @Description  Record scout activity history when viewing/analyzing athlete profiles.
// @Tags         Scouts
// @Accept       json
// @Produce      json
// @Param        scout_id query string                         true "Scout ID (UUID)"
// @Param        request  body  dto.LogScoutActivityRequest    true "Activity Log Data"
// @Success      204
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/logs [post]
// @Security     BearerAuth
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

// ListScoutActivities godoc
// @Summary      Get talent scout activity history
// @Description  Retrieve activity logs performed by a talent scout.
// @Tags         Scouts
// @Produce      json
// @Param        scout_id  query string true  "Scout ID (UUID)"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /scouts/logs [get]
// @Security     BearerAuth
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
