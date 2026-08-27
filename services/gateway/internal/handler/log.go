package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	logv1 "microservice-golang/gen/log/v1"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type LogHandler struct {
	logClient logv1.LogServiceClient
}

func NewLogHandler(logClient logv1.LogServiceClient) *LogHandler {
	return &LogHandler{
		logClient: logClient,
	}
}

func (h *LogHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	if h == nil || h.logClient == nil {
		return
	}
	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin/logs", adminMiddlewares...)

	adminGroup.Post("/audit", h.CreateAuditLog)
	adminGroup.Get("/audit", h.ListAuditLogs)
	adminGroup.Get("/audit/:id", h.GetAuditLog)

	adminGroup.Post("/activity", h.CreateActivityLog)
	adminGroup.Get("/activity", h.ListActivityLogs)
	adminGroup.Get("/activity/:id", h.GetActivityLog)

	adminGroup.Get("/stats", h.GetLogStats)

	// User personal activity logs
	userLogs := router.Group("/logs", auth)
	userLogs.Get("/my-activity", h.ListMyActivityLogs)
}

// CreateAuditLog godoc
// @Summary      Create audit log (Admin)
// @Description  Record critical data change / system audit trail to log-service.
// @Tags         Logs
// @Accept       json
// @Produce      json
// @Param        request body object true "Audit Log Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/logs/audit [post]
// @Security     BearerAuth
func (h *LogHandler) CreateAuditLog(c *fiber.Ctx) error {
	var body struct {
		ServiceName  string  `json:"service_name"`
		Module       string  `json:"module"`
		Action       string  `json:"action"`
		EntityType   string  `json:"entity_type"`
		EntityID     *string `json:"entity_id"`
		UserID       *string `json:"user_id"`
		UserEmail    *string `json:"user_email"`
		UserRole     *string `json:"user_role"`
		IPAddress    *string `json:"ip_address"`
		UserAgent    *string `json:"user_agent"`
		Status       string  `json:"status"`
		Severity     string  `json:"severity"`
		OldValue     *string `json:"old_value"`
		NewValue     *string `json:"new_value"`
		Metadata     *string `json:"metadata"`
		ErrorMessage *string `json:"error_message"`
		DurationMs   int64   `json:"duration_ms"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	ip := c.IP()
	ua := c.Get("User-Agent")
	if body.IPAddress == nil {
		body.IPAddress = &ip
	}
	if body.UserAgent == nil {
		body.UserAgent = &ua
	}

	resp, err := h.logClient.CreateAuditLog(c.Context(), &logv1.CreateAuditLogRequest{
		ServiceName:  body.ServiceName,
		Module:       body.Module,
		Action:       body.Action,
		EntityType:   body.EntityType,
		EntityId:     body.EntityID,
		UserId:       body.UserID,
		UserEmail:    body.UserEmail,
		UserRole:     body.UserRole,
		IpAddress:    body.IPAddress,
		UserAgent:    body.UserAgent,
		Status:       body.Status,
		Severity:     body.Severity,
		OldValue:     body.OldValue,
		NewValue:     body.NewValue,
		Metadata:     body.Metadata,
		ErrorMessage: body.ErrorMessage,
		DurationMs:   body.DurationMs,
	})
	if err != nil {
		return err
	}

	return response.OK(c, resp.Log)
}

// GetAuditLog godoc
// @Summary      Get audit log details (Admin)
// @Description  Retrieve audit trail record by log ID.
// @Tags         Logs
// @Produce      json
// @Param        id   path      string  true  "Audit Log ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /admin/logs/audit/{id} [get]
// @Security     BearerAuth
func (h *LogHandler) GetAuditLog(c *fiber.Ctx) error {
	id := c.Params("id")
	resp, err := h.logClient.GetAuditLog(c.Context(), &logv1.GetAuditLogRequest{Id: id})
	if err != nil {
		return err
	}
	return response.OK(c, resp.Log)
}

// ListAuditLogs godoc
// @Summary      Get list of audit logs (Admin)
// @Description  Retrieve system audit log list filtered by service_name, module, entity_type, action, status, and time range.
// @Tags         Logs
// @Produce      json
// @Param        service_name query string false "Filter Service Name"
// @Param        module       query string false "Filter Module"
// @Param        entity_type  query string false "Filter Entity Type"
// @Param        entity_id    query string false "Filter Entity ID"
// @Param        action       query string false "Filter Action"
// @Param        user_id      query string false "Filter User ID"
// @Param        status       query string false "Filter Status"
// @Param        severity     query string false "Filter Severity"
// @Param        search       query string false "Search query"
// @Param        start_time   query string false "Start Time (RFC3339)"
// @Param        end_time     query string false "End Time (RFC3339)"
// @Param        page         query int    false "Page number (default 1)"
// @Param        limit        query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/logs/audit [get]
// @Security     BearerAuth
func (h *LogHandler) ListAuditLogs(c *fiber.Ctx) error {
	page, limit := parsePagination(c)

	req := &logv1.ListAuditLogsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	if s := c.Query("service_name"); s != "" {
		req.ServiceName = &s
	}
	if m := c.Query("module"); m != "" {
		req.Module = &m
	}
	if e := c.Query("entity_type"); e != "" {
		req.EntityType = &e
	}
	if eId := c.Query("entity_id"); eId != "" {
		req.EntityId = &eId
	}
	if a := c.Query("action"); a != "" {
		req.Action = &a
	}
	if u := c.Query("user_id"); u != "" {
		req.UserId = &u
	}
	if st := c.Query("status"); st != "" {
		req.Status = &st
	}
	if sev := c.Query("severity"); sev != "" {
		req.Severity = &sev
	}
	if q := c.Query("search"); q != "" {
		req.SearchQuery = &q
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = timestamppb.New(t)
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = timestamppb.New(t)
		}
	}

	resp, err := h.logClient.ListAuditLogs(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"logs":        resp.Logs,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"limit":       resp.Limit,
	})
}

// CreateActivityLog godoc
// @Summary      Record user activity log (Admin)
// @Description  Save user action history into the activity log.
// @Tags         Logs
// @Accept       json
// @Produce      json
// @Param        request body object true "Activity Log Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/logs/activity [post]
// @Security     BearerAuth
func (h *LogHandler) CreateActivityLog(c *fiber.Ctx) error {
	var body struct {
		UserID       string  `json:"user_id"`
		Action       string  `json:"action"`
		ResourceType string  `json:"resource_type"`
		ResourceID   *string `json:"resource_id"`
		Description  string  `json:"description"`
		Metadata     *string `json:"metadata"`
		IPAddress    *string `json:"ip_address"`
		UserAgent    *string `json:"user_agent"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	ip := c.IP()
	ua := c.Get("User-Agent")
	if body.IPAddress == nil {
		body.IPAddress = &ip
	}
	if body.UserAgent == nil {
		body.UserAgent = &ua
	}

	resp, err := h.logClient.CreateActivityLog(c.Context(), &logv1.CreateActivityLogRequest{
		UserId:       body.UserID,
		Action:       body.Action,
		ResourceType: body.ResourceType,
		ResourceId:   body.ResourceID,
		Description:  body.Description,
		Metadata:     body.Metadata,
		IpAddress:    body.IPAddress,
		UserAgent:    body.UserAgent,
	})
	if err != nil {
		return err
	}

	return response.OK(c, resp.Log)
}

// GetActivityLog godoc
// @Summary      Get activity log details (Admin)
// @Description  Retrieve user activity data by log ID.
// @Tags         Logs
// @Produce      json
// @Param        id   path      string  true  "Activity Log ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /admin/logs/activity/{id} [get]
// @Security     BearerAuth
func (h *LogHandler) GetActivityLog(c *fiber.Ctx) error {
	id := c.Params("id")
	resp, err := h.logClient.GetActivityLog(c.Context(), &logv1.GetActivityLogRequest{Id: id})
	if err != nil {
		return err
	}
	return response.OK(c, resp.Log)
}

// ListActivityLogs godoc
// @Summary      Get list of activity logs (Admin)
// @Description  Retrieve all user activity logs on the platform with various filters.
// @Tags         Logs
// @Produce      json
// @Param        user_id       query string false "Filter User ID"
// @Param        action        query string false "Filter Action"
// @Param        resource_type query string false "Filter Resource Type"
// @Param        resource_id   query string false "Filter Resource ID"
// @Param        start_time    query string false "Start Time (RFC3339)"
// @Param        end_time      query string false "End Time (RFC3339)"
// @Param        page          query int    false "Page number (default 1)"
// @Param        limit         query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/logs/activity [get]
// @Security     BearerAuth
func (h *LogHandler) ListActivityLogs(c *fiber.Ctx) error {
	page, limit := parsePagination(c)

	req := &logv1.ListActivityLogsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	if u := c.Query("user_id"); u != "" {
		req.UserId = &u
	}
	if a := c.Query("action"); a != "" {
		req.Action = &a
	}
	if r := c.Query("resource_type"); r != "" {
		req.ResourceType = &r
	}
	if rId := c.Query("resource_id"); rId != "" {
		req.ResourceId = &rId
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = timestamppb.New(t)
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = timestamppb.New(t)
		}
	}

	resp, err := h.logClient.ListActivityLogs(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"logs":        resp.Logs,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"limit":       resp.Limit,
	})
}

