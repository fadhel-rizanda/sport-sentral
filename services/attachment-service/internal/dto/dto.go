package dto

import (
	"microservice-golang/services/attachment-service/internal/entity"

	"github.com/google/uuid"
)

type CreatePresignedUploadURLInput struct {
	Filename         string
	MimeType         string
	FileSize         int64
	FileCategory     string
	UploadedByUserID *uuid.UUID
}

type CreatePresignedUploadURLOutput struct {
	AttachmentID string
	UploadURL    string
	FilePath     string
	Headers      map[string]string
}

type ConfirmUploadInput struct {
	AttachmentID uuid.UUID
	IsSuccess    bool
}

type UploadAttachmentInput struct {
	Filename         string
	MimeType         string
	FileCategory     string
	Content          []byte
	UploadedByUserID *uuid.UUID
}

type ListAttachmentsInput struct {
	Page             int32
	Limit            int32
	FileCategory     *string
	UploadedByUserID *uuid.UUID
	Status           *string
}

type ListAttachmentsOutput struct {
	Attachments []*entity.Attachment
	TotalCount  int64
	Page        int32
	Limit       int32
}
