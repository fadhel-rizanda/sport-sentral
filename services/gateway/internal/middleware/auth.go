package middleware

import (
	"github.com/gofiber/fiber/v2"
	authv1 "microservice-golang/gen/auth/v1"
	"microservice-golang/services/gateway/internal/response"
	"strings"
)

const (
	ContextUserID         = "user_id"
	ContextEmail          = "email"
	ContextUsername       = "username"
	ContextActiveRoleName = "active_role_name"
	ContextRoleIDs        = "role_ids"
)

func Auth(authClient authv1.AuthServiceClient) fiber.Handler {
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

		resp, err := authClient.ValidateToken(c.Context(), &authv1.ValidateTokenRequest{
			AccessToken: token,
		})
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals(ContextUserID, resp.UserId)
		c.Locals(ContextEmail, resp.Email)
		c.Locals(ContextUsername, resp.Username)
		c.Locals(ContextActiveRoleName, resp.ActiveRoleName)
		c.Locals(ContextRoleIDs, resp.RoleIds)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		activeRoleName, ok := c.Locals(ContextActiveRoleName).(string)
		if !ok || activeRoleName == "" {
			return response.Error(c, fiber.StatusForbidden, "forbidden")
		}

		for _, role := range roles {
			if role == activeRoleName {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, "insufficient permissions")
	}
}
