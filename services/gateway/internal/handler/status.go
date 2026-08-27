package handler

import (
	"github.com/gofiber/fiber/v2"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type StatusHandler struct {
	statusClient metav1.StatusServiceClient
}

func NewStatusHandler(statusClient metav1.StatusServiceClient) *StatusHandler {
	return &StatusHandler{
		statusClient: statusClient,
	}
}

func (h *StatusHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	status := router.Group("/statuses", auth)
	status.Get("/", h.ListStatuses)
	status.Get("/:id", h.GetStatusByID)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin/statuses", adminMiddlewares...)
	adminGroup.Post("/", h.CreateStatus)
	adminGroup.Put("/:id", h.UpdateStatus)
	adminGroup.Delete("/:id", h.DeleteStatus)
}

// CreateStatus godoc
// @Summary      Create new status (Admin)
// @Description  Create new status metadata for a specific entity (e.g. USER, ACADEMY, MATCH).
// @Tags         Statuses
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateStatusRequest true "New Status Data"
// @Success      200  {object}  response.Response{data=dto.StatusResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/statuses [post]
// @Security     BearerAuth
func (h *StatusHandler) CreateStatus(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.CreateStatusRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.statusClient.CreateStatus(c.Context(), &metav1.CreateStatusRequest{
		Type:        body.Type,
		Name:        body.Name,
		CreatedById: userID,
		Slug:        body.Slug,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"status": mapper.ToStatusResponse(resp.Status)})
}

// UpdateStatus godoc
// @Summary      Update status (Admin)
// @Description  Update type, name, or slug of a status.
// @Tags         Statuses
// @Accept       json
// @Produce      json
// @Param        id      path string                  true "Status ID (UUID)"
// @Param        request body dto.UpdateStatusRequest true "Status Update Data"
// @Success      200  {object}  response.Response{data=dto.StatusResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/statuses/{id} [put]
// @Security     BearerAuth
func (h *StatusHandler) UpdateStatus(c *fiber.Ctx) error {
	statusID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	var body dto.UpdateStatusRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.statusClient.UpdateStatus(c.Context(), &metav1.UpdateStatusRequest{
		Id:          statusID,
		Type:        body.Type,
		Name:        body.Name,
		UpdatedById: userID,
		Slug:        body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"status": mapper.ToStatusResponse(resp.Status)})
}

// DeleteStatus godoc
// @Summary      Delete status (Admin)
// @Description  Delete status via soft delete or permanently.
// @Tags         Statuses
// @Produce      json
// @Param        id        path  string true  "Status ID (UUID)"
// @Param        permanent query bool   false "Permanent delete (default false)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/statuses/{id} [delete]
// @Security     BearerAuth
func (h *StatusHandler) DeleteStatus(c *fiber.Ctx) error {
	statusID := c.Params("id")
	isPermanent := c.QueryBool("permanent", false)
	userID := c.Locals(middleware.ContextUserID).(string)
	_, err := h.statusClient.DeleteStatus(c.Context(), &metav1.DeleteStatusRequest{
		Id:          statusID,
		DeletedById: userID,
		IsPermanent: isPermanent,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "status deleted")
}

// GetStatusByID godoc
// @Summary      Get status details
// @Description  Retrieve status data by status ID (UUID).
// @Tags         Statuses
// @Produce      json
// @Param        id   path      string  true  "Status ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.StatusResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /statuses/{id} [get]
// @Security     BearerAuth
func (h *StatusHandler) GetStatusByID(c *fiber.Ctx) error {
	statusID := c.Params("id")

	resp, err := h.statusClient.GetStatusByID(c.Context(), &metav1.GetStatusByIDRequest{
		Id: statusID,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"status": mapper.ToStatusResponse(resp.Status)})
}

// ListStatuses godoc
// @Summary      Get list of statuses
// @Description  Retrieve list of all available statuses, optional filter by entity type (USER, ACADEMY, etc).
// @Tags         Statuses
// @Produce      json
// @Param        type      query string false "Filter by status type (e.g. USER, ACADEMY)"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.StatusResponse}
// @Failure      401  {object}  response.Response
// @Router       /statuses [get]
// @Security     BearerAuth
func (h *StatusHandler) ListStatuses(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
	statusType := c.Query("type", "")
	var statusTypePtr *string
	if statusType != "" {
		statusTypePtr = &statusType
	}

	resp, err := h.statusClient.ListStatuses(c.Context(), &metav1.ListStatusesRequest{
		PageSize: int32(pageSize),
		Type:     statusTypePtr,
		Page:     int32(page),
	})
	if err != nil {
		return err
	}

	statuses := make([]dto.StatusResponse, len(resp.Statuses))
	for i := range statuses {
		statuses[i] = mapper.ToStatusResponse(resp.Statuses[i])
	}

	return response.OKWithMeta(c, statuses, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}
