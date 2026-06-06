package mapper

import (
	"time"

	commonv1 "microservice-golang/gen/common/v1"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToStatusSimpleResponse(s *commonv1.StatusSimple) dto.StatusSimpleResponse {
	if s == nil {
		return dto.StatusSimpleResponse{}
	}
	return dto.StatusSimpleResponse{
		ID:   s.Id,
		Type: s.Type,
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToTagSimpleResponse(t *commonv1.TagSimple) dto.TagSimpleResponse {
	if t == nil {
		return dto.TagSimpleResponse{}
	}
	return dto.TagSimpleResponse{
		ID:   t.Id,
		Type: t.Type,
		Name: t.Name,
		Slug: t.Slug,
	}
}

func ToStatusResponse(status *metav1.Status) dto.StatusResponse {
	res := dto.StatusResponse{
		ID:        status.Id,
		Type:      status.Type,
		Name:      status.Name,
		Slug:      status.Slug,
		CreatedBy: ToUserSimpleResponse(status.CreatedBy),
		CreatedAt: status.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt: status.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	// UpdatedBy is now a pointer
	if status.UpdatedBy != nil {
		updatedBy := ToUserSimpleResponse(status.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if status.DeletedAt != nil {
		deletedAt := status.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if status.DeletedBy != nil {
			deletedBy := ToUserSimpleResponse(status.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}

func ToTagResponse(tag *metav1.Tag) dto.TagResponse {
	res := dto.TagResponse{
		ID:        tag.Id,
		Type:      tag.Type,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedBy: ToUserSimpleResponse(tag.CreatedBy),
		CreatedAt: tag.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt: tag.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}

	if tag.UpdatedBy != nil {
		updatedBy := ToUserSimpleResponse(tag.UpdatedBy)
		res.UpdatedBy = &updatedBy
	}

	if tag.DeletedAt != nil {
		deletedAt := tag.DeletedAt.AsTime().UTC().Format(time.RFC3339)
		res.DeletedAt = &deletedAt
		if tag.DeletedBy != nil {
			deletedBy := ToUserSimpleResponse(tag.DeletedBy)
			res.DeletedBy = &deletedBy
		}
	}

	return res
}
