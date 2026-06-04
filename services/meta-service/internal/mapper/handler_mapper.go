package mapper

import (
	commonv1 "microservice-golang/gen/common/v1"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/meta-service/internal/dto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoUserSimple(u dto.UserSimpleResponse) *commonv1.UserSimple {
	return &commonv1.UserSimple{
		Id:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
	}
}

func ToProtoStatus(s *dto.StatusResponse) *metav1.Status {
	res := &metav1.Status{
		Id:        s.ID.String(),
		Type:      s.Type,
		Name:      s.Name,
		Slug:      s.Slug,
		CreatedBy: ToProtoUserSimple(s.CreatedBy),
		UpdatedBy: ToProtoUserSimple(s.UpdatedBy),
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}

	if s.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*s.DeletedAt)
		if s.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*s.DeletedBy)
		}
	}

	return res
}

func ToProtoTag(t *dto.TagResponse) *metav1.Tag {
	res := &metav1.Tag{
		Id:        t.ID.String(),
		Type:      t.Type,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedBy: ToProtoUserSimple(t.CreatedBy),
		UpdatedBy: ToProtoUserSimple(t.UpdatedBy),
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}

	if t.DeletedAt != nil {
		res.DeletedAt = timestamppb.New(*t.DeletedAt)
		if t.DeletedBy != nil {
			res.DeletedBy = ToProtoUserSimple(*t.DeletedBy)
		}
	}

	return res
}
