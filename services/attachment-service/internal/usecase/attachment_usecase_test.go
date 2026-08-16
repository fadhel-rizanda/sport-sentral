package usecase

import (
	"context"
	"os"
	"testing"

	"microservice-golang/services/attachment-service/internal/config"
	"microservice-golang/services/attachment-service/internal/dto"
	"microservice-golang/services/attachment-service/internal/entity"
	"microservice-golang/services/attachment-service/internal/storage"

	"github.com/google/uuid"
)

type mockAttachmentRepo struct {
	items map[uuid.UUID]*entity.Attachment
}

func newMockAttachmentRepo() *mockAttachmentRepo {
	return &mockAttachmentRepo{
		items: make(map[uuid.UUID]*entity.Attachment),
	}
}

func (m *mockAttachmentRepo) Create(ctx context.Context, att *entity.Attachment) error {
	m.items[att.ID] = att
	return nil
}

func (m *mockAttachmentRepo) Update(ctx context.Context, att *entity.Attachment) error {
	m.items[att.ID] = att
	return nil
}

func (m *mockAttachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockAttachmentRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Attachment, error) {
	if att, ok := m.items[id]; ok {
		return att, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockAttachmentRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
	var result []*entity.Attachment
	for _, id := range ids {
		if att, ok := m.items[id]; ok {
			result = append(result, att)
		}
	}
	return result, nil
}

func (m *mockAttachmentRepo) List(ctx context.Context, input dto.ListAttachmentsInput) ([]*entity.Attachment, int64, error) {
	var result []*entity.Attachment
	for _, att := range m.items {
		if input.FileCategory != nil && att.FileCategory != *input.FileCategory {
			continue
		}
		result = append(result, att)
	}
	return result, int64(len(result)), nil
}

func TestAttachmentUseCase(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "usecase_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	localSP, _ := storage.NewLocalStorage(config.LocalStorageConfig{
		LocalDir: tempDir,
		BaseURL:  "http://localhost:8080/public",
	})
	repo := newMockAttachmentRepo()
	uc := NewAttachmentUseCase(repo, localSP, nil, "attachments")

	ctx := context.Background()

	// 1. Test UploadAttachment (direct byte upload)
	content := []byte("test binary data")
	userID := uuid.New()
	uploadInput := dto.UploadAttachmentInput{
		Filename:         "avatar.png",
		MimeType:         "image/png",
		FileCategory:     "AVATAR",
		Content:          content,
		UploadedByUserID: &userID,
	}

	att, err := uc.UploadAttachment(ctx, uploadInput)
	if err != nil {
		t.Fatalf("UploadAttachment failed: %v", err)
	}
	if att.Status != entity.StatusActive {
		t.Errorf("expected status ACTIVE, got %s", att.Status)
	}
	if att.Filename != "avatar.png" {
		t.Errorf("expected filename avatar.png, got %s", att.Filename)
	}

	// 2. Test GetAttachment
	fetched, err := uc.GetAttachment(ctx, att.ID)
	if err != nil {
		t.Fatalf("GetAttachment failed: %v", err)
	}
	if fetched.ID != att.ID {
		t.Errorf("ID mismatch: got %s, expected %s", fetched.ID, att.ID)
	}

	// 3. Test CreatePresignedUploadURL
	presignedOut, err := uc.CreatePresignedUploadURL(ctx, dto.CreatePresignedUploadURLInput{
		Filename:     "document.pdf",
		MimeType:     "application/pdf",
		FileSize:     1024,
		FileCategory: "DOCUMENT",
	})
	if err != nil {
		t.Fatalf("CreatePresignedUploadURL failed: %v", err)
	}
	if presignedOut.AttachmentID == "" || presignedOut.UploadURL == "" {
		t.Errorf("expected valid presigned response")
	}

	// 4. Test ListAttachments
	listOut, err := uc.ListAttachments(ctx, dto.ListAttachmentsInput{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListAttachments failed: %v", err)
	}
	if listOut.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", listOut.TotalCount)
	}

	// 5. Test Invalid File Category Validation
	_, err = uc.CreatePresignedUploadURL(ctx, dto.CreatePresignedUploadURLInput{
		Filename:     "test.pdf",
		MimeType:     "application/pdf",
		FileSize:     1024,
		FileCategory: "dvsdvsd",
	})
	if err == nil {
		t.Errorf("expected error for invalid file_category 'dvsdvsd', got nil")
	}

	// 6. Test DeleteAttachment
	if err := uc.DeleteAttachment(ctx, att.ID); err != nil {
		t.Fatalf("DeleteAttachment failed: %v", err)
	}

	_, err = uc.GetAttachment(ctx, att.ID)
	if err == nil {
		t.Errorf("expected error getting deleted attachment")
	}
}
