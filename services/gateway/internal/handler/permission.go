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

type PermissionHandler struct {
	permissionClient rbacv1.PermissionServiceClient
}

func NewPermissionHandler(client rbacv1.PermissionServiceClient) *PermissionHandler {
	return &PermissionHandler{
		permissionClient: client,
	}
}

func (h *PermissionHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	permission := router.Group("/permissions")
	permission.Get("/", h.ListPermissions)
	permission.Get("/:id", h.GetPermission)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin/permissions", adminMiddlewares...)
	adminGroup.Post("/", h.CreatePermission)
	adminGroup.Put("/:id", h.UpdatePermission)
	adminGroup.Delete("/:id", h.DeletePermission)
}

// CreatePermission godoc
// @Summary      Create new permission (Admin)
// @Description  Create a new permission entry based on resource and action.
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePermissionRequest true "New Permission Data"
// @Success      200  {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/permissions [post]
// @Security     BearerAuth
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	var body dto.CreatePermissionRequest
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	res, err := h.permissionClient.CreatePermission(c.Context(), &rbacv1.CreatePermissionRequest{
		Resource:    body.Resource,
		Action:      body.Action,
		Description: body.Description,
		CreatedById: userID,
		Slug:        body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"permission": mapper.ToPermissionResponse(res.Permission)})
}

// UpdatePermission godoc
// @Summary      Update permission (Admin)
// @Description  Update resource, action, or description of a permission.
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Permission ID (UUID)"
// @Param        request body dto.UpdatePermissionRequest true "Permission Update Data"
// @Success      200  {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/permissions/{id} [put]
// @Security     BearerAuth
func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	permissionID := c.Params("id")
	var body dto.UpdatePermissionRequest
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	res, err := h.permissionClient.UpdatePermission(c.Context(), &rbacv1.UpdatePermissionRequest{
		Id:          permissionID,
		Resource:    body.Resource,
		Action:      body.Action,
		Description: body.Description,
		UpdatedById: userID,
		Slug:        body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"permission": mapper.ToPermissionResponse(res.Permission)})
}

// DeletePermission godoc
// @Summary      Delete permission (Admin)
// @Description  Delete permission via soft delete or permanently.
// @Tags         Permissions
// @Produce      json
// @Param        id        path  string true  "Permission ID (UUID)"
// @Param        permanent query bool   false "Permanent delete (default false)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/permissions/{id} [delete]
// @Security     BearerAuth
func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	permissionID := c.Params("id")
	isPermanent := c.QueryBool("permanent", false)
	_, err := h.permissionClient.DeletePermission(c.Context(), &rbacv1.DeletePermissionRequest{
		Id:          permissionID,
		DeletedById: userID,
		IsPermanent: isPermanent,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "permission deleted")
}

// ListPermissions godoc
// @Summary      Get list of permissions
// @Description  Retrieve paginated list of all registered permissions.
// @Tags         Permissions
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.PermissionResponse}
// @Router       /permissions [get]
func (h *PermissionHandler) ListPermissions(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)

	resp, err := h.permissionClient.ListPermissions(c.Context(), &rbacv1.ListPermissionsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return err
	}

	permissions := make([]dto.PermissionResponse, len(resp.Permissions))
	for i, permission := range resp.Permissions {
		permissions[i] = mapper.ToPermissionResponse(permission)
	}
	return response.OKWithMeta(c, permissions, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

// GetPermission godoc
// @Summary      Get permission details
// @Description  Retrieve permission data by ID.
// @Tags         Permissions
// @Produce      json
// @Param        id   path      string  true  "Permission ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.PermissionResponse}
// @Failure      404  {object}  response.Response
// @Router       /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *fiber.Ctx) error {
	permissionID := c.Params("id")
	resp, err := h.permissionClient.GetPermission(c.Context(), &rbacv1.GetPermissionRequest{
		Id: permissionID,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"permission": mapper.ToPermissionResponse(resp.Permission)})
}
