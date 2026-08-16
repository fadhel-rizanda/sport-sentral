package storage

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"microservice-golang/services/attachment-service/internal/config"
)

func TestLocalStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ls, err := NewLocalStorage(config.LocalStorageConfig{
		LocalDir: tempDir,
		BaseURL:  "http://localhost:8080/public",
	})
	if err != nil {
		t.Fatalf("failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	bucket := "test-bucket"
	objectKey := "subfolder/test_file.txt"
	content := []byte("hello attachment service")

	// Save
	downloadURL, err := ls.Save(ctx, bucket, objectKey, bytes.NewReader(content), int64(len(content)), "text/plain")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if downloadURL == "" {
		t.Errorf("expected non-empty download URL")
	}

	// Exists
	exists, err := ls.Exists(ctx, bucket, objectKey)
	if err != nil || !exists {
		t.Errorf("expected file to exist, got exists=%v, err=%v", exists, err)
	}

	// GetPresignedDownloadURL
	presignedGet, err := ls.GetPresignedDownloadURL(ctx, bucket, objectKey, 1*time.Hour)
	if err != nil || presignedGet == "" {
		t.Errorf("expected presigned download url, got err=%v", err)
	}

	// GetPresignedUploadURL
	presignedPut, headers, err := ls.GetPresignedUploadURL(ctx, bucket, objectKey, 15*time.Minute)
	if err != nil || presignedPut == "" || len(headers) == 0 {
		t.Errorf("expected presigned upload url, got err=%v", err)
	}

	// Delete
	if err := ls.Delete(ctx, bucket, objectKey); err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Exists after delete
	existsAfterDelete, _ := ls.Exists(ctx, bucket, objectKey)
	if existsAfterDelete {
		t.Errorf("file should not exist after delete")
	}
}
