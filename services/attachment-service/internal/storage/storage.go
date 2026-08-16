package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"microservice-golang/services/attachment-service/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageProvider interface {
	Save(ctx context.Context, bucket string, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	GetPresignedUploadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, map[string]string, error)
	GetPresignedDownloadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, bucket string, objectKey string) error
	Exists(ctx context.Context, bucket string, objectKey string) (bool, error)
}

// ─── Local Storage Implementation ─────────────────────────────────────────────

type LocalStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(cfg config.LocalStorageConfig) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.LocalDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage dir: %w", err)
	}
	return &LocalStorage{
		baseDir: cfg.LocalDir,
		baseURL: cfg.BaseURL,
	}, nil
}

func cleanURL(baseURL, objectKey string) string {
	base := strings.TrimRight(baseURL, "/")
	key := strings.TrimLeft(objectKey, "/")
	if strings.HasSuffix(base, "/uploads") && strings.HasPrefix(key, "uploads/") {
		key = strings.TrimPrefix(key, "uploads/")
	}
	return fmt.Sprintf("%s/%s", base, key)
}

func cleanPath(baseDir, objectKey string) string {
	base := filepath.Clean(baseDir)
	cleanKey := strings.TrimLeft(filepath.ToSlash(objectKey), "/")
	if filepath.Base(base) == "uploads" && strings.HasPrefix(cleanKey, "uploads/") {
		cleanKey = strings.TrimPrefix(cleanKey, "uploads/")
	}
	return filepath.Join(base, filepath.FromSlash(cleanKey))
}

func (l *LocalStorage) Save(_ context.Context, _ string, objectKey string, reader io.Reader, _ int64, _ string) (string, error) {
	fullPath := cleanPath(l.baseDir, objectKey)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage path: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	downloadURL := cleanURL(l.baseURL, objectKey)
	return downloadURL, nil
}

func (l *LocalStorage) GetPresignedUploadURL(_ context.Context, _ string, objectKey string, _ time.Duration) (string, map[string]string, error) {
	uploadURL := cleanURL(l.baseURL, objectKey)
	return uploadURL, map[string]string{"Content-Type": "application/octet-stream"}, nil
}

func (l *LocalStorage) GetPresignedDownloadURL(_ context.Context, _ string, objectKey string, _ time.Duration) (string, error) {
	downloadURL := cleanURL(l.baseURL, objectKey)
	return downloadURL, nil
}

func (l *LocalStorage) Delete(_ context.Context, _ string, objectKey string) error {
	fullPath := cleanPath(l.baseDir, objectKey)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove local file: %w", err)
	}
	return nil
}

func (l *LocalStorage) Exists(_ context.Context, _ string, objectKey string) (bool, error) {
	fullPath := cleanPath(l.baseDir, objectKey)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ─── MinIO / S3 Storage Implementation ───────────────────────────────────────

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(cfg config.MinioConfig) (*MinioStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		// Log warning but proceed if minio server is unreachable during startup
		fmt.Printf("warning: bucket existence check failed: %v\n", err)
	} else if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create minio bucket: %w", err)
		}
	}

	return &MinioStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (m *MinioStorage) Save(ctx context.Context, bucket string, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = m.bucket
	}

	_, err := m.client.PutObject(ctx, targetBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to put object in minio: %w", err)
	}

	return m.GetPresignedDownloadURL(ctx, targetBucket, objectKey, 24*time.Hour)
}

func (m *MinioStorage) GetPresignedUploadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, map[string]string, error) {
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = m.bucket
	}

	presignedURL, err := m.client.PresignedPutObject(ctx, targetBucket, objectKey, expiry)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return presignedURL.String(), map[string]string{"Content-Type": "application/octet-stream"}, nil
}

func (m *MinioStorage) GetPresignedDownloadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error) {
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = m.bucket
	}

	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, targetBucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download url: %w", err)
	}

	return presignedURL.String(), nil
}

func (m *MinioStorage) Delete(ctx context.Context, bucket string, objectKey string) error {
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = m.bucket
	}

	err := m.client.RemoveObject(ctx, targetBucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove object from minio: %w", err)
	}
	return nil
}

func (m *MinioStorage) Exists(ctx context.Context, bucket string, objectKey string) (bool, error) {
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = m.bucket
	}

	_, err := m.client.StatObject(ctx, targetBucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
