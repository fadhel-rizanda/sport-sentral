package handler

import (
	"github.com/gofiber/fiber/v2"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
	"microservice-golang/shared/pkg/utils"
)

type SportHandler struct {
	client *client.SportClient
}

func NewSportHandler(client *client.SportClient) *SportHandler {
	return &SportHandler{client: client}
}

func (h *SportHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	// Sports Public/Authenticated endpoints
	sports := router.Group("/sports", auth)
	sports.Get("/", h.ListSports)
	sports.Get("/:id", h.GetSport)
	sports.Get("/:id/config", h.GetSportConfig)

	// Admin-only endpoints for Sports
	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminSports := router.Group("/admin/sports", adminMiddlewares...)
	adminSports.Post("/", h.CreateSport)
	adminSports.Put("/:id", h.UpdateSport)
	adminSports.Delete("/:id", h.DeleteSport)
	adminSports.Put("/:id/config", h.UpdateSportConfig)

	// Regulators Public/Authenticated endpoints
	regulators := router.Group("/regulators", auth)
	regulators.Get("/:id", h.GetRegulator)

	// Admin-only endpoints for Regulators
	adminRegulators := router.Group("/admin/regulators", adminMiddlewares...)
	adminRegulators.Post("/", h.CreateRegulator)
	adminRegulators.Post("/:id/staff", h.AddRegulatorStaff)
	adminRegulators.Delete("/:id/staff/:staffId", h.RemoveRegulatorStaff)
	adminRegulators.Post("/:id/assign-sport", h.AssignSportToRegulator)
}

// ─── SPORT HANDLERS ──────────────────────────────────────────────────────────

