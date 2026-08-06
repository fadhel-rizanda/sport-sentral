package handler

import (
	"github.com/gofiber/fiber/v2"
	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type RoleHandler struct {
	roleClient rbacv1.RoleServiceClient
}

func NewRoleHandler(roleClient rbacv1.RoleServiceClient) *RoleHandler {
	return &RoleHandler{
		roleClient: roleClient,
	}
}

func (h *RoleHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	role := router.Group("/roles", auth)
	role.Get("/", h.ListRoles)
	role.Get("/:id", h.GetRoles)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin/roles", adminMiddlewares...)
	adminGroup.Post("/", h.CreateRole)
	adminGroup.Put("/:id", h.UpdateRole)
	adminGroup.Delete("/:id", h.DeleteRole)
}

func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	roleId := c.Params("id")
	res, err := h.roleClient.GetRole(c.Context(), &rbacv1.GetRoleRequest{
		Id: roleId,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"role": mapper.ToRoleResponse(res.Role)})
}

func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	var body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
		Slug        string   `json:"slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	res, err := h.roleClient.CreateRole(c.Context(), &rbacv1.CreateRoleRequest{
		Name:          body.Name,
		Description:   body.Description,
		PermissionIds: body.Permissions,
		CreatedById:   userID,
		Slug:          body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"role": mapper.ToRoleResponse(res.Role)})
}

func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	roleID := c.Params("id")
	var body struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Permissions []string `json:"permissions"`
		Slug        *string  `json:"slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	res, err := h.roleClient.UpdateRole(c.Context(), &rbacv1.UpdateRoleRequest{
		Id:            roleID,
		Name:          body.Name,
		Description:   body.Description,
		PermissionIds: body.Permissions,
		UpdatedById:   userID,
		Slug:          body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"role": mapper.ToRoleResponse(res.Role)})
}

func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	roleId := c.Params("id")
	isPermanent := c.QueryBool("permanent", false)
	userID := c.Locals(middleware.ContextUserID).(string)
	_, err := h.roleClient.DeleteRole(c.Context(), &rbacv1.DeleteRoleRequest{
		Id:          roleId,
		DeletedById: userID,
		IsPermanent: isPermanent,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "role deleted")
}

func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)

	resp, err := h.roleClient.ListRoles(c.Context(), &rbacv1.ListRolesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return err
	}

	roles := make([]dto.RoleResponse, len(resp.Roles))
	for i := range resp.Roles {
		roles[i] = mapper.ToRoleResponse(resp.Roles[i])
	}

	return response.OKWithMeta(c, roles, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}
