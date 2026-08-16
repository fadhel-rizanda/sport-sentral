package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/services/attachment-service/internal/delivery/nats"
	"microservice-golang/services/attachment-service/internal/dto"
	"microservice-golang/services/attachment-service/internal/entity"
	"microservice-golang/services/attachment-service/internal/repository"
	"microservice-golang/services/attachment-service/internal/storage"
	"microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AttachmentUseCase interface {
	CreatePresignedUploadURL(ctx context.Context, input dto.CreatePresignedUploadURLInput) (*dto.CreatePresignedUploadURLOutput, error)
	ConfirmUpload(ctx context.Context, input dto.ConfirmUploadInput) (*entity.Attachment, error)
	UploadAttachment(ctx context.Context, input dto.UploadAttachmentInput) (*entity.Attachment, error)
	GetAttachment(ctx context.Context, id uuid.UUID) (*entity.Attachment, error)
	GetAttachments(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error)
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	ListAttachments(ctx context.Context, input dto.ListAttachmentsInput) (*dto.ListAttachmentsOutput, error)
}

type attachmentUseCase struct {
	repo      repository.AttachmentRepository
	storage   storage.StorageProvider
	publisher *nats.AttachmentEventPublisher
	bucket    string
}

func NewAttachmentUseCase(
	repo repository.AttachmentRepository,
	storage storage.StorageProvider,
	publisher *nats.AttachmentEventPublisher,
	bucket string,
) AttachmentUseCase {
	if bucket == "" {
		bucket = "attachments"
	}
	return &attachmentUseCase{
		repo:      repo,
		storage:   storage,
		publisher: publisher,
		bucket:    bucket,
	}
}

func (uc *attachmentUseCase) CreatePresignedUploadURL(ctx context.Context, input dto.CreatePresignedUploadURLInput) (*dto.CreatePresignedUploadURLOutput, error) {
	if strings.TrimSpace(input.Filename) == "" {
		return nil, errors.InvalidArgument("filename is required")
	}
	if strings.TrimSpace(input.MimeType) == "" {
		return nil, errors.InvalidArgument("mime_type is required")
	}
	if input.FileSize <= 0 {
		return nil, errors.InvalidArgument("file_size must be greater than 0")
	}

	category := strings.ToUpper(strings.TrimSpace(input.FileCategory))
	if category == "" {
		category = entity.CategoryGeneral
	} else if !entity.IsValidCategory(category) {
		return nil, errors.InvalidArgument("invalid file_category: must be one of GENERAL, IMAGE, AVATAR, LOGO, DOCUMENT, CERTIFICATE, VIDEO")
	}

	attID := uuid.New()
	now := time.Now()
	cleanFilename := strings.ReplaceAll(input.Filename, " ", "_")
	objectKey := fmt.Sprintf("uploads/%d/%02d/%s/%s", now.Year(), now.Month(), attID.String(), cleanFilename)

	uploadURL, headers, err := uc.storage.GetPresignedUploadURL(ctx, uc.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to generate upload url: %w", err))
	}

	att := &entity.Attachment{
		ID:               attID,
		Filename:         input.Filename,
		FilePath:         objectKey,
		BucketName:       uc.bucket,
		FileSize:         input.FileSize,
		MimeType:         input.MimeType,
		FileCategory:     category,
		UploadedByUserID: input.UploadedByUserID,
		Status:           entity.StatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := uc.repo.Create(ctx, att); err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to save attachment metadata: %w", err))
	}

	if uc.publisher != nil {
		var userIDStr *string
		if input.UploadedByUserID != nil {
			s := input.UploadedByUserID.String()
			userIDStr = &s
		}
		_ = uc.publisher.PublishAttachmentCreated(ctx, &attachmentv1.AttachmentCreatedEvent{
			AttachmentId:     att.ID.String(),
			Filename:         att.Filename,
			FilePath:         att.FilePath,
			BucketName:       att.BucketName,
			FileSize:         att.FileSize,
			MimeType:         att.MimeType,
			FileCategory:     att.FileCategory,
			UploadedByUserId: userIDStr,
			Status:           att.Status,
			CreatedAt:        timestamppb.New(att.CreatedAt),
		})
	}

	return &dto.CreatePresignedUploadURLOutput{
		AttachmentID: attID.String(),
		UploadURL:    uploadURL,
		FilePath:     objectKey,
		Headers:      headers,
	}, nil
}

func (uc *attachmentUseCase) ConfirmUpload(ctx context.Context, input dto.ConfirmUploadInput) (*entity.Attachment, error) {
	att, err := uc.repo.FindByID(ctx, input.AttachmentID)
	if err != nil {
		return nil, errors.NotFound("attachment")
	}

	if !input.IsSuccess {
		att.Status = entity.StatusFailed
		att.UpdatedAt = time.Now()
		_ = uc.repo.Update(ctx, att)
		return att, nil
	}

	exists, err := uc.storage.Exists(ctx, att.BucketName, att.FilePath)
	if err != nil || !exists {
		att.Status = entity.StatusFailed
		att.UpdatedAt = time.Now()
		_ = uc.repo.Update(ctx, att)
		return nil, errors.InvalidArgument("uploaded file not found in storage bucket")
	}

	att.Status = entity.StatusActive
	att.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, att); err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to update attachment status: %w", err))
	}

	downloadURL, err := uc.storage.GetPresignedDownloadURL(ctx, att.BucketName, att.FilePath, 24*time.Hour)
	if err == nil {
		att.URL = downloadURL
	}

	if uc.publisher != nil {
		_ = uc.publisher.PublishAttachmentUpdated(ctx, &attachmentv1.AttachmentUpdatedEvent{
			AttachmentId: att.ID.String(),
			Status:       att.Status,
			UpdatedAt:    timestamppb.New(att.UpdatedAt),
		})
	}

	return att, nil
}