// CreateSport godoc
// @Summary      Create new sport (Admin)
// @Description  Register a new sport into the platform.
// @Tags         Sports
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateSportRequest true "New Sport Data"
// @Success      201  {object}  response.Response{data=dto.SportResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/sports [post]
// @Security     BearerAuth
func (h *SportHandler) CreateSport(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.CreateSportRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	resp, err := h.client.Sport.CreateSport(c.Context(), &sportv1.CreateSportRequest{
		Name:             reqBody.Name,
		Slug:             reqBody.Slug,
		Description:      reqBody.Description,
		IconAttachmentId: reqBody.IconAttachmentID,
		TierTagId:        reqBody.TierTagID,
		StatusId:         reqBody.StatusID,
		CreatedById:      userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToSportResponse(resp))
}

// GetSport godoc
// @Summary      Get sport details
// @Description  Retrieve sport data by ID.
// @Tags         Sports
// @Produce      json
// @Param        id   path      string  true  "Sport ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.SportResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /sports/{id} [get]
// @Security     BearerAuth
func (h *SportHandler) GetSport(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Sport.GetSport(c.Context(), &sportv1.GetSportRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToSportResponse(resp))
}

// ListSports godoc
// @Summary      Get list of sports
// @Description  Retrieve list of sports with filters for tier tag, status, or search query.
// @Tags         Sports
// @Produce      json
// @Param        tier_tag_id query string false "Filter Tier Tag ID"
// @Param        status_id   query string false "Filter Status ID"
// @Param        search      query string false "Search query for sport name"
// @Param        page        query int    false "Page number (default 1)"
// @Param        page_size   query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.SportResponse}
// @Failure      401  {object}  response.Response
// @Router       /sports [get]
// @Security     BearerAuth
func (h *SportHandler) ListSports(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	tierTagID := c.Query("tier_tag_id", "")
	statusID := c.Query("status_id", "")
	search := c.Query("search", "")

	var tierTagIDPtr *string
	if tierTagID != "" {
		tierTagIDPtr = &tierTagID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	resp, err := h.client.Sport.ListSports(c.Context(), &sportv1.ListSportsRequest{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		TierTagId: tierTagIDPtr,
		StatusId:  statusIDPtr,
		Search:    searchPtr,
	})
	if err != nil {
		return err
	}

	sports := make([]dto.SportResponse, len(resp.Sports))
	for i, s := range resp.Sports {
		sports[i] = mapper.ToSportResponse(s)
	}

	return response.OKWithMeta(c, sports, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

// UpdateSport godoc
// @Summary      Update sport (Admin)
// @Description  Update name, description, icon, or status of a sport.
// @Tags         Sports
// @Accept       json
// @Produce      json
// @Param        id      path string                true "Sport ID (UUID)"
// @Param        request body dto.UpdateSportRequest true "Sport Update Data"
// @Success      200  {object}  response.Response{data=dto.SportResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/sports/{id} [put]
// @Security     BearerAuth
func (h *SportHandler) UpdateSport(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.UpdateSportRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	resp, err := h.client.Sport.UpdateSport(c.Context(), &sportv1.UpdateSportRequest{
		Id:               id,
		Name:             reqBody.Name,
		Description:      reqBody.Description,
		IconAttachmentId: reqBody.IconAttachmentID,
		TierTagId:        reqBody.TierTagID,
		StatusId:         reqBody.StatusID,
		UpdatedById:      userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToSportResponse(resp))
}

// DeleteSport godoc
// @Summary      Delete sport (Admin)
// @Description  Delete sport from the system.
// @Tags         Sports
// @Produce      json
// @Param        id   path      string  true  "Sport ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/sports/{id} [delete]
// @Security     BearerAuth
func (h *SportHandler) DeleteSport(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.Sport.DeleteSport(c.Context(), &sportv1.DeleteSportRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "sport deleted successfully")
}

// GetSportConfig godoc
// @Summary      Get sport statistical & roster configuration
// @Description  Retrieve statistical metrics configuration (pts, reb, ast) and roster rules for a sport.
// @Tags         Sports
// @Produce      json
// @Param        id   path      string  true  "Sport ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.SportConfigResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /sports/{id}/config [get]
// @Security     BearerAuth
func (h *SportHandler) GetSportConfig(c *fiber.Ctx) error {
	sportID := c.Params("id")

	resp, err := h.client.Sport.GetSportConfig(c.Context(), &sportv1.GetSportConfigRequest{
		SportId: sportID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToSportConfigResponse(resp))
}

// UpdateSportConfig godoc
// @Summary      Update sport statistical & roster configuration (Admin)
// @Description  Configure statistical metrics, roster size constraints, and rulebook URL.
// @Tags         Sports
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Sport ID (UUID)"
// @Param        request body dto.UpdateSportConfigRequest true "Sport Configuration Data"
// @Success      200  {object}  response.Response{data=dto.SportConfigResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/sports/{id}/config [put]
// @Security     BearerAuth
func (h *SportHandler) UpdateSportConfig(c *fiber.Ctx) error {
	sportID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.UpdateSportConfigRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	statsInput := make([]*sportv1.SportStatConfigInput, len(reqBody.Stats))
	for i, s := range reqBody.Stats {
		statsInput[i] = &sportv1.SportStatConfigInput{
			StatTypeTagId:     s.StatTypeTagID,
			AggregationMethod: utils.ParseAggregationMethod(s.AggregationMethod),
		}
	}

	resp, err := h.client.Sport.UpdateSportConfig(c.Context(), &sportv1.UpdateSportConfigRequest{
		SportId:              sportID,
		Stats:                statsInput,
		ParticipantTypeTagId: reqBody.ParticipantTypeTagID,
		MinRosterSize:        reqBody.MinRosterSize,
		MaxRosterSize:        reqBody.MaxRosterSize,
		TypicalRosterSize:    reqBody.TypicalRosterSize,
		RulesUrl:             reqBody.RulesURL,
		Description:          reqBody.Description,
		UpdatedById:          userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToSportConfigResponse(resp))
}

// ─── REGULATOR HANDLERS ──────────────────────────────────────────────────────

// CreateRegulator godoc
// @Summary      Create sports regulator organization (Admin)
// @Description  Register an official sports regulatory body/association (e.g. PERBASI, FIBA, PSSI).
// @Tags         Regulators
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateRegulatorRequest true "Regulator Organization Data"
// @Success      201  {object}  response.Response{data=dto.RegulatorResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/regulators [post]
// @Security     BearerAuth
func (h *SportHandler) CreateRegulator(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.CreateRegulatorRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	resp, err := h.client.Regulator.CreateRegulator(c.Context(), &sportv1.CreateRegulatorRequest{
		OrganizationName:         reqBody.OrganizationName,
		Code:                     reqBody.Code,
		ContactEmail:             reqBody.ContactEmail,
		PhoneNumber:              reqBody.PhoneNumber,
		WebsiteUrl:               reqBody.WebsiteURL,
		LogoAttachmentId:         reqBody.LogoAttachmentID,
		StreetAddress:            reqBody.StreetAddress,
		AddressNotes:             reqBody.AddressNotes,
		Latitude:                 reqBody.Latitude,
		Longitude:                reqBody.Longitude,
		AdministrativeDivisionId: reqBody.AdministrativeDivisionID,
		StatusId:                 reqBody.StatusID,
		CreatedById:              userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToRegulatorResponse(resp))
}

// GetRegulator godoc
// @Summary      Get regulator organization details
// @Description  Retrieve sports regulator profile by ID.
// @Tags         Regulators
// @Produce      json
// @Param        id   path      string  true  "Regulator ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.RegulatorResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /regulators/{id} [get]
// @Security     BearerAuth
func (h *SportHandler) GetRegulator(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Regulator.GetRegulator(c.Context(), &sportv1.GetRegulatorRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToRegulatorResponse(resp))
}

// AddRegulatorStaff godoc
// @Summary      Add staff to regulator (Admin)
// @Description  Assign staff/officials to a regulator organization.
// @Tags         Regulators
// @Accept       json
// @Produce      json
// @Param        id      path string                         true "Regulator ID (UUID)"
// @Param        request body dto.AddRegulatorStaffRequest   true "Staff Assignment Data"
// @Success      201  {object}  response.Response{data=dto.RegulatorStaffResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/regulators/{id}/staff [post]
// @Security     BearerAuth
func (h *SportHandler) AddRegulatorStaff(c *fiber.Ctx) error {
	regulatorID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.AddRegulatorStaffRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	resp, err := h.client.Regulator.AddRegulatorStaff(c.Context(), &sportv1.AddStaffRequest{
		RegulatorId: regulatorID,
		UserId:      reqBody.UserID,
		RoleTagId:   reqBody.RoleTagID,
		CreatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToRegulatorStaffResponse(resp))
}

// RemoveRegulatorStaff godoc
// @Summary      Remove staff from regulator (Admin)
// @Description  Remove a staff assignment from a regulator organization.
// @Tags         Regulators
// @Produce      json
// @Param        id      path string true "Regulator ID (UUID)"
// @Param        staffId path string true "Staff Record ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/regulators/{id}/staff/{staffId} [delete]
// @Security     BearerAuth
func (h *SportHandler) RemoveRegulatorStaff(c *fiber.Ctx) error {
	staffID := c.Params("staffId")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.client.Regulator.RemoveRegulatorStaff(c.Context(), &sportv1.RemoveStaffRequest{
		StaffId:     staffID,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "regulator staff removed successfully")
}

// AssignSportToRegulator godoc
// @Summary      Assign sport authority to regulator (Admin)
// @Description  Link a sport to a regulator organization along with competition tier approval requirements.
// @Tags         Regulators
// @Accept       json
// @Produce      json
// @Param        id      path string                                  true "Regulator ID (UUID)"
// @Param        request body dto.AssignSportToRegulatorRequest       true "Sport Authority Assignment Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/regulators/{id}/assign-sport [post]
// @Security     BearerAuth
func (h *SportHandler) AssignSportToRegulator(c *fiber.Ctx) error {
	regulatorID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var reqBody dto.AssignSportToRegulatorRequest
	if err := request.Parse(c, &reqBody); err != nil {
		return err
	}

	_, err := h.client.Regulator.AssignSportToRegulator(c.Context(), &sportv1.AssignSportToRegulatorRequest{
		SportId:                     reqBody.SportID,
		RegulatorId:                 regulatorID,
		RequiresApprovalForOfficial: reqBody.RequiresApprovalForOfficial,
		RequiresApprovalForRegional: reqBody.RequiresApprovalForRegional,
		UpdatedById:                 userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "sport assigned to regulator successfully")
}
