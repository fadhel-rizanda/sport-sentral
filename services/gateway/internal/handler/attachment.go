package handler

import (
	"io"
	"path/filepath"
	"strings"

	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"

	"github.com/gofiber/fiber/v2"
)

type AttachmentHandler struct {
	attClient attachmentv1.AttachmentServiceClient
}

func NewAttachmentHandler(attClient attachmentv1.AttachmentServiceClient) *AttachmentHandler {
	return &AttachmentHandler{
		attClient: attClient,
	}
}

func (h *AttachmentHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	if h == nil || h.attClient == nil {
		return
	}

	attGroup := router.Group("/attachments", auth)

	attGroup.Post("/presigned-url", h.CreatePresignedUploadUrl)
	attGroup.Post("/confirm", h.ConfirmUpload)
	attGroup.Post("/upload", h.UploadAttachment)
	attGroup.Get("/:id", h.GetAttachment)
	attGroup.Get("", h.ListAttachments)
	attGroup.Delete("/:id", h.DeleteAttachment)
}

// CreatePresignedUploadUrl godoc
// @Summary      Get presigned upload URL
// @Description  Initialize file upload (avatar, logo, document) and obtain a presigned URL for direct upload to S3/MinIO/Local.
// @Tags         Attachments
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePresignedUploadUrlRequest true "Upload Initialization Data"
// @Success      200  {object}  response.Response{data=dto.PresignedUploadURLResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /attachments/presigned-url [post]
// @Security     BearerAuth
func (h *AttachmentHandler) CreatePresignedUploadUrl(c *fiber.Ctx) error {
	var body dto.CreatePresignedUploadUrlRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	var userID *string
	if val, ok := c.Locals(middleware.ContextUserID).(string); ok && val != "" {
		userID = &val
	}

	res, err := h.attClient.CreatePresignedUploadUrl(c.Context(), &attachmentv1.CreatePresignedUploadUrlRequest{
		Filename:         body.Filename,
		MimeType:         body.MimeType,
		FileSize:         body.FileSize,
		FileCategory:     body.FileCategory,
		UploadedByUserId: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToPresignedUploadURLResponse(res))
}

// ConfirmUpload godoc
// @Summary      Confirm upload status
// @Description  Verify and mark upload status as COMPLETED or FAILED.
// @Tags         Attachments
// @Accept       json
// @Produce      json
// @Param        request body dto.ConfirmUploadRequest true "Upload Confirmation Data"
// @Success      200  {object}  response.Response{data=dto.AttachmentResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /attachments/confirm [post]
// @Security     BearerAuth
func (h *AttachmentHandler) ConfirmUpload(c *fiber.Ctx) error {
	var body dto.ConfirmUploadRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	isSuccess := true
	if body.IsSuccess != nil {
		isSuccess = *body.IsSuccess
	}

	res, err := h.attClient.ConfirmUpload(c.Context(), &attachmentv1.ConfirmUploadRequest{
		AttachmentId: body.AttachmentID,
		IsSuccess:    isSuccess,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAttachmentResponse(res.Attachment))
}

// UploadAttachment godoc
// @Summary      Direct file upload (Multipart)
// @Description  Upload file directly via multipart/form-data.
// @Tags         Attachments
// @Accept       multipart/form-data
// @Produce      json
// @Param        file          formData file   true  "File to upload"
// @Param        file_category formData string false "Category (GENERAL, IMAGE, AVATAR, LOGO, DOCUMENT, CERTIFICATE, VIDEO)"
// @Success      201  {object}  response.Response{data=dto.AttachmentResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /attachments/upload [post]
// @Security     BearerAuth
func (h *AttachmentHandler) UploadAttachment(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return request.NewFieldError("file", "is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to open uploaded file")
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to read uploaded file")
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileCategory := c.FormValue("file_category", "GENERAL")
	if fileCategory != "" {
		cat := strings.ToUpper(strings.TrimSpace(fileCategory))
		validCategories := map[string]bool{
			"GENERAL": true, "IMAGE": true, "AVATAR": true, "LOGO": true, "DOCUMENT": true, "CERTIFICATE": true, "VIDEO": true,
		}
		if !validCategories[cat] {
			return request.NewFieldError("file_category", "must be one of: GENERAL IMAGE AVATAR LOGO DOCUMENT CERTIFICATE VIDEO")
		}
		fileCategory = cat
	}

	var userID *string
	if val, ok := c.Locals(middleware.ContextUserID).(string); ok && val != "" {
		userID = &val
	}

	res, err := h.attClient.UploadAttachment(c.Context(), &attachmentv1.UploadAttachmentRequest{
		Filename:         filepath.Base(fileHeader.Filename),
		MimeType:         mimeType,
		FileCategory:     fileCategory,
		Content:          fileBytes,
		UploadedByUserId: userID,
	})
	if err != nil {
		return err
	}

	return response.Created(c, mapper.ToAttachmentResponse(res.Attachment))
}

// GetAttachment godoc
// @Summary      Get attachment file details
// @Description  Retrieve metadata and URL link for a file by attachment ID.
// @Tags         Attachments
// @Produce      json
// @Param        id   path      string  true  "Attachment ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.AttachmentResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /attachments/{id} [get]
// @Security     BearerAuth
func (h *AttachmentHandler) GetAttachment(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := h.attClient.GetAttachment(c.Context(), &attachmentv1.GetAttachmentRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, mapper.ToAttachmentResponse(res.Attachment))
}

// ListAttachments godoc
// @Summary      Get list of attachments
// @Description  Retrieve list of attachments with filters for category, uploader user, or status.
// @Tags         Attachments
// @Produce      json
// @Param        file_category       query string false "Filter Category (GENERAL, IMAGE, AVATAR, etc.)"
// @Param        uploaded_by_user_id query string false "Filter Uploader User ID (UUID)"
// @Param        status              query string false "Filter Status (PENDING, COMPLETED, FAILED)"
// @Param        page                query int    false "Page number (default 1)"
// @Param        limit               query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.AttachmentResponse}
// @Failure      401  {object}  response.Response
// @Router       /attachments [get]
// @Security     BearerAuth
func (h *AttachmentHandler) ListAttachments(c *fiber.Ctx) error {
	page, limit := request.ParsePagination(c)
	category := c.Query("file_category")
	userID := c.Query("uploaded_by_user_id")
	status := c.Query("status")

	req := &attachmentv1.ListAttachmentsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}
	if category != "" {
		req.FileCategory = &category
	}
	if userID != "" {
		req.UploadedByUserId = &userID
	}
	if status != "" {
		req.Status = &status
	}

	res, err := h.attClient.ListAttachments(c.Context(), req)
	if err != nil {
		return err
	}

	return response.OKWithMeta(c, mapper.ToAttachmentResponses(res.Attachments), response.Meta{
		Page:     int(res.Page),
		PageSize: int(res.Limit),
		Total:    res.TotalCount,
	})
}

// DeleteAttachment godoc
// @Summary      Delete file attachment
// @Description  Delete attachment metadata and file from storage.
// @Tags         Attachments
// @Produce      json
// @Param        id   path      string  true  "Attachment ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /attachments/{id} [delete]
// @Security     BearerAuth
func (h *AttachmentHandler) DeleteAttachment(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := h.attClient.DeleteAttachment(c.Context(), &attachmentv1.DeleteAttachmentRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, res.Message)
}
