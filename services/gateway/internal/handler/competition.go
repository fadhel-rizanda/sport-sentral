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
