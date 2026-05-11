package handler

import (
	"github.com/gofiber/fiber/v2"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
	"time"
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
	adminGroup.Put("/", h.UpdateStatus)
	adminGroup.Delete("/", h.DeleteStatus)
}

func (h *StatusHandler) CreateStatus(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.statusClient.CreateStatus(c.Context(), &metav1.CreateStatusRequest{
		Type:        body.Type,
		Name:        body.Name,
		CreatedById: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"status": toStatusResponse(resp.Status)})
}

func (h *StatusHandler) UpdateStatus(c *fiber.Ctx) error {
	statusID := c.Query("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	var body struct {
		Type *string `json:"type"`
		Name *string `json:"name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.statusClient.UpdateStatus(c.Context(), &metav1.UpdateStatusRequest{
		Id:          statusID,
		Type:        body.Type,
		Name:        body.Name,
		UpdatedById: userID,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"status": toStatusResponse(resp.Status)})
}

func (h *StatusHandler) DeleteStatus(c *fiber.Ctx) error {
	statusID := c.Query("id")
	isPermanent := c.QueryBool("permanent")
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
	return response.OK(c, fiber.Map{"status": toStatusResponse(resp.Status)})
}

func (h *StatusHandler) ListStatuses(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	statusType := c.Query("type", "")

	resp, err := h.statusClient.ListStatuses(c.Context(), &metav1.ListStatusesRequest{
		PageSize: int32(pageSize),
		Type:     &statusType,
		Page:     int32(page),
	})
	if err != nil {
		return err
	}

	statuses := make([]StatusResponse, len(resp.Statuses))
	for i := range statuses {
		statuses[i] = toStatusResponse(resp.Statuses[i])
	}

	return response.OKWithMeta(c, statuses, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

type StatusResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	CreatedByID string  `json:"created_by_id"`
	UpdatedByID string  `json:"updated_by_id"`
	DeletedByID *string `json:"deleted_by_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at"`
}

func toStatusResponse(status *metav1.Status) StatusResponse {
	res := StatusResponse{
		ID:          status.Id,
		Type:        status.Type,
		Name:        status.Name,
		CreatedByID: status.CreatedById,
		UpdatedByID: status.UpdatedById,
		CreatedAt:   status.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:   status.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
	if status.DeletedAt != nil {
		formattedDate := status.DeletedAt.AsTime().Format(time.RFC3339)
		res.DeletedAt = &formattedDate
		res.DeletedByID = status.DeletedById
	}
	return res
}