// ListMyActivityLogs godoc
// @Summary      Get own account activity history
// @Description  Retrieve activity logs performed by the currently logged-in user account.
// @Tags         Logs
// @Produce      json
// @Param        page  query int false "Page number (default 1)"
// @Param        limit query int false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /logs/my-activity [get]
// @Security     BearerAuth
func (h *LogHandler) ListMyActivityLogs(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	page, limit := parsePagination(c)

	req := &logv1.ListActivityLogsRequest{
		UserId: &userID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	resp, err := h.logClient.ListActivityLogs(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"logs":        resp.Logs,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"limit":       resp.Limit,
	})
}

// GetLogStats godoc
// @Summary      Get system log statistics (Admin)
// @Description  Retrieve summary metrics and volume of audit logs.
// @Tags         Logs
// @Produce      json
// @Param        service_name query string false "Filter Service Name"
// @Param        start_time   query string false "Start Time (RFC3339)"
// @Param        end_time     query string false "End Time (RFC3339)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/logs/stats [get]
// @Security     BearerAuth
func (h *LogHandler) GetLogStats(c *fiber.Ctx) error {
	req := &logv1.GetLogStatsRequest{}
	if s := c.Query("service_name"); s != "" {
		req.ServiceName = &s
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = timestamppb.New(t)
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = timestamppb.New(t)
		}
	}

	resp, err := h.logClient.GetLogStats(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, resp)
}

func parsePagination(c *fiber.Ctx) (int, int) {
	return request.ParsePagination(c)
}
