package mapper

import (
	"time"

	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/services/gateway/internal/dto"
)

func ToAttachmentResponse(att *attachmentv1.Attachment) *dto.AttachmentResponse {
	if att == nil {
		return nil
	}

	var createdAt, updatedAt string
	if att.CreatedAt != nil {
		createdAt = att.CreatedAt.AsTime().UTC().Format(time.RFC3339)
	}
	if att.UpdatedAt != nil {
		updatedAt = att.UpdatedAt.AsTime().UTC().Format(time.RFC3339)
	}

	return &dto.AttachmentResponse{
		ID:               att.Id,
		Filename:         att.Filename,
		FilePath:         att.FilePath,
		BucketName:       att.BucketName,
		FileSize:         att.FileSize,
		MimeType:         att.MimeType,
		FileCategory:     att.FileCategory,
		UploadedByUserID: att.UploadedByUserId,
		Status:           att.Status,
		URL:              att.Url,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

func ToAttachmentResponses(atts []*attachmentv1.Attachment) []*dto.AttachmentResponse {
	result := make([]*dto.AttachmentResponse, len(atts))
	for i, att := range atts {
		result[i] = ToAttachmentResponse(att)
	}
	return result
}

func ToPresignedUploadURLResponse(res *attachmentv1.CreatePresignedUploadUrlResponse) *dto.PresignedUploadURLResponse {
	if res == nil {
		return nil
	}
	return &dto.PresignedUploadURLResponse{
		AttachmentID: res.AttachmentId,
		UploadURL:    res.UploadUrl,
		FilePath:     res.FilePath,
		Headers:      res.Headers,
	}
}
