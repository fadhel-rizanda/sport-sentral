package handler

import (
	"context"

	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/services/attachment-service/internal/dto"
	"microservice-golang/services/attachment-service/internal/mapper"
	"microservice-golang/services/attachment-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type AttachmentHandler struct {
	attachmentv1.UnimplementedAttachmentServiceServer
	usecase usecase.AttachmentUseCase
}

func NewAttachmentHandler(usecase usecase.AttachmentUseCase) *AttachmentHandler {
	return &AttachmentHandler{
		usecase: usecase,
	}
}

func (h *AttachmentHandler) RegisterGRPC(s *grpc.Server) {
	attachmentv1.RegisterAttachmentServiceServer(s, h)
}

func (h *AttachmentHandler) CreatePresignedUploadUrl(ctx context.Context, req *attachmentv1.CreatePresignedUploadUrlRequest) (*attachmentv1.CreatePresignedUploadUrlResponse, error) {
	var userID *uuid.UUID
	if req.UploadedByUserId != nil && *req.UploadedByUserId != "" {
		id, err := uuid.Parse(*req.UploadedByUserId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid uploaded_by_user_id"))
		}
		userID = &id
	}

	out, err := h.usecase.CreatePresignedUploadURL(ctx, dto.CreatePresignedUploadURLInput{
		Filename:         req.GetFilename(),
		MimeType:         req.GetMimeType(),
		FileSize:         req.GetFileSize(),
		FileCategory:     req.GetFileCategory(),
		UploadedByUserID: userID,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.CreatePresignedUploadUrlResponse{
		AttachmentId: out.AttachmentID,
		UploadUrl:    out.UploadURL,
		FilePath:     out.FilePath,
		Headers:      out.Headers,
	}, nil
}

func (h *AttachmentHandler) ConfirmUpload(ctx context.Context, req *attachmentv1.ConfirmUploadRequest) (*attachmentv1.ConfirmUploadResponse, error) {
	attID, err := uuid.Parse(req.GetAttachmentId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid attachment_id"))
	}

	att, err := h.usecase.ConfirmUpload(ctx, dto.ConfirmUploadInput{
		AttachmentID: attID,
		IsSuccess:    req.GetIsSuccess(),
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.ConfirmUploadResponse{
		Attachment: mapper.ToProtoAttachment(att),
	}, nil
}

func (h *AttachmentHandler) UploadAttachment(ctx context.Context, req *attachmentv1.UploadAttachmentRequest) (*attachmentv1.UploadAttachmentResponse, error) {
	var userID *uuid.UUID
	if req.UploadedByUserId != nil && *req.UploadedByUserId != "" {
		id, err := uuid.Parse(*req.UploadedByUserId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid uploaded_by_user_id"))
		}
		userID = &id
	}

	att, err := h.usecase.UploadAttachment(ctx, dto.UploadAttachmentInput{
		Filename:         req.GetFilename(),
		MimeType:         req.GetMimeType(),
		FileCategory:     req.GetFileCategory(),
		Content:          req.GetContent(),
		UploadedByUserID: userID,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.UploadAttachmentResponse{
		Attachment: mapper.ToProtoAttachment(att),
	}, nil
}

func (h *AttachmentHandler) GetAttachment(ctx context.Context, req *attachmentv1.GetAttachmentRequest) (*attachmentv1.GetAttachmentResponse, error) {
	attID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid attachment_id"))
	}

	att, err := h.usecase.GetAttachment(ctx, attID)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.GetAttachmentResponse{
		Attachment: mapper.ToProtoAttachment(att),
	}, nil
}

func (h *AttachmentHandler) GetAttachments(ctx context.Context, req *attachmentv1.GetAttachmentsRequest) (*attachmentv1.GetAttachmentsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetIds()))
	for _, idStr := range req.GetIds() {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid attachment_id in list: " + idStr))
		}
		ids = append(ids, id)
	}

	atts, err := h.usecase.GetAttachments(ctx, ids)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.GetAttachmentsResponse{
		Attachments: mapper.ToProtoAttachments(atts),
	}, nil
}

func (h *AttachmentHandler) DeleteAttachment(ctx context.Context, req *attachmentv1.DeleteAttachmentRequest) (*attachmentv1.DeleteAttachmentResponse, error) {
	attID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid attachment_id"))
	}

	if err := h.usecase.DeleteAttachment(ctx, attID); err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.DeleteAttachmentResponse{
		Success: true,
		Message: "attachment deleted successfully",
	}, nil
}

func (h *AttachmentHandler) ListAttachments(ctx context.Context, req *attachmentv1.ListAttachmentsRequest) (*attachmentv1.ListAttachmentsResponse, error) {
	var category *string
	if req.FileCategory != nil && *req.FileCategory != "" {
		category = req.FileCategory
	}

	var userID *uuid.UUID
	if req.UploadedByUserId != nil && *req.UploadedByUserId != "" {
		id, err := uuid.Parse(*req.UploadedByUserId)
		if err != nil {
			return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid uploaded_by_user_id"))
		}
		userID = &id
	}

	var status *string
	if req.Status != nil && *req.Status != "" {
		status = req.Status
	}

	out, err := h.usecase.ListAttachments(ctx, dto.ListAttachmentsInput{
		Page:             req.GetPage(),
		Limit:            req.GetLimit(),
		FileCategory:     category,
		UploadedByUserID: userID,
		Status:           status,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &attachmentv1.ListAttachmentsResponse{
		Attachments: mapper.ToProtoAttachments(out.Attachments),
		TotalCount:  out.TotalCount,
		Page:        out.Page,
		Limit:       out.Limit,
	}, nil
}
