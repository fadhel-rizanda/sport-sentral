package mapper

import (
	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/services/attachment-service/internal/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoAttachment(att *entity.Attachment) *attachmentv1.Attachment {
	if att == nil {
		return nil
	}

	var userIDStr *string
	if att.UploadedByUserID != nil {
		str := att.UploadedByUserID.String()
		userIDStr = &str
	}

	return &attachmentv1.Attachment{
		Id:               att.ID.String(),
		Filename:         att.Filename,
		FilePath:         att.FilePath,
		BucketName:       att.BucketName,
		FileSize:         att.FileSize,
		MimeType:         att.MimeType,
		FileCategory:     att.FileCategory,
		UploadedByUserId: userIDStr,
		Status:           att.Status,
		Url:              att.URL,
		CreatedAt:        timestamppb.New(att.CreatedAt),
		UpdatedAt:        timestamppb.New(att.UpdatedAt),
	}
}

func ToProtoAttachments(atts []*entity.Attachment) []*attachmentv1.Attachment {
	result := make([]*attachmentv1.Attachment, len(atts))
	for i, att := range atts {
		result[i] = ToProtoAttachment(att)
	}
	return result
}
