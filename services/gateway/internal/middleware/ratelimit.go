package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RateLimit(redisClient *redis.Client, max int, expiration time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("rate_limit:ip:%s", c.IP())
		if userID, ok := c.Locals(ContextUserID).(string); ok && userID != "" {
			key = fmt.Sprintf("rate_limit:user:%s", userID)
		}

		// Use request context with timeout to avoid hanging on slow Redis connections
		ctx, cancel := context.WithTimeout(c.UserContext(), 500*time.Millisecond)
		defer cancel()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		if count == 1 {
			redisClient.Expire(ctx, key, expiration)
		}

		ttl, _ := redisClient.TTL(ctx, key).Result()

		if count > int64(max) {
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests",
			})
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(max)-count))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		return c.Next()
	}
}
