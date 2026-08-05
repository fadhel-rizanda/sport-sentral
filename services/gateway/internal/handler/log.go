package handler

import (
	"strconv"
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

func (h *LogHandler) GetAuditLog(c *fiber.Ctx) error {
	id := c.Params("id")
	resp, err := h.logClient.GetAuditLog(c.Context(), &logv1.GetAuditLogRequest{Id: id})
	if err != nil {
		return err
	}
	return response.OK(c, resp.Log)
}

func (h *LogHandler) ListAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

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

func (h *LogHandler) GetActivityLog(c *fiber.Ctx) error {
	id := c.Params("id")
	resp, err := h.logClient.GetActivityLog(c.Context(), &logv1.GetActivityLogRequest{Id: id})
	if err != nil {
		return err
	}
	return response.OK(c, resp.Log)
}

func (h *LogHandler) ListActivityLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

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

func (h *LogHandler) ListMyActivityLogs(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

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
