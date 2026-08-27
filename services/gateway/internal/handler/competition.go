package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type CompetitionHandler struct {
	client *client.CompetitionClient
}

func NewCompetitionHandler(client *client.CompetitionClient) *CompetitionHandler {
	return &CompetitionHandler{client: client}
}

func (h *CompetitionHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	// Competitions
	competitions := router.Group("/competitions", auth)
	competitions.Get("/", h.ListCompetitions)
	competitions.Get("/:id", h.GetCompetition)
	competitions.Get("/:id/branches", h.ListBranchesByCompetition)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminCompetitions := router.Group("/admin/competitions", adminMiddlewares...)
	adminCompetitions.Post("/", h.CreateCompetition)
	adminCompetitions.Put("/:id", h.UpdateCompetition)
	adminCompetitions.Delete("/:id", h.DeleteCompetition)
	adminCompetitions.Put("/:id/status", h.UpdateCompetitionStatus)

	// Branches
	branches := router.Group("/branches", auth)
	branches.Get("/:id", h.GetBranch)
	branches.Get("/:id/matches", h.ListMatchesByBranch)

	adminBranches := router.Group("/admin/branches", adminMiddlewares...)
	adminBranches.Post("/", h.CreateBranch)
	adminBranches.Put("/:id", h.UpdateBranch)
	adminBranches.Delete("/:id", h.DeleteBranch)

	// Matches
	matches := router.Group("/matches", auth)
	matches.Get("/:id", h.GetMatch)
	matches.Get("/:id/stats", h.GetMatchStats)

	adminMatches := router.Group("/admin/matches", adminMiddlewares...)
	adminMatches.Post("/", h.CreateMatch)
	adminMatches.Put("/:id/status", h.UpdateMatchStatus)
	adminMatches.Put("/:id/score", h.UpdateMatchScore)
	adminMatches.Delete("/:id", h.DeleteMatch)

	// Stats
	stats := router.Group("/stats", auth)
	stats.Get("/athlete/:athleteId", h.GetAthleteAggregate)

	adminStats := router.Group("/admin/stats", adminMiddlewares...)
	adminStats.Post("/", h.RecordMatchStat)
	adminStats.Put("/:id", h.UpdateMatchStat)
	adminStats.Delete("/:id", h.DeleteMatchStat)
	adminStats.Post("/athlete/:athleteId/recalculate", h.RecalculateAggregate)
}

// ─── COMPETITION HANDLERS ────────────────────────────────────────────────────

// CreateCompetition godoc
// @Summary      Create new tournament / competition (Admin)
// @Description  Register a new championship / tournament event into the system.
// @Tags         Competitions
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCompetitionRequest true "New Competition Data"
// @Success      201  {object}  response.Response{data=dto.CompetitionResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/competitions [post]
// @Security     BearerAuth
func (h *CompetitionHandler) CreateCompetition(c *fiber.Ctx) error {
	var body dto.CreateCompetitionRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	startTime, err := time.Parse(time.RFC3339, body.StartDate)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid start_date format, must be RFC3339")
	}

	var endTime *timestamppb.Timestamp
	if body.EndDate != nil && *body.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *body.EndDate)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid end_date format, must be RFC3339")
		}
		endTime = timestamppb.New(t)
	}

	resp, err := h.client.Competition.CreateCompetition(c.Context(), &competitionv1.CreateCompetitionRequest{
		SportId:             body.SportID,
		HostAcademyBranchId: body.HostAcademyBranchID,
		Name:                body.Name,
		Description:         body.Description,
		TierId:              body.TierID,
		StartDate:           timestamppb.New(startTime),
		EndDate:             endTime,
		StatusId:            body.StatusID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToCompetitionResponse(resp.Competition))
}

// GetCompetition godoc
// @Summary      Get competition details
// @Description  Retrieve competition profile data including schedule and host.
// @Tags         Competitions
// @Produce      json
// @Param        id   path      string  true  "Competition ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.CompetitionResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /competitions/{id} [get]
// @Security     BearerAuth
func (h *CompetitionHandler) GetCompetition(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Competition.GetCompetition(c.Context(), &competitionv1.GetCompetitionRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCompetitionResponse(resp.Competition))
}