func (uc *attachmentUseCase) UploadAttachment(ctx context.Context, input dto.UploadAttachmentInput) (*entity.Attachment, error) {
	if strings.TrimSpace(input.Filename) == "" {
		return nil, errors.InvalidArgument("filename is required")
	}
	if strings.TrimSpace(input.MimeType) == "" {
		return nil, errors.InvalidArgument("mime_type is required")
	}
	if len(input.Content) == 0 {
		return nil, errors.InvalidArgument("file content is empty")
	}

	category := strings.ToUpper(strings.TrimSpace(input.FileCategory))
	if category == "" {
		category = entity.CategoryGeneral
	} else if !entity.IsValidCategory(category) {
		return nil, errors.InvalidArgument("invalid file_category: must be one of GENERAL, IMAGE, AVATAR, LOGO, DOCUMENT, CERTIFICATE, VIDEO")
	}

	attID := uuid.New()
	now := time.Now()
	cleanFilename := strings.ReplaceAll(input.Filename, " ", "_")
	objectKey := fmt.Sprintf("uploads/%d/%02d/%s/%s", now.Year(), now.Month(), attID.String(), cleanFilename)

	downloadURL, err := uc.storage.Save(ctx, uc.bucket, objectKey, bytes.NewReader(input.Content), int64(len(input.Content)), input.MimeType)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to store file: %w", err))
	}

	att := &entity.Attachment{
		ID:               attID,
		Filename:         input.Filename,
		FilePath:         objectKey,
		BucketName:       uc.bucket,
		FileSize:         int64(len(input.Content)),
		MimeType:         input.MimeType,
		FileCategory:     category,
		UploadedByUserID: input.UploadedByUserID,
		Status:           entity.StatusActive,
		URL:              downloadURL,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := uc.repo.Create(ctx, att); err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to save attachment record: %w", err))
	}

	if uc.publisher != nil {
		var userIDStr *string
		if input.UploadedByUserID != nil {
			s := input.UploadedByUserID.String()
			userIDStr = &s
		}
		_ = uc.publisher.PublishAttachmentCreated(ctx, &attachmentv1.AttachmentCreatedEvent{
			AttachmentId:     att.ID.String(),
			Filename:         att.Filename,
			FilePath:         att.FilePath,
			BucketName:       att.BucketName,
			FileSize:         att.FileSize,
			MimeType:         att.MimeType,
			FileCategory:     att.FileCategory,
			UploadedByUserId: userIDStr,
			Status:           att.Status,
			CreatedAt:        timestamppb.New(att.CreatedAt),
		})
	}

	return att, nil
}

func (uc *attachmentUseCase) GetAttachment(ctx context.Context, id uuid.UUID) (*entity.Attachment, error) {
	att, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NotFound("attachment")
	}

	downloadURL, err := uc.storage.GetPresignedDownloadURL(ctx, att.BucketName, att.FilePath, 24*time.Hour)
	if err == nil {
		att.URL = downloadURL
	}

	return att, nil
}

func (uc *attachmentUseCase) GetAttachments(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
	attachments, err := uc.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to fetch attachments: %w", err))
	}

	for _, att := range attachments {
		downloadURL, err := uc.storage.GetPresignedDownloadURL(ctx, att.BucketName, att.FilePath, 24*time.Hour)
		if err == nil {
			att.URL = downloadURL
		}
	}

	return attachments, nil
}

func (uc *attachmentUseCase) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	att, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return errors.NotFound("attachment")
	}

	_ = uc.storage.Delete(ctx, att.BucketName, att.FilePath)

	if err := uc.repo.Delete(ctx, id); err != nil {
		return errors.Internal(fmt.Errorf("failed to delete attachment: %w", err))
	}

	if uc.publisher != nil {
		_ = uc.publisher.PublishAttachmentDeleted(ctx, &attachmentv1.AttachmentDeletedEvent{
			AttachmentId: id.String(),
			FilePath:     att.FilePath,
			BucketName:   att.BucketName,
			DeletedAt:    timestamppb.New(time.Now()),
		})
	}

	return nil
}

func (uc *attachmentUseCase) ListAttachments(ctx context.Context, input dto.ListAttachmentsInput) (*dto.ListAttachmentsOutput, error) {
	if input.FileCategory != nil && *input.FileCategory != "" {
		cat := strings.ToUpper(strings.TrimSpace(*input.FileCategory))
		if !entity.IsValidCategory(cat) {
			return nil, errors.InvalidArgument("invalid file_category filter: must be one of GENERAL, IMAGE, AVATAR, LOGO, DOCUMENT, CERTIFICATE, VIDEO")
		}
		input.FileCategory = &cat
	}

	attachments, total, err := uc.repo.List(ctx, input)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to list attachments: %v", err))
	}

	for _, att := range attachments {
		downloadURL, err := uc.storage.GetPresignedDownloadURL(ctx, att.BucketName, att.FilePath, 24*time.Hour)
		if err == nil {
			att.URL = downloadURL
		}
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}

	return &dto.ListAttachmentsOutput{
		Attachments: attachments,
		TotalCount:  total,
		Page:        page,
		Limit:       limit,
	}, nil
}
