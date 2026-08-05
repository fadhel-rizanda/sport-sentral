package handler

import (
	"context"
	logv1 "microservice-golang/gen/log/v1"
	"microservice-golang/services/log-service/internal/mapper"
	"microservice-golang/services/log-service/internal/repository"
	"microservice-golang/services/log-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type LogHandler struct {
	logv1.UnimplementedLogServiceServer
	auditUC    usecase.AuditLogUsecase
	activityUC usecase.ActivityLogUsecase
}

func NewLogHandler(auditUC usecase.AuditLogUsecase, activityUC usecase.ActivityLogUsecase) *LogHandler {
	return &LogHandler{
		auditUC:    auditUC,
		activityUC: activityUC,
	}
}

func (h *LogHandler) RegisterGRPC(s *grpc.Server) {
	logv1.RegisterLogServiceServer(s, h)
}

func (h *LogHandler) CreateAuditLog(ctx context.Context, req *logv1.CreateAuditLogRequest) (*logv1.CreateAuditLogResponse, error) {
	entity := mapper.CreateAuditLogReqToEntity(req)
	res, err := h.auditUC.CreateAuditLog(ctx, entity)
	if err != nil {
		return nil, err
	}

	return &logv1.CreateAuditLogResponse{
		Log: mapper.AuditLogToProto(res),
	}, nil
}

func (h *LogHandler) GetAuditLog(ctx context.Context, req *logv1.GetAuditLogRequest) (*logv1.GetAuditLogResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.InvalidArgument("invalid audit log ID")
	}

	res, err := h.auditUC.GetAuditLogByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, apperr.NotFound("audit log")
	}

	return &logv1.GetAuditLogResponse{
		Log: mapper.AuditLogToProto(res),
	}, nil
}

func (h *LogHandler) ListAuditLogs(ctx context.Context, req *logv1.ListAuditLogsRequest) (*logv1.ListAuditLogsResponse, error) {
	filter := repository.AuditLogFilter{
		ServiceName: req.ServiceName,
		Module:      req.Module,
		EntityType:  req.EntityType,
		EntityID:    req.EntityId,
		Action:      req.Action,
		Status:      req.Status,
		Severity:    req.Severity,
		SearchQuery: req.SearchQuery,
		Page:        int(req.GetPage()),
		Limit:       int(req.GetLimit()),
	}

	if req.UserId != nil {
		if u, err := uuid.Parse(*req.UserId); err == nil {
			filter.UserID = &u
		}
	}
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		filter.StartTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		filter.EndTime = &t
	}

	logs, total, err := h.auditUC.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	protoLogs := make([]*logv1.AuditLog, len(logs))
	for i, l := range logs {
		protoLogs[i] = mapper.AuditLogToProto(&l)
	}

	return &logv1.ListAuditLogsResponse{
		Logs:       protoLogs,
		TotalCount: total,
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
	}, nil
}

func (h *LogHandler) CreateActivityLog(ctx context.Context, req *logv1.CreateActivityLogRequest) (*logv1.CreateActivityLogResponse, error) {
	entity, err := mapper.CreateActivityLogReqToEntity(req)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid activity log request")
	}

	res, err := h.activityUC.CreateActivityLog(ctx, entity)
	if err != nil {
		return nil, err
	}

	return &logv1.CreateActivityLogResponse{
		Log: mapper.ActivityLogToProto(res),
	}, nil
}

func (h *LogHandler) GetActivityLog(ctx context.Context, req *logv1.GetActivityLogRequest) (*logv1.GetActivityLogResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.InvalidArgument("invalid activity log ID")
	}

	res, err := h.activityUC.GetActivityLogByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, apperr.NotFound("activity log")
	}

	return &logv1.GetActivityLogResponse{
		Log: mapper.ActivityLogToProto(res),
	}, nil
}

func (h *LogHandler) ListActivityLogs(ctx context.Context, req *logv1.ListActivityLogsRequest) (*logv1.ListActivityLogsResponse, error) {
	filter := repository.ActivityLogFilter{
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceId,
		Page:         int(req.GetPage()),
		Limit:        int(req.GetLimit()),
	}

	if req.UserId != nil {
		if u, err := uuid.Parse(*req.UserId); err == nil {
			filter.UserID = &u
		}
	}
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		filter.StartTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		filter.EndTime = &t
	}

	logs, total, err := h.activityUC.ListActivityLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	protoLogs := make([]*logv1.ActivityLog, len(logs))
	for i, l := range logs {
		protoLogs[i] = mapper.ActivityLogToProto(&l)
	}

	return &logv1.ListActivityLogsResponse{
		Logs:       protoLogs,
		TotalCount: total,
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
	}, nil
}

func (h *LogHandler) GetLogStats(ctx context.Context, req *logv1.GetLogStatsRequest) (*logv1.GetLogStatsResponse, error) {
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	stats, err := h.auditUC.GetLogStats(ctx, req.ServiceName, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return mapper.LogStatsToProto(stats), nil
}
