package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	authv1 "microservice-golang/gen/auth/v1"
	"microservice-golang/services/gateway/internal/response"
	"microservice-golang/shared/pkg/jwt"
)

const (
	ContextUserID         = "user_id"
	ContextEmail          = "email"
	ContextUsername       = "username"
	ContextActiveRoleName = "active_role_name"
	ContextRoleIDs        = "role_ids"
)

func Auth(jwtManager *jwt.Manager, authClient authv1.AuthServiceClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "missing Authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return response.Error(c, fiber.StatusUnauthorized, "invalid authorization header")
		}

		token := parts[1]

		if jwtManager != nil {
			if claims, err := jwtManager.ValidateAccess(token); err == nil {
				c.Locals(ContextUserID, claims.UserID)
				c.Locals(ContextEmail, claims.Email)
				c.Locals(ContextUsername, claims.Username)
				c.Locals(ContextActiveRoleName, claims.ActiveRoleName)
				c.Locals(ContextRoleIDs, claims.Roles)
				return c.Next()
			}
		}

		if authClient != nil {
			resp, err := authClient.ValidateToken(c.UserContext(), &authv1.ValidateTokenRequest{
				AccessToken: token,
			})
			if err == nil && resp.User != nil {
				c.Locals(ContextUserID, resp.User.Id)
				c.Locals(ContextEmail, resp.User.Email)
				c.Locals(ContextUsername, resp.User.Username)
				if resp.ActiveRole != nil {
					c.Locals(ContextActiveRoleName, resp.ActiveRole.Name)
				}
				c.Locals(ContextRoleIDs, resp.User.RoleIds)
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusUnauthorized, "invalid or expired token")
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		activeRoleName, _ := c.Locals(ContextActiveRoleName).(string)
		roleIDs, _ := c.Locals(ContextRoleIDs).([]string)

		for _, requiredRole := range roles {
			if activeRoleName != "" && activeRoleName == requiredRole {
				return c.Next()
			}
			for _, r := range roleIDs {
				if r == requiredRole {
					return c.Next()
				}
			}
		}

		return response.Error(c, fiber.StatusForbidden, "insufficient permissions")
	}
}