// ListCompetitions godoc
// @Summary      Get list of competitions
// @Description  Retrieve list of competitions with filters for sport, tier, host academy branch, or status.
// @Tags         Competitions
// @Produce      json
// @Param        branch_id query string false "Filter Host Academy Branch ID (UUID)"
// @Param        sport_id  query string false "Filter Sport ID (UUID)"
// @Param        tier_id   query string false "Filter Tier ID (UUID)"
// @Param        status_id query string false "Filter Status ID (UUID)"
// @Param        search    query string false "Search competition name"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.CompetitionResponse}
// @Failure      401  {object}  response.Response
// @Router       /competitions [get]
// @Security     BearerAuth
func (h *CompetitionHandler) ListCompetitions(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	branchID := c.Query("branch_id", "")
	sportID := c.Query("sport_id", "")
	tierID := c.Query("tier_id", "")
	statusID := c.Query("status_id", "")
	search := c.Query("search", "")

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	var sportIDPtr *string
	if sportID != "" {
		sportIDPtr = &sportID
	}

	var tierIDPtr *string
	if tierID != "" {
		tierIDPtr = &tierID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	resp, err := h.client.Competition.ListCompetitions(c.Context(), &competitionv1.ListCompetitionsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		BranchId: branchIDPtr,
		SportId:  sportIDPtr,
		TierId:   tierIDPtr,
		StatusId: statusIDPtr,
		Search:   searchPtr,
	})
	if err != nil {
		return err
	}

	competitions := make([]dto.CompetitionResponse, len(resp.Competitions))
	for i, comp := range resp.Competitions {
		competitions[i] = mapper.ToCompetitionResponse(comp)
	}

	return response.OKWithMeta(c, competitions, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

// UpdateCompetition godoc
// @Summary      Update competition (Admin)
// @Description  Update competition name, dates, or description.
// @Tags         Competitions
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Competition ID (UUID)"
// @Param        request body dto.UpdateCompetitionRequest true "Competition Update Data"
// @Success      200  {object}  response.Response{data=dto.CompetitionResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/competitions/{id} [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateCompetition(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdateCompetitionRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	var startDate *timestamppb.Timestamp
	if body.StartDate != nil && *body.StartDate != "" {
		t, err := time.Parse(time.RFC3339, *body.StartDate)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid start_date format, must be RFC3339")
		}
		startDate = timestamppb.New(t)
	}

	var endDate *timestamppb.Timestamp
	if body.EndDate != nil && *body.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *body.EndDate)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid end_date format, must be RFC3339")
		}
		endDate = timestamppb.New(t)
	}

	resp, err := h.client.Competition.UpdateCompetition(c.Context(), &competitionv1.UpdateCompetitionRequest{
		Id:                  id,
		SportId:             body.SportID,
		HostAcademyBranchId: body.HostAcademyBranchID,
		Name:                body.Name,
		Description:         body.Description,
		TierId:              body.TierID,
		StartDate:           startDate,
		EndDate:             endDate,
		StatusId:            body.StatusID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCompetitionResponse(resp.Competition))
}

// DeleteCompetition godoc
// @Summary      Delete competition (Admin)
// @Description  Delete competition data and its branches.
// @Tags         Competitions
// @Produce      json
// @Param        id   path      string  true  "Competition ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/competitions/{id} [delete]
// @Security     BearerAuth
func (h *CompetitionHandler) DeleteCompetition(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Competition.DeleteCompetition(c.Context(), &competitionv1.DeleteCompetitionRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "competition deleted successfully")
}

// UpdateCompetitionStatus godoc
// @Summary      Update competition status (Admin)
// @Description  Change competition phase status (e.g. Registration, Ongoing, Completed).
// @Tags         Competitions
// @Accept       json
// @Produce      json
// @Param        id      path string true "Competition ID (UUID)"
// @Param        request body object true "New Status ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/competitions/{id}/status [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateCompetitionStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		StatusID string `json:"status_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.client.Competition.UpdateCompetitionStatus(c.Context(), &competitionv1.UpdateCompetitionStatusRequest{
		Id:       id,
		StatusId: body.StatusID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "competition status updated successfully")
}

// ─── BRANCH HANDLERS ─────────────────────────────────────────────────────────

