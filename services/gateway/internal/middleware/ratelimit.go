package middleware

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"time"
)

func RateLimit(redisClient *redis.Client, max int, expiration time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

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
