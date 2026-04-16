package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	rbacv1 "microservice-golang/gen/rbac/v1"
	"microservice-golang/services/user-service/internal/entity"
)

func toProtoPermission(p *entity.Permission) *rbacv1.Permission {
	return &rbacv1.Permission{
		Id:          p.ID.String(),
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
