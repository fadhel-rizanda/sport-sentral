package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	commonv1 "microservice-golang/gen/common/v1"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/usecase"
)

func toProtoUserSimple(u usecase.UserSimpleResponse) *commonv1.UserSimple {
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func toProtoStatus(s *usecase.StatusResponse) *metav1.Status {
	res := &metav1.Status{
		Id:        s.ID.String(),
		Type:      s.Type,
		Name:      s.Name,
		Slug:      s.Slug,
		CreatedBy: toProtoUserSimple(s.CreatedBy),
		UpdatedBy: toProtoUserSimple(s.UpdatedBy),
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}

	if s.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*s.DeletedAt)
		if s.DeletedBy != nil {
			res.DeletedBy = toProtoUserSimple(*s.DeletedBy)
		}
	}

	return res
}

func toProtoTag(t *usecase.TagResponse) *metav1.Tag {
	res := &metav1.Tag{
		Id:        t.ID.String(),
		Type:      t.Type,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedBy: toProtoUserSimple(t.CreatedBy),
		UpdatedBy: toProtoUserSimple(t.UpdatedBy),
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}

	if t.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*t.DeletedAt)
		if t.DeletedBy != nil {
			res.DeletedBy = toProtoUserSimple(*t.DeletedBy)
		}
	}

	return res
}
