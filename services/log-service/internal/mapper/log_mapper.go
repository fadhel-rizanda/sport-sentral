package mapper

import (
	logv1 "microservice-golang/gen/log/v1"
	"microservice-golang/services/log-service/internal/entity"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AuditLogToProto(l *entity.AuditLog) *logv1.AuditLog {
	if l == nil {
		return nil
	}

	res := &logv1.AuditLog{
		Id:          l.ID.String(),
		ServiceName: l.ServiceName,
		Module:      l.Module,
		Action:      l.Action,
		EntityType:  l.EntityType,
		Status:      l.Status,
		Severity:    l.Severity,
		DurationMs:  l.DurationMs,
		CreatedAt:   timestamppb.New(l.CreatedAt),
	}

	if l.EntityID != nil {
		res.EntityId = l.EntityID
	}
	if l.UserID != nil {
		str := l.UserID.String()
		res.UserId = &str
	}
	if l.UserEmail != nil {
		res.UserEmail = l.UserEmail
	}
	if l.UserRole != nil {
		res.UserRole = l.UserRole
	}
	if l.IPAddress != nil {
		res.IpAddress = l.IPAddress
	}
	if l.UserAgent != nil {
		res.UserAgent = l.UserAgent
	}
	if l.OldValue != nil {
		res.OldValue = l.OldValue
	}
	if l.NewValue != nil {
		res.NewValue = l.NewValue
	}
	if l.Metadata != nil {
		res.Metadata = l.Metadata
	}
	if l.ErrorMessage != nil {
		res.ErrorMessage = l.ErrorMessage
	}

	return res
}

func CreateAuditLogReqToEntity(req *logv1.CreateAuditLogRequest) *entity.AuditLog {
	if req == nil {
		return nil
	}

	var userIDPtr *uuid.UUID
	if req.UserId != nil {
		if u, err := uuid.Parse(*req.UserId); err == nil {
			userIDPtr = &u
		}
	}

	return &entity.AuditLog{
		ServiceName:  req.ServiceName,
		Module:       req.Module,
		Action:       req.Action,
		EntityType:   req.EntityType,
		EntityID:     req.EntityId,
		UserID:       userIDPtr,
		UserEmail:    req.UserEmail,
		UserRole:     req.UserRole,
		IPAddress:    req.IpAddress,
		UserAgent:    req.UserAgent,
		Status:       req.Status,
		Severity:     req.Severity,
		OldValue:     req.OldValue,
		NewValue:     req.NewValue,
		Metadata:     req.Metadata,
		ErrorMessage: req.ErrorMessage,
		DurationMs:   req.DurationMs,
	}
}

func ActivityLogToProto(l *entity.ActivityLog) *logv1.ActivityLog {
	if l == nil {
		return nil
	}

	res := &logv1.ActivityLog{
		Id:           l.ID.String(),
		UserId:       l.UserID.String(),
		Action:       l.Action,
		ResourceType: l.ResourceType,
		Description:  l.Description,
		CreatedAt:    timestamppb.New(l.CreatedAt),
	}

	if l.ResourceID != nil {
		res.ResourceId = l.ResourceID
	}
	if l.Metadata != nil {
		res.Metadata = l.Metadata
	}
	if l.IPAddress != nil {
		res.IpAddress = l.IPAddress
	}
	if l.UserAgent != nil {
		res.UserAgent = l.UserAgent
	}

	return res
}

func CreateActivityLogReqToEntity(req *logv1.CreateActivityLogRequest) (*entity.ActivityLog, error) {
	if req == nil {
		return nil, nil
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	return &entity.ActivityLog{
		UserID:       userID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceId,
		Description:  req.Description,
		Metadata:     req.Metadata,
		IPAddress:    req.IpAddress,
		UserAgent:    req.UserAgent,
	}, nil
}

func LogStatsToProto(stats *entity.LogStats) *logv1.GetLogStatsResponse {
	if stats == nil {
		return &logv1.GetLogStatsResponse{}
	}

	sevCounts := make([]*logv1.SeverityCount, len(stats.CountBySeverity))
	for i, c := range stats.CountBySeverity {
		sevCounts[i] = &logv1.SeverityCount{
			Severity: c.Severity,
			Count:    c.Count,
		}
	}

	actCounts := make([]*logv1.ActionCount, len(stats.CountByAction))
	for i, c := range stats.CountByAction {
		actCounts[i] = &logv1.ActionCount{
			Action: c.Action,
			Count:  c.Count,
		}
	}

	srvCounts := make([]*logv1.ServiceCount, len(stats.CountByService))
	for i, c := range stats.CountByService {
		srvCounts[i] = &logv1.ServiceCount{
			ServiceName: c.ServiceName,
			Count:       c.Count,
		}
	}

	return &logv1.GetLogStatsResponse{
		TotalAuditLogs:    stats.TotalAuditLogs,
		TotalActivityLogs: stats.TotalActivityLogs,
		CountBySeverity:   sevCounts,
		CountByAction:     actCounts,
		CountByService:    srvCounts,
	}
}
