package handler

import (
	"time"

	academyv1 "microservice-golang/gen/academy/v1"
	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AcademyHandler struct {
	client     *client.AcademyClient
	compClient *client.CompetitionClient
}

func NewAcademyHandler(client *client.AcademyClient, compClient *client.CompetitionClient) *AcademyHandler {
	return &AcademyHandler{client: client, compClient: compClient}
}

func (h *AcademyHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	// Group
	academy := router.Group("/academy", auth)

	// Academy Holdings
	academy.Post("/holdings", h.CreateAcademyHolding)
	academy.Get("/holdings/:id", h.GetAcademyHolding)
	academy.Get("/holdings", h.ListAcademyHoldings)
	academy.Put("/holdings/:id", h.UpdateAcademyHolding)
	academy.Delete("/holdings/:id", h.DeleteAcademyHolding)

	// Academy Branches
	academy.Post("/branches", h.CreateAcademyBranch)
	academy.Get("/branches/:id", h.GetAcademyBranch)
	academy.Get("/branches", h.ListAcademyBranches)
	academy.Put("/branches/:id", h.UpdateAcademyBranch)
	academy.Delete("/branches/:id", h.DeleteAcademyBranch)

	// Academy Admins
	academy.Post("/admins", h.CreateAcademyAdmin)
	academy.Post("/admins/assign", h.AssignAcademyAdmin)
	academy.Get("/admins/:id", h.GetAcademyAdmin)
	academy.Get("/admins", h.ListAcademyAdmins)
	academy.Put("/admins/:id", h.UpdateAcademyAdmin)
	academy.Delete("/admins/:id", h.DeleteAcademyAdmin)
	academy.Post("/admins/:id/revoke", h.RevokeAcademyAdmin)
	academy.Get("/admins/user/:userID", h.GetAcademyAdminByUser)
	academy.Get("/admins/check/:userID", h.CheckUserIsAcademyAdmin)

	// Enrollments
	academy.Post("/enrollments", h.CreateEnrollment)
	academy.Get("/enrollments/:id", h.GetEnrollment)
	academy.Get("/enrollments/branch/:branchID/athlete/:athleteID", h.GetEnrollmentByBranchAndAthlete)
	academy.Get("/enrollments", h.ListEnrollments)
	academy.Put("/enrollments/:id", h.UpdateEnrollment)
	academy.Delete("/enrollments/:id", h.DeleteEnrollment)

	// Rosters
	academy.Post("/rosters", h.CreateRoster)
	academy.Get("/rosters/:id", h.GetRoster)
	academy.Get("/rosters", h.ListRosters)
	academy.Put("/rosters/:id", h.UpdateRoster)
	academy.Delete("/rosters/:id", h.DeleteRoster)
	academy.Get("/rosters/:id/members", h.GetRosterMembers)
	academy.Get("/rosters/:id/members/:memberID", h.GetRosterMember)
	academy.Post("/rosters/:id/members", h.AddRosterMember)
	academy.Post("/rosters/:id/members/:memberID/remove", h.RemoveRosterMember)
	academy.Delete("/rosters/:id/members/:memberID", h.DeleteRosterMember)
}

// ─── Academy Holding Handlers ──────────────────────────────────────────────────

func (h *AcademyHandler) CreateAcademyHolding(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Name              string   `json:"name" validate:"required"`
		Description       string   `json:"description"`
		Email             string   `json:"email" validate:"required,email"`
		PhoneNumber       string   `json:"phone_number" validate:"required"`
		ImageAttachmentID *string  `json:"image_attachment_id"`
		StatusID          string   `json:"status_id" validate:"required,uuid"`
		StreetAddress     string   `json:"street_address" validate:"required"`
		Notes             *string  `json:"notes"`
		Latitude          *float64 `json:"latitude"`
		Longitude         *float64 `json:"longitude"`
		AdminDivisionID   string   `json:"admin_division_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyHolding.CreateAcademyHolding(c.Context(), &academyv1.CreateAcademyHoldingRequest{
		Name:              body.Name,
		Description:       body.Description,
		Email:             body.Email,
		PhoneNumber:       body.PhoneNumber,
		ImageAttachmentId: body.ImageAttachmentID,
		StatusId:          body.StatusID,
		StreetAddress:     body.StreetAddress,
		Notes:             body.Notes,
		Latitude:          body.Latitude,
		Longitude:         body.Longitude,
		AdminDivisionId:   body.AdminDivisionID,
		CreatedById:       userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToAcademyHoldingResponse(resp.Holding))
}

func (h *AcademyHandler) GetAcademyHolding(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.AcademyHolding.GetAcademyHolding(c.Context(), &academyv1.GetAcademyHoldingRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyHoldingResponse(resp.Holding))
}

func (h *AcademyHandler) ListAcademyHoldings(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	search := c.Query("search", "")
	statusID := c.Query("status_id", "")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	resp, err := h.client.AcademyHolding.ListAcademyHoldings(c.Context(), &academyv1.ListAcademyHoldingsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   searchPtr,
		StatusId: statusIDPtr,
	})
	if err != nil {
		return err
	}

	holdings := make([]dto.AcademyHoldingResponse, len(resp.Holdings))
	for i, holding := range resp.Holdings {
		holdings[i] = mapper.ToAcademyHoldingResponse(holding)
	}

	return response.OKWithMeta(c, holdings, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *AcademyHandler) UpdateAcademyHolding(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Name              *string `json:"name"`
		Description       *string `json:"description"`
		Email             *string `json:"email"`
		PhoneNumber       *string `json:"phone_number"`
		ImageAttachmentID *string `json:"image_attachment_id"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyHolding.UpdateAcademyHolding(c.Context(), &academyv1.UpdateAcademyHoldingRequest{
		Id:                id,
		Name:              body.Name,
		Description:       body.Description,
		Email:             body.Email,
		PhoneNumber:       body.PhoneNumber,
		ImageAttachmentId: body.ImageAttachmentID,
		UpdatedById:       userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyHoldingResponse(resp.Holding))
}

func (h *AcademyHandler) DeleteAcademyHolding(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.client.AcademyHolding.DeleteAcademyHolding(c.Context(), &academyv1.DeleteAcademyHoldingRequest{
		Id:          id,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "academy holding deleted")
}

// ─── Academy Branch Handlers ───────────────────────────────────────────────────

func (h *AcademyHandler) CreateAcademyBranch(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		HoldingID       string   `json:"holding_id" validate:"required,uuid"`
		SportID         string   `json:"sport_id" validate:"required,uuid"`
		Name            string   `json:"name" validate:"required"`
		Email           string   `json:"email" validate:"required,email"`
		PhoneNumber     string   `json:"phone_number" validate:"required"`
		StatusID        string   `json:"status_id" validate:"required,uuid"`
		StreetAddress   string   `json:"street_address" validate:"required"`
		Notes           *string  `json:"notes"`
		Latitude        *float64 `json:"latitude"`
		Longitude       *float64 `json:"longitude"`
		AdminDivisionID string   `json:"admin_division_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyBranch.CreateAcademyBranch(c.Context(), &academyv1.CreateAcademyBranchRequest{
		HoldingId:       body.HoldingID,
		SportId:         body.SportID,
		Name:            body.Name,
		Email:           body.Email,
		PhoneNumber:     body.PhoneNumber,
		StatusId:        body.StatusID,
		StreetAddress:   body.StreetAddress,
		Notes:           body.Notes,
		Latitude:        body.Latitude,
		Longitude:       body.Longitude,
		AdminDivisionId: body.AdminDivisionID,
		CreatedById:     userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToAcademyBranchResponse(resp.Branch))
}

func (h *AcademyHandler) GetAcademyBranch(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.AcademyBranch.GetAcademyBranch(c.Context(), &academyv1.GetAcademyBranchRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyBranchResponse(resp.Branch))
}

func (h *AcademyHandler) ListAcademyBranches(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	search := c.Query("search", "")
	holdingID := c.Query("holding_id", "")
	sportID := c.Query("sport_id", "")
	statusID := c.Query("status_id", "")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var holdingIDPtr *string
	if holdingID != "" {
		holdingIDPtr = &holdingID
	}

	var sportIDPtr *string
	if sportID != "" {
		sportIDPtr = &sportID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	resp, err := h.client.AcademyBranch.ListAcademyBranches(c.Context(), &academyv1.ListAcademyBranchesRequest{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		Search:    searchPtr,
		HoldingId: holdingIDPtr,
		SportId:   sportIDPtr,
		StatusId:  statusIDPtr,
	})
	if err != nil {
		return err
	}

	branches := make([]dto.AcademyBranchResponse, len(resp.Branches))
	for i, b := range resp.Branches {
		branches[i] = mapper.ToAcademyBranchResponse(b)
	}

	return response.OKWithMeta(c, branches, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *AcademyHandler) UpdateAcademyBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Name        *string `json:"name"`
		Email       *string `json:"email"`
		PhoneNumber *string `json:"phone_number"`
		StatusID    *string `json:"status_id"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyBranch.UpdateAcademyBranch(c.Context(), &academyv1.UpdateAcademyBranchRequest{
		Id:          id,
		Name:        body.Name,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		StatusId:    body.StatusID,
		UpdatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyBranchResponse(resp.Branch))
}

func (h *AcademyHandler) DeleteAcademyBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.client.AcademyBranch.DeleteAcademyBranch(c.Context(), &academyv1.DeleteAcademyBranchRequest{
		Id:          id,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "academy branch deleted")
}

// ─── Academy Admin Handlers ────────────────────────────────────────────────────

func (h *AcademyHandler) CreateAcademyAdmin(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		AcademyID string  `json:"academy_id" validate:"required,uuid"`
		BranchID  *string `json:"branch_id"`
		UserID    string  `json:"user_id" validate:"required,uuid"`
		RoleID    string  `json:"role_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyAdmin.CreateAcademyAdmin(c.Context(), &academyv1.CreateAcademyAdminRequest{
		AcademyId:   body.AcademyID,
		BranchId:    body.BranchID,
		UserId:      body.UserID,
		RoleId:      body.RoleID,
		CreatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToAcademyAdminResponse(resp.Admin))
}

func (h *AcademyHandler) GetAcademyAdmin(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.AcademyAdmin.GetAcademyAdmin(c.Context(), &academyv1.GetAcademyAdminRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyAdminResponse(resp.Admin))
}

func (h *AcademyHandler) ListAcademyAdmins(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	search := c.Query("search", "")
	academyID := c.Query("academy_id", "")
	branchID := c.Query("branch_id", "")
	statusID := c.Query("status_id", "")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var academyIDPtr *string
	if academyID != "" {
		academyIDPtr = &academyID
	}

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	resp, err := h.client.AcademyAdmin.ListAcademyAdmins(c.Context(), &academyv1.ListAcademyAdminsRequest{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		Search:    searchPtr,
		AcademyId: academyIDPtr,
		BranchId:  branchIDPtr,
		StatusId:  statusIDPtr,
	})
	if err != nil {
		return err
	}

	admins := make([]dto.AcademyAdminResponse, len(resp.Admins))
	for i, a := range resp.Admins {
		admins[i] = mapper.ToAcademyAdminResponse(a)
	}

	return response.OKWithMeta(c, admins, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *AcademyHandler) UpdateAcademyAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		RoleID string `json:"role_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyAdmin.UpdateAcademyAdmin(c.Context(), &academyv1.UpdateAcademyAdminRequest{
		Id:          id,
		RoleId:      body.RoleID,
		UpdatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyAdminResponse(resp.Admin))
}

func (h *AcademyHandler) DeleteAcademyAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.client.AcademyAdmin.DeleteAcademyAdmin(c.Context(), &academyv1.DeleteAcademyAdminRequest{
		Id:          id,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "academy admin deleted")
}

func (h *AcademyHandler) AssignAcademyAdmin(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		AcademyID string  `json:"academy_id" validate:"required,uuid"`
		BranchID  *string `json:"branch_id"`
		UserID    string  `json:"user_id" validate:"required,uuid"`
		RoleID    string  `json:"role_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.client.AcademyAdmin.AssignAcademyAdmin(c.Context(), &academyv1.AssignAcademyAdminRequest{
		AcademyId:    body.AcademyID,
		BranchId:     body.BranchID,
		UserId:       body.UserID,
		RoleId:       body.RoleID,
		AssignedById: userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToAcademyAdminResponse(resp.Admin))
}

func (h *AcademyHandler) RevokeAcademyAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	resp, err := h.client.AcademyAdmin.RevokeAcademyAdmin(c.Context(), &academyv1.RevokeAcademyAdminRequest{
		Id:          id,
		RevokedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"id": resp.Id,
	})
}

func (h *AcademyHandler) GetAcademyAdminByUser(c *fiber.Ctx) error {
	userID := c.Params("userID")
	holdingID := c.Query("holding_id", "")
	branchID := c.Query("branch_id", "")

	var holdingIDPtr *string
	if holdingID != "" {
		holdingIDPtr = &holdingID
	}

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	resp, err := h.client.AcademyAdmin.GetAcademyAdminByUser(c.Context(), &academyv1.GetAcademyAdminByUserRequest{
		HoldingId: holdingIDPtr,
		BranchId:  branchIDPtr,
		UserId:    userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAcademyAdminResponse(resp.Admin))
}

func (h *AcademyHandler) CheckUserIsAcademyAdmin(c *fiber.Ctx) error {
	userID := c.Params("userID")
	holdingID := c.Query("holding_id", "")
	branchID := c.Query("branch_id", "")

	var holdingIDPtr *string
	if holdingID != "" {
		holdingIDPtr = &holdingID
	}

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	resp, err := h.client.AcademyAdmin.CheckUserIsAcademyAdmin(c.Context(), &academyv1.CheckUserIsAcademyAdminRequest{
		HoldingId: holdingIDPtr,
		BranchId:  branchIDPtr,
		UserId:    userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"is_admin": resp.IsAdmin})
}

// ─── Enrollment Handlers ───────────────────────────────────────────────────────

func (h *AcademyHandler) CreateEnrollment(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		AcademyBranchID string  `json:"academy_branch_id" validate:"required,uuid"`
		AthleteID       string  `json:"athlete_id" validate:"required,uuid"`
		JoinedAt        string  `json:"joined_at" validate:"required"`
		ExpiresAt       *string `json:"expires_at"`
		StatusID        string  `json:"status_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	joinedTime, err := time.Parse(time.RFC3339, body.JoinedAt)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid joined_at format, must be RFC3339")
	}

	var expiresTime *timestamppb.Timestamp
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid expires_at format, must be RFC3339")
		}
		expiresTime = timestamppb.New(t)
	}

	resp, err := h.client.Enrollment.CreateEnrollment(c.Context(), &academyv1.CreateEnrollmentRequest{
		AcademyBranchId: body.AcademyBranchID,
		AthleteId:       body.AthleteID,
		JoinedAt:        timestamppb.New(joinedTime),
		ExpiresAt:       expiresTime,
		StatusId:        body.StatusID,
		CreatedById:     userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToEnrollmentResponse(resp.Enrollment))
}

