package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"microservice-golang/services/attachment-service/internal/dto"
	"microservice-golang/services/attachment-service/internal/entity"
	"microservice-golang/shared/pkg/redisclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	attachmentCacheTTL    = 15 * time.Minute
	attachmentCachePrefix = "attachment:id:"
)

func attachmentCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", attachmentCachePrefix, id.String())
}

type AttachmentRepository interface {
	Create(ctx context.Context, attachment *entity.Attachment) error
	Update(ctx context.Context, attachment *entity.Attachment) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Attachment, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error)
	List(ctx context.Context, input dto.ListAttachmentsInput) ([]*entity.Attachment, int64, error)
}

type attachmentRepository struct {
	db    *gorm.DB
	redis redisclient.Client
}

func NewAttachmentRepository(db *gorm.DB, redis redisclient.Client) AttachmentRepository {
	return &attachmentRepository{
		db:    db,
		redis: redis,
	}
}

func (r *attachmentRepository) Create(ctx context.Context, attachment *entity.Attachment) error {
	if err := r.db.WithContext(ctx).Create(attachment).Error; err != nil {
		return err
	}
	return nil
}

func (r *attachmentRepository) Update(ctx context.Context, attachment *entity.Attachment) error {
	if err := r.db.WithContext(ctx).Save(attachment).Error; err != nil {
		return err
	}
	r.invalidateCache(ctx, attachment.ID)
	return nil
}

func (r *attachmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Attachment{}, "id = ?", id).Error; err != nil {
		return err
	}
	r.invalidateCache(ctx, id)
	return nil
}

func (r *attachmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Attachment, error) {
	key := attachmentCacheKey(id)
	if r.redis != nil {
		if cached, err := r.redis.Get(ctx, key); err == nil && cached != "" {
			var att entity.Attachment
			if err := json.Unmarshal([]byte(cached), &att); err == nil {
				return &att, nil
			}
		}
	}

	var att entity.Attachment
	if err := r.db.WithContext(ctx).First(&att, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if r.redis != nil {
		if data, err := json.Marshal(att); err == nil {
			_ = r.redis.Set(ctx, key, string(data), attachmentCacheTTL)
		}
	}

	return &att, nil
}

func (r *attachmentRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
	if len(ids) == 0 {
		return []*entity.Attachment{}, nil
	}

	var attachments []*entity.Attachment
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&attachments).Error; err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *attachmentRepository) List(ctx context.Context, input dto.ListAttachmentsInput) ([]*entity.Attachment, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Attachment{})

	if input.FileCategory != nil && *input.FileCategory != "" {
		query = query.Where("file_category = ?", *input.FileCategory)
	}

	if input.UploadedByUserID != nil {
		query = query.Where("uploaded_by_user_id = ?", *input.UploadedByUserID)
	}

	if input.Status != nil && *input.Status != "" {
		query = query.Where("status = ?", *input.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var attachments []*entity.Attachment
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(limit)).Find(&attachments).Error; err != nil {
		return nil, 0, err
	}

	return attachments, total, nil
}

func (r *attachmentRepository) invalidateCache(ctx context.Context, id uuid.UUID) {
	if r.redis != nil {
		_ = r.redis.Del(ctx, attachmentCacheKey(id))
	}
}
