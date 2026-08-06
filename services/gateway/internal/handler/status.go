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

func (h *StatusHandler) CreateStatus(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
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

func (h *StatusHandler) UpdateStatus(c *fiber.Ctx) error {
	statusID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	var body struct {
		Type *string `json:"type"`
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
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
