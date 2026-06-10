package usecase

import (
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/entity"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func buildSportEvent(
	eventType sportv1.SportEventType,
	s *entity.Sport,
) *sportv1.SportEvent {
	evtID, _ := uuid.NewV7()

	var iconID *string
	if s.IconAttachmentID != nil {
		str := s.IconAttachmentID.String()
		iconID = &str
	}

	var regulatorID *string
	if s.RegulatorID != nil {
		str := s.RegulatorID.String()
		regulatorID = &str
	}

	var tier string
	if s.TierTag != nil {
		tier = s.TierTag.Slug
	}

	evt := &sportv1.SportEvent{
		EventId:          evtID.String(),
		EventType:        eventType,
		OccurredAt:       timestamppb.Now(),
		SportId:          s.ID.String(),
		Name:             s.Name,
		Slug:             s.Slug,
		IconAttachmentId: iconID,
		IsVerified:       s.RegulatorID != nil,
		RegulatorId:      regulatorID,
		Tier:             tier,
	}

	if s.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(s.DeletedAt.Time)
	}

	return evt
}
