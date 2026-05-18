package usecase

import (
	"microservice-golang/services/meta-service/internal/entity"
)

func toStatusResponse(status *entity.Status) *StatusResponse {
	res := &StatusResponse{
		ID:        status.ID,
		Type:      status.Type,
		Name:      status.Name,
		Slug:      status.Slug,
		CreatedAt: status.CreatedAt,
		UpdatedAt: status.UpdatedAt,
	}

	if status.CreatedBy != nil {
		res.CreatedBy = UserSimpleResponse{
			ID:       status.CreatedByID,
			Email:    status.CreatedBy.Email,
			Username: status.CreatedBy.Username,
			FullName: status.CreatedBy.FullName,
		}
	}

	if status.UpdatedBy != nil {
		res.UpdatedBy = UserSimpleResponse{
			ID:       status.UpdatedByID,
			Email:    status.UpdatedBy.Email,
			Username: status.UpdatedBy.Username,
			FullName: status.UpdatedBy.FullName,
		}
	}

	if status.DeletedAt.Valid {
		res.DeletedAt = &status.DeletedAt.Time
		if status.DeletedByID != nil && status.DeletedBy != nil {
			res.DeletedBy = &UserSimpleResponse{
				ID:       *status.DeletedByID,
				Email:    status.DeletedBy.Email,
				Username: status.DeletedBy.Username,
				FullName: status.DeletedBy.FullName,
			}
		}
	}

	return res
}

func toTagResponse(tag *entity.Tag) *TagResponse {
	res := &TagResponse{
		ID:        tag.ID,
		Type:      tag.Type,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}

	if tag.CreatedBy != nil {
		res.CreatedBy = UserSimpleResponse{
			ID:       tag.CreatedByID,
			Email:    tag.CreatedBy.Email,
			Username: tag.CreatedBy.Username,
			FullName: tag.CreatedBy.FullName,
		}
	}

	if tag.UpdatedBy != nil {
		res.UpdatedBy = UserSimpleResponse{
			ID:       tag.UpdatedByID,
			Email:    tag.UpdatedBy.Email,
			Username: tag.UpdatedBy.Username,
			FullName: tag.UpdatedBy.FullName,
		}
	}

	if tag.DeletedAt.Valid {
		res.DeletedAt = &tag.DeletedAt.Time
		if tag.DeletedByID != nil && tag.DeletedBy != nil {
			res.DeletedBy = &UserSimpleResponse{
				ID:       *tag.DeletedByID,
				Email:    tag.DeletedBy.Email,
				Username: tag.DeletedBy.Username,
				FullName: tag.DeletedBy.FullName,
			}
		}
	}

	return res
}