func (h *AcademyHandler) GetEnrollment(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.Enrollment.GetEnrollment(c.Context(), &academyv1.GetEnrollmentRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToEnrollmentResponse(resp.Enrollment))
}

func (h *AcademyHandler) GetEnrollmentByBranchAndAthlete(c *fiber.Ctx) error {
	branchID := c.Params("branchID")
	athleteID := c.Params("athleteID")

	resp, err := h.client.Enrollment.GetEnrollmentByBranchAndAthlete(c.Context(), &academyv1.GetEnrollmentByBranchAndAthleteRequest{
		BranchId:  branchID,
		AthleteId: athleteID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToEnrollmentResponse(resp.Enrollment))
}

func (h *AcademyHandler) ListEnrollments(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	search := c.Query("search", "")
	branchID := c.Query("branch_id", "")
	athleteID := c.Query("athlete_id", "")
	statusID := c.Query("status_id", "")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	var athleteIDPtr *string
	if athleteID != "" {
		athleteIDPtr = &athleteID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	resp, err := h.client.Enrollment.ListEnrollments(c.Context(), &academyv1.ListEnrollmentsRequest{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		Search:    searchPtr,
		BranchId:  branchIDPtr,
		AthleteId: athleteIDPtr,
		StatusId:  statusIDPtr,
	})
	if err != nil {
		return err
	}

	enrollments := make([]dto.EnrollmentResponse, len(resp.Enrollments))
	for i, e := range resp.Enrollments {
		enrollments[i] = mapper.ToEnrollmentResponse(e)
	}

	return response.OKWithMeta(c, enrollments, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *AcademyHandler) UpdateEnrollment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		LeftAt    *string `json:"left_at"`
		ExpiresAt *string `json:"expires_at"`
		StatusID  *string `json:"status_id"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	var leftTime *timestamppb.Timestamp
	if body.LeftAt != nil && *body.LeftAt != "" {
		t, err := time.Parse(time.RFC3339, *body.LeftAt)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid left_at format, must be RFC3339")
		}
		leftTime = timestamppb.New(t)
	}

	var expiresTime *timestamppb.Timestamp
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid expires_at format, must be RFC3339")
		}
		expiresTime = timestamppb.New(t)
	}

	resp, err := h.client.Enrollment.UpdateEnrollment(c.Context(), &academyv1.UpdateEnrollmentRequest{
		Id:          id,
		LeftAt:      leftTime,
		ExpiresAt:   expiresTime,
		StatusId:    body.StatusID,
		UpdatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToEnrollmentResponse(resp.Enrollment))
}

func (h *AcademyHandler) DeleteEnrollment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.client.Enrollment.DeleteEnrollment(c.Context(), &academyv1.DeleteEnrollmentRequest{
		Id:          id,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "enrollment deleted")
}

// ─── Roster Handlers ───────────────────────────────────────────────────────────

func (h *AcademyHandler) CreateRoster(c *fiber.Ctx) error {
	var body struct {
		AcademyBranchID string  `json:"academy_branch_id" validate:"required,uuid"`
		CompetitionID   *string `json:"competition_id"`
		Name            string  `json:"name" validate:"required"`
		TagID           string  `json:"tag_id" validate:"required,uuid"`
		StatusID        string  `json:"status_id" validate:"required,uuid"`
		MaxSize         int32   `json:"max_size" validate:"required"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.compClient.Roster.CreateRoster(c.Context(), &competitionv1.CreateRosterRequest{
		AcademyBranchId: body.AcademyBranchID,
		CompetitionId:   body.CompetitionID,
		Name:            body.Name,
		TagId:           body.TagID,
		StatusId:        body.StatusID,
		MaxSize:         body.MaxSize,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToRosterResponse(resp.Roster))
}

func (h *AcademyHandler) GetRoster(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.compClient.Roster.GetRoster(c.Context(), &competitionv1.GetRosterRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToRosterResponse(resp.Roster))
}

func (h *AcademyHandler) ListRosters(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	search := c.Query("search", "")
	competitionID := c.Query("competition_id", "")
	branchID := c.Query("branch_id", "")
	statusID := c.Query("status_id", "")
	tagID := c.Query("tag_id", "")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var competitionIDPtr *string
	if competitionID != "" {
		competitionIDPtr = &competitionID
	}

	var branchIDPtr *string
	if branchID != "" {
		branchIDPtr = &branchID
	}

	var statusIDPtr *string
	if statusID != "" {
		statusIDPtr = &statusID
	}

	var tagIDPtr *string
	if tagID != "" {
		tagIDPtr = &tagID
	}

	resp, err := h.compClient.Roster.ListRosters(c.Context(), &competitionv1.ListRostersRequest{
		Page:          int32(page),
		PageSize:      int32(pageSize),
		Search:        searchPtr,
		CompetitionId: competitionIDPtr,
		BranchId:      branchIDPtr,
		StatusId:      statusIDPtr,
		TagId:         tagIDPtr,
	})
	if err != nil {
		return err
	}

	rosters := make([]dto.RosterResponse, len(resp.Rosters))
	for i, r := range resp.Rosters {
		rosters[i] = mapper.ToRosterResponse(r)
	}

	return response.OKWithMeta(c, rosters, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *AcademyHandler) UpdateRoster(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		CompetitionID *string `json:"competition_id"`
		Name          *string `json:"name"`
		TagID         *string `json:"tag_id"`
		StatusID      *string `json:"status_id"`
		MaxSize       *int32  `json:"max_size"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.compClient.Roster.UpdateRoster(c.Context(), &competitionv1.UpdateRosterRequest{
		Id:            id,
		CompetitionId: body.CompetitionID,
		Name:          body.Name,
		TagId:         body.TagID,
		StatusId:      body.StatusID,
		MaxSize:       body.MaxSize,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToRosterResponse(resp.Roster))
}

func (h *AcademyHandler) DeleteRoster(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	_, err := h.compClient.Roster.DeleteRoster(c.Context(), &competitionv1.DeleteRosterRequest{
		Id:          id,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "roster deleted")
}

func (h *AcademyHandler) GetRosterMembers(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.compClient.Roster.GetRosterMembers(c.Context(), &competitionv1.GetRosterMembersRequest{
		RosterId: id,
	})
	if err != nil {
		return err
	}

	members := make([]dto.RosterMemberResponse, len(resp.Members))
	for i, m := range resp.Members {
		members[i] = mapper.ToRosterMemberResponse(m)
	}

	return response.OK(c, members)
}

func (h *AcademyHandler) GetRosterMember(c *fiber.Ctx) error {
	id := c.Params("id")
	memberID := c.Params("memberID")

	resp, err := h.compClient.Roster.GetRosterMember(c.Context(), &competitionv1.GetRosterMemberRequest{
		RosterId: id,
		MemberId: memberID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToRosterMemberResponse(resp.Member))
}

func (h *AcademyHandler) AddRosterMember(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		AthleteID    string `json:"athlete_id" validate:"required,uuid"`
		JerseyNumber *int32 `json:"jersey_number"`
		PositionID   string `json:"position_id" validate:"required,uuid"`
		StatusID     string `json:"status_id" validate:"required,uuid"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.compClient.Roster.AddRosterMember(c.Context(), &competitionv1.AddRosterMemberRequest{
		RosterId:     id,
		AthleteId:    body.AthleteID,
		JerseyNumber: body.JerseyNumber,
		PositionId:   body.PositionID,
		StatusId:     body.StatusID,
		AddedById:    userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToRosterMemberResponse(resp.Member))
}

func (h *AcademyHandler) RemoveRosterMember(c *fiber.Ctx) error {
	id := c.Params("id")
	memberID := c.Params("memberID")
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		RemovalReason string `json:"removal_reason" validate:"required"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.compClient.Roster.RemoveRosterMember(c.Context(), &competitionv1.RemoveRosterMemberRequest{
		RosterId:      id,
		MemberId:      memberID,
		RemovedById:   userID,
		RemovalReason: body.RemovalReason,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"member_id": resp.MemberId})
}

func (h *AcademyHandler) DeleteRosterMember(c *fiber.Ctx) error {
	id := c.Params("id")
	memberID := c.Params("memberID")

	resp, err := h.compClient.Roster.DeleteRosterMember(c.Context(), &competitionv1.DeleteRosterMemberRequest{
		RosterId: id,
		MemberId: memberID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"member_id": resp.MemberId})
}
