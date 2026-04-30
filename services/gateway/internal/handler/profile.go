package handler

import (
	"github.com/gofiber/fiber/v2"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type ProfileHandler struct {
	rbacClient rbacv1.RbacServiceClient
}

func NewProfileHandler(rbacClient rbacv1.RbacServiceClient) *ProfileHandler {
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

func (h *ProfileHandler) ApplyProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		RoleName string `json:"role_name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ApplyProfile(c.Context(), &rbacv1.ApplyProfileRequest{
		UserId:   userID,
		RoleName: body.RoleName,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "profile application submitted")
}

func (h *ProfileHandler) ToggleProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		RoleName string `json:"role_name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ToggleProfile(c.Context(), &rbacv1.ToggleProfileRequest{
		UserId:   userID,
		RoleName: body.RoleName,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "profile toggled")
}

func (h *ProfileHandler) ApproveProfile(c *fiber.Ctx) error {
	var body struct {
		UserId   string `json:"user_id"`
		RoleName string `json:"role_name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.ApproveProfile(c.Context(), &rbacv1.ApproveProfileRequest{
		UserId:   body.UserId,
		RoleName: body.RoleName,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "profile approved")
}

func (h *ProfileHandler) RejectProfile(c *fiber.Ctx) error {
	var body struct {
		UserId   string `json:"user_id"`
		RoleName string `json:"role_name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.rbacClient.RejectProfile(c.Context(), &rbacv1.RejectProfileRequest{
		UserId:   body.UserId,
		RoleName: body.RoleName,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "profile rejected")
}
