package usecase

import (
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/services/sport-service/internal/entity"
	"microservice-golang/shared/pkg/utils"

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

	var stats []*sportv1.SportStatEvent
	if s.Config != nil && len(s.Config.Stats) > 0 {
		stats = make([]*sportv1.SportStatEvent, len(s.Config.Stats))
		for i, stat := range s.Config.Stats {
			stats[i] = &sportv1.SportStatEvent{
				Id:                stat.ID.String(),
				StatTypeTagId:     stat.StatTypeTagID.String(),
				AggregationMethod: utils.ParseAggregationMethod(stat.AggregationMethod),
			}
		}
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
		Stats:            stats,
	}

	if s.DeletedAt.Valid {
		evt.DeletedAt = timestamppb.New(s.DeletedAt.Time)
	}

	return evt
}
