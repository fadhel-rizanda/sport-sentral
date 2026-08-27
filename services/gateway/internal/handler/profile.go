package handler

import (
	"github.com/gofiber/fiber/v2"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type ProfileHandler struct {
	rbacClient rbacv1.RBACServiceClient
}

func NewProfileHandler(rbacClient rbacv1.RBACServiceClient) *ProfileHandler {
	return &ProfileHandler{
		rbacClient: rbacClient,
	}
}

func (h *ProfileHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	me := router.Group("/me", auth)
	me.Post("/profiles/apply", h.ApplyProfile)
	me.Post("/profiles/toggle", h.ToggleProfile)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin", adminMiddlewares...)
	adminGroup.Post("/profiles/approve", h.ApproveProfile)
	adminGroup.Post("/profiles/reject", h.RejectProfile)
}

// ApplyProfile godoc
// @Summary      Apply for new profile role
// @Description  Submit an application for a new profile role (e.g. coach, scout, court owner) for verification.
// @Tags         Profiles
// @Accept       json
// @Produce      json
// @Param        request body dto.ApplyProfileRequest true "Profile Application Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /me/profiles/apply [post]
// @Security     BearerAuth
func (h *ProfileHandler) ApplyProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.ApplyProfileRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ApplyProfile(c.Context(), &rbacv1.ApplyProfileRequest{
		UserId: userID,
		RoleId: body.RoleID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "profile application submitted")
}

// ToggleProfile godoc
// @Summary      Switch active role
// @Description  Switch between approved profile roles for multi-role users.
// @Tags         Profiles
// @Accept       json
// @Produce      json
// @Param        request body dto.ToggleProfileRequest true "Profile Toggle Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /me/profiles/toggle [post]
// @Security     BearerAuth
func (h *ProfileHandler) ToggleProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.ToggleProfileRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ToggleProfile(c.Context(), &rbacv1.ToggleProfileRequest{
		UserId: userID,
		RoleId: body.RoleID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "profile toggled")
}

// ApproveProfile godoc
// @Summary      Approve profile role application (Admin)
// @Description  Approve a user's role/profile application (platform admin only).
// @Tags         Profiles
// @Accept       json
// @Produce      json
// @Param        request body dto.ReviewProfileRequest true "Profile Approval Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/profiles/approve [post]
// @Security     BearerAuth
func (h *ProfileHandler) ApproveProfile(c *fiber.Ctx) error {
	var body dto.ReviewProfileRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ApproveProfile(c.Context(), &rbacv1.ApproveProfileRequest{
		UserId: body.UserID,
		RoleId: body.RoleID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "profile approved")
}

// RejectProfile godoc
// @Summary      Reject profile role application (Admin)
// @Description  Reject a user's role/profile application (platform admin only).
// @Tags         Profiles
// @Accept       json
// @Produce      json
// @Param        request body dto.ReviewProfileRequest true "Profile Rejection Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/profiles/reject [post]
// @Security     BearerAuth
func (h *ProfileHandler) RejectProfile(c *fiber.Ctx) error {
	var body dto.ReviewProfileRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.RejectProfile(c.Context(), &rbacv1.RejectProfileRequest{
		UserId: body.UserID,
		RoleId: body.RoleID,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "profile rejected")
}
