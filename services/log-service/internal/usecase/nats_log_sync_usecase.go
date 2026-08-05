package usecase

import (
	"context"
	"encoding/json"
	"microservice-golang/services/log-service/internal/entity"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NatsLogSyncUsecase interface {
	ProcessSystemEvent(ctx context.Context, subject string, payload []byte) error
	ProcessActivityEvent(ctx context.Context, subject string, payload []byte) error
}

type natsLogSyncUsecase struct {
	auditUC    AuditLogUsecase
	activityUC ActivityLogUsecase
	log        *zap.Logger
}

func NewNatsLogSyncUseCase(auditUC AuditLogUsecase, activityUC ActivityLogUsecase, log *zap.Logger) NatsLogSyncUsecase {
	return &natsLogSyncUsecase{
		auditUC:    auditUC,
		activityUC: activityUC,
		log:        log,
	}
}

func (u *natsLogSyncUsecase) ProcessSystemEvent(ctx context.Context, subject string, payload []byte) error {
	u.log.Info("processing system event for audit log service", zap.String("subject", subject))

	// Construct automatic audit log based on NATS subject
	payloadStr := string(payload)

	// Parse optional user_id if present in payload
	var genericMap map[string]interface{}
	var userIDPtr *uuid.UUID
	var entityIDPtr *string

	if err := json.Unmarshal(payload, &genericMap); err == nil {
		if uidStr, ok := genericMap["id"].(string); ok && uidStr != "" {
			entityIDPtr = &uidStr
			if parsedUUID, err := uuid.Parse(uidStr); err == nil {
				userIDPtr = &parsedUUID
			}
		} else if uidStr, ok := genericMap["user_id"].(string); ok && uidStr != "" {
			if parsedUUID, err := uuid.Parse(uidStr); err == nil {
				userIDPtr = &parsedUUID
			}
		}
	}

	auditEntry := &entity.AuditLog{
		ID:          uuid.New(),
		ServiceName: entity.ServiceNameNatsBus,
		Module:      entity.ModuleEventSync,
		Action:      subject,
		EntityType:  entity.EntityTypeEventPayload,
		EntityID:    entityIDPtr,
		UserID:      userIDPtr,
		Status:      entity.StatusSuccess,
		Severity:    entity.SeverityInfo,
		Metadata:    &payloadStr,
		CreatedAt:   time.Now(),
	}

	if _, err := u.auditUC.CreateAuditLog(ctx, auditEntry); err != nil {
		u.log.Error("failed to record audit log from NATS event", zap.String("subject", subject), zap.Error(err))
		return err
	}

	return nil
}

func (u *natsLogSyncUsecase) ProcessActivityEvent(ctx context.Context, subject string, payload []byte) error {
	u.log.Info("processing activity event for log service", zap.String("subject", subject))

	var req struct {
		UserID       string  `json:"user_id"`
		Action       string  `json:"action"`
		ResourceType string  `json:"resource_type"`
		ResourceID   *string `json:"resource_id,omitempty"`
		Description  string  `json:"description"`
		Metadata     *string `json:"metadata,omitempty"`
		IPAddress    *string `json:"ip_address,omitempty"`
		UserAgent    *string `json:"user_agent,omitempty"`
	}

	payloadStr := string(payload)
	if err := json.Unmarshal(payload, &req); err != nil || req.UserID == "" {
		u.log.Warn("invalid activity event payload, skipping activity record", zap.String("subject", subject))
		return nil
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		u.log.Warn("invalid user_id UUID in activity event", zap.String("user_id", req.UserID))
		return err
	}

	actionStr := req.Action
	if actionStr == "" {
		actionStr = subject
	}

	resourceTypeStr := req.ResourceType
	if resourceTypeStr == "" {
		resourceTypeStr = "SystemResource"
	}

	descStr := req.Description
	if descStr == "" {
		descStr = "Activity event: " + subject
	}

	metaPtr := req.Metadata
	if metaPtr == nil {
		metaPtr = &payloadStr
	}

	activityEntry := &entity.ActivityLog{
		ID:           uuid.New(),
		UserID:       userUUID,
		Action:       actionStr,
		ResourceType: resourceTypeStr,
		ResourceID:   req.ResourceID,
		Description:  descStr,
		Metadata:     metaPtr,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		CreatedAt:    time.Now(),
	}

	if _, err := u.activityUC.CreateActivityLog(ctx, activityEntry); err != nil {
		u.log.Error("failed to record activity log from NATS event", zap.String("subject", subject), zap.Error(err))
		return err
	}

	return nil
}
