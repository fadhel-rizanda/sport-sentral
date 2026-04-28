package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type client struct {
	r *redis.Client
}

func New(r *redis.Client) Client {
	return &client{r: r}
}

func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.r.Set(ctx, key, value, expiration).Err()
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	return c.r.Get(ctx, key).Result()
}

func (c *client) Del(ctx context.Context, keys ...string) error {
	return c.r.Del(ctx, keys...).Err()
}
