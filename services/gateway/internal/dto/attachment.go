package dto

type AttachmentResponse struct {
	ID               string  `json:"id"`
	Filename         string  `json:"filename"`
	FilePath         string  `json:"file_path"`
	BucketName       string  `json:"bucket_name"`
	FileSize         int64   `json:"file_size"`
	MimeType         string  `json:"mime_type"`
	FileCategory     string  `json:"file_category"`
	UploadedByUserID *string `json:"uploaded_by_user_id,omitempty"`
	Status           string  `json:"status"`
	URL              string  `json:"url"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type PresignedUploadURLResponse struct {
	AttachmentID string            `json:"attachment_id"`
	UploadURL    string            `json:"upload_url"`
	FilePath     string            `json:"file_path"`
	Headers      map[string]string `json:"headers,omitempty"`
}

type CreatePresignedUploadUrlRequest struct {
	Filename     string `json:"filename" validate:"required,min=1" example:"avatar.png"`
	MimeType     string `json:"mime_type" validate:"required,min=1" example:"image/png"`
	FileSize     int64  `json:"file_size" validate:"required,gt=0" example:"1048576"`
	FileCategory string `json:"file_category" validate:"omitempty,oneof=GENERAL IMAGE AVATAR LOGO DOCUMENT CERTIFICATE VIDEO" example:"AVATAR"`
}

type ConfirmUploadRequest struct {
	AttachmentID string `json:"attachment_id" validate:"required,uuid" example:"01950000-0000-7000-8000-000000000001"`
	IsSuccess    *bool  `json:"is_success" example:"true"`
}
