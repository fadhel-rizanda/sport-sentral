package grpc

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	apperr "microservice-golang/shared/pkg/errors"
)

// ExtractUserID extracts the user ID from the gRPC incoming context metadata.
func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, apperr.Unauthorized("missing metadata")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		userIDs = md.Get("x-user-id")
	}

	if len(userIDs) == 0 || userIDs[0] == "" {
		return uuid.Nil, apperr.Unauthorized("user id is required in metadata")
	}

	id, err := uuid.Parse(userIDs[0])
	if err != nil {
		return uuid.Nil, apperr.InvalidArgument("invalid user id in metadata")
	}

	return id, nil
}

// ExtractActiveRole extracts the active role name from the gRPC incoming context metadata.
func ExtractActiveRole(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", apperr.Unauthorized("missing metadata")
	}

	roles := md.Get("active-role")
	if len(roles) == 0 {
		roles = md.Get("x-active-role")
	}

	if len(roles) == 0 || roles[0] == "" {
		return "", apperr.Unauthorized("active role is required in metadata")
	}

	return roles[0], nil
}

// ValidatePermission checks if the user in the context has the required permission slug in the replicated database tables.
func ValidatePermission(ctx context.Context, db *gorm.DB, permissionSlug string) error {
	userID, err := ExtractUserID(ctx)
	if err != nil {
		return err
	}

	activeRole, err := ExtractActiveRole(ctx)
	if err != nil {
		return err
	}

	// Platform admin bypasses all permission checks
	if activeRole == "platform_admin" {
		return nil
	}

	var count int64
	err = db.Table("replicated_permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = replicated_permissions.id").
		Joins("JOIN replicated_users ON replicated_users.active_role_id = role_permissions.role_id").
		Where("replicated_users.id = ? AND replicated_permissions.slug = ?", userID, permissionSlug).
		Where("replicated_users.deleted_at IS NULL").
		Where("replicated_permissions.deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return apperr.Internal(err)
	}

	if count == 0 {
		return apperr.Forbidden("insufficient permissions")
	}

	return nil
}