// CreateBranch godoc
// @Summary      Create competition category/stage (Admin)
// @Description  Register a phase/stage (e.g. Group Stage, Semifinals, U-16 Division) for a competition.
// @Tags         Competitions
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCompetitionBranchRequest true "Competition Category/Stage Data"
// @Success      201  {object}  response.Response{data=dto.CompetitionBranchResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/branches [post]
// @Security     BearerAuth
func (h *CompetitionHandler) CreateBranch(c *fiber.Ctx) error {
	var body dto.CreateCompetitionBranchRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.Branch.CreateBranch(c.Context(), &competitionv1.CreateBranchRequest{
		CompetitionId:  body.CompetitionID,
		Name:           body.Name,
		Description:    body.Description,
		StatusId:       body.StatusID,
		ParentBranchId: body.ParentBranchID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToCompetitionBranchResponse(resp.Branch))
}

// GetBranch godoc
// @Summary      Get competition stage details
// @Description  Retrieve detailed stage/category information for a competition by ID.
// @Tags         Competitions
// @Produce      json
// @Param        id   path      string  true  "Branch ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.CompetitionBranchResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /branches/{id} [get]
// @Security     BearerAuth
func (h *CompetitionHandler) GetBranch(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Branch.GetBranch(c.Context(), &competitionv1.GetBranchRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCompetitionBranchResponse(resp.Branch))
}

// ListBranchesByCompetition godoc
// @Summary      Get list of stages/categories for a competition
// @Tags         Competitions
// @Produce      json
// @Param        id path string true "Competition ID (UUID)"
// @Success      200  {object}  response.Response{data=[]dto.CompetitionBranchResponse}
// @Failure      401  {object}  response.Response
// @Router       /competitions/{id}/branches [get]
// @Security     BearerAuth
func (h *CompetitionHandler) ListBranchesByCompetition(c *fiber.Ctx) error {
	competitionID := c.Params("id")

	resp, err := h.client.Branch.ListBranchesByCompetition(c.Context(), &competitionv1.ListBranchesByCompetitionRequest{
		CompetitionId: competitionID,
	})
	if err != nil {
		return err
	}

	branches := make([]dto.CompetitionBranchResponse, len(resp.Branches))
	for i, b := range resp.Branches {
		branches[i] = mapper.ToCompetitionBranchResponse(b)
	}

	return response.OK(c, branches)
}

// UpdateBranch godoc
// @Summary      Update competition stage (Admin)
// @Tags         Competitions
// @Accept       json
// @Produce      json
// @Param        id      path string                              true "Branch ID (UUID)"
// @Param        request body dto.UpdateCompetitionBranchRequest true "Stage Update Data"
// @Success      200  {object}  response.Response{data=dto.CompetitionBranchResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/branches/{id} [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateBranch(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdateCompetitionBranchRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.Branch.UpdateBranch(c.Context(), &competitionv1.UpdateBranchRequest{
		Id:             id,
		Name:           body.Name,
		Description:    body.Description,
		StatusId:       body.StatusID,
		ParentBranchId: body.ParentBranchID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToCompetitionBranchResponse(resp.Branch))
}

// DeleteBranch godoc
// @Summary      Delete competition stage (Admin)
// @Tags         Competitions
// @Produce      json
// @Param        id   path      string  true  "Branch ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/branches/{id} [delete]
// @Security     BearerAuth
func (h *CompetitionHandler) DeleteBranch(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Branch.DeleteBranch(c.Context(), &competitionv1.DeleteBranchRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "competition branch deleted successfully")
}

// ─── MATCH HANDLERS ──────────────────────────────────────────────────────────

// CreateMatch godoc
// @Summary      Create new match schedule (Admin)
// @Description  Register a match between two or more participating teams/rosters.
// @Tags         Matches
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMatchRequest true "New Match Data"
// @Success      201  {object}  response.Response{data=dto.MatchResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/matches [post]
// @Security     BearerAuth
func (h *CompetitionHandler) CreateMatch(c *fiber.Ctx) error {
	var body dto.CreateMatchRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	scheduledTime, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid scheduled_at format, must be RFC3339")
	}

	participantsInput := make([]*competitionv1.CreateMatchParticipantRequest, len(body.Participants))
	for i, p := range body.Participants {
		participantsInput[i] = &competitionv1.CreateMatchParticipantRequest{
			RosterId:    p.RosterID,
			FormatTagId: p.FormatTagID,
			ResultTagId: p.ResultTagID,
			Score:       p.Score,
		}
	}

	resp, err := h.client.Match.CreateMatch(c.Context(), &competitionv1.CreateMatchRequest{
		BranchId:     body.BranchID,
		ScheduledAt:  timestamppb.New(scheduledTime),
		Location:     body.Location,
		Referee:      body.Referee,
		Notes:        body.Notes,
		Participants: participantsInput,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToMatchResponse(resp.Match))
}

