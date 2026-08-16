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
