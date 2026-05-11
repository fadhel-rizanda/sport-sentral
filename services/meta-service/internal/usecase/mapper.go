package usecase

import "microservice-golang/services/meta-service/internal/entity"

func toStatusResponse(s *entity.Status) *StatusResponse {
	return &StatusResponse{
		ID:          s.ID,
		Type:        s.Type,
		Name:        s.Name,
		CreatedByID: s.CreatedByID,
		UpdatedByID: s.UpdatedByID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toTagResponse(t *entity.Tag) *TagResponse {
	res := &TagResponse{
		ID:          t.ID,
		Type:        t.Type,
		Name:        t.Name,
		Slug:        t.Slug,
		CreatedByID: t.CreatedBy,
		UpdatedByID: t.UpdatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DeletedAt.Valid {
		res.DeletedAt = &t.DeletedAt.Time
		res.DeletedByID = t.DeletedBy
	}
	return res
}
