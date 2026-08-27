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

// GetRoles godoc
// @Summary      Get role details
// @Description  Retrieve role data by role ID.
// @Tags         Roles
// @Produce      json
// @Param        id   path      string  true  "Role ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.RoleResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /roles/{id} [get]
// @Security     BearerAuth
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

// CreateRole godoc
// @Summary      Create new role (Admin)
// @Description  Create a new role definition and its permission associations.
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateRoleRequest true "New Role Data"
// @Success      200  {object}  response.Response{data=dto.RoleResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/roles [post]
// @Security     BearerAuth
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	var body dto.CreateRoleRequest
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

// UpdateRole godoc
// @Summary      Update role (Admin)
// @Description  Update name, description, or permissions of a role.
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id      path string                true "Role ID (UUID)"
// @Param        request body dto.UpdateRoleRequest true "Role Update Data"
// @Success      200  {object}  response.Response{data=dto.RoleResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/roles/{id} [put]
// @Security     BearerAuth
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	roleID := c.Params("id")
	var body dto.UpdateRoleRequest
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

// DeleteRole godoc
// @Summary      Delete role (Admin)
// @Description  Delete role via soft delete or permanently.
// @Tags         Roles
// @Produce      json
// @Param        id        path  string true  "Role ID (UUID)"
// @Param        permanent query bool   false "Permanent delete (default false)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/roles/{id} [delete]
// @Security     BearerAuth
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

// ListRoles godoc
// @Summary      Get list of roles
// @Description  Retrieve paginated list of all available roles.
// @Tags         Roles
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.RoleResponse}
// @Failure      401  {object}  response.Response
// @Router       /roles [get]
// @Security     BearerAuth
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