// GetMatch godoc
// @Summary      Get match details
// @Description  Retrieve schedule, location, referee, participant scores, and match status.
// @Tags         Matches
// @Produce      json
// @Param        id   path      string  true  "Match ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.MatchResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /matches/{id} [get]
// @Security     BearerAuth
func (h *CompetitionHandler) GetMatch(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Match.GetMatch(c.Context(), &competitionv1.GetMatchRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToMatchResponse(resp.Match))
}

// ListMatchesByBranch godoc
// @Summary      Get list of matches by competition stage
// @Tags         Matches
// @Produce      json
// @Param        id        path  string true  "Competition Branch ID (UUID)"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.MatchResponse}
// @Failure      401  {object}  response.Response
// @Router       /branches/{id}/matches [get]
// @Security     BearerAuth
func (h *CompetitionHandler) ListMatchesByBranch(c *fiber.Ctx) error {
	branchID := c.Params("id")
	page, pageSize := request.ParsePagination(c)

	resp, err := h.client.Match.ListMatchesByBranch(c.Context(), &competitionv1.ListMatchesByBranchRequest{
		BranchId: branchID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return err
	}

	matches := make([]dto.MatchResponse, len(resp.Matches))
	for i, m := range resp.Matches {
		matches[i] = mapper.ToMatchResponse(m)
	}

	return response.OKWithMeta(c, matches, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

// UpdateMatchStatus godoc
// @Summary      Update match status (Admin)
// @Description  Update match status (e.g. scheduled, live, completed, cancelled).
// @Tags         Matches
// @Accept       json
// @Produce      json
// @Param        id      path string true "Match ID (UUID)"
// @Param        request body object true "New Match Status"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/matches/{id}/status [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateMatchStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Status string `json:"status" validate:"required,min=1"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.client.Match.UpdateMatchStatus(c.Context(), &competitionv1.UpdateMatchStatusRequest{
		Id:     id,
		Status: body.Status,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "match status updated successfully")
}

// UpdateMatchScore godoc
// @Summary      Update final match score (Admin)
// @Description  Record final scores for match participants.
// @Tags         Matches
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Match ID (UUID)"
// @Param        request body dto.UpdateMatchScoreRequest true "Match Participant Scores"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/matches/{id}/score [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateMatchScore(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdateMatchScoreRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	participantsInput := make([]*competitionv1.UpdateParticipantScoreRequest, len(body.Participants))
	for i, p := range body.Participants {
		participantsInput[i] = &competitionv1.UpdateParticipantScoreRequest{
			ParticipantId: p.ParticipantID,
			Score:         p.Score,
		}
	}

	_, err := h.client.Match.UpdateMatchScore(c.Context(), &competitionv1.UpdateMatchScoreRequest{
		Id:           id,
		Participants: participantsInput,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "match score updated successfully")
}

// DeleteMatch godoc
// @Summary      Delete match (Admin)
// @Tags         Matches
// @Produce      json
// @Param        id   path      string  true  "Match ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/matches/{id} [delete]
// @Security     BearerAuth
func (h *CompetitionHandler) DeleteMatch(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Match.DeleteMatch(c.Context(), &competitionv1.DeleteMatchRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "match deleted successfully")
}

// ─── STATS HANDLERS ──────────────────────────────────────────────────────────

// RecordMatchStat godoc
// @Summary      Record player performance statistics in match (Admin)
// @Description  Record stats (points, assists, rebounds, yellow cards, etc.) for an athlete in a specific match.
// @Tags         Match Stats
// @Accept       json
// @Produce      json
// @Param        request body dto.RecordMatchStatRequest true "Match Statistic Data"
// @Success      201  {object}  response.Response{data=dto.MatchStatResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/stats [post]
// @Security     BearerAuth
func (h *CompetitionHandler) RecordMatchStat(c *fiber.Ctx) error {
	var body dto.RecordMatchStatRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.Stat.RecordMatchStat(c.Context(), &competitionv1.RecordMatchStatRequest{
		MatchId:    body.MatchID,
		AthleteId:  body.AthleteID,
		StatTypeId: body.StatTypeID,
		Value:      body.Value,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToMatchStatResponse(resp.Stat))
}

// UpdateMatchStat godoc
// @Summary      Update match performance statistic value (Admin)
// @Tags         Match Stats
// @Accept       json
// @Produce      json
// @Param        id      path string                    true "Match Stat ID (UUID)"
// @Param        request body dto.UpdateMatchStatRequest true "Statistic Update Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/stats/{id} [put]
// @Security     BearerAuth
func (h *CompetitionHandler) UpdateMatchStat(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdateMatchStatRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.client.Stat.UpdateMatchStat(c.Context(), &competitionv1.UpdateMatchStatRequest{
		Id:    id,
		Value: body.Value,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "match statistic updated successfully")
}

// DeleteMatchStat godoc
// @Summary      Delete match statistic record (Admin)
// @Tags         Match Stats
// @Produce      json
// @Param        id   path      string  true  "Match Stat ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/stats/{id} [delete]
// @Security     BearerAuth
func (h *CompetitionHandler) DeleteMatchStat(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Stat.DeleteMatchStat(c.Context(), &competitionv1.DeleteMatchStatRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "match statistic deleted successfully")
}

// GetMatchStats godoc
// @Summary      Get all player statistics for a match
// @Description  Retrieve complete box score match statistics for all players.
// @Tags         Match Stats
// @Produce      json
// @Param        id   path      string  true  "Match ID (UUID)"
// @Success      200  {object}  response.Response{data=[]dto.MatchStatResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /matches/{id}/stats [get]
// @Security     BearerAuth
func (h *CompetitionHandler) GetMatchStats(c *fiber.Ctx) error {
	matchID := c.Params("id")

	resp, err := h.client.Stat.GetMatchStats(c.Context(), &competitionv1.GetMatchStatsRequest{
		MatchId: matchID,
	})
	if err != nil {
		return err
	}

	stats := make([]dto.MatchStatResponse, len(resp.Stats))
	for i, st := range resp.Stats {
		stats[i] = mapper.ToMatchStatResponse(st)
	}

	return response.OK(c, stats)
}

// GetAthleteAggregate godoc
// @Summary      Get athlete statistical aggregates in a competition
// @Description  Retrieve average or total individual statistics for an athlete across a tournament.
// @Tags         Match Stats
// @Produce      json
// @Param        athleteId      path  string true "Athlete User ID (UUID)"
// @Param        competition_id query string true "Competition ID (UUID)"
// @Success      200  {object}  response.Response{data=[]dto.AthleteStatsAggregateResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /stats/athlete/{athleteId} [get]
// @Security     BearerAuth
func (h *CompetitionHandler) GetAthleteAggregate(c *fiber.Ctx) error {
	athleteID := c.Params("athleteId")
	competitionID := c.Query("competition_id", "")

	if competitionID == "" {
		return response.Error(c, fiber.StatusBadRequest, "competition_id query param is required")
	}

	resp, err := h.client.Stat.GetAthleteAggregate(c.Context(), &competitionv1.GetAthleteAggregateRequest{
		AthleteId:     athleteID,
		CompetitionId: competitionID,
	})
	if err != nil {
		return err
	}

	aggregates := make([]dto.AthleteStatsAggregateResponse, len(resp.Aggregates))
	for i, agg := range resp.Aggregates {
		aggregates[i] = mapper.ToAthleteStatsAggregateResponse(agg)
	}

	return response.OK(c, aggregates)
}

// RecalculateAggregate godoc
// @Summary      Recalculate athlete statistical aggregates in a competition (Admin)
// @Description  Recalculate and re-aggregate average / total performance data for an athlete in a tournament.
// @Tags         Match Stats
// @Accept       json
// @Produce      json
// @Param        athleteId path string true "Athlete User ID (UUID)"
// @Param        request   body object true "Competition ID Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/stats/athlete/{athleteId}/recalculate [post]
// @Security     BearerAuth
func (h *CompetitionHandler) RecalculateAggregate(c *fiber.Ctx) error {
	athleteID := c.Params("athleteId")

	var body struct {
		CompetitionID string `json:"competition_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.client.Stat.RecalculateAggregate(c.Context(), &competitionv1.RecalculateAggregateRequest{
		AthleteId:     athleteID,
		CompetitionId: body.CompetitionID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "aggregates recalculated successfully")
}
