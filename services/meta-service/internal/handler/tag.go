package handler

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/usecase"
	apperr "microservice-golang/shared/pkg/errors"
)

type TagHandler struct {
	metav1.UnimplementedTagServiceServer
	uc usecase.TagUseCase
}

func NewTagHandler(uc usecase.TagUseCase) *TagHandler {
	return &TagHandler{uc: uc}
}

func (h *TagHandler) RegisterGRPC(s *grpc.Server) {
	metav1.RegisterTagServiceServer(s, h)
}

func (h *TagHandler) CreateTag(ctx context.Context, req *metav1.CreateTagRequest) (*metav1.CreateTagResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedBy())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid created_by id"))
	}

	res, err := h.uc.Create(ctx, usecase.CreateTagRequest{
		Type:      req.GetType(),
		Name:      req.GetName(),
		Slug:      req.Slug,
		CreatedBy: createdBy,
	})

	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.CreateTagResponse{
		Tag: toProtoTag(res),
	}, nil
}

func (h *TagHandler) GetTag(ctx context.Context, req *metav1.GetTagRequest) (*metav1.GetTagResponse, error) {
	res, err := h.uc.GetByTypeAndName(ctx, req.GetType(), req.GetName())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.GetTagResponse{
		Tag: toProtoTag(res),
	}, nil
}

func (h *TagHandler) GetTagByID(ctx context.Context, req *metav1.GetTagByIDRequest) (*metav1.GetTagResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(apperr.InvalidArgument("invalid status id"))
	}

	res, err := h.uc.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.GetTagResponse{
		Tag: toProtoTag(res),
	}, nil
}

func (h *TagHandler) ListTag(ctx context.Context, req *metav1.ListTagsByTypeRequest) (*metav1.ListTagsByTypeResponse, error) {
	res, err := h.uc.ListByType(ctx, req.GetType())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	tags := make([]*metav1.Tag, 0, len(res))
	for _, tag := range res {
		tags = append(tags, toProtoTag(tag))
	}
	return &metav1.ListTagsByTypeResponse{
		Tags: tags,
	}, nil
}

func (h *TagHandler) UpdateTag(ctx context.Context, req *metav1.UpdateTagRequest) (*metav1.UpdateTagResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	updatedBy, err := uuid.Parse(req.GetUpdatedBy())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	res, err := h.uc.Update(ctx, id, usecase.UpdateTagRequest{
		Type:      req.Type,
		Name:      req.Name,
		Slug:      req.Slug,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	return &metav1.UpdateTagResponse{
		Tag: toProtoTag(res),
	}, nil
}

func (h *TagHandler) DeleteTag(ctx context.Context, req *metav1.DeleteTagRequest) (*metav1.DeleteTagResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}

	if req.GetIsPermanent() {
		err = h.uc.HardDelete(ctx, usecase.DeleteTagRequest{
			ID: id,
		})
	} else {
		deletedBy, err := uuid.Parse(req.GetDeletedBy())
		if err != nil {
			return nil, apperr.ToGRPC(err)
		}
		err = h.uc.SoftDelete(ctx, usecase.DeleteTagRequest{
			ID:        id,
			DeletedBy: deletedBy,
		})
	}
	if err != nil {
		return nil, apperr.ToGRPC(err)
	}
	return &metav1.DeleteTagResponse{}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func toProtoTag(t *usecase.TagResponse) *metav1.Tag {
	return &metav1.Tag{
		Id:        t.ID.String(),
		Type:      t.Type,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedBy: t.CreatedBy.String(),
		UpdatedBy: t.UpdatedBy.String(),
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}
